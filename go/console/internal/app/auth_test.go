package app_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFlow(t *testing.T) {
	env := newTestEnv(t)

	resp, body := env.do(http.MethodGet, "/api/health", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, body.Success)

	resp, body = env.do(http.MethodGet, "/api/auth/me", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, 40100, body.Code)

	resp, body = env.do(http.MethodPost, "/api/auth/login", map[string]string{"username": "admin", "password": "wrong"})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.False(t, body.Success)

	env.login()
	resp, body = env.do(http.MethodGet, "/api/auth/me", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var me struct {
		User          struct{ Username string } `json:"user"`
		SetupRequired bool                      `json:"setup_required"`
	}
	env.decode(body, &me)
	assert.Equal(t, "admin", me.User.Username)
	assert.True(t, me.SetupRequired)

	resp, _ = env.do(http.MethodPost, "/api/auth/logout", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, _ = env.do(http.MethodGet, "/api/auth/me", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUIFallbackAndAPINotFound(t *testing.T) {
	env := newTestEnv(t)
	resp, err := http.Get(env.server.URL + "/some/client/route")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	resp2, body := env.do(http.MethodGet, "/api/does-not-exist", nil)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
	assert.Equal(t, 40400, body.Code)
}
