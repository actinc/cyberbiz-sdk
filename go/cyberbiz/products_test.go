package cyberbiz

import (
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// productsRecorded is what the recording test server saw on the last request.
type productsRecorded struct {
	Method      string
	Path        string
	Query       string
	ContentType string
	Body        []byte
}

// productsRecorder returns a client whose server records every request and
// answers with body.
func productsRecorder(t *testing.T, status int, body string) (*Client, *productsRecorded) {
	t.Helper()
	rec := &productsRecorded{}
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.Method, rec.Path, rec.Query = r.Method, r.URL.EscapedPath(), r.URL.RawQuery
		rec.ContentType = r.Header.Get("Content-Type")
		rec.Body, _ = io.ReadAll(r.Body)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}, WithMaxRetries(0))
	return c, rec
}

func (r *productsRecorded) assert(t *testing.T, method, path, query string) {
	t.Helper()
	if r.Method != method || r.Path != path || r.Query != query {
		t.Errorf("got %s %s?%s, want %s %s?%s", r.Method, r.Path, r.Query, method, path, query)
	}
}

// assertJSON compares the recorded body with want as JSON values.
func (r *productsRecorded) assertJSON(t *testing.T, want string) {
	t.Helper()
	var got, exp any
	if err := json.Unmarshal(r.Body, &got); err != nil {
		t.Fatalf("body %q is not JSON: %v", r.Body, err)
	}
	if err := json.Unmarshal([]byte(want), &exp); err != nil {
		t.Fatalf("want %q is not JSON: %v", want, err)
	}
	g, _ := json.Marshal(got, json.Deterministic(true))
	e, _ := json.Marshal(exp, json.Deterministic(true))
	if string(g) != string(e) {
		t.Errorf("body = %s, want %s", g, e)
	}
}

func productsPtr[T any](v T) *T { return &v }

// Golden decode tests.

func TestProductGoldenList(t *testing.T) {
	var out []Product
	decodeGolden(t, "v1/GET_v1_products.json", &out)
	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	p := out[0]
	if p.ID != 69458894 || p.Title == "" || p.Published {
		t.Errorf("product = %+v", p)
	}
	if p.Price != MoneyFromInt(200) || p.TaxTypeID != TaxTypeInclusive {
		t.Errorf("price = %v tax = %q", p.Price, p.TaxTypeID)
	}
	if p.SellFrom != nil || p.SellTo != nil || p.PosShop != nil || p.SpecialCollection != nil {
		t.Errorf("nullable fields not nil: %+v", p)
	}
	if p.CreatedAt.String() != "2026-07-10 20:22:05" {
		t.Errorf("created_at = %v", p.CreatedAt)
	}
	if len(p.TemperatureTypes) != 1 || p.TemperatureTypes[0] != TemperatureTypeRoom {
		t.Errorf("temperature_types = %v", p.TemperatureTypes)
	}
	if p.SEOMetaTags == nil || p.SEOMetaTags.Title != "" {
		t.Errorf("seo_meta_tags = %+v", p.SEOMetaTags)
	}
	if !p.Searchable || p.GoogleProductCategoryID != 0 || len(p.PhotoURLs) != 0 {
		t.Errorf("misc = %+v", p)
	}
	if out[1].PosShop == nil || out[1].PosShop.ID != 1554 {
		t.Errorf("pos_shop = %+v", out[1].PosShop)
	}
}

