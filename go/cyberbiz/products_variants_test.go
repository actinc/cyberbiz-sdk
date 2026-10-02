package cyberbiz

import (
	"testing"
)

func TestProductVariantGolden(t *testing.T) {
	var list []ProductVariant
	decodeGolden(t, "v1/GET_v1_products_{id}_product_variants.json", &list)
	if len(list) != 1 || list[0].ID != 84683248 || list[0].Price != MoneyFromInt(200) {
		t.Fatalf("list = %+v", list)
	}
	if list[0].Weight != 0 || list[0].Meas != 0 || list[0].Sold != 0 || list[0].Position != 1 {
		t.Errorf("numbers = %+v", list[0])
	}

	var v ProductVariant
	decodeGolden(t, "v1/GET_v1_products_{id}_product_variants_{id}.json", &v)
	if v.ID != 0 {
		t.Errorf("detail should omit id, got %d", v.ID)
	}
	if v.ProductID != 69458894 || v.Cost != MoneyFromInt(190) || v.InventoryPolicy != InventoryPolicyDeny {
		t.Errorf("detail = %+v", v)
	}
	if len(v.PhotoURLs) != 1 || v.CreatedAt.String() != "2026-07-10 20:22:05" {
		t.Errorf("photo_urls = %v created_at = %v", v.PhotoURLs, v.CreatedAt)
	}

	var bySKU []ProductVariant
	decodeGolden(t, "v1/GET_v1_products_sku_{id}_product_variants.json", &bySKU)
	if len(bySKU) != 2 || bySKU[0].ProductID != 65995816 || bySKU[1].ID != 80028076 {
		t.Fatalf("by sku = %+v", bySKU)
	}
	if bySKU[0].InventoryPolicy != InventoryPolicyContinue || bySKU[0].SKU != "SKU-5ac62f" {
		t.Errorf("by sku[0] = %+v", bySKU[0])
	}
}

func TestProductOptionGolden(t *testing.T) {
	var list []ProductOption
	decodeGolden(t, "v1/GET_v1_products_{id}_product_options.json", &list)
	if list == nil || len(list) != 0 {
		t.Errorf("list = %v", list)
	}
	var o ProductOption
	decodeGolden(t, "v1/GET_v1_products_{id}_product_options_{id}.json", &o)
	if o.ID != 0 || o.Position != 1 || o.Types != "頸枕,眼罩" || o.Name == "" {
		t.Errorf("option = %+v", o)
	}
}

func TestProductVariantsRoundTrip(t *testing.T) {
	c := goldenServer(t)
	list, _, err := c.Products.ListVariants(testCtx, 69458894)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != 84683248 {
		t.Errorf("list = %+v", list)
	}
	v, _, err := c.Products.GetVariant(testCtx, 69458894, 84683248)
	if err != nil {
		t.Fatal(err)
	}
	if v.ID != 84683248 || v.ProductID != 69458894 {
		t.Errorf("variant = %+v", v)
	}
	page, err := c.Products.ListVariantsBySKU(testCtx, "1101050001", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Pagination.Total != 1 {
		t.Errorf("page = %+v", page.Pagination)
	}
	o, _, err := c.Products.GetOption(testCtx, 69458894, 77)
	if err != nil {
		t.Fatal(err)
	}
	if o.ID != 77 || o.Types != "頸枕,眼罩" {
		t.Errorf("option = %+v", o)
	}
}

func TestProductVariantsRequests(t *testing.T) {
	c, rec := productsRecorder(t, 200, `{"product_id":5,"sku":"S"}`)
	v, _, err := c.Products.CreateVariant(testCtx, 5, &ProductVariantCreateRequest{
		Position: 1, InventoryManagement: false, InventoryQuantity: 0,
		InventoryPolicy: InventoryPolicyContinue, RequiresShipping: true,
		Price: productsPtr(MoneyFromFloat(199.5)), Cost: productsPtr(Money(0)), Weight: productsPtr(0.25), SKU: productsPtr("S"),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products/5/product_variants", "")
	rec.assertJSON(t, `{"position":1,"inventory_management":false,"inventory_quantity":0,
		"inventory_policy":"continue","requires_shipping":true,"price":199.5,"cost":0,"weight":0.25,"sku":"S"}`)
	if v.SKU != "S" {
		t.Errorf("variant = %+v", v)
	}

	v, _, err = c.Products.UpdateVariant(testCtx, 5, 6, &ProductVariantUpdateRequest{
		InventoryQuantity: productsPtr(0), HoneycombSync: productsPtr(false), Option1: productsPtr("紅"),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v1/products/5/product_variants/6", "")
	rec.assertJSON(t, `{"inventory_quantity":0,"honeycomb_sync":false,"option1":"紅"}`)
	if v.ID != 6 {
		t.Errorf("id not set from path: %d", v.ID)
	}

	if _, err := c.Products.DeleteVariant(testCtx, 5, 6); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "DELETE", "/v1/products/5/product_variants/6", "")

	c, rec = productsRecorder(t, 200, `[]`)
	if _, err := c.Products.ListVariantsBySKU(testCtx, "A/B C", &ListOptions{Page: 3}); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/sku/A%2FB%20C/product_variants", "page=3")
	for _, err := range c.Products.AllVariantsBySKU(testCtx, "X", nil) {
		if err != nil {
			t.Fatal(err)
		}
	}
	rec.assert(t, "GET", "/v1/products/sku/X/product_variants", "page=1&per_page=50")
}

func TestProductOptionsRequests(t *testing.T) {
	c, rec := productsRecorder(t, 200, `[{"name":"size","position":1,"types":"S,M","id":8}]`)
	opts, _, err := c.Products.ListOptions(testCtx, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/5/product_options", "")
	if len(opts) != 1 || opts[0].ID != 8 {
		t.Errorf("options = %+v", opts)
	}

	c, rec = productsRecorder(t, 200, `{"name":"size","position":1,"types":"S,M"}`)
	var o *ProductOption

	o, _, err = c.Products.CreateOption(testCtx, 5, &ProductOptionCreateRequest{Name: "size", Types: "S,M"})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products/5/product_options", "")
	rec.assertJSON(t, `{"name":"size","types":"S,M"}`)
	if o.Types != "S,M" {
		t.Errorf("option = %+v", o)
	}

	o, _, err = c.Products.UpdateOption(testCtx, 5, 8, &ProductOptionUpdateRequest{Types: productsPtr("")})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v1/products/5/product_options/8", "")
	rec.assertJSON(t, `{"types":""}`)
	if o.ID != 8 {
		t.Errorf("id not set from path: %d", o.ID)
	}

	if _, err := c.Products.DeleteOption(testCtx, 5, 8); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "DELETE", "/v1/products/5/product_options/8", "")
}
