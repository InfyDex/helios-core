package jwt

import (
	"strings"
	"testing"
	"time"
)

func TestSignAndValidate(t *testing.T) {
	t.Parallel()
	secret := "01234567890123456789012345678901"
	userID := "550e8400-e29b-41d4-a716-446655440000"
	email := "u@example.com"
	exp := time.Hour

	tok, err := Sign(userID, email, secret, exp)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Validate(secret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != userID || claims.Email != email {
		t.Fatalf("claims UserID=%q Email=%q", claims.UserID, claims.Email)
	}
	if claims.Subject != userID {
		t.Fatalf("Subject=%q", claims.Subject)
	}
}

func TestValidateWrongSecret(t *testing.T) {
	t.Parallel()
	secret := "01234567890123456789012345678901"
	tok, err := Sign("550e8400-e29b-41d4-a716-446655440000", "u@x", secret, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Validate(strings.Repeat("b", 32), tok)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestValidateGarbage(t *testing.T) {
	t.Parallel()
	_, err := Validate("01234567890123456789012345678901", "not-a-jwt")
	if err == nil {
		t.Fatal("expected error")
	}
}
