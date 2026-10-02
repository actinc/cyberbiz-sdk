//go:build integration && write

package cyberbiz_test

import (
	"encoding/json/v2"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

// TestLiveWriteFormat settles whether the platform accepts JSON bodies on v1
// writes (the Postman collection says yes; the swagger documents form data).
// It performs a no-op update: it reads a product and PUTs its current title
// back, once as JSON and once as a form, and reports which the API accepts.
// It mutates nothing observable, but it is a write, so it only runs with
// both the `integration` and `write` build tags:
//
//	go test -tags 'integration write' -run TestLiveWriteFormat -v ./cyberbiz
func TestLiveWriteFormat(t *testing.T) {
	c := liveClient(t)
	ctx := liveContext(t)

	var products []struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	q := url.Values{"per_page": {"1"}}
	if _, err := c.Do(ctx, &cyberbiz.Request{Method: http.MethodGet, Path: "v1/products", Query: q}, &products); err != nil {
		t.Fatal(err)
	}
	if len(products) == 0 {
		t.Skip("shop has no products")
	}
	p := products[0]
	path := "v1/products/" + itoa(p.ID)

	jsonBody, _ := json.Marshal(map[string]string{"title": p.Title})
	resp, err := c.Do(ctx, &cyberbiz.Request{Method: http.MethodPut, Path: path, Body: jsonBody}, nil)
	report(t, "JSON body", resp, err)

	form := url.Values{"title": {p.Title}}
	resp, err = c.Do(ctx, &cyberbiz.Request{
		Method: http.MethodPut, Path: path, Body: []byte(form.Encode()),
		Header: http.Header{"Content-Type": {"application/x-www-form-urlencoded"}},
	}, nil)
	report(t, "form body", resp, err)
}

func report(t *testing.T, label string, resp *cyberbiz.Response, err error) {
	t.Helper()
	switch {
	case err == nil:
		t.Logf("%s: accepted (HTTP %d)", label, resp.StatusCode)
	case resp != nil:
		t.Logf("%s: rejected: %v", label, err)
	default:
		t.Logf("%s: transport error: %v", label, err)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
