package config

import (
	"fmt"
	"os"
)

// Config contains runtime settings for the API server and external providers.
type Config struct {
	HTTPAddr              string
	DatabaseDriver        string
	DatabaseURL           string
	EncryptionKeyHex      string
	FrontendURL           string
	StaticPath            string
	MonzoClientID         string
	MonzoClientSecret     string
	MonzoRedirectURL      string
	TrueLayerClientID     string
	TrueLayerClientSecret string
	TrueLayerRedirectURL  string
	GoCardlessSecretID    string
	GoCardlessSecretKey   string
}

// Load reads configuration from environment variables, applying safe local defaults where possible.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:              getEnv("HTTP_ADDR", ":8080"),
		DatabaseDriver:        getEnv("DATABASE_DRIVER", "postgres"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://pennypilot:pennypilot@localhost:5432/pennypilot?sslmode=disable"),
		EncryptionKeyHex:      os.Getenv("ENCRYPTION_KEY_HEX"),
		FrontendURL:           getEnv("FRONTEND_URL", "http://localhost:3000"),
		StaticPath:            os.Getenv("STATIC_PATH"),
		MonzoClientID:         os.Getenv("MONZO_CLIENT_ID"),
		MonzoClientSecret:     os.Getenv("MONZO_CLIENT_SECRET"),
		MonzoRedirectURL:      getEnv("MONZO_REDIRECT_URL", "http://localhost:8080/auth/monzo/callback"),
		TrueLayerClientID:     os.Getenv("TRUELAYER_CLIENT_ID"),
		TrueLayerClientSecret: os.Getenv("TRUELAYER_CLIENT_SECRET"),
		TrueLayerRedirectURL:  getEnv("TRUELAYER_REDIRECT_URL", "http://localhost:8080/auth/truelayer/callback"),
		GoCardlessSecretID:    os.Getenv("GOCARDLESS_SECRET_ID"),
		GoCardlessSecretKey:   os.Getenv("GOCARDLESS_SECRET_KEY"),
	}

	if cfg.EncryptionKeyHex != "" && len(cfg.EncryptionKeyHex) != 64 {
		return Config{}, fmt.Errorf("ENCRYPTION_KEY_HEX must be 64 hex characters when set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
