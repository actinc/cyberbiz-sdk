package cyberbiz

import (
	"errors"
	"testing"
)

// productsV2SearchSample is the documented response of POST /v2/products/search,
// trimmed to the fields that differ from v1.
const productsV2SearchSample = `{
  "products": [{
    "id": 14, "title": "玩偶", "handle": "catup-20250529145127", "published": true,
    "sell_from": null, "sell_to": null, "price": 1000.0, "tax_type_id": "inclusive_tax",
    "custom_collections": [{"title": "首頁商品", "handle": "frontpage", "published": true,
      "body_html": null, "products_order_name": "按標題拼音升序: A-Z", "position": 0, "id": 1}],
    "special_collection": null, "tags": [],
    "product_variants": [{"product_id": 14, "name": "(多款) - 大 / 酷酷風", "position": 1,
      "price": 1000.0, "cost": null, "compare_at_price": 15000.0, "meas": 0.0, "max_usable_bonus": 0.0,
      "weight": 0.0, "option1": "大", "option2": "酷酷風", "option3": null,
      "inventory_management": false, "inventory_quantity": null, "sold": 0,
      "safety_inventory_quantity": null, "inventory_policy": "deny", "sku": "P0001", "qc": null,
      "requires_shipping": true, "created_at": "2025-05-29 14:51:27", "updated_at": "2025-06-17 16:03:58",
      "honeycomb_sync": true, "vendor": null, "id": 34, "photo_urls": ["//url/media/x.jpeg"]}],
    "product_options": [{"name": "大小", "position": 1, "types": "大,中,小", "id": 7}],
    "pos_shop": null, "photo_urls": ["//url/media/x.jpeg"], "related_collections": [],
    "created_at": "2025-05-29 14:51:27", "updated_at": "2025-05-29 14:51:57",
    "temperature_types": ["常溫"], "searchable": true, "google_product_category_id": null,
    "seo_meta_tags": {"title": null, "description": null, "keywords": null}
  }],
  "total_count": 1,
  "total_pages": 1
}`

