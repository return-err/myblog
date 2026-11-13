package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setValidEnv(t *testing.T) {
	t.Helper()

	env := map[string]string{
		"JWT_SECRET":         "test-secret-key-for-unit-tests-123456",
		"DB_HOST":            "localhost",
		"DB_PORT":            "5432",
		"DB_USER":            "testuser",
		"DB_PASSWORD":        "test-db-password",
		"DB_NAME":            "testdb",
		"INITIAL_ADMIN_USER": "testadmin",
		"INITIAL_ADMIN_PASS": "test-password-123",
	}

	for key, value := range env {
		t.Setenv(key, value)
	}
}

func TestValidateStartup(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T)
		wantErr  string
		wantPort string
	}{
		{
			name: "valid environment",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("PORT", "")
			},
			wantErr:  "",
			wantPort: "8080",
		},
		{
			name: "missing JWT_SECRET",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("JWT_SECRET", "")
			},
			wantErr: "JWT_SECRET is required",
		},
		{
			name: "short JWT_SECRET",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("JWT_SECRET", "short")
			},
			wantErr: "JWT_SECRET must be at least 16 bytes",
		},
		{
			name: "missing database host",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("DB_HOST", "")
			},
			wantErr: "DB_HOST is required",
		},
		{
			name: "missing initial admin password",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("INITIAL_ADMIN_PASS", "")
			},
			wantErr: "INITIAL_ADMIN_PASS is required",
		},
		{
			name: "missing database password",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("DB_PASSWORD", "")
			},
			wantErr: "DB_PASSWORD is required",
		},
		{
			name: "custom port",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("PORT", "9090")
			},
			wantErr:  "",
			wantPort: "9090",
		},
		{
			name: "non-numeric PORT",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("PORT", "abc")
			},
			wantErr: `POST must be an integer between 1 and 65535, got "abc"`,
		},
		{
			name: "PORT out of range",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("PORT", "70000")
			},
			wantErr: `POST must be an integer between 1 and 65535, got "70000"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			cfg, err := Load()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "test-secret-key-for-unit-tests-123456", cfg.JWTSecret)
				assert.Equal(t, tt.wantPort, cfg.Port)
			}
		})
	}
}
