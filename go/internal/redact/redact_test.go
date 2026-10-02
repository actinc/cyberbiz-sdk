package redact

import (
	"encoding/json/v2"
	"net/http"
	"strings"
	"testing"
)

func TestKey(t *testing.T) {
	yes := []string{"name", "Email", "mobile", "billing_address", "address1", "paper_invoice_no",
		"checkout_referral_code", "carrier_number", "customer_email", "access_token"}
	no := []string{"id", "title", "sku", "price", "order_number", "created_at", "vendor", "zip"}
	for _, k := range yes {
		if !Key(k) {
			t.Errorf("%q should be sensitive", k)
		}
	}
	for _, k := range no {
		if Key(k) {
			t.Errorf("%q should not be sensitive", k)
		}
	}
}

func TestValue(t *testing.T) {
	yes := []string{"someone@example.org", "0912345678", "+886912345678", "0912-345-678",
		"eyJhbGciOiJIUzI1NiJ9.eyJzaG9wX2lkIjoxfQ.abc_def-123"}
	no := []string{"hello", "2026-07-10 20:22:05", "A12345", "12345678", "https://x.y/z"}
	for _, v := range yes {
		if !Value(v) {
			t.Errorf("%q should be sensitive", v)
		}
	}
	for _, v := range no {
		if Value(v) {
			t.Errorf("%q should not be sensitive", v)
		}
	}
}

