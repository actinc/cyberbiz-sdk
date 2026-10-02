package cyberbiz

import (
	"context"
	"fmt"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CustomerCouponType is the kind of discount a coupon grants.
type CustomerCouponType string

// Known CustomerCouponType values (coupon_type of a coupon).
const (
	CustomerCouponTypeAmount       CustomerCouponType = "amount"        // fixed amount off
	CustomerCouponTypePercent      CustomerCouponType = "percent"       // percentage off
	CustomerCouponTypeFreeShipping CustomerCouponType = "free_shipping" // waives shipping
	CustomerCouponTypeGift         CustomerCouponType = "gift"          // grants a gift item
)

// CustomerCouponRestrictStrategy says how a coupon combines with other
// campaigns; it is the alternative to the older concurrently_apply flag.
type CustomerCouponRestrictStrategy string

const (
	// CustomerCouponRestrictUnrestricted lets every campaign product use the coupon.
	CustomerCouponRestrictUnrestricted CustomerCouponRestrictStrategy = "unrestricted"
	// CustomerCouponRestrictRestrict excludes the listed campaigns' products.
	CustomerCouponRestrictRestrict CustomerCouponRestrictStrategy = "restrict"
	// CustomerCouponRestrictForbidden rejects the coupon when the order holds a listed campaign's product.
	CustomerCouponRestrictForbidden CustomerCouponRestrictStrategy = "forbidden"
)

// CustomerCoupon is a coupon as seen from a customer: either a personal
// coupon (GET /v1/customers/{id}/coupons) or a shop-wide coupon with the
// customer's own usage attached (GET /v1/customers/{id}/shop_coupons).
// CustomerID is zero for shop-wide coupons.
type CustomerCoupon struct {
	ID                  int64  `json:"id"`
	CustomerID          int64  `json:"customer_id"`
	Title               string `json:"title"`
	Code                string `json:"code"`
	CouponTypeName      string `json:"coupon_type_name"` // localised type, e.g. "金額"
	CouponValue         string `json:"coupon_value"`     // localised value, e.g. "5.0元"
	OrderPriceThreshold Money  `json:"order_price_threshold"`
	StartDate           Date   `json:"start_date"`
	EndDate             Date   `json:"end_date"`
	ConcurrentlyApply   bool   `json:"concurrently_apply"` // may combine with collection or shop-wide campaigns
	UsageLimit          int    `json:"usage_limit"`
	CanAccumulateBonus  bool   `json:"can_accumulate_bonus"`
	UsageUnlimited      bool   `json:"usage_unlimited"`
	UsedTimes           int    `json:"used_times"`
	GiftOrderID         int64  `json:"gift_order_id"` // the order that granted the coupon; zero when none
	GiftDays            int    `json:"gift_days"`     // days the gifted coupon stays usable
	// AccountUsageLimitEnabled caps how often each account may use the coupon.
	AccountUsageLimitEnabled bool                           `json:"account_usage_limit_enabled"`
	AccountUsageLimit        int                            `json:"account_usage_limit"`
	RestrictStrategy         CustomerCouponRestrictStrategy `json:"restrict_strategy"`
	RestrictCampaigns        []string                       `json:"restrict_campaigns"` // campaign codes, e.g. "vip_discount"
	Tags                     []string                       `json:"tags"`               // product tags the coupon is bound to
	PosShopIDs               []int64                        `json:"pos_shop_ids"`
	CouponStatus             CouponStatus                   `json:"coupon_status"`
	GiftOrderStatus          GiftOrderStatus                `json:"gift_order_status"` // empty when not granted by an order
	Valid                    bool                           `json:"valid"`
	CustomerUsedTimes        int                            `json:"customer_used_times"` // this customer's uses of a shop-wide coupon
	CustomerUsable           bool                           `json:"customer_usable"`     // this customer still has uses left
	ProductIDs               []int64                        `json:"product_ids"`
}

// Usable applies the documented CouponStatus / GiftOrderStatus combination
// rule; see CouponUsable.
func (c *CustomerCoupon) Usable() bool {
	return CouponUsable(c.CouponStatus, c.GiftOrderStatus)
}

// CustomerShopCoupon is a shop-wide coupon with the customer's usage; it has
// the same shape as CustomerCoupon.
type CustomerShopCoupon = CustomerCoupon

// CustomerCouponListOptions filter GET /v1/customers/{id}/coupons.
type CustomerCouponListOptions struct {
	ListOptions
	// OrderPriceThresholdLTEQ keeps coupons whose minimum order amount is at most this.
	OrderPriceThresholdLTEQ Money                `url:"order_price_threshold_lteq,omitempty"`
	CouponTypes             []CustomerCouponType `url:"coupon_types,omitempty,comma"`
	Tags                    []string             `url:"tags,omitempty,comma"`
	OnlyValid               *bool                `url:"only_valid,omitempty"`
}

// CustomerShopCouponListOptions filter GET /v1/customers/{id}/shop_coupons.
type CustomerShopCouponListOptions struct {
	ListOptions
	OnlyValid *bool `url:"only_valid,omitempty"`
}

// CustomerCouponCreateRequest is the body of POST /v1/customers/{id}/coupons.
// Use either ConcurrentlyApply or RestrictStrategy, not both.
type CustomerCouponCreateRequest struct {
	Title               string                         `json:"title"`
	Code                string                         `json:"code"`
	CouponType          CustomerCouponType             `json:"coupon_type"`
	Value               Money                          `json:"value"` // amount or percent, per CouponType
	OrderPriceThreshold Money                          `json:"order_price_threshold"`
	UsageLimit          int                            `json:"usage_limit"`
	StartDate           Date                           `json:"start_date,omitzero"`
	EndDate             Date                           `json:"end_date,omitzero"`
	ConcurrentlyApply   *bool                          `json:"concurrently_apply,omitzero"`
	RestrictStrategy    CustomerCouponRestrictStrategy `json:"restrict_strategy,omitzero"`
	RestrictCampaigns   []string                       `json:"restrict_campaigns,omitzero"`
	Tags                []string                       `json:"tags,omitzero"`
	ProductIDs          []int64                        `json:"product_ids,omitzero"`
}

// CustomerCouponUpdateRequest is the body of
// PUT /v1/customers/{id}/coupons/{coupon_id}.
type CustomerCouponUpdateRequest struct {
	Title               *string                        `json:"title,omitzero"`
	Code                *string                        `json:"code,omitzero"`
	CouponType          CustomerCouponType             `json:"coupon_type,omitzero"`
	Value               *Money                         `json:"value,omitzero"`
	OrderPriceThreshold *Money                         `json:"order_price_threshold,omitzero"`
	StartDate           Date                           `json:"start_date,omitzero"`
	EndDate             Date                           `json:"end_date,omitzero"`
	UsageLimit          *int                           `json:"usage_limit,omitzero"`
	UsedTimes           *int                           `json:"used_times,omitzero"`
	ConcurrentlyApply   *bool                          `json:"concurrently_apply,omitzero"`
	RestrictStrategy    CustomerCouponRestrictStrategy `json:"restrict_strategy,omitzero"`
	RestrictCampaigns   []string                       `json:"restrict_campaigns,omitzero"`
	Tags                []string                       `json:"tags,omitzero"`
}

// ListCoupons returns one page of the customer's personal coupons
// (GET /v1/customers/{id}/coupons).
func (s *CustomersService) ListCoupons(ctx context.Context, id int64, opts *CustomerCouponListOptions) (*Page[CustomerCoupon], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerCoupon](ctx, s.client, fmt.Sprintf("v1/customers/%d/coupons", id), q)
}

