package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestGetLimiter(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		wantReuse bool
	}{
		{
			name:      "creates limiter for new IP",
			ip:        "192.168.1.1",
			wantReuse: false,
		},
		{
			name:      "reuses limiter for existing IP",
			ip:        "192.168.1.1",
			wantReuse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &IPLimiter{
				limiter: make(map[string]*rate.Limiter),
				r:       rate.Limit(1),
				burst:   5,
			}

			first := u.GetLimiter(tt.ip)
			require.NotNil(t, first)
			_, ok := u.limiter[tt.ip]
			require.True(t, ok)

			if tt.wantReuse {
				second := u.GetLimiter(tt.ip)
				require.Same(t, first, second)
			} else {
				require.Len(t, u.limiter, 1)
			}
		})
	}
}
