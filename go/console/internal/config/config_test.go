package config

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromEnvRequiresKey(t *testing.T) {
	_, err := FromEnv(func(string) string { return "" })
	require.ErrorIs(t, err, ErrMissingEncryptionKey)
	assert.Contains(t, err.Error(), "openssl rand -hex 32")
}

func TestParseKey(t *testing.T) {
	raw := bytes.Repeat([]byte{1}, 32)
	hexKey, err := ParseKey("0101010101010101010101010101010101010101010101010101010101010101")
	require.NoError(t, err)
	assert.Equal(t, raw, hexKey)
	b64Key, err := ParseKey(base64.StdEncoding.EncodeToString(raw))
	require.NoError(t, err)
	assert.Equal(t, raw, b64Key)
	_, err = ParseKey("abc")
	assert.Error(t, err)
}

func TestDefaults(t *testing.T) {
	cfg, err := FromEnv(func(k string) string {
		if k == "CONSOLE_ENCRYPTION_KEY" {
			return "0101010101010101010101010101010101010101010101010101010101010101"
		}
		return ""
	})
	require.NoError(t, err)
	assert.Equal(t, ":8787", cfg.Addr)
	assert.Equal(t, 30, cfg.LogRetentionDays)
	assert.Equal(t, "admin", cfg.AdminUser)
	assert.Equal(t, "../../testdata/golden", cfg.GoldenDir)
}
