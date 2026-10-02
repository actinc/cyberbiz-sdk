package app_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoldenExport(t *testing.T) {
	env := newTestEnv(t)
	env.login()
	env.createShop()

	resp, body := env.do(http.MethodPost, "/api/shops/1/requests", map[string]any{"method": "GET", "path": "/v1/orders/123"})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	var row db.OutboundLog
	env.decode(body, &row)

	resp, body = env.do(http.MethodPost, "/api/outbound/"+itoa(row.ID)+"/golden", map[string]any{})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	want := filepath.Join(env.cfg.GoldenDir, "v1", "GET_v1_orders_{id}.json")
	assert.Contains(t, string(body.Data), "GET_v1_orders_{id}.json")
	data, err := os.ReadFile(want)
	require.NoError(t, err)
	assert.Contains(t, string(data), "redacted@example.com")
	assert.NotContains(t, string(data), "王小明")
	headers, err := os.ReadFile(filepath.Join(env.cfg.GoldenDir, "v1", "GET_v1_orders_{id}.headers.json"))
	require.NoError(t, err)
	assert.Contains(t, string(headers), "Content-Type")

	resp, body = env.do(http.MethodPost, "/api/outbound/"+itoa(row.ID)+"/golden", map[string]any{})
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	assert.Equal(t, 40900, body.Code)

	resp, body = env.do(http.MethodPost, "/api/outbound/"+itoa(row.ID)+"/golden", map[string]any{"name": "include", "overwrite": true})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	_, err = os.Stat(filepath.Join(env.cfg.GoldenDir, "v1", "GET_v1_orders_{id}_include.json"))
	assert.NoError(t, err)

	// Inbound golden.
	payload := []byte(samplePayload)
	wresp := env.postWebhook(payload, signedHeaders(payload, appSecret, testDomain, "orders/paid"))
	require.Equal(t, http.StatusOK, wresp.StatusCode)
	inRow := lastInbound(t, env)
	resp, body = env.do(http.MethodPost, "/api/inbound/"+itoa(inRow.ID)+"/golden", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	got, err := os.ReadFile(filepath.Join(env.cfg.WebhookGoldenDir, "orders_paid.json"))
	require.NoError(t, err)
	assert.Contains(t, string(got), "redacted@example.com")
	hdr, err := os.ReadFile(filepath.Join(env.cfg.WebhookGoldenDir, "orders_paid.headers.json"))
	require.NoError(t, err)
	assert.Contains(t, string(hdr), "X-Cyberbiz-Event")
}
