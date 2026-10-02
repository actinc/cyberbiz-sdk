package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// CustomCollectionCreateRequest is the body of CreateCustom. Title, Handle
// and Published are required.
type CustomCollectionCreateRequest struct {
	Title         string        `json:"title"`
	Handle        string        `json:"handle"` // storefront path /collections/{handle}
	Published     bool          `json:"published"`
	BodyHTML      *string       `json:"body_html,omitzero"`
	ProductsOrder ProductsOrder `json:"products_order,omitzero"`
}

// CustomCollectionUpdateRequest is the body of UpdateCustom; every field is
// optional and unset fields are left unchanged.
type CustomCollectionUpdateRequest struct {
	Title         *string       `json:"title,omitzero"`
	Handle        *string       `json:"handle,omitzero"`
	Published     *bool         `json:"published,omitzero"`
	BodyHTML      *string       `json:"body_html,omitzero"`
	ProductsOrder ProductsOrder `json:"products_order,omitzero"`
}

// ListCustom returns one page of custom collections; list items carry no
// products (GET /v1/custom_collections).
func (s *CollectionsService) ListCustom(ctx context.Context, opts *CollectionListOptions) (*Page[CustomCollection], error) {
	return collectionsList[CustomCollection](ctx, s.client, "v1/custom_collections", opts)
}

// AllCustom walks every page of custom collections (GET /v1/custom_collections).
func (s *CollectionsService) AllCustom(ctx context.Context, opts *CollectionListOptions) iter.Seq2[CustomCollection, error] {
	return collectionsAll[CustomCollection](ctx, s.client, "v1/custom_collections", opts)
}

// ListCustomV2 returns one page of custom collections matched by title or
// handle; items carry no products (GET /v2/custom_collections).
func (s *CollectionsService) ListCustomV2(ctx context.Context, opts *CollectionSearchOptions) (*Page[CustomCollection], error) {
	return collectionsList[CustomCollection](ctx, s.client, "v2/custom_collections", opts)
}

// AllCustomV2 walks every page of the v2 custom collection search
// (GET /v2/custom_collections).
func (s *CollectionsService) AllCustomV2(ctx context.Context, opts *CollectionSearchOptions) iter.Seq2[CustomCollection, error] {
	return collectionsAll[CustomCollection](ctx, s.client, "v2/custom_collections", opts)
}

// GetCustom returns one custom collection with its products
// (GET /v1/custom_collections/{id}).
func (s *CollectionsService) GetCustom(ctx context.Context, id int64) (*CustomCollection, *Response, error) {
	return collectionsGet(ctx, s.client, fmt.Sprintf("v1/custom_collections/%d", id),
		func(c *CustomCollection) { c.ID = id })
}

// CreateCustom creates a custom collection (POST /v1/custom_collections).
func (s *CollectionsService) CreateCustom(ctx context.Context, req *CustomCollectionCreateRequest) (*CustomCollection, *Response, error) {
	return collectionsWrite[CustomCollection](ctx, s.client, http.MethodPost, "v1/custom_collections", req)
}

// UpdateCustom changes a custom collection (PUT /v1/custom_collections/{id}).
func (s *CollectionsService) UpdateCustom(ctx context.Context, id int64, req *CustomCollectionUpdateRequest) (*CustomCollection, *Response, error) {
	return collectionsWrite[CustomCollection](ctx, s.client, http.MethodPut, fmt.Sprintf("v1/custom_collections/%d", id), req)
}

// DeleteCustom removes a custom collection (DELETE /v1/custom_collections/{id}).
func (s *CollectionsService) DeleteCustom(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/custom_collections/%d", id), nil)
}

// AddCustomProducts adds products to a custom collection
// (POST /v1/custom_collections/{id}/products).
func (s *CollectionsService) AddCustomProducts(ctx context.Context, id int64, productIDs IDList) (*CustomCollection, *Response, error) {
	return collectionsWrite[CustomCollection](ctx, s.client, http.MethodPost,
		fmt.Sprintf("v1/custom_collections/%d/products", id), collectionsProductIDsBody{ProductIDs: productIDs})
}

// ReplaceCustomProducts replaces the whole product list of a custom
// collection (PUT /v1/custom_collections/{id}/products).
func (s *CollectionsService) ReplaceCustomProducts(ctx context.Context, id int64, productIDs IDList) (*CustomCollection, *Response, error) {
	return collectionsWrite[CustomCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/custom_collections/%d/products", id), collectionsProductIDsBody{ProductIDs: productIDs})
}

// RemoveCustomProducts removes products from a custom collection; the ids
// travel in the query string (DELETE /v1/custom_collections/{id}/products).
func (s *CollectionsService) RemoveCustomProducts(ctx context.Context, id int64, productIDs IDList) (*CustomCollection, *Response, error) {
	return collectionsDeleteWithQuery[CustomCollection](ctx, s.client,
		fmt.Sprintf("v1/custom_collections/%d/products", id), collectionsProductIDsQuery(productIDs))
}

// ReorderCustomProducts sets the manual product order of a custom
// collection to the given sequence
// (PUT /v1/custom_collections/{id}/products/manual_order).
func (s *CollectionsService) ReorderCustomProducts(ctx context.Context, id int64, productIDs IDList) (*CustomCollection, *Response, error) {
	return collectionsWrite[CustomCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/custom_collections/%d/products/manual_order", id), collectionsProductIDsBody{ProductIDs: productIDs})
}
