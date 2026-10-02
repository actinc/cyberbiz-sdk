package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// DiscountsService exposes shop-wide discounts (/v1/discounts) and shop-wide
// coupons (/v1/shop_coupons).
type DiscountsService struct {
	client *Client
}

// DiscountCreateRequest is the body of Create.
type DiscountCreateRequest struct {
	Name         string       `json:"name"`
	DiscountType DiscountType `json:"discount_type"`
	// Value is the amount (TWD) for amount discounts or the percentage for
	// percent discounts.
	Value Money `json:"value"`
	// Threshold is the minimum order total the discount applies from.
	Threshold Money `json:"threshold"`
	StartAt   Time  `json:"start_at,omitzero"`
	EndAt     Time  `json:"end_at,omitzero"`
	// Days is how many days a coupon-type discount stays usable.
	Days              *int         `json:"days,omitzero"`
	ConcurrentlyApply *bool        `json:"concurrently_apply,omitzero"`
	SalesChannelID    SalesChannel `json:"sales_channel_id,omitzero"`
}

// DiscountUpdateRequest is the body of Update; every field is optional.
type DiscountUpdateRequest struct {
	Name              *string      `json:"name,omitzero"`
	DiscountType      DiscountType `json:"discount_type,omitzero"`
	Value             *Money       `json:"value,omitzero"`
	Threshold         *Money       `json:"threshold,omitzero"`
	StartAt           Time         `json:"start_at,omitzero"`
	EndAt             Time         `json:"end_at,omitzero"`
	Days              *int         `json:"days,omitzero"`
	ConcurrentlyApply *bool        `json:"concurrently_apply,omitzero"`
	SalesChannelID    SalesChannel `json:"sales_channel_id,omitzero"`
}

// RegisterCouponRuleRequest is the body of UpdateRegisterCouponRule. Every
// field is required by the platform.
type RegisterCouponRuleRequest struct {
	Enabled bool `json:"enabled"`
	// Value is the coupon amount.
	Value               Money `json:"value"`
	OrderPriceThreshold Money `json:"order_price_threshold"`
	UsageLimit          int   `json:"usage_limit"`
	// ExpireDay is the number of days the coupon stays usable.
	ExpireDay         int  `json:"expire_day"`
	ConcurrentlyApply bool `json:"concurrently_apply"`
}

// List returns one page of discounts (GET /v1/discounts).
func (s *DiscountsService) List(ctx context.Context, opts *ListOptions) (*Page[Discount], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Discount](ctx, s.client, "v1/discounts", q)
}

// All walks every page of discounts (GET /v1/discounts).
func (s *DiscountsService) All(ctx context.Context, opts *ListOptions) iter.Seq2[Discount, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Discount, error) bool) { yield(Discount{}, err) }
	}
	return listAll[Discount](ctx, s.client, "v1/discounts", q)
}

