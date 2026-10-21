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
		name    string
		setup   func(t *testing.T)
		wantErr string
	}{
		{
			name: "valid environment",
			setup: func(t *testing.T) {
				setValidEnv(t)
			},
			wantErr: "",
		},
		{
			name: "missing JWT_SECRET",
			setup: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("JWT_SECRET", "")
			},
			wantErr: "JWT_SECRT is required",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			secret, err := ValidateStartup()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, "", secret)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
