package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// SpecialCollectionCreateRequest is the body of CreateSpecial. Title,
// Handle, Published, SpecialCollectionType and RestIncludeDiscount are
// required; quantity tiers are added afterwards with CreateSpecialRule.
type SpecialCollectionCreateRequest struct {
	Title                 string                `json:"title"`
	Handle                string                `json:"handle"` // storefront path /collections/{handle}
	Published             bool                  `json:"published"`
	StartDate             Time                  `json:"start_date,omitzero"`
	EndDate               Time                  `json:"end_date,omitzero"`
	BodyHTML              *string               `json:"body_html,omitzero"`
	SpecialCollectionType SpecialCollectionType `json:"special_collection_type"`
	// RestIncludeDiscount is whether items beyond the tier quantity are
	// still discounted.
	RestIncludeDiscount bool `json:"rest_include_discount"`
}

// SpecialCollectionUpdateRequest is the body of UpdateSpecial; every field
// is optional and unset fields are left unchanged.
type SpecialCollectionUpdateRequest struct {
	Title                 *string               `json:"title,omitzero"`
	Handle                *string               `json:"handle,omitzero"`
	Published             *bool                 `json:"published,omitzero"`
	StartDate             Time                  `json:"start_date,omitzero"`
	EndDate               Time                  `json:"end_date,omitzero"`
	BodyHTML              *string               `json:"body_html,omitzero"`
	SpecialCollectionType SpecialCollectionType `json:"special_collection_type,omitzero"`
	RestIncludeDiscount   *bool                 `json:"rest_include_discount,omitzero"`
}

// SpecialCollectionRuleRequest is the body of CreateSpecialRule and
// UpdateSpecialRule: buying Quantity items costs Price (amount, discount and
// per_discount types) or Percentage of the list price (percentage type).
// Quantity is required on create; on update an unset field is unchanged.
type SpecialCollectionRuleRequest struct {
	Quantity   *int     `json:"quantity,omitzero"`
	Price      *Money   `json:"price,omitzero"`
	Percentage *float64 `json:"percentage,omitzero"`
}

// ListSpecial returns one page of special collections (GET /v1/special_collections).
func (s *CollectionsService) ListSpecial(ctx context.Context, opts *CollectionListOptions) (*Page[SpecialCollection], error) {
	return collectionsList[SpecialCollection](ctx, s.client, "v1/special_collections", opts)
}

// AllSpecial walks every page of special collections (GET /v1/special_collections).
func (s *CollectionsService) AllSpecial(ctx context.Context, opts *CollectionListOptions) iter.Seq2[SpecialCollection, error] {
	return collectionsAll[SpecialCollection](ctx, s.client, "v1/special_collections", opts)
}

// GetSpecial returns one special collection (GET /v1/special_collections/{id}).
func (s *CollectionsService) GetSpecial(ctx context.Context, id int64) (*SpecialCollection, *Response, error) {
	return collectionsGet(ctx, s.client, fmt.Sprintf("v1/special_collections/%d", id),
		func(c *SpecialCollection) { c.ID = id })
}

// CreateSpecial creates a special collection (POST /v1/special_collections).
func (s *CollectionsService) CreateSpecial(ctx context.Context, req *SpecialCollectionCreateRequest) (*SpecialCollection, *Response, error) {
	return collectionsWrite[SpecialCollection](ctx, s.client, http.MethodPost, "v1/special_collections", req)
}

// UpdateSpecial changes a special collection (PUT /v1/special_collections/{id}).
func (s *CollectionsService) UpdateSpecial(ctx context.Context, id int64, req *SpecialCollectionUpdateRequest) (*SpecialCollection, *Response, error) {
	return collectionsWrite[SpecialCollection](ctx, s.client, http.MethodPut, fmt.Sprintf("v1/special_collections/%d", id), req)
}

// DeleteSpecial removes a special collection (DELETE /v1/special_collections/{id}).
func (s *CollectionsService) DeleteSpecial(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/special_collections/%d", id), nil)
}

// AddSpecialProducts adds products to a special collection
// (POST /v1/special_collections/{id}/products).
func (s *CollectionsService) AddSpecialProducts(ctx context.Context, id int64, productIDs IDList) (*SpecialCollection, *Response, error) {
	return collectionsWrite[SpecialCollection](ctx, s.client, http.MethodPost,
		fmt.Sprintf("v1/special_collections/%d/products", id), collectionsProductIDsBody{ProductIDs: productIDs})
}

// RemoveSpecialProducts removes products from a special collection; the ids
// travel in the query string (DELETE /v1/special_collections/{id}/products).
func (s *CollectionsService) RemoveSpecialProducts(ctx context.Context, id int64, productIDs IDList) (*SpecialCollection, *Response, error) {
	return collectionsDeleteWithQuery[SpecialCollection](ctx, s.client,
		fmt.Sprintf("v1/special_collections/%d/products", id), collectionsProductIDsQuery(productIDs))
}

// CreateSpecialRule adds a quantity tier to a special collection and
// returns the updated collection (POST /v1/special_collections/{id}/rules).
func (s *CollectionsService) CreateSpecialRule(ctx context.Context, id int64, req *SpecialCollectionRuleRequest) (*SpecialCollection, *Response, error) {
	return collectionsWrite[SpecialCollection](ctx, s.client, http.MethodPost,
		fmt.Sprintf("v1/special_collections/%d/rules", id), req)
}

// UpdateSpecialRule changes a quantity tier and returns the updated
// collection (PUT /v1/special_collections/{id}/rules/{rule_id}).
func (s *CollectionsService) UpdateSpecialRule(ctx context.Context, id, ruleID int64, req *SpecialCollectionRuleRequest) (*SpecialCollection, *Response, error) {
	return collectionsWrite[SpecialCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/special_collections/%d/rules/%d", id, ruleID), req)
}

// DeleteSpecialRule removes a quantity tier and returns the updated
// collection (DELETE /v1/special_collections/{id}/rules/{rule_id}).
func (s *CollectionsService) DeleteSpecialRule(ctx context.Context, id, ruleID int64) (*SpecialCollection, *Response, error) {
	var out SpecialCollection
	resp, err := s.client.delete(ctx, fmt.Sprintf("v1/special_collections/%d/rules/%d", id, ruleID), &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
