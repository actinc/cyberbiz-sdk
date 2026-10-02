package gendocs

import "testing"

func TestGoldenPathPattern(t *testing.T) {
	cases := []struct {
		path, name string
		match      bool
	}{
		{"/v1/orders", "v1_orders", true},
		{"/v1/orders/{order_id}", "v1_orders_{id}", true},
		{"/v1/customers/{customer_id}/uid_providers/{provider_type}", "v1_customers_{id}_uid_providers_line", true},
		{"/v1/custom_fields/{custom_field_name}", "v1_custom_fields_foo", true},
		{"/v1/products/sku/{product_sku}/product_variants", "v1_products_sku_{id}_product_variants", true},
		{"/v1/orders", "v1_orders_{id}", false},
		{"/v1/orders/{order_id}", "v1_orders_{id}_fulfillments", false},
		{"/v1/custom_fields", "v1_custom_fields_foo", false},
	}
	for _, c := range cases {
		if got := goldenPathPattern(c.path).MatchString(c.name); got != c.match {
			t.Errorf("%s vs %s: got %v want %v", c.path, c.name, got, c.match)
		}
	}
}

func TestGoldenFindPrefersExact(t *testing.T) {
	set := &goldenSet{files: []*goldenFile{
		{Group: "v1", Method: "GET", Name: "v1_customers_{id}_uid_providers_line"},
		{Group: "v1", Method: "GET", Name: "v1_customers_{id}_uid_providers_{id}"},
		{Group: "errors", Method: "GET", Name: "v1_customers_{id}_uid_providers_facebook", Status: 422},
	}}
	hits := set.find("GET", "/v1/customers/{customer_id}/uid_providers/{provider_type}", "v1")
	if len(hits) != 2 || hits[0].Name != "v1_customers_{id}_uid_providers_{id}" {
		t.Fatalf("exact spelling must come first: %+v", hits)
	}
	if errs := set.errorsFor("GET", "/v1/customers/{customer_id}/uid_providers/{provider_type}"); errs[422] == nil {
		t.Error("error golden not found")
	}
}

func TestInferErrorStatus(t *testing.T) {
	cases := map[string]int{
		`{"error": ["無權使用該 API"]}`:                      401,
		`{"error": ["無此資源"]}`:                           404,
		`{"error": ["custom_field_type 無效值"]}`:          422,
		`{"error": "Must have pos_shop_coupon plugin"}`: 403,
		`{"messages": "Invoice is not ready."}`:         403,
		`{"message": "無效的 Provider Type"}`:              422,
		`[{"id": 1}]`:                                   0,
	}
	for body, want := range cases {
		v, err := DecodeJSON([]byte(body))
		if err != nil {
			t.Fatal(err)
		}
		if got := inferErrorStatus(v); got != want {
			t.Errorf("%s: got %d want %d", body, got, want)
		}
	}
}
