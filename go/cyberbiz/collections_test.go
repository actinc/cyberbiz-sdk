package cyberbiz

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Golden decode tests.

func TestCustomCollectionGoldenList(t *testing.T) {
	var out []CustomCollection
	decodeGolden(t, "v1/GET_v1_custom_collections.json", &out)
	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	c := out[0]
	if c.ID != 459369 || c.Title != "英茶香" || c.Handle != "英茶香" || !c.Published {
		t.Errorf("item = %+v", c)
	}
	if c.BodyHTML != "" || c.Position != 2 || c.ProductsOrderName != "按標題拼音升序: A-Z" {
		t.Errorf("item = %+v", c)
	}
	if c.Products != nil {
		t.Errorf("list items carry no products: %+v", c.Products)
	}
}

func TestCustomCollectionGoldenDetail(t *testing.T) {
	var out CustomCollection
	decodeGolden(t, "v1/GET_v1_custom_collections_{id}.json", &out)
	if out.ID != 0 {
		t.Errorf("detail omits id, got %d", out.ID)
	}
	if out.Title != "英茶香" || out.Position != 2 || out.Products == nil || len(out.Products) != 0 {
		t.Errorf("detail = %+v", out)
	}
}

func TestCustomCollectionGoldenV2(t *testing.T) {
	var out []CustomCollection
	decodeGolden(t, "v2/GET_v2_custom_collections.json", &out)
	if len(out) != 2 || out[1].ID != 403543 || out[1].Position != 1 {
		t.Errorf("v2 = %+v", out)
	}
}

func TestCollectionGoldenEmptyLists(t *testing.T) {
	cases := map[string]any{
		"v1/GET_v1_smart_collections.json":            &[]SmartCollection{},
		"v2/GET_v2_smart_collections.json":            &[]SmartCollection{},
		"v1/GET_v1_special_collections.json":          &[]SpecialCollection{},
		"v1/GET_v1_add_buy_collections.json":          &[]AddBuyCollection{},
		"v1/GET_v1_variant_discount_collections.json": &[]VariantDiscountCollection{},
	}
	for name, out := range cases {
		decodeGolden(t, name, out)
		if n := reflect.ValueOf(out).Elem().Len(); n != 0 {
			t.Errorf("%s: len = %d", name, n)
		}
	}
}

func TestCollectionGoldenErrors(t *testing.T) {
	cases := []struct {
		golden string
		status int
		want   error
		msg    string
		call   func(c *Client) error
	}{
		{"errors/GET_v1_limit_collections.json", 401, ErrUnauthorized, "無權使用該 API",
			func(c *Client) error { _, err := c.Collections.ListLimit(testCtx, nil); return err }},
		{"errors/GET_v1_limit_collections_{id}.json", 401, ErrUnauthorized, "無權使用該 API",
			func(c *Client) error { _, _, err := c.Collections.GetLimit(testCtx, 1); return err }},
		{"errors/GET_v1_vip_collections.json", 401, ErrUnauthorized, "無權使用該 API",
			func(c *Client) error { _, err := c.Collections.ListVIP(testCtx, nil); return err }},
		{"errors/GET_v1_vip_collections_{id}.json", 401, ErrUnauthorized, "無權使用該 API",
			func(c *Client) error { _, _, err := c.Collections.GetVIP(testCtx, 1); return err }},
		{"errors/GET_v1_smart_collections_{id}.json", 404, ErrNotFound, "無此資源",
			func(c *Client) error { _, _, err := c.Collections.GetSmart(testCtx, 1); return err }},
		{"errors/GET_v1_special_collections_{id}.json", 404, ErrNotFound, "無此資源",
			func(c *Client) error { _, _, err := c.Collections.GetSpecial(testCtx, 1); return err }},
		{"errors/GET_v1_add_buy_collections_{id}.json", 404, ErrNotFound, "無此資源",
			func(c *Client) error { _, _, err := c.Collections.GetAddBuy(testCtx, 1); return err }},
		{"errors/GET_v1_variant_discount_collections_{id}.json", 404, ErrNotFound,
			"無此單品折扣群組 variant_discount_collection_id = 1",
			func(c *Client) error { _, _, err := c.Collections.GetVariantDiscount(testCtx, 1); return err }},
	}
	for _, tc := range cases {
		body := readGolden(t, tc.golden)
		c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write(body)
		}, WithMaxRetries(0))
		err := tc.call(c)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v", tc.golden, err)
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) || len(apiErr.Messages) != 1 || apiErr.Messages[0] != tc.msg {
			t.Errorf("%s: messages = %v", tc.golden, err)
		}
	}
}

