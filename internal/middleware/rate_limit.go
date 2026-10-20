package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/handle"
	"golang.org/x/time/rate"
)

type IPLimiter struct {
	mu      sync.Mutex
	limiter map[string]*rate.Limiter
	r       rate.Limit
	burst   int
}

func NewIPLimiter(r rate.Limit, burst int) *IPLimiter {
	return &IPLimiter{
		limiter: make(map[string]*rate.Limiter),
		r:       r,
		burst:   burst,
	}
}

func (u *IPLimiter) GetLimiter(ip string) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	limiter, ok := u.limiter[ip]
	if !ok {
		limiter = rate.NewLimiter(u.r, u.burst)
		u.limiter[ip] = limiter
	}

	return limiter
}

func (u *IPLimiter) Allow(ip string) bool {
	limiter := u.GetLimiter(ip)
	return limiter.Allow()
}

func LoginRateLimit(limiter *IPLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		if !limiter.Allow(ip) {
			handle.RespondWithError(ctx, apperror.ErrLoginRateLimit)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
