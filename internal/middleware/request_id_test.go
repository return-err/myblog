package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		wantSame  bool
	}{
		{
			name:      "new request",
			requestID: "",
			wantSame:  false,
		},
		{
			name:      "request with request id",
			requestID: "test-123",
			wantSame:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter(RequestID(), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.requestID != "" {
				req.Header.Set("X-Request-ID", tt.requestID)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			got := w.Header().Get("X-Request-ID")
			if tt.wantSame {
				require.Equal(t, tt.requestID, got)
			} else {
				require.NotEmpty(t, got)
			}
		})
	}
}