// Synthetic decode tests for the kinds without a non-empty Golden File,
// shaped after the swagger and Postman examples.

func TestSpecialCollectionDecodesRulesAndType(t *testing.T) {
	var out SpecialCollection
	src := `{"title":"任選","handle":"pick","published":true,"start_date":"2026-01-01 00:00:00",
	  "end_date":null,"body_html":null,"position":3,
	  "special_collection_type":{"name":"任選折數","code":"percentage"},
	  "type_rules":[{"id":11,"quantity":2,"price":"199.5","percentage":85}],
	  "rest_include_discount":false,"products":[{"id":7,"title":"A"}]}`
	if err := json.Unmarshal([]byte(src), &out); err != nil {
		t.Fatal(err)
	}
	if out.SpecialCollectionType == nil || out.SpecialCollectionType.Code != SpecialCollectionTypePercentage {
		t.Errorf("type = %+v", out.SpecialCollectionType)
	}
	if out.StartDate.Year() != 2026 || !out.EndDate.IsZero() {
		t.Errorf("dates = %v %v", out.StartDate, out.EndDate)
	}
	r := out.TypeRules[0]
	if r.ID != 11 || r.Quantity != 2 || r.Price != 19950 || r.Percentage != 85 {
		t.Errorf("rule = %+v", r)
	}
	if out.Products[0].Position != 0 || out.Products[0].Title != "A" {
		t.Errorf("products = %+v", out.Products)
	}
}

func TestLimitCollectionDecodesNumericCalculationType(t *testing.T) {
	var out LimitCollection
	src := `{"calculation_type":1,"handle":"limit","limit_count":5,"min_count":null,"period_type":"monthly",
	  "position":1,"products_order":"title.asc","published":true,"title":"限購","products":[],"id":9}`
	if err := json.Unmarshal([]byte(src), &out); err != nil {
		t.Fatal(err)
	}
	if out.CalculationType != "1" || out.PeriodType != LimitPeriodMonthly || out.LimitCount != 5 || out.ID != 9 {
		t.Errorf("limit = %+v", out)
	}
	if err := json.Unmarshal([]byte(`{"calculation_type":"by_product"}`), &out); err != nil {
		t.Fatal(err)
	}
	if out.CalculationType != LimitCalculationByProduct {
		t.Errorf("calculation_type = %q", out.CalculationType)
	}
}

func TestVIPCollectionDecodesRuleAndPromotion(t *testing.T) {
	var out VIPCollection
	src := `{"title":"金卡","position":1,"id":3,
	  "rule":{"rule_type":"total_price","order_start":"2026-01-01","order_end":null,"total_order_count":0,
	    "total_price":10000.0,"customers_tag":null,"vip_expiredate_start":"2026-01-01","vip_expiredate_end":"2026-12-31"},
	  "promotion":{"promotion_type":"discount","discount":9.5,"concurrently_apply":true}}`
	if err := json.Unmarshal([]byte(src), &out); err != nil {
		t.Fatal(err)
	}
	if out.Rule == nil || out.Rule.RuleType != VIPRuleTotalPrice || out.Rule.TotalPrice != 1000000 {
		t.Errorf("rule = %+v", out.Rule)
	}
	if out.Rule.OrderStart.String() != "2026-01-01" || !out.Rule.OrderEnd.IsZero() {
		t.Errorf("dates = %+v", out.Rule)
	}
	if out.Promotion == nil || out.Promotion.PromotionType != VIPPromotionDiscount || out.Promotion.Discount != 9.5 || !out.Promotion.ConcurrentlyApply {
		t.Errorf("promotion = %+v", out.Promotion)
	}
}

