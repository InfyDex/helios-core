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
}

var allowedIssuers = map[string]struct{}{
	"https://accounts.google.com": {},
	"accounts.google.com":         {},
}

// VerifyIDToken validates a Google ID token (signature, expiry, audience) and returns profile fields.
func VerifyIDToken(ctx context.Context, rawToken, clientID string) (Profile, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Profile{}, fmt.Errorf("google: empty id token")
	}
	payload, err := idtoken.Validate(ctx, rawToken, clientID)
	if err != nil {
		return Profile{}, err
	}
	if _, ok := allowedIssuers[payload.Issuer]; !ok {
		return Profile{}, fmt.Errorf("google: unexpected issuer %q", payload.Issuer)
	}
	sub := strings.TrimSpace(payload.Subject)
	email := stringClaim(payload.Claims, "email")
	name := stringClaim(payload.Claims, "name")
	picture := stringClaim(payload.Claims, "picture")
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
