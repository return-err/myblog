package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/healthz" || path == "/readyz" {
			ctx.Next()
			return
		}

		start := time.Now()
		ctx.Next()

		status := ctx.Writer.Status()
		attrs := []any{
			"request_id", ctx.GetString("request_id"),
			"method", ctx.Request.Method,
			"path", path,
			"route", ctx.FullPath(),
			"status", status,
			"latency_ms", float64(time.Since(start).Microseconds()) / 1000,
			"ip", ctx.ClientIP(),
		}

		switch {
		case status >= 500:
			slog.Error("request", attrs...)
		case status >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	}
}
