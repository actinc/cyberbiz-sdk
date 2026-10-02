package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// BranchStoresService exposes branch stores (門市) and their payments,
// shipping rates, staff users and delivery preparation settings. Every
// endpoint here answers 404 with {"error":["無此資源"]} for an unknown store.
type BranchStoresService struct {
	client *Client
}

// BranchStoreShippingRateCreateRequest is one shipping rate tier when
// creating a store or a rate.
type BranchStoreShippingRateCreateRequest struct {
	Name             string `json:"name"` // courier or delivery method name
	MinOrderSubtotal Money  `json:"min_order_subtotal"`
	Price            Money  `json:"price"`
}

// BranchStoreShippingRateUpdateRequest is the body of UpdateShippingRate.
type BranchStoreShippingRateUpdateRequest struct {
	Name             *string `json:"name,omitzero"`
	MinOrderSubtotal *Money  `json:"min_order_subtotal,omitzero"`
	Price            *Money  `json:"price,omitzero"`
}

// BranchStoreCreateRequest is the body of Create. The swagger declares no
// required fields; the platform needs at least a name.
type BranchStoreCreateRequest struct {
	Name string `json:"name"`
	// SourceType and SourceID link the store to a POS shop, if any.
	SourceType   BranchStoreSourceType `json:"source_type,omitzero"`
	SourceID     *int64                `json:"source_id,omitzero"`
	Phone        *string               `json:"phone,omitzero"`
	Zip          *string               `json:"zip,omitzero"`
	County       *string               `json:"county,omitzero"`
	District     *string               `json:"district,omitzero"`
	Address      *string               `json:"address,omitzero"`
	OpeningHours *string               `json:"opening_hours,omitzero"`
	StoreNo      *string               `json:"store_no,omitzero"`
	Lng          *float64              `json:"lng,omitzero"`
	Lat          *float64              `json:"lat,omitzero"`
	Enabled      *bool                 `json:"enabled,omitzero"`
	// PaymentIDs is the comma-separated list of allowed payment method ids,
	// from ListPayments; the platform takes it as one string.
	PaymentIDs    string                                 `json:"payment_ids,omitzero"`
	ShippingRates []BranchStoreShippingRateCreateRequest `json:"shipping_rates,omitzero"`
}

// BranchStoreUpdateRequest is the body of Update.
type BranchStoreUpdateRequest struct {
	Name         *string  `json:"name,omitzero"`
	Phone        *string  `json:"phone,omitzero"`
	Zip          *string  `json:"zip,omitzero"`
	County       *string  `json:"county,omitzero"`
	District     *string  `json:"district,omitzero"`
	Address      *string  `json:"address,omitzero"`
	OpeningHours *string  `json:"opening_hours,omitzero"`
	StoreNo      *string  `json:"store_no,omitzero"`
	Lng          *float64 `json:"lng,omitzero"`
	Lat          *float64 `json:"lat,omitzero"`
	Enabled      *bool    `json:"enabled,omitzero"`
	// PaymentIDs is the comma-separated list of allowed payment method ids.
	PaymentIDs string `json:"payment_ids,omitzero"`
}

// BranchStoreUserCreateRequest is the body of CreateUser.
type BranchStoreUserCreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// BranchStoreUserUpdateRequest is the body of UpdateUser.
type BranchStoreUserUpdateRequest struct {
	Name  *string `json:"name,omitzero"`
	Email *string `json:"email,omitzero"`
}

// PrepareDeliveryConfigUpdateRequest is the body of
// UpdatePrepareDeliveryConfig.
type PrepareDeliveryConfigUpdateRequest struct {
	DeliveryDateEnabled  *bool `json:"delivery_date_enabled,omitzero"`
	DeliveryDateRequired *bool `json:"delivery_date_required,omitzero"`
	// PrepareDeliveryDay is the lead time in days, 1 to 90.
	PrepareDeliveryDay *int `json:"prepare_delivery_day,omitzero"`
	SelectableRange    *int `json:"selectable_range,omitzero"`
}

