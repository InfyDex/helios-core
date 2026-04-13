// Package google verifies Google ID tokens for Helios login.
package google

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/idtoken"
)

// Profile holds Google-verified identity fields used to upsert a user.
type Profile struct {
	Sub     string
	Email   string
	Name    string
	Picture string
	// Phone is the OIDC phone_number claim when the client requests phone scope.
	Phone string
}

var allowedIssuers = map[string]struct{}{
	"https://accounts.google.com": {},
	"accounts.google.com":         {},
}

// VerifyIDToken validates a Google ID token (signature, expiry, issuer) and checks that
// the JWT aud claim is one of allowedAudiences (e.g. web + Android + iOS OAuth client IDs).
func VerifyIDToken(ctx context.Context, rawToken string, allowedAudiences []string) (Profile, error) {
	if len(allowedAudiences) == 0 {
		return Profile{}, fmt.Errorf("google: no allowed OAuth client IDs configured")
	}
	allowed := make(map[string]struct{}, len(allowedAudiences))
	for _, a := range allowedAudiences {
		a = strings.TrimSpace(a)
		if a != "" {
			allowed[a] = struct{}{}
		}
	}
	if len(allowed) == 0 {
		return Profile{}, fmt.Errorf("google: no allowed OAuth client IDs configured")
	}
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Profile{}, fmt.Errorf("google: empty id token")
	}
	// Signature and expiry are validated; aud is enforced against allowed below (multi-client).
	payload, err := idtoken.Validate(ctx, rawToken, "")
	if err != nil {
		return Profile{}, err
	}
	if _, ok := allowed[payload.Audience]; !ok {
		return Profile{}, fmt.Errorf("google: token audience %q is not an allowed OAuth client ID", payload.Audience)
	}
	if _, ok := allowedIssuers[payload.Issuer]; !ok {
		return Profile{}, fmt.Errorf("google: unexpected issuer %q", payload.Issuer)
	}
	sub := strings.TrimSpace(payload.Subject)
	email := stringClaim(payload.Claims, "email")
	name := stringClaim(payload.Claims, "name")
	picture := stringClaim(payload.Claims, "picture")
	phone := stringClaim(payload.Claims, "phone_number")
	if sub == "" || email == "" {
		return Profile{}, fmt.Errorf("google: missing sub or email in token")
	}
	if name == "" {
		name = email
	}
	return Profile{
		Sub:     sub,
		Email:   email,
		Name:    name,
		Picture: picture,
		Phone:   phone,
	}, nil
}

func stringClaim(claims map[string]interface{}, key string) string {
	if claims == nil {
		return ""
	}
	v, ok := claims[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
