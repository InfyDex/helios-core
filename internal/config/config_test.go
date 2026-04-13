package config

import (
	"os"
	"strings"
	"testing"
)

func TestParseGoogleClientIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{"single", "abc.apps.googleusercontent.com", []string{"abc.apps.googleusercontent.com"}, false},
		{"comma separated", "a.apps.googleusercontent.com, b.apps.googleusercontent.com", []string{"a.apps.googleusercontent.com", "b.apps.googleusercontent.com"}, false},
		{"dedupe", "same.apps.googleusercontent.com, same.apps.googleusercontent.com", []string{"same.apps.googleusercontent.com"}, false},
		{"empty", "", nil, true},
		{"only commas", ", ,", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseGoogleClientIDs(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseGoogleClientIDs(%q) err=%v wantErr=%v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len=%d want %d: got %#v want %#v", len(got), len(tt.want), got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got[%d]=%q want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	unsetEnvs := func(keys ...string) {
		t.Helper()
		for _, k := range keys {
			_ = os.Unsetenv(k)
		}
	}
	const goodSecret = "01234567890123456789012345678901" // 32 chars

	t.Run("missing database", func(t *testing.T) {
		t.Cleanup(func() { unsetEnvs("DATABASE_URL", "JWT_SECRET", "GOOGLE_CLIENT_ID") })
		_ = os.Setenv("JWT_SECRET", goodSecret)
		_ = os.Setenv("GOOGLE_CLIENT_ID", "x.apps.googleusercontent.com")
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
			t.Fatalf("expected DATABASE_URL error, got %v", err)
		}
	})

	t.Run("short jwt secret", func(t *testing.T) {
		t.Cleanup(func() { unsetEnvs("DATABASE_URL", "JWT_SECRET", "GOOGLE_CLIENT_ID") })
		_ = os.Setenv("DATABASE_URL", "postgres://x")
		_ = os.Setenv("JWT_SECRET", "short")
		_ = os.Setenv("GOOGLE_CLIENT_ID", "x.apps.googleusercontent.com")
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Fatalf("expected JWT_SECRET error, got %v", err)
		}
	})

	t.Run("ok", func(t *testing.T) {
		t.Cleanup(func() { unsetEnvs("PORT", "DATABASE_URL", "JWT_SECRET", "JWT_EXPIRY", "GOOGLE_CLIENT_ID") })
		_ = os.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
		_ = os.Setenv("JWT_SECRET", goodSecret)
		_ = os.Setenv("JWT_EXPIRY", "120")
		_ = os.Setenv("GOOGLE_CLIENT_ID", "a.apps.googleusercontent.com,b.apps.googleusercontent.com")
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Port != "8080" {
			t.Fatalf("Port=%q want 8080", cfg.Port)
		}
		if len(cfg.GoogleClientIDs) != 2 {
			t.Fatalf("GoogleClientIDs=%v", cfg.GoogleClientIDs)
		}
		if cfg.JWTExpiry.String() != "2m0s" {
			t.Fatalf("JWTExpiry=%v want 2m", cfg.JWTExpiry)
		}
	})
}
