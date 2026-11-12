package config

import (
	"fmt"
	"os"
	"strings"
)

// type Config struct {
// 	Port       string
// 	LogFormat  string
// 	JWTSecret  string
// 	DB         DBConfig
// 	AccessTTL  time.Duration
// 	RefreshTTL time.Duration
// 	AdminUser  string
// 	AdminPass  string
// }

func ValidateStartup() (string, error) {
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

	DBPass := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	if DBPass == "" {
		return "", fmt.Errorf("DB_PASSWORD is required")
	}

	return secret, nil
}
