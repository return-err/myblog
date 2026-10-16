package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/handle"
	"golang.org/x/time/rate"
)

func LoginRateLimit(limiter *rate.Limiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !limiter.Allow() {
			handle.RespondWithError(ctx, apperror.ErrLoginRateLimit)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
