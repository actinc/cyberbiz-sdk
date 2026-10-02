// Package stats serves the dashboard counters.
package stats

import (
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/inbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/outbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/shops"
	"github.com/gin-gonic/gin"
)

// Handlers exposes GET /api/stats.
type Handlers struct {
	shops    *shops.Service
	outbound *outbound.Repository
	inbound  *inbound.Repository
}

// NewHandlers wires the counters.
func NewHandlers(s *shops.Service, o *outbound.Repository, i *inbound.Repository) *Handlers {
	return &Handlers{shops: s, outbound: o, inbound: i}
}

// Register mounts the route on an authenticated group.
func (h *Handlers) Register(r gin.IRoutes) { r.GET("/stats", h.get) }

func (h *Handlers) get(c *gin.Context) {
	ctx := c.Request.Context()
	since := time.Now().Add(-24 * time.Hour)
	nShops, err := h.shops.Count(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	nOut, err := h.outbound.CountSince(ctx, since)
	if err != nil {
		response.Fail(c, err)
		return
	}
	nIn, err := h.inbound.CountSince(ctx, since, false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	nInvalid, err := h.inbound.CountSince(ctx, since, true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"shops": nShops, "outbound_24h": nOut, "inbound_24h": nIn, "inbound_invalid_24h": nInvalid})
}
