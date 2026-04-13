package google

import (
	"context"
	"strings"
	"testing"
)

func TestVerifyIDToken_emptyAudiences(t *testing.T) {
	t.Parallel()
	_, err := VerifyIDToken(context.Background(), "x", nil)
	if err == nil || !strings.Contains(err.Error(), "no allowed") {
		t.Fatalf("expected no allowed error, got %v", err)
	}
}

func TestVerifyIDToken_emptyToken(t *testing.T) {
	t.Parallel()
	_, err := VerifyIDToken(context.Background(), "  ", []string{"a.apps.googleusercontent.com"})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty token error, got %v", err)
	}
}

func TestVerifyIDToken_invalidJWT(t *testing.T) {
	t.Parallel()
	_, err := VerifyIDToken(context.Background(), "header.payload.sig", []string{"a.apps.googleusercontent.com"})
	if err == nil {
		t.Fatal("expected error for fake token")
	}
}

func TestStringClaim_types(t *testing.T) {
	t.Parallel()
	if got := stringClaim(nil, "x"); got != "" {
		t.Fatalf("nil claims: %q", got)
	}
	if got := stringClaim(map[string]interface{}{"k": " v "}, "k"); got != "v" {
		t.Fatalf("string trim: %q", got)
	}
	if got := stringClaim(map[string]interface{}{"k": 42}, "k"); got != "42" {
		t.Fatalf("non-string: %q", got)
	}
}
