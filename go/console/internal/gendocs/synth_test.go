package gendocs

import (
	"encoding/json"
	"testing"
)

func TestSynthesizerReplacesPersonalData(t *testing.T) {
	src := `{
	  "id": 56943817, "order_number": 1101, "order_name": "#1101",
	  "customer": {"name": "REDACTED", "email": "someone@gmail.com", "mobile": "0933123456", "id": 42058619,
	    "address": {"address": "REDACTED", "detail_address": {"zip": "REDACTED", "city": "REDACTED"}}},
	  "line_items": [{"id": 1, "product_id": 56943817, "title": "真實商品", "sku": "GP.LUG11.02K", "price": 9999.0}],
	  "product_url": "//acme.cyberbiz.co/products/ware0014",
	  "photo": "/media/W1siZiIsIjI2NzIxL3Byb2R1Y3RzIl1d.jpeg?sha=4f630ab34394042e",
	  "account_activation_url": "http://acme.cyberbiz.co/account/customer/activate?confirmation_token=xxxxxxxxxxxxxxxxxxxx",
	  "vendor_type": "acme-webhook-test", "empty": "", "note": null, "flag": true
	}`
	v, err := DecodeJSON([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	out := newSynthesizer().Value("", v, nil).(*OMap)
	get := func(path ...string) any {
		var cur any = out
		for _, p := range path {
			cur, _ = cur.(*OMap).Get(p)
		}
		return cur
	}
	if get("id") != json.Number("1") || get("customer", "id") != json.Number("2") {
		t.Errorf("ids must be renumbered from 1: %v %v", get("id"), get("customer", "id"))
	}
	if get("line_items").([]any)[0].(*OMap).Keys()[0] != "id" {
		t.Error("key order must be preserved")
	}
	if pid, _ := get("line_items").([]any)[0].(*OMap).Get("product_id"); pid != json.Number("1") {
		t.Errorf("the same source id must map to the same synthetic id, got %v", pid)
	}
	if get("order_number") != json.Number("1001") || get("order_name") != "#1001" {
		t.Errorf("order number/name: %v %v", get("order_number"), get("order_name"))
	}
	if get("customer", "name") != synthName || get("customer", "email") != synthEmail || get("customer", "mobile") != synthMobile {
		t.Errorf("customer: %v", get("customer"))
	}
	if get("customer", "address", "detail_address", "zip") != "110" {
		t.Error("zip not synthesized")
	}
	li := get("line_items").([]any)[0].(*OMap)
	if title, _ := li.Get("title"); title != "範例商品" {
		t.Errorf("title = %v", title)
	}
	if sku, _ := li.Get("sku"); sku != "SKU-001" {
		t.Errorf("sku = %v", sku)
	}
	if price, _ := li.Get("price"); price != json.Number("9999.0") {
		t.Errorf("money must be kept: %v", price)
	}
	for _, k := range []string{"product_url", "photo", "account_activation_url", "vendor_type"} {
		if err := scanOutput(k, []byte(get(k).(string))); err != nil {
			t.Error(err)
		}
	}
	if get("empty") != "" || get("note") != nil || get("flag") != true {
		t.Error("empty strings, nulls and booleans must keep their shape")
	}
}

func TestSynthesizerPostmanPlaceholders(t *testing.T) {
	v, _ := DecodeJSON([]byte(`{"name": "string", "confirmed_at": "1991-09-20T01:46:33.257Z", "points": 8502.041422141616, "birthday": "1951-06-30"}`))
	out := newSynthesizer().Value("", v, nil).(*OMap)
	if n, _ := out.Get("name"); n != synthName {
		t.Errorf("name placeholder: %v", n)
	}
	if at, _ := out.Get("confirmed_at"); at != synthTimestamp {
		t.Errorf("ISO timestamp placeholder: %v", at)
	}
	if p, _ := out.Get("points"); p != json.Number("8502.04") {
		t.Errorf("random float must be rounded: %v", p)
	}
}
