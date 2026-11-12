package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	JWTSecret string
	// Port       string
	// LogFormat  string
	// JWTSecret  string
	// DB         DBConfig
	// AccessTTL  time.Duration
	// RefreshTTL time.Duration
	// AdminUser  string
	// AdminPass  string
}

func Load() (*Config, error) {
	require := []string{
		"JWT_SECRET",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"INITIAL_ADMIN_USER",
		"INITIAL_ADMIN_PASS",
	}
	for _, key := range require {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return nil, fmt.Errorf("%s is required", key)
		}
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 bytes")
	}
	if secret == "secret-key" {
		return nil, fmt.Errorf("JWT_SECRET must not use a default value")
	}

	return &Config{JWTSecret: secret}, nil
}
