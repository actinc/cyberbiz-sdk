package cyberbiz

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// MaxBranchStoreInventoryItems is the documented cap on items per
// BatchUpdateBranchStoreInventory call.
const MaxBranchStoreInventoryItems = 1000

// ProductSearchV2Request is the JSON body of SearchV2. Every field is
// optional. The endpoint pages with Limit and Offset and reports totals in
// the response body.
type ProductSearchV2Request struct {
	Q                 string             `json:"q,omitzero"` // fuzzy match on the title
	Limit             int                `json:"limit,omitzero"`
	Offset            int                `json:"offset,omitzero"`
	FilterPublished   *bool              `json:"filter_published,omitzero"` // defaults to true on the platform
	Vendors           []string           `json:"vendors,omitzero"`
	ProductTypes      []string           `json:"product_types,omitzero"`
	Tags              []string           `json:"tags,omitzero"`
	StoreTypes        []ProductStoreType `json:"store_types,omitzero"`
	FilterBranchStore *bool              `json:"filter_branch_store,omitzero"`
	FilterOnSell      *bool              `json:"filter_on_sell,omitzero"`
	IDs               []int64            `json:"ids,omitzero"`
	ExcludeTags       []string           `json:"exclude_tags,omitzero"`
	SKUs              []string           `json:"skus,omitzero"`          // custom feature
	BranchStores      []int64            `json:"branch_stores,omitzero"` // branch store ids, custom feature
	StoreNos          []string           `json:"store_nos,omitzero"`     // branch store numbers, custom feature
}

// BranchStoreInventoryItem sets the stock of one SKU in a branch store.
type BranchStoreInventoryItem struct {
	SKU               string `json:"sku"`
	InventoryQuantity int    `json:"inventory_quantity"`
}

// BranchStoreInventoryUpdateRequest is the body of
// BatchUpdateBranchStoreInventory. Both fields are required and Items holds
// at most MaxBranchStoreInventoryItems entries.
type BranchStoreInventoryUpdateRequest struct {
	StoreNumber string                     `json:"store_number"`
	Items       []BranchStoreInventoryItem `json:"items"`
}

// SearchV2 finds products with the richer v2 filters (POST /v2/products/search).
func (s *ProductsService) SearchV2(ctx context.Context, req *ProductSearchV2Request) (*ProductSearchResult, *Response, error) {
	if req == nil {
		req = &ProductSearchV2Request{}
	}
	var out ProductSearchResult
	resp, err := s.client.post(ctx, "v2/products/search", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// BatchUpdateBranchStoreInventory queues an asynchronous stock update for
// the SKUs of one branch store and returns the job to poll with
// GetBranchStoreInventoryUpdateStatus. Requires the "batch update branch
// store inventory by store number and SKU" feature
// (PUT /v2/products/batch_update_branch_store_inventory).
func (s *ProductsService) BatchUpdateBranchStoreInventory(ctx context.Context, req *BranchStoreInventoryUpdateRequest) (*BranchStoreInventoryJob, *Response, error) {
	if req == nil {
		return nil, nil, errors.New("cyberbiz: request must not be nil")
	}
	if len(req.Items) > MaxBranchStoreInventoryItems {
		return nil, nil, fmt.Errorf("cyberbiz: at most %d inventory items per call, got %d", MaxBranchStoreInventoryItems, len(req.Items))
	}
	var out BranchStoreInventoryJob
	resp, err := s.client.put(ctx, "v2/products/batch_update_branch_store_inventory", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// GetBranchStoreInventoryUpdateStatus polls a batch inventory update job
// (GET /v2/products/branch_store_inventory_update_status?job_id=).
func (s *ProductsService) GetBranchStoreInventoryUpdateStatus(ctx context.Context, jobID string) (*BranchStoreInventoryJobState, *Response, error) {
	var out BranchStoreInventoryJobState
	q := url.Values{"job_id": {jobID}}
	resp, err := s.client.getOne(ctx, "v2/products/branch_store_inventory_update_status", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