func TestProductGoldenDetail(t *testing.T) {
	var p Product
	decodeGolden(t, "v1/GET_v1_products_{id}.json", &p)
	if p.ID != 0 {
		t.Errorf("detail should omit id, got %d", p.ID)
	}
	if len(p.ProductVariants) != 1 {
		t.Fatalf("variants = %d", len(p.ProductVariants))
	}
	v := p.ProductVariants[0]
	if v.ID != 84683248 || v.ProductID != 69458894 || v.SKU != "SKU-5ac62f" || v.QC != "WARE0014" {
		t.Errorf("variant = %+v", v)
	}
	if v.Cost != MoneyFromInt(190) || v.CompareAtPrice != MoneyFromInt(200) || v.MaxUsableBonus != 0 {
		t.Errorf("variant money = %+v", v)
	}
	if v.InventoryPolicy != InventoryPolicyDeny || !v.InventoryManagement || !v.HoneycombSync {
		t.Errorf("variant flags = %+v", v)
	}
	if v.SafetyInventoryQuantity != 0 || v.Option1 != "" {
		t.Errorf("variant nulls = %+v", v)
	}
	if v.UpdatedAt.String() != "2026-07-10 20:22:07" {
		t.Errorf("variant updated_at = %v", v.UpdatedAt)
	}
	if p.Photos == nil || len(p.Photos) != 0 || p.Channel != nil || p.BranchStore != nil {
		t.Errorf("detail nested = %+v", p)
	}
	if p.UpdatedAt.String() != "2026-07-29 09:08:40" {
		t.Errorf("updated_at = %v", p.UpdatedAt)
	}
}

func TestProductGoldenSearch(t *testing.T) {
	var out []Product
	decodeGolden(t, "v1/GET_v1_products_search_query.json", &out)
	if len(out) != 2 || out[0].ID != 69221930 || !out[0].Published {
		t.Fatalf("search = %+v", out)
	}
	if len(out[0].PhotoURLs) != 1 || !strings.HasPrefix(out[0].PhotoURLs[0], "//cdn-general") {
		t.Errorf("photo_urls = %v", out[0].PhotoURLs)
	}
	if out[0].Price != MoneyFromInt(12000) {
		t.Errorf("price = %v", out[0].Price)
	}

	var coll []Product
	decodeGolden(t, "v1/GET_v1_products_search_collection.json", &coll)
	p := coll[0]
	if p.ID != 63639028 || p.SellFrom == nil || p.SellFrom.String() != "2025-12-19 00:00:00" {
		t.Fatalf("collection product = %+v", p)
	}
	if p.SellTo == nil || p.SellTo.String() != "2027-04-30 00:00:00" || p.BriefText == "" {
		t.Errorf("sell_to = %v brief_text = %q", p.SellTo, p.BriefText)
	}
	if len(p.CustomCollections) != 1 || p.CustomCollections[0].ID != 403543 || !p.CustomCollections[0].Published {
		t.Errorf("custom_collections = %+v", p.CustomCollections)
	}
	if p.CustomCollections[0].Position != 1 || p.CustomCollections[0].ProductsOrderName == "" {
		t.Errorf("custom_collection ref = %+v", p.CustomCollections[0])
	}
}

func TestProductGoldenSubResources(t *testing.T) {
	var tags []ProductTag
	decodeGolden(t, "v1/GET_v1_products_{id}_product_tags.json", &tags)
	if tags == nil || len(tags) != 0 {
		t.Errorf("tags = %v", tags)
	}
	var descs []ProductDescription
	decodeGolden(t, "v1/GET_v1_products_{id}_product_descriptions.json", &descs)
	if descs == nil || len(descs) != 0 {
		t.Errorf("descriptions = %v", descs)
	}
	var names []ProductDescriptionSettingName
	decodeGolden(t, "v1/GET_v1_products_get_product_description_setting_names.json", &names)
	if names == nil || len(names) != 0 {
		t.Errorf("setting names = %v", names)
	}
	var all ProductShippingNames
	decodeGolden(t, "v1/GET_v1_products_bind_shippings.json", &all)
	if len(all.ShippingNames) != 4 || all.ShippingNames[0] != "黑貓宅急便" {
		t.Errorf("bind_shippings = %v", all.ShippingNames)
	}
	var one ProductShippingNames
	decodeGolden(t, "v1/GET_v1_products_{id}_bind_shippings.json", &one)
	if len(one.ShippingNames) != 4 || one.ShippingNames[0] != "門市取貨（預設）" {
		t.Errorf("product bind_shippings = %v", one.ShippingNames)
	}
}