// Get returns one discount (GET /v1/discounts/{id}).
func (s *DiscountsService) Get(ctx context.Context, id int64) (*Discount, *Response, error) {
	var out Discount
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/discounts/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Create creates a discount (POST /v1/discounts).
func (s *DiscountsService) Create(ctx context.Context, req *DiscountCreateRequest) (*Discount, *Response, error) {
	var out Discount
	resp, err := s.client.post(ctx, "v1/discounts", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Update changes a discount (PUT /v1/discounts/{id}).
func (s *DiscountsService) Update(ctx context.Context, id int64, req *DiscountUpdateRequest) (*Discount, *Response, error) {
	var out Discount
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/discounts/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Delete removes a discount (DELETE /v1/discounts/{id}).
func (s *DiscountsService) Delete(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/discounts/%d", id), nil)
}

// UpdateRegisterCouponRule sets the coupon granted on customer registration
// (PUT /v1/discounts/register_coupon_rule).
func (s *DiscountsService) UpdateRegisterCouponRule(ctx context.Context, req *RegisterCouponRuleRequest) (*RegisterCouponRule, *Response, error) {
	var out RegisterCouponRule
	resp, err := s.client.put(ctx, "v1/discounts/register_coupon_rule", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ShopCouponListOptions filters ListCoupons and AllCoupons.
type ShopCouponListOptions struct {
	ListOptions
	// OrderPriceThresholdLTE keeps coupons whose threshold is at most this.
	OrderPriceThresholdLTE Money `url:"order_price_threshold_lteq,omitempty"`
	// CouponTypes filters by coupon type, comma-separated by the platform.
	CouponTypes []CouponType `url:"coupon_types,omitempty,comma"`
	// Tags filters by bound product tag, comma-separated by the platform.
	Tags      []string `url:"tags,omitempty,comma"`
	OnlyValid *bool    `url:"only_valid,omitempty"`
}

// ShopCouponCreateRequest is the body of CreateCoupon.
type ShopCouponCreateRequest struct {
	Title      string     `json:"title"`
	Code       string     `json:"code"`
	CouponType CouponType `json:"coupon_type"`
	// Value is the amount (TWD) or percentage, depending on CouponType.
	Value               Money `json:"value"`
	OrderPriceThreshold Money `json:"order_price_threshold"`
	StartDate           Date  `json:"start_date,omitzero"`
	EndDate             Date  `json:"end_date,omitzero"`
	ConcurrentlyApply   bool  `json:"concurrently_apply"`
	UsageLimit          int   `json:"usage_limit"`
	// AccountUsageLimitEnabled turns on the per-account usage limit.
	AccountUsageLimitEnabled *bool              `json:"account_usage_limit_enabled,omitzero"`
	AccountUsageLimit        *int               `json:"account_usage_limit,omitzero"`
	RestrictStrategy         RestrictStrategy   `json:"restrict_strategy,omitzero"`
	RestrictCampaigns        []RestrictCampaign `json:"restrict_campaigns,omitzero"`
	Tags                     []string           `json:"tags,omitzero"`
	ProductIDs               []int64            `json:"product_ids,omitzero"`
}

// ShopCouponUpdateRequest is the body of UpdateCoupon; every field is
// optional.
type ShopCouponUpdateRequest struct {
	Title                    *string            `json:"title,omitzero"`
	Code                     *string            `json:"code,omitzero"`
	CouponType               CouponType         `json:"coupon_type,omitzero"`
	Value                    *Money             `json:"value,omitzero"`
	OrderPriceThreshold      *Money             `json:"order_price_threshold,omitzero"`
	StartDate                Date               `json:"start_date,omitzero"`
	EndDate                  Date               `json:"end_date,omitzero"`
	ConcurrentlyApply        *bool              `json:"concurrently_apply,omitzero"`
	UsageLimit               *int               `json:"usage_limit,omitzero"`
	AccountUsageLimitEnabled *bool              `json:"account_usage_limit_enabled,omitzero"`
	AccountUsageLimit        *int               `json:"account_usage_limit,omitzero"`
	UsedTimes                *int               `json:"used_times,omitzero"`
	RestrictStrategy         RestrictStrategy   `json:"restrict_strategy,omitzero"`
	RestrictCampaigns        []RestrictCampaign `json:"restrict_campaigns,omitzero"`
	Tags                     []string           `json:"tags,omitzero"`
}

// ListCoupons returns one page of shop coupons (GET /v1/shop_coupons).
func (s *DiscountsService) ListCoupons(ctx context.Context, opts *ShopCouponListOptions) (*Page[ShopCoupon], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[ShopCoupon](ctx, s.client, "v1/shop_coupons", q)
}

// AllCoupons walks every page of shop coupons (GET /v1/shop_coupons).
func (s *DiscountsService) AllCoupons(ctx context.Context, opts *ShopCouponListOptions) iter.Seq2[ShopCoupon, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(ShopCoupon, error) bool) { yield(ShopCoupon{}, err) }
	}
	return listAll[ShopCoupon](ctx, s.client, "v1/shop_coupons", q)
}

// GetCoupon returns one shop coupon (GET /v1/shop_coupons/{id}).
func (s *DiscountsService) GetCoupon(ctx context.Context, id int64) (*ShopCoupon, *Response, error) {
	var out ShopCoupon
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/shop_coupons/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// CreateCoupon creates a shop coupon (POST /v1/shop_coupons).
func (s *DiscountsService) CreateCoupon(ctx context.Context, req *ShopCouponCreateRequest) (*ShopCoupon, *Response, error) {
	var out ShopCoupon
	resp, err := s.client.post(ctx, "v1/shop_coupons", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateCoupon changes a shop coupon (PUT /v1/shop_coupons/{id}).
func (s *DiscountsService) UpdateCoupon(ctx context.Context, id int64, req *ShopCouponUpdateRequest) (*ShopCoupon, *Response, error) {
	var out ShopCoupon
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/shop_coupons/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// DeleteCoupon removes a shop coupon (DELETE /v1/shop_coupons/{id}).
func (s *DiscountsService) DeleteCoupon(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/shop_coupons/%d", id), nil)
}
