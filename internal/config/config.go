// Package config loads process environment for Helios Core.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime settings from the environment.
type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiry       time.Duration
	GoogleClientIDs []string // OAuth client IDs (web, Android, iOS); token aud must match one
}

// Load reads required environment variables and returns Config or an error.
func Load() (Config, error) {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	expSec, err := strconv.Atoi(strings.TrimSpace(envOrDefault("JWT_EXPIRY", "3600")))
	if err != nil || expSec < 60 {
		return Config{}, fmt.Errorf("JWT_EXPIRY must be an integer >= 60 (seconds)")
	}
	clientIDs, err := parseGoogleClientIDs(os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		return Config{}, err
	}
	return Config{
		Port:            port,
		DatabaseURL:     dbURL,
		JWTSecret:       secret,
		JWTExpiry:       time.Duration(expSec) * time.Second,
		GoogleClientIDs: clientIDs,
	}, nil
}

// parseGoogleClientIDs splits GOOGLE_CLIENT_ID on commas (and trims whitespace).
// One ID is enough for web-only; list all platform client IDs for Android + iOS + web.
func parseGoogleClientIDs(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID is required (one or more OAuth client IDs, comma-separated)")
	}
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID must contain at least one non-empty client ID")
	}
	return out, nil
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
