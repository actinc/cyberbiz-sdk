// Package router assembles the Gin engine from the feature handlers.
package router

import (
	"github.com/actinc/cyberbiz-sdk/go/console/internal/auth"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/catalog"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/inbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/outbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/shops"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/stats"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/ui"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Deps are the handlers the router mounts.
type Deps struct {
	Version  string
	Auth     *auth.Service
	AuthH    *auth.Handlers
	Shops    *shops.Handlers
	Outbound *outbound.Handlers
	Inbound  *inbound.Handlers
	Receiver *inbound.Receiver
	Stats    *stats.Handlers
	Catalog  *catalog.Catalog
}

// New builds the engine: public routes, the session-guarded API, the
// webhook endpoint, and the embedded UI with SPA fallback.
func New(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), requestLog())
	r.RedirectTrailingSlash = false

	public := r.Group("/api")
	public.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok", "version": d.Version})
	})
	private := r.Group("/api", d.Auth.Require())
	d.AuthH.Register(public, private)
	d.Shops.Register(private)
	d.Outbound.Register(private)
	d.Inbound.Register(private)
	d.Stats.Register(private)
	private.GET("/catalog", func(c *gin.Context) { response.OK(c, d.Catalog) })

	r.POST("/webhooks/cyberbiz", d.Receiver.Handle)
	r.NoRoute(ui.Handler())
	return r
}

func requestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		log.Debug().Str("method", c.Request.Method).Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).Msg("http")
	}
}
