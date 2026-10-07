package cyberbiz

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient returns a client pointed at handler with retries kept but
// sleeps and the rate limiter removed so tests run instantly.
func newTestClient(t *testing.T, handler http.HandlerFunc, opts ...Option) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	base := []Option{WithBaseURL(srv.URL), WithRateLimit(0)}
	c, err := New("test-token", append(base, opts...)...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c.sleep = func(time.Duration) {}
	return c, srv
}

func TestNewRejectsEmptyToken(t *testing.T) {
	if _, err := New(" "); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestNewRejectsBadOptions(t *testing.T) {
	cases := map[string]Option{
		"base url":  WithBaseURL(""),
		"rate":      WithRateLimit(-1),
		"retries":   WithMaxRetries(-1),
		"logger":    WithLogger(nil),
		"transport": WithTransport(nil),
	}
	for name, opt := range cases {
		if _, err := New("tok", opt); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestDoSendsAuthAndAcceptHeaders(t *testing.T) {
	var got http.Header
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	var out struct {
		OK bool `json:"ok"`
	}
	if _, err := c.get(context.Background(), "/v1/thing", nil, &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "Bearer test-token" {
		t.Errorf("Authorization = %q", got.Get("Authorization"))
	}
	if got.Get("Accept") != "application/json" {
		t.Errorf("Accept = %q", got.Get("Accept"))
	}
	if got.Get("User-Agent") != defaultUserAgent {
		t.Errorf("User-Agent = %q", got.Get("User-Agent"))
	}
	if !out.OK {
		t.Error("response not decoded")
	}
}

func TestDoResolvesPathAgainstBaseURL(t *testing.T) {
	var path, query string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		_, _ = w.Write([]byte(`[]`))
	})
	q := map[string][]string{"per_page": {"2"}}
	if _, err := c.get(context.Background(), "v1/orders", q, nil); err != nil {
		t.Fatal(err)
	}
	if path != "/v1/orders" || query != "per_page=2" {
		t.Errorf("got %s?%s", path, query)
	}
}

func TestDoParsesPaginationHeaders(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Page", "2")
		h.Set("X-Per-Page", "50")
		h.Set("X-Offset", "0")
		h.Set("X-Total", "195")
		h.Set("X-Total-Pages", "4")
		h.Set("X-Next-Page", "3")
		h.Set("X-Prev-Page", "1")
		h.Set("X-Request-Id", "req-123")
		_, _ = w.Write([]byte(`[{"id":1},{"id":2}]`))
	})
	page, err := list[struct {
		ID int64 `json:"id"`
	}](context.Background(), c, "v1/orders", nil)
	if err != nil {
		t.Fatal(err)
	}
	want := Pagination{Page: 2, PerPage: 50, Total: 195, TotalPages: 4, NextPage: 3, PrevPage: 1}
	if page.Pagination != want {
		t.Errorf("pagination = %+v, want %+v", page.Pagination, want)
	}
	if len(page.Items) != 2 || page.Items[1].ID != 2 {
		t.Errorf("items = %+v", page.Items)
	}
	if page.Response.RequestID != "req-123" {
		t.Errorf("request id = %q", page.Response.RequestID)
	}
}

func TestDoMapsStatusToSentinels(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   error
		msg    string
	}{
		{401, `{"error":["拒絕存取"]}`, ErrUnauthorized, "拒絕存取"},
		{403, `{"error":"Must have pos_shop_coupon plugin"}`, ErrForbidden, "Must have pos_shop_coupon plugin"},
		{404, `{"error":["無此資源"]}`, ErrNotFound, "無此資源"},
		{422, `{"message":"無效的 Provider Type"}`, ErrValidation, "無效的 Provider Type"},
		{403, `{"messages":"Invoice is not ready."}`, ErrForbidden, "Invoice is not ready."},
		{500, `oops`, ErrServer, ""},
	}
	for _, tc := range cases {
		c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}, WithMaxRetries(0))
		_, err := c.get(context.Background(), "v1/x", nil, nil)
		if !errors.Is(err, tc.want) {
			t.Errorf("status %d: errors.Is(%v, %v) = false", tc.status, err, tc.want)
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("status %d: not an *APIError: %T", tc.status, err)
		}
		if tc.msg != "" && (len(apiErr.Messages) != 1 || apiErr.Messages[0] != tc.msg) {
			t.Errorf("status %d: messages = %q, want %q", tc.status, apiErr.Messages, tc.msg)
		}
	}
}

func TestDoTreats200ErrorObjectAsError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error":["商品不存在"]}`))
	})
	var out map[string]any
	_, err := c.get(context.Background(), "v1/x", nil, &out)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 200 || apiErr.Messages[0] != "商品不存在" {
		t.Fatalf("got %v", err)
	}
}

func TestDoLeavesLegitimateMessageObjectsAlone(t *testing.T) {
	// uid_providers/{type} legitimately returns {customer_id, uid, message}.
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"customer_id":1,"uid":"U1","message":"ok"}`))
	})
	var out struct {
		CustomerID int64 `json:"customer_id"`
	}
	if _, err := c.get(context.Background(), "v1/x", nil, &out); err != nil {
		t.Fatal(err)
	}
	if out.CustomerID != 1 {
		t.Error("not decoded")
	}
}

func TestGetOneMapsNullBodyToNotFound(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`null`))
	})
	var out struct{}
	_, err := c.getOne(context.Background(), "v2/customers/by_uid_provider", nil, &out)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestDoRetriesOn429ThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"id":1}`))
	})
	var out struct {
		ID int `json:"id"`
	}
	if _, err := c.get(context.Background(), "v1/x", nil, &out); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 3 || out.ID != 1 {
		t.Errorf("calls = %d, out = %+v", calls.Load(), out)
	}
}

