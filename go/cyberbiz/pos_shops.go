package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// PosShopsService exposes POS shops and their terminals, products and
// variants (/v1/pos_shops), POS shop coupons (/v1/pos_shop_coupons, which
// need the pos_shop_coupon plugin or every call is a 403), and stored-value
// wallets (/v2/pos_wallets, which need the POS 儲值金 feature).
type PosShopsService struct {
	client *Client
}

// PosShopUpdateRequest is the body of Update.
type PosShopUpdateRequest struct {
	Name     *string `json:"name,omitzero"`
	Phone    *string `json:"phone,omitzero"`
	County   *string `json:"county,omitzero"`
	District *string `json:"district,omitzero"`
	Address  *string `json:"address,omitzero"`
	// VATNumber is the company tax id (統一編號).
	VATNumber *string `json:"VAT_number,omitzero"`
	// AESKey is the encryption key for the shop's invoice device.
	AESKey             *string `json:"aes_key,omitzero"`
	CanFindOthersOrder *bool   `json:"can_find_others_order,omitzero"`
	CanAccessCustomers *bool   `json:"can_access_customers,omitzero"`
}

// PosUpdateRequest is the body of UpdatePos.
type PosUpdateRequest struct {
	Name                 *string `json:"name,omitzero"`
	PettyCash            *Money  `json:"petty_cash,omitzero"`
	AdminPasswordEnabled *bool   `json:"admin_password_enabled,omitzero"`
}

// PosShopCouponCreateRequest is the body of CreateCoupon.
type PosShopCouponCreateRequest struct {
	Title      string            `json:"title"`
	Code       string            `json:"code"`
	CouponType PosShopCouponType `json:"coupon_type"`
	// Value is the discount: an amount for amount coupons, a percentage
	// (10 means 10%) for percent coupons.
	Value               Money `json:"value"`
	OrderPriceThreshold Money `json:"order_price_threshold"`
	StartDate           Date  `json:"start_date,omitzero"`
	EndDate             Date  `json:"end_date,omitzero"`
	// ConcurrentlyApply allows use together with bundle discounts or shop
	// campaigns.
	ConcurrentlyApply  bool  `json:"concurrently_apply"`
	UsageLimit         int   `json:"usage_limit"`
	CanAccumulateBonus *bool `json:"can_accumulate_bonus,omitzero"`
	UsageUnlimited     *bool `json:"usage_unlimited,omitzero"`
	// PosShopIDs are the shops that accept the coupon.
	PosShopIDs []int64 `json:"pos_shop_ids,omitzero"`
}

// PosShopCouponUpdateRequest is the body of UpdateCoupon.
type PosShopCouponUpdateRequest struct {
	Title               *string           `json:"title,omitzero"`
	Code                *string           `json:"code,omitzero"`
	CouponType          PosShopCouponType `json:"coupon_type,omitzero"`
	Value               *Money            `json:"value,omitzero"`
	OrderPriceThreshold *Money            `json:"order_price_threshold,omitzero"`
	StartDate           Date              `json:"start_date,omitzero"`
	EndDate             Date              `json:"end_date,omitzero"`
	ConcurrentlyApply   *bool             `json:"concurrently_apply,omitzero"`
	UsageLimit          *int              `json:"usage_limit,omitzero"`
	PosShopIDs          []int64           `json:"pos_shop_ids,omitzero"`
}

// List returns one page of POS shops (GET /v1/pos_shops).
func (s *PosShopsService) List(ctx context.Context, opts *ListOptions) (*Page[PosShop], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[PosShop](ctx, s.client, "v1/pos_shops", q)
}

// All walks every page of POS shops (GET /v1/pos_shops).
func (s *PosShopsService) All(ctx context.Context, opts *ListOptions) iter.Seq2[PosShop, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(PosShop, error) bool) { yield(PosShop{}, err) }
	}
	return listAll[PosShop](ctx, s.client, "v1/pos_shops", q)
}

// Get returns one POS shop (GET /v1/pos_shops/{id}).
func (s *PosShopsService) Get(ctx context.Context, id int64) (*PosShop, *Response, error) {
	var out PosShop
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/pos_shops/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Update changes a POS shop (PUT /v1/pos_shops/{id}).
func (s *PosShopsService) Update(ctx context.Context, id int64, req *PosShopUpdateRequest) (*PosShop, *Response, error) {
	var out PosShop
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/pos_shops/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ListPoses returns one page of a shop's POS terminals
// (GET /v1/pos_shops/{id}/poses).
func (s *PosShopsService) ListPoses(ctx context.Context, shopID int64, opts *ListOptions) (*Page[Pos], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Pos](ctx, s.client, fmt.Sprintf("v1/pos_shops/%d/poses", shopID), q)
}

// AllPoses walks every page of a shop's POS terminals
// (GET /v1/pos_shops/{id}/poses).
func (s *PosShopsService) AllPoses(ctx context.Context, shopID int64, opts *ListOptions) iter.Seq2[Pos, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Pos, error) bool) { yield(Pos{}, err) }
	}
	return listAll[Pos](ctx, s.client, fmt.Sprintf("v1/pos_shops/%d/poses", shopID), q)
}

