package app_test

import (
	"net/http"
	"testing"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShopCRUD(t *testing.T) {
	env := newTestEnv(t)
	env.login()

	shop := env.createShop()
	assert.Equal(t, int64(12345), shop.CyberbizShopID)
	assert.Equal(t, "shop.example.com", shop.CustomDomain)
	assert.True(t, shop.HasAppSecret)
	assert.Regexp(t, `^sha256:[0-9a-f]{8} len=\d+$`, shop.TokenFingerprint)
	assert.Equal(t, "https://tunnel.example.test/webhooks/cyberbiz", shop.WebhookURL)

	// The validation call was recorded as an Outbound tagged with the shop.
	var rows []db.OutboundLog
	require.NoError(t, env.app.DB.Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Equal(t, "/shop", rows[0].Path)
	assert.Equal(t, shop.ID, rows[0].ShopID)
	assert.Equal(t, 200, rows[0].ResponseStatus)
	assert.Equal(t, "req-shop-1", rows[0].RequestID)
	assert.NotContains(t, rows[0].RequestHeaders, goodToken)
	assert.Contains(t, rows[0].RequestHeaders, "Authorization")

	// Duplicate domain → 409.
	resp, body := env.do(http.MethodPost, "/api/shops", map[string]string{
		"name": "Dup", "shop_domain": testDomain, "api_token": goodToken,
	})
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	assert.Equal(t, 40900, body.Code)

	resp, body = env.do(http.MethodGet, "/api/shops", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list []shopJSON
	env.decode(body, &list)
	require.Len(t, list, 1)

	// Update without token keeps the fingerprint.
	resp, body = env.do(http.MethodPut, "/api/shops/1", map[string]string{"name": "Renamed"})
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	var updated shopJSON
	env.decode(body, &updated)
	assert.Equal(t, "Renamed", updated.Name)
	assert.Equal(t, shop.TokenFingerprint, updated.TokenFingerprint)

	resp, body = env.do(http.MethodPost, "/api/shops/1/verify", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, body.Error)
	assert.Contains(t, string(body.Data), `"shop_info"`)

	resp, _ = env.do(http.MethodDelete, "/api/shops/1", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, body = env.do(http.MethodGet, "/api/shops/1", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, 40400, body.Code)

	// Logs survive the delete.
	require.NoError(t, env.app.DB.Find(&rows).Error)
	assert.NotEmpty(t, rows)
}

func TestShopCreateRejectsBadToken(t *testing.T) {
	env := newTestEnv(t)
	env.login()
	resp, body := env.do(http.MethodPost, "/api/shops", map[string]string{
		"name": "Bad", "shop_domain": "bad.cyberbiz.co", "api_token": "wrong-token",
	})
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	assert.Equal(t, 50200, body.Code)
	assert.Contains(t, string(body.Data), `"status":401`)

	var n int64
	require.NoError(t, env.app.DB.Model(&db.Shop{}).Count(&n).Error)
	assert.Zero(t, n)

	resp, body = env.do(http.MethodPost, "/api/shops", map[string]string{"name": "x"})
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Equal(t, 42200, body.Code)
}
