package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type AdminConfig struct {
	Username string
	Password string
}

type Config struct {
	Port string
	// LogFormat  string
	JWTSecret string
	DB        DBConfig
	// AccessTTL  time.Duration
	// RefreshTTL time.Duration
	Admin AdminConfig
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

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return nil, fmt.Errorf("POST must be an integer between 1 and 65535, got %q", port)
	}
	port = strconv.Itoa(n)

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 bytes")
	}
	if secret == "secret-key" {
		return nil, fmt.Errorf("JWT_SECRET must not use a default value")
	}

	return &Config{
		Port:      port,
		JWTSecret: secret,
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
		Admin: AdminConfig{
			Username: os.Getenv("INITIAL_ADMIN_USER"),
			Password: os.Getenv("INITIAL_ADMIN_PASS"),
		},
	}, nil
}
