package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const samplePayload = `{"id":1001,"order_number":"TEST-0001","email":"buyer@example.com","total":"100.0"}`

func (e *testEnv) postWebhook(body []byte, headers map[string]string) *http.Response {
	e.t.Helper()
	req, err := http.NewRequest(http.MethodPost, e.server.URL+"/webhooks/cyberbiz", bytes.NewReader(body))
	require.NoError(e.t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", webhook.UserAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(e.t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp
}

func signedHeaders(body []byte, secret, domain, event string) map[string]string {
	return map[string]string{
		webhook.HeaderDomain:     domain,
		webhook.HeaderShopDomain: "shop.example.com",
		webhook.HeaderEvent:      event,
		webhook.HeaderSignature:  webhook.Sign(body, secret),
		webhook.HeaderDomainHMAC: webhook.SignDomain(domain, secret),
	}
}

func lastInbound(t *testing.T, env *testEnv) db.InboundLog {
	t.Helper()
	var row db.InboundLog
	require.NoError(t, env.app.DB.Order("id DESC").First(&row).Error)
	return row
}

func TestWebhookOutcomes(t *testing.T) {
	env := newTestEnv(t)
	env.login()
	shop := env.createShop()
	body := []byte(samplePayload)

	resp := env.postWebhook(body, signedHeaders(body, appSecret, testDomain, "orders/paid"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	row := lastInbound(t, env)
	assert.Equal(t, db.InboundValid, row.Status)
	assert.True(t, row.SignatureValid)
	require.NotNil(t, row.DomainSignatureValid)
	assert.True(t, *row.DomainSignatureValid)
	require.NotNil(t, row.ShopID)
	assert.Equal(t, shop.ID, *row.ShopID)
	assert.Equal(t, "orders/paid", row.Event)
	assert.Nil(t, row.DuplicateOf)
	firstID := row.ID

	// Same delivery again → duplicate_of the first row.
	resp = env.postWebhook(body, signedHeaders(body, appSecret, testDomain, "orders/paid"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	row = lastInbound(t, env)
	require.NotNil(t, row.DuplicateOf)
	assert.Equal(t, firstID, *row.DuplicateOf)

	// Same body under a different Event is not a redelivery.
	resp = env.postWebhook(body, signedHeaders(body, appSecret, testDomain, "orders/closed"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	row = lastInbound(t, env)
	assert.Equal(t, "orders/closed", row.Event)
	assert.Nil(t, row.DuplicateOf)

	resp = env.postWebhook(body, signedHeaders(body, "wrong-secret", testDomain, "orders/paid"))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	row = lastInbound(t, env)
	assert.Equal(t, db.InboundInvalidSignature, row.Status)
	assert.False(t, row.SignatureValid)
	assert.Equal(t, 401, row.ResponseStatus)

	resp = env.postWebhook(body, signedHeaders(body, appSecret, "nobody.cyberbiz.co", "orders/paid"))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	row = lastInbound(t, env)
	assert.Equal(t, db.InboundUnknownShop, row.Status)
	assert.Nil(t, row.ShopID)

	resp = env.postWebhook(body, map[string]string{webhook.HeaderDomain: testDomain})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	row = lastInbound(t, env)
	assert.Equal(t, db.InboundMalformed, row.Status)

	notJSON := []byte(`{"broken":`)
	resp = env.postWebhook(notJSON, signedHeaders(notJSON, appSecret, testDomain, "orders/paid"))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, db.InboundMalformed, lastInbound(t, env).Status)

	huge := bytes.Repeat([]byte("a"), 2<<20+1)
	resp = env.postWebhook(huge, signedHeaders(huge, appSecret, testDomain, "orders/paid"))
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
	assert.Equal(t, db.InboundMalformed, lastInbound(t, env).Status)

	apiResp, env2 := env.do(http.MethodGet, "/api/inbound?status=valid", nil)
	require.Equal(t, http.StatusOK, apiResp.StatusCode)
	var page pageJSON
	env.decode(env2, &page)
	assert.Equal(t, int64(3), page.Total) // paid, its redelivery, and closed
	assert.NotContains(t, string(page.Items[0]), `"body"`)

	apiResp, env2 = env.do(http.MethodGet, "/api/inbound/1", nil)
	require.Equal(t, http.StatusOK, apiResp.StatusCode)
	var full db.InboundLog
	require.NoError(t, json.Unmarshal(env2.Data, &full))
	assert.Equal(t, samplePayload, full.Body)
	// Headers reach the UI as an object, not as a string of JSON text.
	var shape struct {
		Headers map[string]any `json:"headers"`
	}
	require.NoError(t, json.Unmarshal(env2.Data, &shape))
	assert.NotEmpty(t, shape.Headers)
}
