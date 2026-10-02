// Command console runs the CYBERBIZ Console: the API, the webhook endpoint
// and the embedded UI on one port.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/app"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/config"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/logging"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "console:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := logging.Setup(cfg.LogLevel); err != nil {
		return fmt.Errorf("CONSOLE_LOG_LEVEL: %w", err)
	}
	a, err := app.New(cfg, app.Options{})
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Info().Str("db", cfg.DBPath).Int("retention_days", cfg.LogRetentionDays).Msg("console starting")
	return a.Run(ctx)
}
