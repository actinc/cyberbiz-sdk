package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var testNow = time.Date(2026, 10, 8, 15, 21, 4, 0, time.UTC)

func TestParseKinds(t *testing.T) {
	tests := []struct {
		raw     string
		want    []kind
		wantErr bool
	}{
		{"all", allKinds, false},
		{"", allKinds, false},
		{"products, customers", []kind{kindProducts, kindCustomers}, false},
		{"orders", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := parseKinds(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d kinds, want %d", len(got), len(tt.want))
			}
			for _, k := range tt.want {
				if !got.has(k) {
					t.Errorf("missing kind %s", k)
				}
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}
	full := map[string]string{"CYBERBIZ_API_TOKEN": "t", "CYBERBIZ_SHOP": "https://Demo.cyberbiz.co/"}
	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		wantErr string
	}{
		{"ok", nil, full, ""},
		{"dry run needs no env", []string{"-dry-run"}, nil, ""},
		{"missing env", nil, nil, "must be set"},
		{"n too big", []string{"-n", "51"}, full, "-n must be"},
		{"n zero", []string{"-n", "0"}, full, "-n must be"},
		{"bad kind", []string{"-only", "orders"}, full, "unknown kind"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := parseOptions(tt.args, env(tt.env))
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				if tt.env != nil && o.shop != "demo.cyberbiz.co" {
					t.Errorf("shop = %q, want normalized domain", o.shop)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeDomain(t *testing.T) {
	tests := map[string]string{
		"demo.cyberbiz.co":           "demo.cyberbiz.co",
		" HTTPS://Demo.cyberbiz.co/": "demo.cyberbiz.co",
		"http://demo.cyberbiz.co/x":  "demo.cyberbiz.co",
		"":                           "",
	}
	for in, want := range tests {
		if got := normalizeDomain(in); got != want {
			t.Errorf("normalizeDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildPlanIsMarkedAndUnique(t *testing.T) {
	kinds, _ := parseKinds("all")
	p := buildPlan(newRunID(testNow), 5, kinds, testNow)
	if p.RunID != "s20261008232104" {
		t.Errorf("run id = %q, want Taipei time", p.RunID)
	}
	seen := map[string]bool{}
	unique := func(what, v string) {
		t.Helper()
		if seen[v] {
			t.Errorf("duplicate %s %q", what, v)
		}
		seen[v] = true
	}
	titles := []string{}
	for _, pp := range p.Products {
		unique("handle", pp.Product.Handle)
		titles = append(titles, pp.Product.Title)
		for _, v := range pp.Variants {
			unique("sku", *v.SKU)
		}
	}
	for _, c := range p.Customers {
		unique("email", c.Email)
		titles = append(titles, c.Name)
		if !strings.HasSuffix(c.Email, "@example.com") {
			t.Errorf("customer email %q is not on the reserved example.com domain", c.Email)
		}
	}
	for _, c := range p.Coupons {
		unique("coupon code", c.Code)
		titles = append(titles, c.Title)
	}
	for _, c := range p.Collections {
		unique("handle", c.Handle)
		titles = append(titles, c.Title)
	}
	for _, d := range p.Discounts {
		titles = append(titles, d.Name)
	}
	for _, pg := range p.Pages {
		titles = append(titles, pg.Title)
	}
	for _, title := range titles {
		if !strings.HasPrefix(title, titlePrefix) {
			t.Errorf("title %q lacks %s prefix", title, titlePrefix)
		}
	}
	if len(p.Products) != 5 || len(p.Blogs) != 1 || len(p.Blogs[0].Articles) != 5 {
		t.Errorf("unexpected plan sizes: %d products, %d blogs", len(p.Products), len(p.Blogs))
	}
}

// fakeShop serves GET /shop and answers every POST with a fresh id. It
// records the path of each write.
type fakeShop struct {
	domain   string
	failPath string // writes to this path answer 422
	mu       sync.Mutex
	nextID   int64
	writes   []string
	products []map[string]any // decoded POST /v1/products bodies
}

func (f *fakeShop) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet && r.URL.Path == "/shop" {
		fmt.Fprintf(w, `{"shop_info":{"id":1,"primary_domain":%q}}`, f.domain)
		return
	}
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.nextID++
	id := f.nextID
	f.writes = append(f.writes, r.Method+" "+r.URL.Path)
	if r.URL.Path == "/v1/products" {
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		f.products = append(f.products, m)
	}
	f.mu.Unlock()
	if r.URL.Path == f.failPath {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"error":"款式的 SKU 為必填欄位"}`)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":%d}`, id)
}

func (f *fakeShop) count(prefix string) int {
	n := 0
	for _, w := range f.writes {
		if strings.HasPrefix(w, prefix) {
			n++
		}
	}
	return n
}

func runAgainst(t *testing.T, f *fakeShop, shop string) (manifest, error) {
	t.Helper()
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	kinds, _ := parseKinds("all")
	path := filepath.Join(t.TempDir(), "m.json")
	o := options{n: 2, kinds: kinds, manifest: path, token: "t", shop: shop, baseURL: srv.URL}
	err := run(context.Background(), o, testNow, io.Discard)
	var m manifest
	if data, readErr := os.ReadFile(path); readErr == nil {
		if jsonErr := json.Unmarshal(data, &m); jsonErr != nil {
			t.Fatal(jsonErr)
		}
	}
	return m, err
}

func TestRunRefusesAnotherShop(t *testing.T) {
	f := &fakeShop{domain: "production.cyberbiz.co"}
	_, err := runAgainst(t, f, "test.cyberbiz.co")
	if err == nil || !strings.Contains(err.Error(), "refusing to write") {
		t.Fatalf("err = %v, want refusal", err)
	}
	if len(f.writes) != 0 {
		t.Fatalf("made %d writes to the wrong shop", len(f.writes))
	}
}

func TestRunCreatesEveryKind(t *testing.T) {
	f := &fakeShop{domain: "test.cyberbiz.co"}
	m, err := runAgainst(t, f, "test.cyberbiz.co")
	if err != nil {
		t.Fatal(err)
	}
	wantWrites := map[string]int{
		"POST /v1/products":           2 + 2*3, // products plus variants
		"POST /v1/customers":          2,
		"POST /v1/discounts":          2,
		"POST /v1/shop_coupons":       2,
		"POST /v1/custom_collections": 2 + 2, // collections plus product links
		"POST /v1/blogs":              1 + 2, // blog plus articles
		"POST /v2/pages":              2,
	}
	for prefix, want := range wantWrites {
		if got := f.count(prefix); got != want {
			t.Errorf("%s: %d writes, want %d (all writes: %v)", prefix, got, want, f.writes)
		}
	}
	if len(m.Products) != 2 || len(m.Variants) != 6 || len(m.Customers) != 2 ||
		len(m.Coupons) != 2 || len(m.Collections) != 2 || len(m.Articles) != 2 || len(m.Pages) != 2 {
		t.Errorf("manifest is incomplete: %+v", m)
	}
}

func TestRunSendsProductSKU(t *testing.T) {
	f := &fakeShop{domain: "test.cyberbiz.co"}
	if _, err := runAgainst(t, f, "test.cyberbiz.co"); err != nil {
		t.Fatal(err)
	}
	if len(f.products) != 2 {
		t.Fatalf("got %d product bodies, want 2", len(f.products))
	}
	for _, body := range f.products {
		if sku, _ := body["sku"].(string); !strings.HasPrefix(sku, "TEST-") {
			t.Errorf("product body sku = %v, want TEST-...; body: %v", body["sku"], body)
		}
		if title, _ := body["title"].(string); !strings.HasPrefix(title, titlePrefix) {
			t.Errorf("product body title = %v; embedded request was not inlined", body["title"])
		}
	}
}

func TestRunContinuesAfterAFailedKind(t *testing.T) {
	f := &fakeShop{domain: "test.cyberbiz.co", failPath: "/v1/products"}
	m, err := runAgainst(t, f, "test.cyberbiz.co")
	if err == nil || !strings.Contains(err.Error(), "create product") {
		t.Fatalf("err = %v, want the product failure", err)
	}
	if len(m.Products) != 0 || len(m.Customers) != 2 || len(m.Pages) != 2 {
		t.Errorf("other kinds should still be created: %+v", m)
	}
}
