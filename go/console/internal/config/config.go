// Package config loads the Console configuration from environment variables
// (and an optional .env file) and validates it at boot.
package config

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config is the validated Console configuration. See console/docs/spec.md
// for the meaning and default of every variable.
type Config struct {
	Addr             string
	DBPath           string
	AdminUser        string
	AdminPassword    string
	EncryptionKey    []byte
	SessionTTL       time.Duration
	LogRetentionDays int
	GoldenDir        string
	WebhookGoldenDir string
	PublicURL        string
	LogLevel         string
	CyberbizBaseURL  string
}

// ErrMissingEncryptionKey is returned when CONSOLE_ENCRYPTION_KEY is unset.
var ErrMissingEncryptionKey = errors.New(
	"CONSOLE_ENCRYPTION_KEY is required: 32 bytes as hex or base64; generate one with `openssl rand -hex 32`")

// Load reads .env when present, then the environment.
func Load() (Config, error) {
	_ = godotenv.Load() // a missing .env is not an error
	return FromEnv(os.Getenv)
}

// FromEnv builds a Config from the given lookup function, which lets tests
// supply values without touching the process environment.
func FromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{
		Addr:             envOr(getenv, "CONSOLE_ADDR", ":8787"),
		DBPath:           envOr(getenv, "CONSOLE_DB_PATH", "./console.db"),
		AdminUser:        envOr(getenv, "CONSOLE_ADMIN_USER", "admin"),
		AdminPassword:    getenv("CONSOLE_ADMIN_PASSWORD"),
		GoldenDir:        envOr(getenv, "CONSOLE_GOLDEN_DIR", "../../testdata/golden"),
		WebhookGoldenDir: envOr(getenv, "CONSOLE_WEBHOOK_GOLDEN_DIR", "../webhook/testdata"),
		PublicURL:        getenv("CONSOLE_PUBLIC_URL"),
		LogLevel:         envOr(getenv, "CONSOLE_LOG_LEVEL", "info"),
		CyberbizBaseURL:  getenv("CYBERBIZ_BASE_URL"),
	}
	key, err := ParseKey(getenv("CONSOLE_ENCRYPTION_KEY"))
	if err != nil {
		return Config{}, err
	}
	cfg.EncryptionKey = key
	ttl, err := time.ParseDuration(envOr(getenv, "CONSOLE_SESSION_TTL", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("CONSOLE_SESSION_TTL: %w", err)
	}
	cfg.SessionTTL = ttl
	days, err := strconv.Atoi(envOr(getenv, "CONSOLE_LOG_RETENTION_DAYS", "30"))
	if err != nil || days < 0 {
		return Config{}, fmt.Errorf("CONSOLE_LOG_RETENTION_DAYS: must be a non-negative integer")
	}
	cfg.LogRetentionDays = days
	return cfg, nil
}

// ParseKey decodes a 32-byte key given as hex or base64 (std or URL, with
// or without padding).
func ParseKey(s string) ([]byte, error) {
	if s == "" {
		return nil, ErrMissingEncryptionKey
	}
	if b, err := hex.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(s); err == nil && len(b) == 32 {
			return b, nil
		}
	}
	return nil, errors.New("CONSOLE_ENCRYPTION_KEY: must decode to exactly 32 bytes (hex or base64)")
}

func envOr(getenv func(string) string, key, def string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return def
}