func TestProductsSearchV2(t *testing.T) {
	c, rec := productsRecorder(t, 200, productsV2SearchSample)
	res, _, err := c.Products.SearchV2(testCtx, &ProductSearchV2Request{
		Q: "玩偶", Limit: 20, FilterPublished: productsPtr(false),
		StoreTypes: []ProductStoreType{ProductStoreTypeEC, ProductStoreTypeBranchStore},
		IDs:        []int64{14}, SKUs: []string{"P0001"}, BranchStores: []int64{10003}, StoreNos: []string{"AL001"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v2/products/search", "")
	rec.assertJSON(t, `{"q":"玩偶","limit":20,"filter_published":false,"store_types":[1,4],
		"ids":[14],"skus":["P0001"],"branch_stores":[10003],"store_nos":["AL001"]}`)
	if res.TotalCount != 1 || res.TotalPages != 1 || len(res.Products) != 1 {
		t.Fatalf("result = %+v", res)
	}
	p := res.Products[0]
	if p.ID != 14 || p.Handle != "catup-20250529145127" || p.Price != MoneyFromInt(1000) {
		t.Errorf("product = %+v", p)
	}
	v := p.ProductVariants[0]
	if v.ID != 34 || v.Option1 != "大" || v.CompareAtPrice != MoneyFromInt(15000) || v.InventoryQuantity != 0 || v.InventoryManagement {
		t.Errorf("variant = %+v", v)
	}
	if len(p.ProductOptions) != 1 || p.ProductOptions[0].ID != 7 || p.ProductOptions[0].Types != "大,中,小" {
		t.Errorf("options = %+v", p.ProductOptions)
	}
	if p.CustomCollections[0].ID != 1 || p.CustomCollections[0].Handle != "frontpage" {
		t.Errorf("custom_collections = %+v", p.CustomCollections)
	}

	if _, _, err := c.Products.SearchV2(testCtx, nil); err != nil {
		t.Fatal(err)
	}
	rec.assertJSON(t, `{}`)
}

func TestProductsBatchUpdateBranchStoreInventory(t *testing.T) {
	c, rec := productsRecorder(t, 200, `{"job_id":"60f9a9c01fe97b19d517c5d9","status":"QUEUED",
		"message":"Inventory update job has been queued","store_number":"123","total_items":2}`)
	job, _, err := c.Products.BatchUpdateBranchStoreInventory(testCtx, &BranchStoreInventoryUpdateRequest{
		StoreNumber: "123",
		Items:       []BranchStoreInventoryItem{{SKU: "sku1", InventoryQuantity: 10}, {SKU: "sku2", InventoryQuantity: 0}},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v2/products/batch_update_branch_store_inventory", "")
	rec.assertJSON(t, `{"store_number":"123","items":[{"sku":"sku1","inventory_quantity":10},{"sku":"sku2","inventory_quantity":0}]}`)
	if job.JobID != "60f9a9c01fe97b19d517c5d9" || job.Status != BranchStoreInventoryJobQueued || job.TotalItems != 2 {
		t.Errorf("job = %+v", job)
	}
	if job.Status.IsDone() {
		t.Error("queued job reported done")
	}

	tooMany := &BranchStoreInventoryUpdateRequest{StoreNumber: "1", Items: make([]BranchStoreInventoryItem, MaxBranchStoreInventoryItems+1)}
	if _, _, err := c.Products.BatchUpdateBranchStoreInventory(testCtx, tooMany); err == nil {
		t.Error("oversized batch accepted")
	}
	if _, _, err := c.Products.BatchUpdateBranchStoreInventory(testCtx, nil); err == nil {
		t.Error("nil request accepted")
	}

	c, _ = productsRecorder(t, 422, string(readGolden(t, "errors/GET_v2_products_branch_store_inventory_update_status.json")))
	_, _, err = c.Products.BatchUpdateBranchStoreInventory(testCtx, &BranchStoreInventoryUpdateRequest{StoreNumber: "1"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !errors.Is(err, ErrValidation) || len(apiErr.Messages) != 1 {
		t.Fatalf("got %v", err)
	}
}

func TestProductsGetBranchStoreInventoryUpdateStatus(t *testing.T) {
	c, rec := productsRecorder(t, 200, `{"job_id":"ad21ae26a2c7bb3af88f082d","status":"SUCCESS",
		"message":"Inventory update completed successfully",
		"created_at":"2025-07-10 17:34:12 +0800","completed_at":"2025-07-10 17:34:29 +0800",
		"result":{"total":2,"succeeded":1,"failed":1,"failed_items":[{"sku":"sku1","error":"xxxxxxxxx"}]},
		"error":null}`)
	st, _, err := c.Products.GetBranchStoreInventoryUpdateStatus(testCtx, "ad21ae26a2c7bb3af88f082d")
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v2/products/branch_store_inventory_update_status", "job_id=ad21ae26a2c7bb3af88f082d")
	if st.Status != BranchStoreInventoryJobSuccess || !st.Status.IsDone() || st.Error != "" {
		t.Errorf("state = %+v", st)
	}
	if st.CreatedAt.String() != "2025-07-10 17:34:12" || st.CompletedAt == nil || st.CompletedAt.String() != "2025-07-10 17:34:29" {
		t.Errorf("times = %v %v", st.CreatedAt, st.CompletedAt)
	}
	if st.Result == nil || st.Result.Failed != 1 || len(st.Result.FailedItems) != 1 || st.Result.FailedItems[0].SKU != "sku1" {
		t.Errorf("result = %+v", st.Result)
	}

	c, _ = productsRecorder(t, 200, `{"job_id":"6dd64c003ed9cb33e868db15","status":"QUEUE","message":"Job is queued for processing",
		"created_at":null,"completed_at":null,"result":null,"error":null}`)
	st, _, err = c.Products.GetBranchStoreInventoryUpdateStatus(testCtx, "6dd64c003ed9cb33e868db15")
	if err != nil {
		t.Fatal(err)
	}
	if st.Status != BranchStoreInventoryJobQueue || !st.CreatedAt.IsZero() || st.CompletedAt != nil || st.Result != nil {
		t.Errorf("queued state = %+v", st)
	}

	c, _ = productsRecorder(t, 200, `{"job_id":"c069e31be38362a817ffb3c2","status":"FAILED","message":"Inventory update failed",
		"created_at":"2025-07-10 17:36:16 +0800","completed_at":"2025-07-10 17:36:16 +0800","result":null,
		"error":"Store 123 is currently offline"}`)
	st, _, err = c.Products.GetBranchStoreInventoryUpdateStatus(testCtx, "c069e31be38362a817ffb3c2")
	if err != nil {
		t.Fatal(err)
	}
	if st.Status != BranchStoreInventoryJobFailed || st.Error != "Store 123 is currently offline" {
		t.Errorf("failed state = %+v", st)
	}
}
