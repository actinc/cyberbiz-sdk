package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// Limit collections require a feature licence; the test shop answers every
// endpoint with 401 "無權使用該 API", so nothing here is verified against a
// Golden File. Shapes follow the swagger and Postman examples.

// LimitCollectionCreateRequest is the body of CreateLimit. Only Title is
// required. ProductsOrder does not accept [ProductsOrderManual].
type LimitCollectionCreateRequest struct {
	Title           string               `json:"title"`
	Published       *bool                `json:"published,omitzero"`
	PeriodType      LimitPeriodType      `json:"period_type,omitzero"`
	CalculationType LimitCalculationType `json:"calculation_type,omitzero"`
	LimitCount      *int                 `json:"limit_count,omitzero"`
	MinCount        *int                 `json:"min_count,omitzero"` // enterprise plans only
	ProductsOrder   ProductsOrder        `json:"products_order,omitzero"`
	ProductIDs      IDList               `json:"product_ids,omitzero"`
}

// LimitCollectionUpdateRequest is the body of UpdateLimit; every field is
// optional and unset fields are left unchanged.
type LimitCollectionUpdateRequest struct {
	Title           *string              `json:"title,omitzero"`
	Published       *bool                `json:"published,omitzero"`
	PeriodType      LimitPeriodType      `json:"period_type,omitzero"`
	CalculationType LimitCalculationType `json:"calculation_type,omitzero"`
	LimitCount      *int                 `json:"limit_count,omitzero"`
	MinCount        *int                 `json:"min_count,omitzero"`
	ProductsOrder   ProductsOrder        `json:"products_order,omitzero"`
}

// ListLimit returns one page of limit collections (GET /v1/limit_collections).
func (s *CollectionsService) ListLimit(ctx context.Context, opts *CollectionListOptions) (*Page[LimitCollection], error) {
	return collectionsList[LimitCollection](ctx, s.client, "v1/limit_collections", opts)
}

// AllLimit walks every page of limit collections (GET /v1/limit_collections).
func (s *CollectionsService) AllLimit(ctx context.Context, opts *CollectionListOptions) iter.Seq2[LimitCollection, error] {
	return collectionsAll[LimitCollection](ctx, s.client, "v1/limit_collections", opts)
}

// GetLimit returns one limit collection (GET /v1/limit_collections/{id}).
func (s *CollectionsService) GetLimit(ctx context.Context, id int64) (*LimitCollection, *Response, error) {
	return collectionsGet(ctx, s.client, fmt.Sprintf("v1/limit_collections/%d", id),
		func(c *LimitCollection) { c.ID = id })
}

// CreateLimit creates a limit collection (POST /v1/limit_collections).
func (s *CollectionsService) CreateLimit(ctx context.Context, req *LimitCollectionCreateRequest) (*LimitCollection, *Response, error) {
	return collectionsWrite[LimitCollection](ctx, s.client, http.MethodPost, "v1/limit_collections", req)
}

// UpdateLimit changes a limit collection (PUT /v1/limit_collections/{id}).
func (s *CollectionsService) UpdateLimit(ctx context.Context, id int64, req *LimitCollectionUpdateRequest) (*LimitCollection, *Response, error) {
	return collectionsWrite[LimitCollection](ctx, s.client, http.MethodPut, fmt.Sprintf("v1/limit_collections/%d", id), req)
}

// DeleteLimit removes a limit collection (DELETE /v1/limit_collections/{id}).
func (s *CollectionsService) DeleteLimit(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/limit_collections/%d", id), nil)
}

// AddLimitProducts adds products to a limit collection
// (POST /v1/limit_collections/{id}/products).
func (s *CollectionsService) AddLimitProducts(ctx context.Context, id int64, productIDs IDList) (*LimitCollection, *Response, error) {
	return collectionsWrite[LimitCollection](ctx, s.client, http.MethodPost,
		fmt.Sprintf("v1/limit_collections/%d/products", id), collectionsProductIDsBody{ProductIDs: productIDs})
}

// ReplaceLimitProducts replaces the whole product list of a limit
// collection (PUT /v1/limit_collections/{id}/products).
func (s *CollectionsService) ReplaceLimitProducts(ctx context.Context, id int64, productIDs IDList) (*LimitCollection, *Response, error) {
	return collectionsWrite[LimitCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/limit_collections/%d/products", id), collectionsProductIDsBody{ProductIDs: productIDs})
}

// RemoveLimitProducts removes products from a limit collection; the ids
// travel in the query string (DELETE /v1/limit_collections/{id}/products).
func (s *CollectionsService) RemoveLimitProducts(ctx context.Context, id int64, productIDs IDList) (*LimitCollection, *Response, error) {
	return collectionsDeleteWithQuery[LimitCollection](ctx, s.client,
		fmt.Sprintf("v1/limit_collections/%d/products", id), collectionsProductIDsQuery(productIDs))
}

// ReorderLimitProducts sets the manual product order of a limit collection
// (PUT /v1/limit_collections/{id}/products/manual_order).
func (s *CollectionsService) ReorderLimitProducts(ctx context.Context, id int64, productIDs IDList) (*LimitCollection, *Response, error) {
	return collectionsWrite[LimitCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/limit_collections/%d/products/manual_order", id), collectionsProductIDsBody{ProductIDs: productIDs})
}
