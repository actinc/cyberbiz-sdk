// Package app wires configuration, database, services and the router into
// one runnable Console.
package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/auth"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/catalog"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/config"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/crypto"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/inbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/logging"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/outbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/retention"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/router"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/shops"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/stats"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Version is reported by GET /api/health.
const Version = "0.1.0"

// App is a fully wired Console.
type App struct {
	Config    config.Config
	DB        *gorm.DB
	Engine    *gin.Engine
	Shops     *shops.Service
	Auth      *auth.Service
	Retention *retention.Job
}

// Options adjust wiring for tests.
type Options struct {
	// Stdout receives the generated admin password; defaults to os.Stdout.
	Stdout io.Writer
	// Transport is the base RoundTripper the recorder wraps; nil means
	// http.DefaultTransport.
	Transport http.RoundTripper
	// RateLimit overrides the SDK rate limit (tests pass 0 to disable).
	RateLimit *float64
}

// New opens the database, runs migrations, creates the admin user, and
// builds the router.
func New(cfg config.Config, opts Options) (*App, error) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	cipher, err := crypto.New(cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}
	g, err := db.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	cat, err := catalog.Load()
	if err != nil {
		return nil, err
	}

	authSvc := auth.NewService(auth.NewRepository(g), cfg.SessionTTL)
	if err := authSvc.EnsureAdmin(context.Background(), cfg.AdminUser, cfg.AdminPassword, opts.Stdout); err != nil {
		return nil, err
	}
	outRepo := outbound.NewRepository(g)
	inRepo := inbound.NewRepository(g)
	shopSvc := shops.NewService(shops.NewRepository(g), cipher, shops.Options{
		Transport: outbound.NewRecorder(outRepo, opts.Transport),
		BaseURL:   cfg.CyberbizBaseURL,
		PublicURL: cfg.PublicURL,
		Logger:    logging.Slog(log.Logger),
		RateLimit: opts.RateLimit,
	})
	setupRequired := func(c *gin.Context) (bool, error) {
		n, err := shopSvc.Count(c.Request.Context())
		return n == 0, err
	}
	engine := router.New(router.Deps{
		Version:  Version,
		Auth:     authSvc,
		AuthH:    auth.NewHandlers(authSvc, setupRequired),
		Shops:    shops.NewHandlers(shopSvc),
		Outbound: outbound.NewHandlers(outRepo, outbound.NewExecutor(shopSvc, outRepo), cfg.GoldenDir),
		Inbound:  inbound.NewHandlers(inRepo, cfg.WebhookGoldenDir),
		Receiver: inbound.NewReceiver(inRepo, shopSvc),
		Stats:    stats.NewHandlers(shopSvc, outRepo, inRepo),
		Catalog:  cat,
	})
	return &App{
		Config: cfg, DB: g, Engine: engine, Shops: shopSvc, Auth: authSvc,
		Retention: retention.New(outRepo, inRepo, cfg.LogRetentionDays),
	}, nil
}

// Run starts the retention job and serves until ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	a.Retention.Start(ctx, 24*time.Hour)
	srv := &http.Server{Addr: a.Config.Addr, Handler: a.Engine, ReadHeaderTimeout: 10 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", a.Config.Addr).Msg("console listening")
		errCh <- srv.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