// List returns one page of branch stores (GET /v1/branch_stores).
func (s *BranchStoresService) List(ctx context.Context, opts *ListOptions) (*Page[BranchStore], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[BranchStore](ctx, s.client, "v1/branch_stores", q)
}

// All walks every page of branch stores (GET /v1/branch_stores).
func (s *BranchStoresService) All(ctx context.Context, opts *ListOptions) iter.Seq2[BranchStore, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(BranchStore, error) bool) { yield(BranchStore{}, err) }
	}
	return listAll[BranchStore](ctx, s.client, "v1/branch_stores", q)
}

// Get returns one branch store. The detail response omits the id, so it is
// filled in from the argument (GET /v1/branch_stores/{id}).
func (s *BranchStoresService) Get(ctx context.Context, id int64) (*BranchStore, *Response, error) {
	var out BranchStore
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/branch_stores/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Create creates a branch store (POST /v1/branch_stores).
func (s *BranchStoresService) Create(ctx context.Context, req *BranchStoreCreateRequest) (*BranchStore, *Response, error) {
	var out BranchStore
	resp, err := s.client.post(ctx, "v1/branch_stores", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Update changes a branch store. The response omits the id, so it is filled
// in from the argument (PUT /v1/branch_stores/{id}).
func (s *BranchStoresService) Update(ctx context.Context, id int64, req *BranchStoreUpdateRequest) (*BranchStore, *Response, error) {
	var out BranchStore
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/branch_stores/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Delete removes a branch store (DELETE /v1/branch_stores/{id}).
func (s *BranchStoresService) Delete(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/branch_stores/%d", id), nil)
}

// ListPayments returns the payment methods that store pickup may use. The
// list is not paginated (GET /v1/branch_stores/payments).
func (s *BranchStoresService) ListPayments(ctx context.Context) ([]BranchStorePayment, *Response, error) {
	var out []BranchStorePayment
	resp, err := s.client.get(ctx, "v1/branch_stores/payments", nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// SetExpressDelivery turns express delivery (快速到貨) on or off for a store.
// The swagger documents the reply as a list of stores; it is left undecoded
// in the Response (PUT /v1/branch_stores/{id}/express_delivery_setting).
func (s *BranchStoresService) SetExpressDelivery(ctx context.Context, id int64, enabled bool) (*Response, error) {
	body := struct {
		Status bool `json:"express_delivery_status"`
	}{Status: enabled}
	return s.client.put(ctx, fmt.Sprintf("v1/branch_stores/%d/express_delivery_setting", id), body, nil)
}

// ListShippingRates returns every shipping rate tier of a store. The list
// is not paginated (GET /v1/branch_stores/{id}/shipping_rates).
func (s *BranchStoresService) ListShippingRates(ctx context.Context, storeID int64) ([]BranchStoreShippingRate, *Response, error) {
	var out []BranchStoreShippingRate
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/branch_stores/%d/shipping_rates", storeID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// GetShippingRate returns one shipping rate tier. The detail response omits
// the id, so it is filled in from the argument
// (GET /v1/branch_stores/{id}/shipping_rates/{rate_id}).
func (s *BranchStoresService) GetShippingRate(ctx context.Context, storeID, rateID int64) (*BranchStoreShippingRate, *Response, error) {
	var out BranchStoreShippingRate
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/branch_stores/%d/shipping_rates/%d", storeID, rateID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = rateID
	return &out, resp, nil
}

// CreateShippingRate adds a shipping rate tier to a store
// (POST /v1/branch_stores/{id}/shipping_rates).
func (s *BranchStoresService) CreateShippingRate(ctx context.Context, storeID int64, req *BranchStoreShippingRateCreateRequest) (*BranchStoreShippingRate, *Response, error) {
	var out BranchStoreShippingRate
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/branch_stores/%d/shipping_rates", storeID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateShippingRate changes a shipping rate tier. The response omits the
// id, so it is filled in from the argument
// (PUT /v1/branch_stores/{id}/shipping_rates/{rate_id}).
func (s *BranchStoresService) UpdateShippingRate(ctx context.Context, storeID, rateID int64, req *BranchStoreShippingRateUpdateRequest) (*BranchStoreShippingRate, *Response, error) {
	var out BranchStoreShippingRate
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/branch_stores/%d/shipping_rates/%d", storeID, rateID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = rateID
	return &out, resp, nil
}

// DeleteShippingRate removes a shipping rate tier
// (DELETE /v1/branch_stores/{id}/shipping_rates/{rate_id}).
func (s *BranchStoresService) DeleteShippingRate(ctx context.Context, storeID, rateID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/branch_stores/%d/shipping_rates/%d", storeID, rateID), nil)
}

// ListUsers returns one page of a store's staff users
// (GET /v1/branch_stores/{id}/users).
func (s *BranchStoresService) ListUsers(ctx context.Context, storeID int64, opts *ListOptions) (*Page[BranchStoreUser], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[BranchStoreUser](ctx, s.client, fmt.Sprintf("v1/branch_stores/%d/users", storeID), q)
}

// AllUsers walks every page of a store's staff users
// (GET /v1/branch_stores/{id}/users).
func (s *BranchStoresService) AllUsers(ctx context.Context, storeID int64, opts *ListOptions) iter.Seq2[BranchStoreUser, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(BranchStoreUser, error) bool) { yield(BranchStoreUser{}, err) }
	}
	return listAll[BranchStoreUser](ctx, s.client, fmt.Sprintf("v1/branch_stores/%d/users", storeID), q)
}

// GetUser returns one staff user (GET /v1/branch_stores/{id}/users/{user_id}).
func (s *BranchStoresService) GetUser(ctx context.Context, storeID, userID int64) (*BranchStoreUser, *Response, error) {
	var out BranchStoreUser
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/branch_stores/%d/users/%d", storeID, userID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = userID
	return &out, resp, nil
}

// CreateUser adds a staff user to a store
// (POST /v1/branch_stores/{id}/users).
func (s *BranchStoresService) CreateUser(ctx context.Context, storeID int64, req *BranchStoreUserCreateRequest) (*BranchStoreUser, *Response, error) {
	var out BranchStoreUser
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/branch_stores/%d/users", storeID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateUser changes a staff user. The response omits the id, so it is
// filled in from the argument
// (PUT /v1/branch_stores/{id}/users/{user_id}).
func (s *BranchStoresService) UpdateUser(ctx context.Context, storeID, userID int64, req *BranchStoreUserUpdateRequest) (*BranchStoreUser, *Response, error) {
	var out BranchStoreUser
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/branch_stores/%d/users/%d", storeID, userID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = userID
	return &out, resp, nil
}

// DeleteUser removes a staff user
// (DELETE /v1/branch_stores/{id}/users/{user_id}).
func (s *BranchStoresService) DeleteUser(ctx context.Context, storeID, userID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/branch_stores/%d/users/%d", storeID, userID), nil)
}

// GetPrepareDeliveryConfig returns a store's preparation-time settings
// (GET /v1/branch_stores/{id}/prepare_delivery_config).
func (s *BranchStoresService) GetPrepareDeliveryConfig(ctx context.Context, storeID int64) (*PrepareDeliveryConfig, *Response, error) {
	var out PrepareDeliveryConfig
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/branch_stores/%d/prepare_delivery_config", storeID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdatePrepareDeliveryConfig changes a store's preparation-time settings
// (PUT /v1/branch_stores/{id}/prepare_delivery_config).
func (s *BranchStoresService) UpdatePrepareDeliveryConfig(ctx context.Context, storeID int64, req *PrepareDeliveryConfigUpdateRequest) (*PrepareDeliveryConfig, *Response, error) {
	var out PrepareDeliveryConfig
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/branch_stores/%d/prepare_delivery_config", storeID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