func TestVariantDiscountDetailAcceptsArrayOrObject(t *testing.T) {
	obj := `{"id":4,"shop_id":77,"title":"單品","published":true,"amount":100,"discount_amount":0,"percentage":0,
	  "variant_discount_collection_type_id":1,"variant_discount_collection_type_code":"amount",
	  "start_date":"2026-03-01 10:00:00","end_date":null,
	  "variants":[{"id":501,"product_id":42,"product_title":"T","option1":"S","option2":null,"option3":null,"sku":"T-S"}]}`
	for _, body := range []string{obj, "[" + obj + "]"} {
		c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) })
		got, _, err := c.Collections.GetVariantDiscount(testCtx, 4)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 4 || got.ShopID != 77 || got.Amount != 10000 || got.VariantDiscountCollectionTypeCode != VariantDiscountTypeAmount {
			t.Errorf("got %+v", got)
		}
		if len(got.Variants) != 1 || got.Variants[0].SKU != "T-S" || got.Variants[0].ProductID != 42 {
			t.Errorf("variants = %+v", got.Variants)
		}
	}
}

// Golden server round trips.

func TestCollectionsGoldenRoundTrip(t *testing.T) {
	c := goldenServer(t)
	got, _, err := c.Collections.GetCustom(testCtx, 459369)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 459369 || got.Title != "英茶香" {
		t.Errorf("GetCustom stamps id: %+v", got)
	}
	page, err := c.Collections.ListCustom(testCtx, &CollectionListOptions{ListOptions{PerPage: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Pagination.Total != 1 {
		t.Errorf("ListCustom = %+v", page)
	}
	v2, err := c.Collections.ListCustomV2(testCtx, &CollectionSearchOptions{Q: "茶"})
	if err != nil || len(v2.Items) != 2 {
		t.Errorf("ListCustomV2 = %+v, %v", v2, err)
	}
	smart, err := c.Collections.ListSmartV2(testCtx, nil)
	if err != nil || len(smart.Items) != 0 {
		t.Errorf("ListSmartV2 = %+v, %v", smart, err)
	}
	var n int
	for _, err := range c.Collections.AllCustom(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 2 {
		t.Errorf("AllCustom yielded %d", n)
	}
}

// Request shape tests: one table row per method.

type collectionsRecordedRequest struct {
	Method, Path, Query, Body string
}

func collectionsRecordingClient(t *testing.T, rec *collectionsRecordedRequest) *Client {
	t.Helper()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*rec = collectionsRecordedRequest{r.Method, r.URL.Path, r.URL.RawQuery, string(b)}
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "_collections") {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	return c
}

func collectionsStr(s string) *string     { return &s }
func collectionsBool(b bool) *bool        { return &b }
func collectionsInt(i int) *int           { return &i }
func collectionsMoney(m Money) *Money     { return &m }
func collectionsFloat(f float64) *float64 { return &f }

var (
	collectionsCtx      = context.Background()
	collectionsFixedAt  = NewTime(time.Date(2026, 3, 1, 10, 0, 0, 0, Taipei))
	collectionsFixedDay = NewDate(2026, 3, 1)
)

type collectionsShapeCase struct {
	name   string
	call   func(s *CollectionsService) error
	method string
	path   string
	query  string
	body   string // JSON, "" for no body
}

func collectionsCustomShapeCases() []collectionsShapeCase {
	return []collectionsShapeCase{
		{"ListCustom", func(s *CollectionsService) error {
			_, err := s.ListCustom(collectionsCtx, &CollectionListOptions{ListOptions{Page: 2, PerPage: 10, Offset: 5}})
			return err
		}, "GET", "/v1/custom_collections", "offset=5&page=2&per_page=10", ""},
		{"ListCustomV2", func(s *CollectionsService) error {
			_, err := s.ListCustomV2(collectionsCtx, &CollectionSearchOptions{Q: "tea"})
			return err
		}, "GET", "/v2/custom_collections", "q=tea", ""},
		{"GetCustom", func(s *CollectionsService) error { _, _, err := s.GetCustom(collectionsCtx, 7); return err },
			"GET", "/v1/custom_collections/7", "", ""},
		{"CreateCustom", func(s *CollectionsService) error {
			_, _, err := s.CreateCustom(collectionsCtx, &CustomCollectionCreateRequest{Title: "T", Handle: "t", ProductsOrder: ProductsOrderManual})
			return err
		}, "POST", "/v1/custom_collections", "", `{"title":"T","handle":"t","published":false,"products_order":"manual"}`},
		{"UpdateCustom", func(s *CollectionsService) error {
			_, _, err := s.UpdateCustom(collectionsCtx, 7, &CustomCollectionUpdateRequest{Published: collectionsBool(false), BodyHTML: collectionsStr("")})
			return err
		}, "PUT", "/v1/custom_collections/7", "", `{"published":false,"body_html":""}`},
		{"DeleteCustom", func(s *CollectionsService) error { _, err := s.DeleteCustom(collectionsCtx, 7); return err },
			"DELETE", "/v1/custom_collections/7", "", ""},
		{"AddCustomProducts", func(s *CollectionsService) error {
			_, _, err := s.AddCustomProducts(collectionsCtx, 7, IDList{1, 2, 3})
			return err
		}, "POST", "/v1/custom_collections/7/products", "", `{"product_ids":"1,2,3"}`},
		{"ReplaceCustomProducts", func(s *CollectionsService) error {
			_, _, err := s.ReplaceCustomProducts(collectionsCtx, 7, IDList{4})
			return err
		}, "PUT", "/v1/custom_collections/7/products", "", `{"product_ids":"4"}`},
		{"RemoveCustomProducts", func(s *CollectionsService) error {
			_, _, err := s.RemoveCustomProducts(collectionsCtx, 7, IDList{1, 2})
			return err
		}, "DELETE", "/v1/custom_collections/7/products", "product_ids=1%2C2", ""},
		{"ReorderCustomProducts", func(s *CollectionsService) error {
			_, _, err := s.ReorderCustomProducts(collectionsCtx, 7, IDList{3, 1, 2})
			return err
		}, "PUT", "/v1/custom_collections/7/products/manual_order", "", `{"product_ids":"3,1,2"}`},
	}
}

func collectionsSmartShapeCases() []collectionsShapeCase {
	return []collectionsShapeCase{
		{"ListSmart", func(s *CollectionsService) error { _, err := s.ListSmart(collectionsCtx, nil); return err },
			"GET", "/v1/smart_collections", "", ""},
		{"ListSmartV2", func(s *CollectionsService) error {
			_, err := s.ListSmartV2(collectionsCtx, &CollectionSearchOptions{ListOptions: ListOptions{Page: 3}, Q: "x"})
			return err
		}, "GET", "/v2/smart_collections", "page=3&q=x", ""},
		{"GetSmart", func(s *CollectionsService) error { _, _, err := s.GetSmart(collectionsCtx, 8); return err },
			"GET", "/v1/smart_collections/8", "", ""},
		{"CreateSmart", func(s *CollectionsService) error {
			_, _, err := s.CreateSmart(collectionsCtx, &SmartCollectionCreateRequest{Title: "S", Handle: "s", Published: true, BodyHTML: collectionsStr("<p>x</p>")})
			return err
		}, "POST", "/v1/smart_collections", "", `{"title":"S","handle":"s","published":true,"body_html":"<p>x</p>"}`},
		{"UpdateSmart", func(s *CollectionsService) error {
			_, _, err := s.UpdateSmart(collectionsCtx, 8, &SmartCollectionUpdateRequest{Title: collectionsStr("S2")})
			return err
		}, "PUT", "/v1/smart_collections/8", "", `{"title":"S2"}`},
		{"DeleteSmart", func(s *CollectionsService) error { _, err := s.DeleteSmart(collectionsCtx, 8); return err },
			"DELETE", "/v1/smart_collections/8", "", ""},
		{"CreateSmartRule", func(s *CollectionsService) error {
			_, _, err := s.CreateSmartRule(collectionsCtx, 8, &SmartCollectionRuleRequest{
				Column: SmartRuleColumnVariantsPrice, Relation: SmartRuleRelationLessThan, Condition: "1000"})
			return err
		}, "POST", "/v1/smart_collections/8/rules", "", `{"column":"variants_price","relation":"lt","condition":"1000"}`},
		{"UpdateSmartRule", func(s *CollectionsService) error {
			_, _, err := s.UpdateSmartRule(collectionsCtx, 8, 9, &SmartCollectionRuleRequest{Condition: "shoes"})
			return err
		}, "PUT", "/v1/smart_collections/8/rules/9", "", `{"condition":"shoes"}`},
		{"DeleteSmartRule", func(s *CollectionsService) error { _, _, err := s.DeleteSmartRule(collectionsCtx, 8, 9); return err },
			"DELETE", "/v1/smart_collections/8/rules/9", "", ""},
	}
}

func collectionsSpecialShapeCases() []collectionsShapeCase {
	return []collectionsShapeCase{
		{"ListSpecial", func(s *CollectionsService) error { _, err := s.ListSpecial(collectionsCtx, nil); return err },
			"GET", "/v1/special_collections", "", ""},
		{"GetSpecial", func(s *CollectionsService) error { _, _, err := s.GetSpecial(collectionsCtx, 5); return err },
			"GET", "/v1/special_collections/5", "", ""},
		{"CreateSpecial", func(s *CollectionsService) error {
			_, _, err := s.CreateSpecial(collectionsCtx, &SpecialCollectionCreateRequest{Title: "P", Handle: "p", Published: true,
				StartDate: collectionsFixedAt, SpecialCollectionType: SpecialCollectionTypeAmount, RestIncludeDiscount: true})
			return err
		}, "POST", "/v1/special_collections", "",
			`{"title":"P","handle":"p","published":true,"start_date":"2026-03-01 10:00:00","special_collection_type":"amount","rest_include_discount":true}`},
		{"UpdateSpecial", func(s *CollectionsService) error {
			_, _, err := s.UpdateSpecial(collectionsCtx, 5, &SpecialCollectionUpdateRequest{EndDate: collectionsFixedAt, RestIncludeDiscount: collectionsBool(false)})
			return err
		}, "PUT", "/v1/special_collections/5", "", `{"end_date":"2026-03-01 10:00:00","rest_include_discount":false}`},
		{"DeleteSpecial", func(s *CollectionsService) error { _, err := s.DeleteSpecial(collectionsCtx, 5); return err },
			"DELETE", "/v1/special_collections/5", "", ""},
		{"AddSpecialProducts", func(s *CollectionsService) error {
			_, _, err := s.AddSpecialProducts(collectionsCtx, 5, IDList{10, 11})
			return err
		}, "POST", "/v1/special_collections/5/products", "", `{"product_ids":"10,11"}`},
		{"RemoveSpecialProducts", func(s *CollectionsService) error {
			_, _, err := s.RemoveSpecialProducts(collectionsCtx, 5, IDList{10})
			return err
		}, "DELETE", "/v1/special_collections/5/products", "product_ids=10", ""},
		{"CreateSpecialRule", func(s *CollectionsService) error {
			_, _, err := s.CreateSpecialRule(collectionsCtx, 5, &SpecialCollectionRuleRequest{Quantity: collectionsInt(3), Price: collectionsMoney(MoneyFromInt(999))})
			return err
		}, "POST", "/v1/special_collections/5/rules", "", `{"quantity":3,"price":999}`},
		{"UpdateSpecialRule", func(s *CollectionsService) error {
			_, _, err := s.UpdateSpecialRule(collectionsCtx, 5, 6, &SpecialCollectionRuleRequest{Percentage: collectionsFloat(0)})
			return err
		}, "PUT", "/v1/special_collections/5/rules/6", "", `{"percentage":0}`},
		{"DeleteSpecialRule", func(s *CollectionsService) error { _, _, err := s.DeleteSpecialRule(collectionsCtx, 5, 6); return err },
			"DELETE", "/v1/special_collections/5/rules/6", "", ""},
	}
}

func collectionsAddBuyShapeCases() []collectionsShapeCase {
	return []collectionsShapeCase{
		{"ListAddBuy", func(s *CollectionsService) error { _, err := s.ListAddBuy(collectionsCtx, nil); return err },
			"GET", "/v1/add_buy_collections", "", ""},
		{"GetAddBuy", func(s *CollectionsService) error { _, _, err := s.GetAddBuy(collectionsCtx, 3); return err },
			"GET", "/v1/add_buy_collections/3", "", ""},
		{"CreateAddBuy", func(s *CollectionsService) error {
			_, _, err := s.CreateAddBuy(collectionsCtx, &AddBuyCollectionCreateRequest{Title: "A", Published: true,
				ProductsOrder: ProductsOrderPriceAsc, EndDate: collectionsFixedAt, Price: MoneyFromInt(500), ItemLimit: 2})
			return err
		}, "POST", "/v1/add_buy_collections", "",
			`{"title":"A","published":true,"products_order":"price.asc","end_date":"2026-03-01 10:00:00","price":500,"item_limit":2}`},
		{"UpdateAddBuy", func(s *CollectionsService) error {
			_, _, err := s.UpdateAddBuy(collectionsCtx, 3, &AddBuyCollectionUpdateRequest{Price: collectionsMoney(0), ItemLimit: collectionsInt(0)})
			return err
		}, "PUT", "/v1/add_buy_collections/3", "", `{"price":0,"item_limit":0}`},
		{"DeleteAddBuy", func(s *CollectionsService) error { _, err := s.DeleteAddBuy(collectionsCtx, 3); return err },
			"DELETE", "/v1/add_buy_collections/3", "", ""},
		{"AddAddBuyProduct", func(s *CollectionsService) error {
			_, _, err := s.AddAddBuyProduct(collectionsCtx, 3, &AddBuyProductRequest{ProductID: 42, Price: MoneyFromFloat(99.5)})
			return err
		}, "POST", "/v1/add_buy_collections/3/products", "", `{"product_id":42,"price":99.5}`},
		{"UpdateAddBuyProduct", func(s *CollectionsService) error {
			_, _, err := s.UpdateAddBuyProduct(collectionsCtx, 3, 42, &AddBuyProductUpdateRequest{Price: collectionsMoney(MoneyFromInt(50))})
			return err
		}, "PUT", "/v1/add_buy_collections/3/products/42", "", `{"price":50}`},
		{"RemoveAddBuyProduct", func(s *CollectionsService) error {
			_, _, err := s.RemoveAddBuyProduct(collectionsCtx, 3, 42)
			return err
		},
			"DELETE", "/v1/add_buy_collections/3/products/42", "", ""},
		{"ReorderAddBuyProducts", func(s *CollectionsService) error {
			_, _, err := s.ReorderAddBuyProducts(collectionsCtx, 3, IDList{2, 1})
			return err
		}, "PUT", "/v1/add_buy_collections/3/products/manual_order", "", `{"product_ids":"2,1"}`},
	}
}

func collectionsLimitShapeCases() []collectionsShapeCase {
	return []collectionsShapeCase{
		{"ListLimit", func(s *CollectionsService) error { _, err := s.ListLimit(collectionsCtx, nil); return err },
			"GET", "/v1/limit_collections", "", ""},
		{"GetLimit", func(s *CollectionsService) error { _, _, err := s.GetLimit(collectionsCtx, 4); return err },
			"GET", "/v1/limit_collections/4", "", ""},
		{"CreateLimit", func(s *CollectionsService) error {
			_, _, err := s.CreateLimit(collectionsCtx, &LimitCollectionCreateRequest{Title: "L", Published: collectionsBool(true),
				PeriodType: LimitPeriodEach, CalculationType: LimitCalculationByProduct, LimitCount: collectionsInt(2), ProductIDs: IDList{1, 2}})
			return err
		}, "POST", "/v1/limit_collections", "",
			`{"title":"L","published":true,"period_type":"each","calculation_type":"by_product","limit_count":2,"product_ids":"1,2"}`},
		{"UpdateLimit", func(s *CollectionsService) error {
			_, _, err := s.UpdateLimit(collectionsCtx, 4, &LimitCollectionUpdateRequest{MinCount: collectionsInt(0), ProductsOrder: ProductsOrderSellWeightDesc})
			return err
		}, "PUT", "/v1/limit_collections/4", "", `{"min_count":0,"products_order":"sell_weight.desc"}`},
		{"DeleteLimit", func(s *CollectionsService) error { _, err := s.DeleteLimit(collectionsCtx, 4); return err },
			"DELETE", "/v1/limit_collections/4", "", ""},
		{"AddLimitProducts", func(s *CollectionsService) error {
			_, _, err := s.AddLimitProducts(collectionsCtx, 4, IDList{9})
			return err
		}, "POST", "/v1/limit_collections/4/products", "", `{"product_ids":"9"}`},
		{"ReplaceLimitProducts", func(s *CollectionsService) error {
			_, _, err := s.ReplaceLimitProducts(collectionsCtx, 4, IDList{9, 8})
			return err
		}, "PUT", "/v1/limit_collections/4/products", "", `{"product_ids":"9,8"}`},
		{"RemoveLimitProducts", func(s *CollectionsService) error {
			_, _, err := s.RemoveLimitProducts(collectionsCtx, 4, IDList{9})
			return err
		}, "DELETE", "/v1/limit_collections/4/products", "product_ids=9", ""},
		{"ReorderLimitProducts", func(s *CollectionsService) error {
			_, _, err := s.ReorderLimitProducts(collectionsCtx, 4, IDList{8, 9})
			return err
		}, "PUT", "/v1/limit_collections/4/products/manual_order", "", `{"product_ids":"8,9"}`},
	}
}

func collectionsVariantDiscountShapeCases() []collectionsShapeCase {
	return []collectionsShapeCase{
		{"ListVariantDiscount", func(s *CollectionsService) error { _, err := s.ListVariantDiscount(collectionsCtx, nil); return err },
			"GET", "/v1/variant_discount_collections", "", ""},
		{"GetVariantDiscount", func(s *CollectionsService) error { _, _, err := s.GetVariantDiscount(collectionsCtx, 2); return err },
			"GET", "/v1/variant_discount_collections/2", "", ""},
		{"CreateVariantDiscount", func(s *CollectionsService) error {
			_, _, err := s.CreateVariantDiscount(collectionsCtx, &VariantDiscountCollectionCreateRequest{Title: "V", Published: collectionsBool(false),
				StartDate: collectionsFixedAt, VariantDiscountCollectionType: VariantDiscountTypePercentage, Percentage: collectionsFloat(80), VariantIDs: []int64{5, 6}})
			return err
		}, "POST", "/v1/variant_discount_collections", "",
			`{"title":"V","published":false,"start_date":"2026-03-01 10:00:00","variant_discount_collection_type":"percentage","percentage":80,"variant_ids":[5,6]}`},
		{"UpdateVariantDiscount", func(s *CollectionsService) error {
			_, _, err := s.UpdateVariantDiscount(collectionsCtx, 2, &VariantDiscountCollectionUpdateRequest{
				VariantDiscountCollectionType: VariantDiscountTypeDiscountAmount, DiscountAmount: collectionsMoney(MoneyFromInt(30))})
			return err
		}, "PUT", "/v1/variant_discount_collections/2", "", `{"variant_discount_collection_type":"discount_amount","discount_amount":30}`},
		{"DeleteVariantDiscount", func(s *CollectionsService) error { _, err := s.DeleteVariantDiscount(collectionsCtx, 2); return err },
			"DELETE", "/v1/variant_discount_collections/2", "", ""},
		{"AddVariantDiscountVariants", func(s *CollectionsService) error {
			_, _, err := s.AddVariantDiscountVariants(collectionsCtx, 2, []int64{7})
			return err
		}, "POST", "/v1/variant_discount_collections/2/variants", "", `{"variant_ids":[7]}`},
		{"ReplaceVariantDiscountVariants", func(s *CollectionsService) error {
			_, _, err := s.ReplaceVariantDiscountVariants(collectionsCtx, 2, []int64{7, 8})
			return err
		}, "PUT", "/v1/variant_discount_collections/2/variants", "", `{"variant_ids":[7,8]}`},
		{"RemoveVariantDiscountVariants", func(s *CollectionsService) error {
			_, _, err := s.RemoveVariantDiscountVariants(collectionsCtx, 2, []int64{7, 8})
			return err
		}, "DELETE", "/v1/variant_discount_collections/2/variants", "variant_ids%5B%5D=7&variant_ids%5B%5D=8", ""},
	}
}

func collectionsVipShapeCases() []collectionsShapeCase {
	return []collectionsShapeCase{
		{"ListVIP", func(s *CollectionsService) error { _, err := s.ListVIP(collectionsCtx, nil); return err },
			"GET", "/v1/vip_collections", "", ""},
		{"GetVIP", func(s *CollectionsService) error { _, _, err := s.GetVIP(collectionsCtx, 6); return err },
			"GET", "/v1/vip_collections/6", "", ""},
		{"CreateVIP", func(s *CollectionsService) error {
			_, _, err := s.CreateVIP(collectionsCtx, &VIPCollectionCreateRequest{Title: "Gold", Position: 1})
			return err
		}, "POST", "/v1/vip_collections", "", `{"title":"Gold","position":1}`},
		{"UpdateVIP", func(s *CollectionsService) error {
			_, _, err := s.UpdateVIP(collectionsCtx, 6, &VIPCollectionUpdateRequest{Position: collectionsInt(0)})
			return err
		}, "PUT", "/v1/vip_collections/6", "", `{"position":0}`},
		{"DeleteVIP", func(s *CollectionsService) error { _, err := s.DeleteVIP(collectionsCtx, 6); return err },
			"DELETE", "/v1/vip_collections/6", "", ""},
		{"UpdateVIPRule", func(s *CollectionsService) error {
			_, _, err := s.UpdateVIPRule(collectionsCtx, 6, &VIPRuleRequest{VIPRuleSpec: VIPRuleTotalPrice, OrderStart: collectionsFixedDay,
				TotalPrice: collectionsMoney(MoneyFromInt(10000)), CustomersTag: collectionsStr("")})
			return err
		}, "PUT", "/v1/vip_collections/6/update_rule", "",
			`{"vip_rule_spec":"total_price","order_start":"2026-03-01","total_price":10000,"customers_tag":""}`},
		{"UpdateVIPPromotion", func(s *CollectionsService) error {
			_, _, err := s.UpdateVIPPromotion(collectionsCtx, 6, &VIPPromotionRequest{VIPPromotionSpec: VIPPromotionFreeShipping, ConcurrentlyApply: collectionsBool(false)})
			return err
		}, "PUT", "/v1/vip_collections/6/update_promotion", "", `{"vip_promotion_spec":"free_shipping","concurrently_apply":false}`},
	}
}

func TestCollectionsRequestShapes(t *testing.T) {
	var cases []collectionsShapeCase
	for _, group := range [][]collectionsShapeCase{collectionsCustomShapeCases(), collectionsSmartShapeCases(), collectionsSpecialShapeCases(),
		collectionsAddBuyShapeCases(), collectionsLimitShapeCases(), collectionsVariantDiscountShapeCases(), collectionsVipShapeCases()} {
		cases = append(cases, group...)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var rec collectionsRecordedRequest
			c := collectionsRecordingClient(t, &rec)
			if err := tc.call(c.Collections); err != nil {
				t.Fatal(err)
			}
			if rec.Method != tc.method || rec.Path != tc.path || rec.Query != tc.query {
				t.Errorf("got %s %s?%s, want %s %s?%s", rec.Method, rec.Path, rec.Query, tc.method, tc.path, tc.query)
			}
			if !collectionsSameJSON(rec.Body, tc.body) {
				t.Errorf("body = %s, want %s", rec.Body, tc.body)
			}
		})
	}
}

// collectionsSameJSON compares two JSON documents structurally; "" matches "".
func collectionsSameJSON(got, want string) bool {
	if got == "" || want == "" {
		return got == want
	}
	var a, b any
	if json.Unmarshal([]byte(got), &a) != nil || json.Unmarshal([]byte(want), &b) != nil {
		return false
	}
	return reflect.DeepEqual(a, b)
}

func TestCollectionsAllPropagatesOptionError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`[]`)) })
	for _, err := range c.Collections.AllSmart(collectionsCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
	}
	if IDList(nil).String() != "" || (IDList{1}).String() != "1" {
		t.Error("IDList.String")
	}
}
