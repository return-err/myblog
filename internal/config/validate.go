package config

import (
	"fmt"
	"os"
	"strings"
)

func ValidateStartup() (string, error) {
	require := []string{
		"JWT_SECRET",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_NAME",
		"INITIAL_ADMIN_USER",
		"INITIAL_ADMIN_PASS",
	}
	for _, key := range require {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return "", fmt.Errorf("%s is required", key)
		}
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 16 {
		return "", fmt.Errorf("JWT_SECRET must be at least 16 bytes")
	}
	if secret == "secret-key" {
		return "", fmt.Errorf("JWT_SECRET must not use a default value")
	}

	return secret, nil
}
