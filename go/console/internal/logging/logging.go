// Package logging configures zerolog and bridges it to log/slog for the SDK.
package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup configures the global zerolog logger at the given level.
func Setup(level string) error {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(lvl)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).With().Timestamp().Logger()
	return nil
}

// Slog returns a slog.Logger whose records are written through l.
func Slog(l zerolog.Logger) *slog.Logger {
	return slog.New(&handler{logger: l})
}

type handler struct {
	logger zerolog.Logger
	attrs  []slog.Attr
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return toZerolog(level) >= h.logger.GetLevel() && toZerolog(level) >= zerolog.GlobalLevel()
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	ev := h.logger.WithLevel(toZerolog(r.Level))
	for _, a := range h.attrs {
		ev = ev.Interface(a.Key, a.Value.Any())
	}
	r.Attrs(func(a slog.Attr) bool {
		ev = ev.Interface(a.Key, a.Value.Any())
		return true
	})
	ev.Msg(r.Message)
	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &handler{logger: h.logger, attrs: merged}
}

func (h *handler) WithGroup(string) slog.Handler { return h }

func toZerolog(level slog.Level) zerolog.Level {
	switch {
	case level >= slog.LevelError:
		return zerolog.ErrorLevel
	case level >= slog.LevelWarn:
		return zerolog.WarnLevel
	case level >= slog.LevelInfo:
		return zerolog.InfoLevel
	default:
		return zerolog.DebugLevel
	}
}