func TestDoGivesUpAfterMaxRetries(t *testing.T) {
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}, WithMaxRetries(2))
	_, err := c.get(context.Background(), "v1/x", nil, nil)
	if !errors.Is(err, ErrServer) {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3", calls.Load())
	}
}

func TestDoDoesNotRetryPostOnTransportError(t *testing.T) {
	var calls atomic.Int32
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
	})
	srv.Close()
	_, err := c.post(context.Background(), "v1/x", map[string]int{"a": 1}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 0 {
		t.Errorf("calls = %d", calls.Load())
	}
}

func TestDoDoesNotRepeatAWriteAfterAGatewayError(t *testing.T) {
	cases := []struct {
		method string
		status int
	}{
		{http.MethodPost, http.StatusBadGateway},
		{http.MethodPost, http.StatusServiceUnavailable},
		{http.MethodPost, http.StatusGatewayTimeout},
		{http.MethodPatch, http.StatusBadGateway},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s %d", tc.method, tc.status), func(t *testing.T) {
			var calls, sleeps atomic.Int32
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tc.status)
			})
			c.sleep = func(time.Duration) { sleeps.Add(1) }
			req := &Request{Method: tc.method, Path: "v1/orders", Body: map[string]int{"a": 1}}
			_, err := c.Do(context.Background(), req, nil)
			var apiErr *APIError
			if !errors.Is(err, ErrServer) || !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status {
				t.Fatalf("got %v, want APIError %d", err, tc.status)
			}
			if calls.Load() != 1 || sleeps.Load() != 0 {
				t.Errorf("calls = %d, sleeps = %d, want 1 and 0", calls.Load(), sleeps.Load())
			}
		})
	}
}

func TestDoRepeatsAWriteRejectedBy429(t *testing.T) {
	var calls atomic.Int32
	var waits []time.Duration
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"id":7}`))
	})
	c.sleep = func(d time.Duration) { waits = append(waits, d) }
	var out struct {
		ID int `json:"id"`
	}
	if _, err := c.post(context.Background(), "v1/orders", map[string]int{"a": 1}, &out); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || out.ID != 7 {
		t.Errorf("calls = %d, out = %+v", calls.Load(), out)
	}
	if len(waits) != 1 || waits[0] != 2*time.Second {
		t.Errorf("waits = %v, want [2s] from Retry-After", waits)
	}
}

func TestDoRetryHonoursContextCancel(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}, WithBackoff(func(int) time.Duration { return time.Hour }))
	c.sleep = time.Sleep
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := c.get(ctx, "v1/x", nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
}

func TestDoRejectsInvalidUTF8OnlyWhenStrict(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{\"name\":\"bad\xff\"}"))
	}
	var out struct {
		Name string `json:"name"`
	}
	lenient, _ := newTestClient(t, handler)
	if _, err := lenient.get(context.Background(), "v1/x", nil, &out); err != nil {
		t.Errorf("lenient: %v", err)
	}
	strict, _ := newTestClient(t, handler, WithStrictJSON())
	if _, err := strict.get(context.Background(), "v1/x", nil, &out); err == nil {
		t.Error("strict: expected error")
	}
}

func TestListAllWalksEveryPage(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("X-Next-Page", "2")
			_, _ = w.Write([]byte(`[{"id":1},{"id":2}]`))
		case "2":
			w.Header().Set("X-Next-Page", "3")
			_, _ = w.Write([]byte(`[{"id":3}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
		}
	})
	type item struct {
		ID int `json:"id"`
	}
	var ids []int
	for it, err := range listAll[item](context.Background(), c, "v1/x", nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, it.ID)
	}
	if len(ids) != 3 || ids[2] != 3 {
		t.Errorf("ids = %v", ids)
	}
}

func TestListAllStopsEarlyWhenCallerBreaks(t *testing.T) {
	var calls atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("X-Next-Page", "99")
		_, _ = w.Write([]byte(`[{"id":1}]`))
	})
	type item struct {
		ID int `json:"id"`
	}
	for _, err := range listAll[item](context.Background(), c, "v1/x", nil) {
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d", calls.Load())
	}
}

func TestRateLimiterSpacesRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	c, err := New("tok", WithBaseURL(srv.URL), WithRateLimit(20))
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	for i := 0; i < 4; i++ {
		if _, err := c.get(context.Background(), "v1/x", nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	// 4 requests at 20/s with burst 1 need at least 3 intervals of 50ms.
	if elapsed := time.Since(start); elapsed < 140*time.Millisecond {
		t.Errorf("requests were not rate limited: %v", elapsed)
	}
}

func TestParseRetryAfter(t *testing.T) {
	if d, ok := parseRetryAfter("2"); !ok || d != 2*time.Second {
		t.Errorf("seconds: %v %v", d, ok)
	}
	future := time.Now().Add(3 * time.Second).UTC().Format(http.TimeFormat)
	if d, ok := parseRetryAfter(future); !ok || d < 2*time.Second || d > 3*time.Second {
		t.Errorf("date: %v %v", d, ok)
	}
	if _, ok := parseRetryAfter("soon"); ok {
		t.Error("garbage accepted")
	}
}
