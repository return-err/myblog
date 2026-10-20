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
		secondIP  string
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
		{
			name:     "creates separate limiter for different IPs",
			ip:       "192.168.1.1",
			secondIP: "192.168.1.2",
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
			require.Equal(t, rate.Limit(1), first.Limit())
			require.Equal(t, 5, first.Burst())

			if tt.wantReuse {
				second := u.GetLimiter(tt.ip)
				require.Same(t, first, second)
			} else {
				require.Len(t, u.limiter, 1)
			}

			if tt.secondIP != "" {
				second := u.GetLimiter(tt.secondIP)
				require.NotNil(t, second)
				_, ok := u.limiter[tt.secondIP]
				require.True(t, ok)

				require.NotSame(t, first, second)
				require.Len(t, u.limiter, 2)
			}
		})
	}
}
