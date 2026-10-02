package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/app"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/config"
	"github.com/stretchr/testify/require"
)

const (
	testKeyHex   = "1111111111111111111111111111111111111111111111111111111111111111"
	goodToken    = "good-token-for-tests"
	appSecret    = "app-secret-for-tests"
	testDomain   = "example.cyberbiz.co"
	adminPass    = "admin-password"
	fakeShopBody = `{"shop_info":{"id":12345,"name":"Example Shop","primary_domain":"shop.example.com","currency":"TWD"}}`
)

// fakeCyberbiz is a stand-in for app-store-api.cyberbiz.io.
type fakeCyberbiz struct {
	*httptest.Server
	ordersCalls atomic.Int32
}

func newFakeCyberbiz(t *testing.T) *fakeCyberbiz {
	t.Helper()
	f := &fakeCyberbiz{}
	mux := http.NewServeMux()
	f.registerShop(mux)
	f.registerResources(mux)
	f.Server = httptest.NewServer(mux)
	t.Cleanup(f.Close)
	return f
}

func (f *fakeCyberbiz) registerShop(mux *http.ServeMux) {
	mux.HandleFunc("/shop", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+goodToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid token"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req-shop-1")
		_, _ = w.Write([]byte(fakeShopBody))
	})
}

func (f *fakeCyberbiz) registerResources(mux *http.ServeMux) {
	mux.HandleFunc("/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if f.ordersCalls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		w.Header().Set("X-Request-Id", "req-orders-2")
		_, _ = w.Write([]byte(`[{"id":1,"email":"buyer@example.com","total":"100.0"}]`))
	})
	mux.HandleFunc("/v1/orders/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			_, _ = w.Write(append([]byte(`{"echo":`), append(body, '}')...))
			return
		}
		_, _ = w.Write([]byte(`{"id":123,"customer_name":"王小明","email":"buyer@example.com"}`))
	})
	mux.HandleFunc("/v1/labels.zip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write([]byte{0x50, 0x4b, 0x03, 0x04, 0x00, 0xff})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":["無此資源"]}`))
	})
}

// testEnv is one Console under test with a logged-in HTTP client.
type testEnv struct {
	t      *testing.T
	app    *app.App
	server *httptest.Server
	client *http.Client
	fake   *fakeCyberbiz
	cfg    config.Config
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	fake := newFakeCyberbiz(t)
	dir := t.TempDir()
	env := map[string]string{
		"CONSOLE_DB_PATH":            filepath.Join(dir, "console.db"),
		"CONSOLE_ENCRYPTION_KEY":     testKeyHex,
		"CONSOLE_ADMIN_PASSWORD":     adminPass,
		"CONSOLE_GOLDEN_DIR":         filepath.Join(dir, "golden"),
		"CONSOLE_WEBHOOK_GOLDEN_DIR": filepath.Join(dir, "webhook-golden"),
		"CONSOLE_PUBLIC_URL":         "https://tunnel.example.test",
		"CYBERBIZ_BASE_URL":          fake.URL,
	}
	cfg, err := config.FromEnv(func(k string) string { return env[k] })
	require.NoError(t, err)
	noLimit := 0.0
	a, err := app.New(cfg, app.Options{Stdout: io.Discard, RateLimit: &noLimit})
	require.NoError(t, err)
	srv := httptest.NewServer(a.Engine)
	t.Cleanup(srv.Close)
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &testEnv{t: t, app: a, server: srv, client: &http.Client{Jar: jar, Timeout: 10 * time.Second}, fake: fake, cfg: cfg}
}

// envelope is the decoded response envelope.
type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
	Code    int             `json:"code"`
}

func (e *testEnv) do(method, path string, body any, headers ...string) (*http.Response, envelope) {
	e.t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(e.t, err)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.server.URL+path, reader)
	require.NoError(e.t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := e.client.Do(req)
	require.NoError(e.t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(e.t, err)
	var env envelope
	if len(raw) > 0 && bytes.HasPrefix(bytes.TrimSpace(raw), []byte("{")) {
		require.NoError(e.t, json.Unmarshal(raw, &env), "body: %s", raw)
	}
	return resp, env
}

func (e *testEnv) login() {
	e.t.Helper()
	resp, env := e.do(http.MethodPost, "/api/auth/login", map[string]string{"username": "admin", "password": adminPass})
	require.Equal(e.t, http.StatusOK, resp.StatusCode)
	require.True(e.t, env.Success)
}

func (e *testEnv) decode(env envelope, out any) {
	e.t.Helper()
	require.NoError(e.t, json.Unmarshal(env.Data, out), "data: %s", env.Data)
}

type shopJSON struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	ShopDomain       string `json:"shop_domain"`
	CustomDomain     string `json:"custom_domain"`
	CyberbizShopID   int64  `json:"cyberbiz_shop_id"`
	HasAppSecret     bool   `json:"has_app_secret"`
	TokenFingerprint string `json:"token_fingerprint"`
	WebhookURL       string `json:"webhook_url"`
}

func (e *testEnv) createShop() shopJSON {
	e.t.Helper()
	resp, env := e.do(http.MethodPost, "/api/shops", map[string]string{
		"name": "Test", "shop_domain": testDomain, "app_name": "console", "app_id": "app-1",
		"app_secret": appSecret, "api_token": goodToken,
	})
	require.Equal(e.t, http.StatusOK, resp.StatusCode, env.Error)
	var shop shopJSON
	e.decode(env, &shop)
	return shop
}

type pageJSON struct {
	Items []json.RawMessage `json:"items"`
	Total int64             `json:"total"`
}
