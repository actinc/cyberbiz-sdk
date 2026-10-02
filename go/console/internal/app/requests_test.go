package app_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteRecordsEveryAttempt(t *testing.T) {
	env := newTestEnv(t)
	env.login()
	shop := env.createShop()

	resp, body := env.do(http.MethodPost, "/api/shops/1/requests", map[string]any{
		"method": "GET", "path": "/v1/orders", "query": map[string][]string{"page": {"1"}},
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	var last db.OutboundLog
	env.decode(body, &last)
	assert.Equal(t, 2, last.Attempt)
	assert.Equal(t, 200, last.ResponseStatus)
	assert.Equal(t, "page=1", last.Query)
	assert.Equal(t, "req-orders-2", last.RequestID)
	assert.Contains(t, last.ResponseBody, "buyer@example.com")

	resp, body = env.do(http.MethodGet, "/api/outbound?shop_id=1&path=/v1/orders", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var page pageJSON
	env.decode(body, &page)
	assert.Equal(t, int64(2), page.Total)
	var first, second db.OutboundLog
	require.NoError(t, json.Unmarshal(page.Items[1], &first))
	require.NoError(t, json.Unmarshal(page.Items[0], &second))
	assert.Equal(t, 1, first.Attempt)
	assert.Equal(t, 429, first.ResponseStatus)
	assert.Equal(t, 2, second.Attempt)
	assert.Equal(t, shop.ID, second.ShopID)
	assert.NotContains(t, string(page.Items[0]), "response_body")

	resp, body = env.do(http.MethodGet, "/api/outbound?status=429", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	env.decode(body, &page)
	assert.Equal(t, int64(1), page.Total)
}

func TestExecuteBodyAndBinary(t *testing.T) {
	env := newTestEnv(t)
	env.login()
	env.createShop()

	resp, body := env.do(http.MethodPost, "/api/shops/1/requests", map[string]any{
		"method": "PUT", "path": "v1/orders/7", "body": map[string]any{"order": map[string]string{"note": "hi"}},
		"headers": map[string]string{"X-Debug": "1"},
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	var row db.OutboundLog
	env.decode(body, &row)
	assert.JSONEq(t, `{"order":{"note":"hi"}}`, row.RequestBody)
	assert.Contains(t, row.ResponseBody, `"echo"`)
	assert.Contains(t, row.RequestHeaders, "X-Debug")

	resp, body = env.do(http.MethodPost, "/api/shops/1/requests", map[string]any{"method": "GET", "path": "/v1/labels.zip"})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	env.decode(body, &row)
	want := []byte{0x50, 0x4b, 0x03, 0x04, 0x00, 0xff}
	assert.Equal(t, base64.StdEncoding.EncodeToString(want), row.ResponseBody)

	raw, err := env.client.Get(env.server.URL + "/api/outbound/" + itoa(row.ID) + "/body")
	require.NoError(t, err)
	defer raw.Body.Close()
	data, err := io.ReadAll(raw.Body)
	require.NoError(t, err)
	assert.Equal(t, want, data)
	assert.Equal(t, "application/zip", raw.Header.Get("Content-Type"))

	// An upstream error status is still a recorded row, not a 502.
	resp, body = env.do(http.MethodPost, "/api/shops/1/requests", map[string]any{"method": "GET", "path": "/v1/nothing"})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	env.decode(body, &row)
	assert.Equal(t, 404, row.ResponseStatus)

	resp, body = env.do(http.MethodPost, "/api/shops/1/requests", map[string]any{"method": "TRACE", "path": "/v1/orders"})
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Equal(t, 42200, body.Code)

	resp, body = env.do(http.MethodGet, "/api/stats", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body.Data), `"outbound_24h":4`)
}

func itoa(n uint) string { return strconv.FormatUint(uint64(n), 10) }