// Request shape tests.

func TestProductsListAndAll(t *testing.T) {
	c, rec := productsRecorder(t, 200, `[{"id":1,"title":"a"}]`)
	page, err := c.Products.List(testCtx, &ProductListOptions{ListOptions: ListOptions{Page: 2, PerPage: 10}})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products", "page=2&per_page=10")
	if len(page.Items) != 1 || page.Items[0].ID != 1 {
		t.Errorf("items = %+v", page.Items)
	}
	var n int
	for p, err := range c.Products.All(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
		if p.Title != "a" {
			t.Errorf("title = %q", p.Title)
		}
	}
	if n != 1 {
		t.Errorf("all yielded %d", n)
	}
	rec.assert(t, "GET", "/v1/products", "page=1&per_page=50")
}

func TestProductsGetRoundTrip(t *testing.T) {
	c := goldenServer(t)
	p, resp, err := c.Products.Get(testCtx, 69458894)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != 69458894 || p.Title == "" || resp.StatusCode != 200 {
		t.Errorf("product = %+v", p)
	}
	page, err := c.Products.List(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Pagination.Total != 1 {
		t.Errorf("page = %+v", page.Pagination)
	}
}

func TestProductsGetMapsErrorGolden(t *testing.T) {
	c, _ := productsRecorder(t, 404, string(readGolden(t, "errors/GET_v1_products_{id}.json")))
	_, _, err := c.Products.Get(testCtx, 1)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !errors.Is(err, ErrNotFound) || apiErr.Messages[0] != "無此資源" {
		t.Fatalf("got %v", err)
	}
}

func TestProductsCreateBody(t *testing.T) {
	c, rec := productsRecorder(t, 201, `{"title":"T","id":0}`)
	req := &ProductCreateRequest{
		Title: "T", Handle: "t-1", Published: false, Price: MoneyFromFloat(99.5),
		SellFrom:         productsTime("2026-01-02 03:04:05"),
		BriefText:        productsPtr(""),
		Searchable:       productsPtr(false),
		TaxTypeID:        TaxTypeZero,
		TemperatureTypes: []TemperatureType{TemperatureTypeFrozen},
	}
	if _, _, err := c.Products.Create(testCtx, req); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products", "")
	rec.assertJSON(t, `{"title":"T","handle":"t-1","published":false,"price":99.5,
		"sell_from":"2026-01-02 03:04:05","brief_text":"","searchable":false,
		"tax_type_id":"zero_tax","temperature_types":["冷凍"]}`)
	if rec.ContentType != "application/json" {
		t.Errorf("content-type = %q", rec.ContentType)
	}
}

func TestProductsUpdateBody(t *testing.T) {
	c, rec := productsRecorder(t, 200, `{"title":"T"}`)
	req := &ProductUpdateRequest{
		Published: productsPtr(true), Price: productsPtr(MoneyFromInt(0)),
		RelatedCollections: []ProductRelatedCollectionInput{{CollectionID: 7, CollectionType: "SmartCollection"}},
	}
	p, _, err := c.Products.Update(testCtx, 5, req)
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v1/products/5", "")
	rec.assertJSON(t, `{"published":true,"price":0,"related_collections":[{"collection_id":7,"collection_type":"SmartCollection"}]}`)
	if p.ID != 5 {
		t.Errorf("id not set from path: %d", p.ID)
	}
	if _, _, err := c.Products.Update(testCtx, 5, &ProductUpdateRequest{}); err != nil {
		t.Fatal(err)
	}
	rec.assertJSON(t, `{}`)
}

func TestProductsDeleteAndPosShopBatch(t *testing.T) {
	c, rec := productsRecorder(t, 200, `[{"id":9}]`)
	if _, err := c.Products.Delete(testCtx, 5); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "DELETE", "/v1/products/5", "")
	out, _, err := c.Products.CreateForPosShops(testCtx, &ProductPosShopBatchCreateRequest{Title: "T", PosShopIDs: "1,2", Price: productsPtr(MoneyFromInt(10))})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products/pos_shop_batch", "")
	rec.assertJSON(t, `{"title":"T","pos_shop_ids":"1,2","price":10}`)
	if len(out) != 1 || out[0].ID != 9 {
		t.Errorf("out = %+v", out)
	}
}

