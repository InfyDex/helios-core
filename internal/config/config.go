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
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiry      time.Duration
	GoogleClientID string
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
	googleClient := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	if googleClient == "" {
		return Config{}, fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	return Config{
		Port:           port,
		DatabaseURL:    dbURL,
		JWTSecret:      secret,
		JWTExpiry:      time.Duration(expSec) * time.Second,
		GoogleClientID: googleClient,
	}, nil
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
