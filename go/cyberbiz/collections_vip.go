package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// VIP collections require a feature licence; the test shop answers every
// endpoint with 401 "無權使用該 API", so nothing here is verified against a
// Golden File. Shapes follow the swagger and Postman examples.

// VIPCollectionCreateRequest is the body of CreateVIP. Both fields are
// required; the rule and promotion are set afterwards with UpdateVIPRule
// and UpdateVIPPromotion.
type VIPCollectionCreateRequest struct {
	Title    string `json:"title"`
	Position int    `json:"position"` // evaluation priority
}

// VIPCollectionUpdateRequest is the body of UpdateVIP; every field is
// optional and unset fields are left unchanged.
type VIPCollectionUpdateRequest struct {
	Title    *string `json:"title,omitzero"`
	Position *int    `json:"position,omitzero"`
}

// VIPRuleRequest is the body of UpdateVIPRule. VIPRuleSpec is required;
// the swagger marks TotalOrderCount, TotalPrice and CustomersTag required
// too, though only the one matching the spec is meaningful.
type VIPRuleRequest struct {
	VIPRuleSpec        VIPRuleType `json:"vip_rule_spec"`
	VIPExpiredateStart Date        `json:"vip_expiredate_start,omitzero"`
	VIPExpiredateEnd   Date        `json:"vip_expiredate_end,omitzero"`
	OrderStart         Date        `json:"order_start,omitzero"`
	OrderEnd           Date        `json:"order_end,omitzero"`
	TotalOrderCount    *int        `json:"total_order_count,omitzero"`
	TotalPrice         *Money      `json:"total_price,omitzero"`
	CustomersTag       *string     `json:"customers_tag,omitzero"`
}

// VIPPromotionRequest is the body of UpdateVIPPromotion. VIPPromotionSpec
// is required; Discount is the order discount in "折" units (9 means 10%
// off) for the discount type.
type VIPPromotionRequest struct {
	VIPPromotionSpec  VIPPromotionType `json:"vip_promotion_spec"`
	Discount          *float64         `json:"discount,omitzero"`
	ConcurrentlyApply *bool            `json:"concurrently_apply,omitzero"`
}

// ListVIP returns one page of VIP collections (GET /v1/vip_collections).
func (s *CollectionsService) ListVIP(ctx context.Context, opts *CollectionListOptions) (*Page[VIPCollection], error) {
	return collectionsList[VIPCollection](ctx, s.client, "v1/vip_collections", opts)
}

// AllVIP walks every page of VIP collections (GET /v1/vip_collections).
func (s *CollectionsService) AllVIP(ctx context.Context, opts *CollectionListOptions) iter.Seq2[VIPCollection, error] {
	return collectionsAll[VIPCollection](ctx, s.client, "v1/vip_collections", opts)
}

// GetVIP returns one VIP collection (GET /v1/vip_collections/{id}).
func (s *CollectionsService) GetVIP(ctx context.Context, id int64) (*VIPCollection, *Response, error) {
	return collectionsGet(ctx, s.client, fmt.Sprintf("v1/vip_collections/%d", id),
		func(c *VIPCollection) { c.ID = id })
}

// CreateVIP creates a VIP collection (POST /v1/vip_collections).
func (s *CollectionsService) CreateVIP(ctx context.Context, req *VIPCollectionCreateRequest) (*VIPCollection, *Response, error) {
	return collectionsWrite[VIPCollection](ctx, s.client, http.MethodPost, "v1/vip_collections", req)
}

// UpdateVIP changes a VIP collection's title or priority
// (PUT /v1/vip_collections/{id}).
func (s *CollectionsService) UpdateVIP(ctx context.Context, id int64, req *VIPCollectionUpdateRequest) (*VIPCollection, *Response, error) {
	return collectionsWrite[VIPCollection](ctx, s.client, http.MethodPut, fmt.Sprintf("v1/vip_collections/%d", id), req)
}

// DeleteVIP removes a VIP collection (DELETE /v1/vip_collections/{id}).
func (s *CollectionsService) DeleteVIP(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/vip_collections/%d", id), nil)
}

// UpdateVIPRule sets the qualifying condition of a VIP collection and
// returns the stored rule (PUT /v1/vip_collections/{id}/update_rule).
func (s *CollectionsService) UpdateVIPRule(ctx context.Context, id int64, req *VIPRuleRequest) (*VIPCollectionRule, *Response, error) {
	return collectionsWrite[VIPCollectionRule](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/vip_collections/%d/update_rule", id), req)
}

// UpdateVIPPromotion sets the benefit of a VIP collection and returns the
// stored promotion (PUT /v1/vip_collections/{id}/update_promotion).
func (s *CollectionsService) UpdateVIPPromotion(ctx context.Context, id int64, req *VIPPromotionRequest) (*VIPCollectionPromotion, *Response, error) {
	return collectionsWrite[VIPCollectionPromotion](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/vip_collections/%d/update_promotion", id), req)
}
