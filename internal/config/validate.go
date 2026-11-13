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
	Port      int
	LogFormat string
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

	portStr := strings.TrimSpace(os.Getenv("PORT"))
	if portStr == "" {
		portStr = "8080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("PORT must be an integer between 1 and 65535, got %q", portStr)
	}

	logFormat := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	if logFormat == "" {
		logFormat = "json"
	}
	if logFormat != "json" && logFormat != "text" {
		return nil, fmt.Errorf("LOG_FORMAT must be \"json\" or \"text\", got %q", logFormat)
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 bytes")
	}
	if secret == "secret-key" {
		return nil, fmt.Errorf("JWT_SECRET must not use a default value")
	}

	return &Config{
		Port:      port,
		LogFormat: logFormat,
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