// GetPos returns one POS terminal (GET /v1/pos_shops/{id}/poses/{pos_id}).
func (s *PosShopsService) GetPos(ctx context.Context, shopID, posID int64) (*Pos, *Response, error) {
	var out Pos
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/pos_shops/%d/poses/%d", shopID, posID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdatePos changes a POS terminal (PUT /v1/pos_shops/{id}/poses/{pos_id}).
func (s *PosShopsService) UpdatePos(ctx context.Context, shopID, posID int64, req *PosUpdateRequest) (*Pos, *Response, error) {
	var out Pos
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/pos_shops/%d/poses/%d", shopID, posID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ListProducts returns one page of the products sold at a POS shop
// (GET /v1/pos_shops/{id}/products).
func (s *PosShopsService) ListProducts(ctx context.Context, shopID int64, opts *ListOptions) (*Page[PosProduct], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[PosProduct](ctx, s.client, fmt.Sprintf("v1/pos_shops/%d/products", shopID), q)
}

// AllProducts walks every page of the products sold at a POS shop
// (GET /v1/pos_shops/{id}/products).
func (s *PosShopsService) AllProducts(ctx context.Context, shopID int64, opts *ListOptions) iter.Seq2[PosProduct, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(PosProduct, error) bool) { yield(PosProduct{}, err) }
	}
	return listAll[PosProduct](ctx, s.client, fmt.Sprintf("v1/pos_shops/%d/products", shopID), q)
}

// ListProductVariants returns one page of the variants sold at a POS shop
// (GET /v1/pos_shops/{id}/product_variants).
func (s *PosShopsService) ListProductVariants(ctx context.Context, shopID int64, opts *ListOptions) (*Page[PosProductVariant], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[PosProductVariant](ctx, s.client, fmt.Sprintf("v1/pos_shops/%d/product_variants", shopID), q)
}

// AllProductVariants walks every page of the variants sold at a POS shop
// (GET /v1/pos_shops/{id}/product_variants).
func (s *PosShopsService) AllProductVariants(ctx context.Context, shopID int64, opts *ListOptions) iter.Seq2[PosProductVariant, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(PosProductVariant, error) bool) { yield(PosProductVariant{}, err) }
	}
	return listAll[PosProductVariant](ctx, s.client, fmt.Sprintf("v1/pos_shops/%d/product_variants", shopID), q)
}

// ListCoupons returns one page of POS shop coupons
// (GET /v1/pos_shop_coupons).
func (s *PosShopsService) ListCoupons(ctx context.Context, opts *ListOptions) (*Page[PosShopCoupon], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[PosShopCoupon](ctx, s.client, "v1/pos_shop_coupons", q)
}

// AllCoupons walks every page of POS shop coupons
// (GET /v1/pos_shop_coupons).
func (s *PosShopsService) AllCoupons(ctx context.Context, opts *ListOptions) iter.Seq2[PosShopCoupon, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(PosShopCoupon, error) bool) { yield(PosShopCoupon{}, err) }
	}
	return listAll[PosShopCoupon](ctx, s.client, "v1/pos_shop_coupons", q)
}

// GetCoupon returns one POS shop coupon. The detail response omits the id,
// so it is filled in from the argument (GET /v1/pos_shop_coupons/{id}).
func (s *PosShopsService) GetCoupon(ctx context.Context, id int64) (*PosShopCoupon, *Response, error) {
	var out PosShopCoupon
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/pos_shop_coupons/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// CreateCoupon creates a POS shop coupon (POST /v1/pos_shop_coupons).
func (s *PosShopsService) CreateCoupon(ctx context.Context, req *PosShopCouponCreateRequest) (*PosShopCoupon, *Response, error) {
	var out PosShopCoupon
	resp, err := s.client.post(ctx, "v1/pos_shop_coupons", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateCoupon changes a POS shop coupon. The response omits the id, so it
// is filled in from the argument (PUT /v1/pos_shop_coupons/{id}).
func (s *PosShopsService) UpdateCoupon(ctx context.Context, id int64, req *PosShopCouponUpdateRequest) (*PosShopCoupon, *Response, error) {
	var out PosShopCoupon
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/pos_shop_coupons/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// DeleteCoupon removes a POS shop coupon (DELETE /v1/pos_shop_coupons/{id}).
func (s *PosShopsService) DeleteCoupon(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/pos_shop_coupons/%d", id), nil)
}

// WalletBalance returns a customer's stored-value balance
// (GET /v2/pos_wallets/{customer_id}/balance).
func (s *PosShopsService) WalletBalance(ctx context.Context, customerID int64) (*PosWalletBalance, *Response, error) {
	var out PosWalletBalance
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v2/pos_wallets/%d/balance", customerID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// WalletTransactions returns a customer's stored-value history, newest
// first. The platform does not paginate it and caps it at 1000 rows
// (GET /v2/pos_wallets/{customer_id}/transactions).
func (s *PosShopsService) WalletTransactions(ctx context.Context, customerID int64) ([]PosWalletTransaction, *Response, error) {
	var out []PosWalletTransaction
	resp, err := s.client.get(ctx, fmt.Sprintf("v2/pos_wallets/%d/transactions", customerID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}
