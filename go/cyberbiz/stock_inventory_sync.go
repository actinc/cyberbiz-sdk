package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// InventorySyncGroupCreateRequest is the body of CreateInventorySyncGroup.
type InventorySyncGroupCreateRequest struct {
	Title string `json:"title"`
	// InventoryQuantity is the shared stock count and must be positive.
	InventoryQuantity int                 `json:"inventory_quantity"`
	InventoryPolicy   InventorySyncPolicy `json:"inventory_policy"`
}

// InventorySyncGroupUpdateRequest is the body of UpdateInventorySyncGroup.
type InventorySyncGroupUpdateRequest struct {
	Title *string `json:"title,omitzero"`
	// InventoryQuantity is the shared stock count and must be positive.
	InventoryQuantity *int                `json:"inventory_quantity,omitzero"`
	InventoryPolicy   InventorySyncPolicy `json:"inventory_policy,omitzero"`
}

// ListInventorySyncGroups returns one page of sync groups. A shop without
// the feature gets a 422 (GET /v1/inventory_sync_groups).
func (s *StockService) ListInventorySyncGroups(ctx context.Context, opts *ListOptions) (*Page[InventorySyncGroup], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[InventorySyncGroup](ctx, s.client, "v1/inventory_sync_groups", q)
}

// AllInventorySyncGroups walks every page of sync groups
// (GET /v1/inventory_sync_groups).
func (s *StockService) AllInventorySyncGroups(ctx context.Context, opts *ListOptions) iter.Seq2[InventorySyncGroup, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(InventorySyncGroup, error) bool) { yield(InventorySyncGroup{}, err) }
	}
	return listAll[InventorySyncGroup](ctx, s.client, "v1/inventory_sync_groups", q)
}

// GetInventorySyncGroup returns one sync group. The detail response omits
// the id, so it is filled in from the argument
// (GET /v1/inventory_sync_groups/{id}).
func (s *StockService) GetInventorySyncGroup(ctx context.Context, id int64) (*InventorySyncGroup, *Response, error) {
	var out InventorySyncGroup
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/inventory_sync_groups/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// CreateInventorySyncGroup creates a sync group
// (POST /v1/inventory_sync_groups).
func (s *StockService) CreateInventorySyncGroup(ctx context.Context, req *InventorySyncGroupCreateRequest) (*InventorySyncGroup, *Response, error) {
	var out InventorySyncGroup
	resp, err := s.client.post(ctx, "v1/inventory_sync_groups", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateInventorySyncGroup changes a sync group. The response omits the
// id, so it is filled in from the argument
// (PUT /v1/inventory_sync_groups/{id}).
func (s *StockService) UpdateInventorySyncGroup(ctx context.Context, id int64, req *InventorySyncGroupUpdateRequest) (*InventorySyncGroup, *Response, error) {
	var out InventorySyncGroup
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/inventory_sync_groups/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// DeleteInventorySyncGroup removes a sync group
// (DELETE /v1/inventory_sync_groups/{id}).
func (s *StockService) DeleteInventorySyncGroup(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/inventory_sync_groups/%d", id), nil)
}

// AddInventorySyncGroupVariants adds product variants to a sync group
// (POST /v1/inventory_sync_groups/{id}/variants).
func (s *StockService) AddInventorySyncGroupVariants(ctx context.Context, id int64, variantIDs []int64) (*InventorySyncGroup, *Response, error) {
	body := struct {
		VariantIDs string `json:"variant_ids"`
	}{VariantIDs: IDList(variantIDs).String()}
	var out InventorySyncGroup
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/inventory_sync_groups/%d/variants", id), body, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// RemoveInventorySyncGroupVariants removes product variants from a sync
// group. The ids travel in the query string, as the platform requires
// (DELETE /v1/inventory_sync_groups/{id}/variants?variant_ids=1,2).
func (s *StockService) RemoveInventorySyncGroupVariants(ctx context.Context, id int64, variantIDs []int64) (*InventorySyncGroup, *Response, error) {
	var out InventorySyncGroup
	resp, err := s.client.Do(ctx, &Request{
		Method: http.MethodDelete,
		Path:   fmt.Sprintf("v1/inventory_sync_groups/%d/variants", id),
		Query:  url.Values{"variant_ids": {IDList(variantIDs).String()}},
	}, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
