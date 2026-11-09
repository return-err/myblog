package handle

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPinger struct {
	Pinger
	mockPing func(ctx context.Context) error
}

func (p mockPinger) Ping(ctx context.Context) error {
	return p.mockPing(ctx)
}

func TestReadyz(t *testing.T) {
	tests := []struct {
		name     string
		mockPing func(ctx context.Context) error
		wantCode int
		wantBody string
	}{
		{
			name: "success",
			mockPing: func(ctx context.Context) error {
				return ctx.Err()
			},
			wantCode: http.StatusOK,
			wantBody: "ok",
		},
		{
			name: "db down",
			mockPing: func(ctx context.Context) error {
				return errors.New("connection refused")
			},
			wantCode: http.StatusServiceUnavailable,
			wantBody: "db unavailable",
		},
		{
			name: "timeout",
			mockPing: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
			wantCode: http.StatusServiceUnavailable,
			wantBody: "db unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPinger := mockPinger{
				mockPing: tt.mockPing,
			}
			mockHandle := NewHealthHandle(mockPinger)
			r := setupTestRouter(http.MethodGet, "/test", mockHandle.Readyz)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