func TestProductsSearchQuery(t *testing.T) {
	c, rec := productsRecorder(t, 200, `[{"id":3}]`)
	out, _, err := c.Products.Search(testCtx, &ProductSearchOptions{
		Q: "tea", Limit: 5, Offset: 10, FilterPublished: productsPtr(false),
		OrderBy: ProductSearchOrderBySalesVolume, Vendor: "v", FilterBranchStore: productsPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/search", "filter_branch_store=true&filter_published=false&limit=5&offset=10&order_by=sales_volume&q=tea&vendor=v")
	if len(out) != 1 || out[0].ID != 3 {
		t.Errorf("out = %+v", out)
	}
	if _, _, err := c.Products.SearchCollection(testCtx, &ProductCollectionSearchOptions{CollectionHandle: "所有商品", Limit: 2}); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/search/collection", "collection_handle=%E6%89%80%E6%9C%89%E5%95%86%E5%93%81&limit=2")
}

func TestProductsSEOAndTags(t *testing.T) {
	c, rec := productsRecorder(t, 200, `[{"id":1,"name":"new","category":0}]`)
	tags, _, err := c.Products.AddTags(testCtx, 5, []string{"new", "hot"})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v1/products/5/product_tags/add", "")
	rec.assertJSON(t, `{"tags":"new,hot"}`)
	if len(tags) != 1 || tags[0].Name != "new" {
		t.Errorf("tags = %+v", tags)
	}
	if _, _, err := c.Products.RemoveTags(testCtx, 5, []string{"hot"}); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v1/products/5/product_tags/remove", "")
	rec.assertJSON(t, `{"tags":"hot"}`)
	if _, _, err := c.Products.ListTags(testCtx, 5); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/5/product_tags", "")

	c, rec = productsRecorder(t, 200, `{"title":"T","seo_meta_tags":{"title":"S","description":null,"keywords":null}}`)
	p, _, err := c.Products.UpdateSEOMetaTags(testCtx, 5, &ProductSEOMetaTagsUpdateRequest{Title: productsPtr("S"), Keywords: productsPtr("")})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v1/products/5/seo_meta_tags", "")
	rec.assertJSON(t, `{"title":"S","keywords":""}`)
	if p.ID != 5 || p.SEOMetaTags == nil || p.SEOMetaTags.Title != "S" {
		t.Errorf("product = %+v", p)
	}
}

func TestProductsBindShippings(t *testing.T) {
	c, rec := productsRecorder(t, 200, `{"shipping_names":["a","b"]}`)
	names, _, err := c.Products.ListBindableShippings(testCtx)
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/bind_shippings", "")
	if len(names) != 2 || names[1] != "b" {
		t.Errorf("names = %v", names)
	}
	if _, _, err := c.Products.GetBindShippings(testCtx, 5); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/5/bind_shippings", "")
	if _, _, err := c.Products.BindShippings(testCtx, 5, []string{"a"}); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products/5/bind_shippings", "")
	rec.assertJSON(t, `{"shipping_names":["a"]}`)

	c, rec = productsRecorder(t, 200, `[{"setting_name":"product_description_section_spec","title":"規格"}]`)
	names2, _, err := c.Products.ListDescriptionSettingNames(testCtx)
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/get_product_description_setting_names", "")
	if len(names2) != 1 || names2[0].SettingName != ProductDescriptionSettingSpec {
		t.Errorf("names = %+v", names2)
	}
}

func productsTime(s string) Time {
	parsed, err := ParseTime(s)
	if err != nil {
		panic(err)
	}
	return parsed
}