func TestJSONPreservesShapeAndOrder(t *testing.T) {
	in := `{"id":8,"name":"王小明","email":"a@b.co","prices":{"total_price":9999.0},
	"line_items":[{"sku":"SKU-1","note":"call me","title":"Tee"}],
	"buyer":{"name":"X","phone":"0912345678","address":{"city":"Taipei","address1":"No. 1"}},
	"tags":["vip","a@b.co"],"birthday":"1990-05-01","created_at":"2026-07-10 20:22:05",
	"payment_url":"https://pay.example/abc","closed_at":null,"vip":true}`
	out, err := JSON([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	checks := map[string]any{
		"id":          8.0,
		"name":        Placeholder,
		"email":       "redacted@example.com",
		"birthday":    "1990-01-01",
		"created_at":  "2026-07-10 20:22:05",
		"payment_url": "https://example.com/redacted",
		"closed_at":   nil,
		"vip":         true,
	}
	for k, want := range checks {
		if got := doc[k]; got != want {
			t.Errorf("%s = %v, want %v", k, got, want)
		}
	}
	if doc["prices"].(map[string]any)["total_price"] != 9999.0 {
		t.Error("nested number changed")
	}
	item := doc["line_items"].([]any)[0].(map[string]any)
	if item["sku"] != Hashed("sku", "SKU-1") || item["title"] != "Tee" || item["note"] != Placeholder {
		t.Errorf("line item = %v", item)
	}
	buyer := doc["buyer"].(map[string]any)
	if buyer["phone"] != "0912345678" || buyer["name"] != Placeholder {
		t.Errorf("buyer = %v", buyer)
	}
	addr := buyer["address"].(map[string]any)
	if addr["city"] != Placeholder || addr["address1"] != Placeholder {
		t.Errorf("address = %v", addr)
	}
	tags := doc["tags"].([]any)
	if tags[0] != "vip" || tags[1] != "redacted@example.com" {
		t.Errorf("tags = %v", tags)
	}
	// Key order must survive.
	if !strings.HasPrefix(strings.TrimSpace(string(out)), "{\n  \"id\": 8,\n  \"name\"") {
		t.Errorf("order not preserved:\n%s", out)
	}
}

func TestJSONHandlesArraysAndNonJSON(t *testing.T) {
	out, err := JSON([]byte(`[{"email":"x@y.z"},{"email":null}]`))
	if err != nil || !strings.Contains(string(out), "redacted@example.com") {
		t.Errorf("array: %s %v", out, err)
	}
	raw := []byte("PK\x03\x04 binary zip")
	if out, err := JSON(raw); err != nil || string(out) != string(raw) {
		t.Errorf("non-json: %s %v", out, err)
	}
	if out, err := JSON([]byte(`null`)); err != nil || string(out) != "null" {
		t.Errorf("null: %s %v", out, err)
	}
}

func TestHeaders(t *testing.T) {
	h := http.Header{"Authorization": {"Bearer x"}, "X-Real-Ip": {"1.2.3.4"}, "Accept": {"application/json"}}
	out := Headers(h)
	if out.Get("Authorization") != Placeholder || out.Get("X-Real-Ip") != Placeholder {
		t.Errorf("not redacted: %v", out)
	}
	if out.Get("Accept") != "application/json" {
		t.Error("accept changed")
	}
	out.Set("Accept", "x")
	if h.Get("Accept") != "application/json" {
		t.Error("input mutated")
	}
}

func TestHosts(t *testing.T) {
	cases := map[string]string{
		"acmeco.cyberbiz.co":                                        ExampleShopHost,
		"//acmeco.cyberbiz.co/products/ware0014":                    "//" + ExampleShopHost + "/products/ware0014",
		"https://admin.uat.globexco.acmeco.io/api/webhook":          "https://" + ExampleHost + "/api/webhook",
		"www.acmestoreworld.select":                                 ExampleHost,
		"https://app-store-api.cyberbiz.io/v1/orders":               "https://app-store-api.cyberbiz.io/v1/orders",
		"//cdn-general.cybassets.com/theme_src/x.png":               "//cdn-general.cybassets.com/theme_src/x.png",
		"https://api-doc.cyberbiz.co/v1/api_document":               "https://api-doc.cyberbiz.co/v1/api_document",
		"https://example.com/redacted":                              "https://example.com/redacted",
		"plain text with 1.5 in it":                                 "plain text with 1.5 in it",
		"2026-07-10 20:22:05":                                       "2026-07-10 20:22:05",
		"http://localhost:8787/x":                                   "http://localhost:8787/x",
		"[{\"url\":\"https://shop.example.tw/a\"}] see cyberbiz.co": "[{\"url\":\"https://" + ExampleHost + "/a\"}] see cyberbiz.co",
	}
	for in, want := range cases {
		if got := Hosts(in); got != want {
			t.Errorf("Hosts(%q)\n got %q\nwant %q", in, got, want)
		}
	}
}

func TestHashedIsStableAndDistinct(t *testing.T) {
	a, b := Hashed("sku", "acmesku000001"), Hashed("sku", "acmesku000002")
	if a == b || a != Hashed("sku", "acmesku000001") || !strings.HasPrefix(a, "SKU-") || len(a) != 10 {
		t.Errorf("got %q %q", a, b)
	}
	if Hashed("sku", "") != "" {
		t.Error("empty must stay empty")
	}
}

func TestHeadersRewriteHosts(t *testing.T) {
	h := http.Header{"X-Cyberbiz-Domain": {"acmeco.cyberbiz.co"}, "X-Cyberbiz-Shop-Domain": {"www.acmestoreworld.select"}}
	out := Headers(h)
	if out.Get("X-Cyberbiz-Domain") != ExampleShopHost || out.Get("X-Cyberbiz-Shop-Domain") != ExampleHost {
		t.Errorf("got %v", out)
	}
}

func TestQueryTokensAreRedacted(t *testing.T) {
	in := []byte(`{"account_activation_url":"http://acmeco.cyberbiz.co/account/customer/activate?confirmation_token=aB3dE5fG7hJ9kL1mN2pQ&x=1"}`)
	out, err := JSON(in)
	if err != nil {
		t.Fatal(err)
	}
	want := `"http://` + ExampleShopHost + `/account/customer/activate?confirmation_token=` + Placeholder + `&x=1"`
	if !strings.Contains(string(out), want) {
		t.Errorf("got %s, want it to contain %s", out, want)
	}
}
