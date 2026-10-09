package cyberbiz

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// These tests pin request shapes the platform source requires (CBSDK-45).

// sequenceServer answers each request with the body registered for its
// "METHOD /path" and records the requests in order.
type sequenceServer struct {
	mu       sync.Mutex
	requests []string // "METHOD /path body"
}

func newSequenceClient(t *testing.T, replies map[string]string) (*Client, *sequenceServer) {
	t.Helper()
	seq := &sequenceServer{}
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		key := r.Method + " " + r.URL.Path
		seq.mu.Lock()
		seq.requests = append(seq.requests, strings.TrimSpace(key+" "+string(body)))
		seq.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		reply, ok := replies[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, reply)
	})
	return c, seq
}

func (s *sequenceServer) assert(t *testing.T, want ...string) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.Join(s.requests, "\n") != strings.Join(want, "\n") {
		t.Errorf("requests:\n%s\nwant:\n%s", strings.Join(s.requests, "\n"), strings.Join(want, "\n"))
	}
}

func TestProductCreateSendsSKU(t *testing.T) {
	c, seq := newSequenceClient(t, map[string]string{
		"POST /v1/products":                `{"id":1}`,
		"POST /v1/products/pos_shop_batch": `[{"id":2}]`,
	})
	sku := "SKU-1"
	if _, _, err := c.Products.Create(testCtx, &ProductCreateRequest{Title: "T", Handle: "t", Price: MoneyFromInt(100), SKU: &sku}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.Products.CreateForPosShops(testCtx, &ProductPosShopBatchCreateRequest{Title: "T", PosShopIDs: "3", SKU: &sku}); err != nil {
		t.Fatal(err)
	}
	seq.assert(t,
		`POST /v1/products {"title":"T","handle":"t","published":false,"price":100,"sku":"SKU-1"}`,
		`POST /v1/products/pos_shop_batch {"title":"T","pos_shop_ids":"3","sku":"SKU-1"}`)
}

func TestProductUpdateKeepsTaxAndTemperature(t *testing.T) {
	c, seq := newSequenceClient(t, map[string]string{
		"GET /v1/products/7": `{"title":"old","tax_type_id":"zero_tax","temperature_types":["冷凍"]}`,
		"PUT /v1/products/7": `{"title":"new"}`,
	})
	title := "new"
	req := &ProductUpdateRequest{Title: &title}
	if _, _, err := c.Products.Update(testCtx, 7, req); err != nil {
		t.Fatal(err)
	}
	seq.assert(t,
		`GET /v1/products/7`,
		`PUT /v1/products/7 {"title":"new","tax_type_id":"zero_tax","temperature_types":["冷凍"]}`)
	if req.TaxTypeID != "" || req.TemperatureTypes != nil {
		t.Errorf("caller's request was modified: %+v", req)
	}
}

func TestProductUpdateSkipsReadWhenBothSet(t *testing.T) {
	c, seq := newSequenceClient(t, map[string]string{"PUT /v1/products/7": `{}`})
	req := &ProductUpdateRequest{TaxTypeID: TaxTypeExclusive, TemperatureTypes: []TemperatureType{TemperatureTypeRefrigerated}}
	if _, _, err := c.Products.Update(testCtx, 7, req); err != nil {
		t.Fatal(err)
	}
	seq.assert(t, `PUT /v1/products/7 {"tax_type_id":"exclusive_tax","temperature_types":["冷藏"]}`)
}

func TestProductUpdateFailsWhenReadFails(t *testing.T) {
	c, seq := newSequenceClient(t, map[string]string{})
	title := "new"
	if _, _, err := c.Products.Update(testCtx, 7, &ProductUpdateRequest{Title: &title}); err == nil {
		t.Fatal("update went ahead without the current tax type")
	}
	seq.assert(t, `GET /v1/products/7`)
}

func TestVariantDiscountUpdateResendsType(t *testing.T) {
	c, seq := newSequenceClient(t, map[string]string{
		"GET /v1/variant_discount_collections/2": `{"title":"V","variant_discount_collection_type_code":"percentage"}`,
		"PUT /v1/variant_discount_collections/2": `{"title":"V2"}`,
	})
	title := "V2"
	if _, _, err := c.Collections.UpdateVariantDiscount(testCtx, 2, &VariantDiscountCollectionUpdateRequest{Title: &title}); err != nil {
		t.Fatal(err)
	}
	seq.assert(t,
		`GET /v1/variant_discount_collections/2`,
		`PUT /v1/variant_discount_collections/2 {"title":"V2","variant_discount_collection_type":"percentage"}`)
}

func TestSpecialRuleCreateSendsAllRequiredFields(t *testing.T) {
	c, seq := newSequenceClient(t, map[string]string{"POST /v1/special_collections/5/rules": `{}`})
	qty, pct := 2, 80.0
	if _, _, err := c.Collections.CreateSpecialRule(testCtx, 5, &SpecialCollectionRuleRequest{Quantity: &qty, Percentage: &pct}); err != nil {
		t.Fatal(err)
	}
	seq.assert(t, `POST /v1/special_collections/5/rules {"quantity":2,"price":0,"percentage":80}`)
}