// GetCoupon returns one personal coupon
// (GET /v1/customers/{id}/coupons/{coupon_id}).
func (s *CustomersService) GetCoupon(ctx context.Context, customerID, couponID int64) (*CustomerCoupon, *Response, error) {
	var out CustomerCoupon
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d/coupons/%d", customerID, couponID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ID == 0 {
		out.ID = couponID
	}
	return &out, resp, nil
}

// CreateCoupon grants the customer a personal coupon
// (POST /v1/customers/{id}/coupons).
func (s *CustomersService) CreateCoupon(ctx context.Context, id int64, req *CustomerCouponCreateRequest) (*CustomerCoupon, *Response, error) {
	var out CustomerCoupon
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/customers/%d/coupons", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateCoupon changes a personal coupon
// (PUT /v1/customers/{id}/coupons/{coupon_id}).
func (s *CustomersService) UpdateCoupon(ctx context.Context, customerID, couponID int64, req *CustomerCouponUpdateRequest) (*CustomerCoupon, *Response, error) {
	var out CustomerCoupon
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/customers/%d/coupons/%d", customerID, couponID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ID == 0 {
		out.ID = couponID
	}
	return &out, resp, nil
}

// DeleteCoupon removes a personal coupon
// (DELETE /v1/customers/{id}/coupons/{coupon_id}).
func (s *CustomersService) DeleteCoupon(ctx context.Context, customerID, couponID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/customers/%d/coupons/%d", customerID, couponID), nil)
}

// ListShopCoupons returns one page of shop-wide coupons with the customer's
// usage (GET /v1/customers/{id}/shop_coupons).
func (s *CustomersService) ListShopCoupons(ctx context.Context, id int64, opts *CustomerShopCouponListOptions) (*Page[CustomerShopCoupon], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerShopCoupon](ctx, s.client, fmt.Sprintf("v1/customers/%d/shop_coupons", id), q)
}

// GetShopCoupon returns one shop-wide coupon with the customer's usage
// (GET /v1/customers/{id}/shop_coupons/{coupon_id}).
func (s *CustomersService) GetShopCoupon(ctx context.Context, customerID, couponID int64) (*CustomerShopCoupon, *Response, error) {
	var out CustomerShopCoupon
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d/shop_coupons/%d", customerID, couponID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ID == 0 {
		out.ID = couponID
	}
	return &out, resp, nil
}

// DecrementShopCoupon records one use of a shop-wide coupon by the customer
// (POST /v1/customers/{id}/shop_coupons/{coupon_id}/decrement).
func (s *CustomersService) DecrementShopCoupon(ctx context.Context, customerID, couponID int64) (*CustomerShopCoupon, *Response, error) {
	var out CustomerShopCoupon
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/customers/%d/shop_coupons/%d/decrement", customerID, couponID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ID == 0 {
		out.ID = couponID
	}
	return &out, resp, nil
}
