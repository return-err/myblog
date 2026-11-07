package handle

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandle struct {
	pinger Pinger
}

func NewHealthHandle(p Pinger) *HealthHandle {
	return &HealthHandle{pinger: p}
}

func (h *HealthHandle) Healthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *HealthHandle) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		c.String(http.StatusServiceUnavailable, "db unavailable")
		return
	}
	c.String(http.StatusOK, "ok")
}
