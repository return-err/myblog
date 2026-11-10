package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func LogFormat(p gin.LogFormatterParams) string {
	return fmt.Sprintf("[%s] %s %s %d %s\n", p.Keys["request_id"], p.Method, p.Path, p.StatusCode, p.Latency)
}
