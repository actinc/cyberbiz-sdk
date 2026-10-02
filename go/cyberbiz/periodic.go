package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// PeriodicService exposes subscription parent orders and their child
// pre-orders (/v1/periodic_orders).
type PeriodicService struct {
	client *Client
}

// PeriodicOrderListOptions filters List and All.
type PeriodicOrderListOptions struct {
	ListOptions
	// PreorderStartAt keeps parents with a child order created at or after
	// this time.
	PreorderStartAt Time `url:"preorder_start_at,omitempty"`
}

// PeriodicOrderUpdateRequest is the body of Update; every field is optional.
// Updating periodic orders is a paid feature enabled per shop.
type PeriodicOrderUpdateRequest struct {
	BillingAddress *PeriodicBillingAddressRequest `json:"billing_address,omitzero"`
	Note           *string                        `json:"note,omitzero"`
	Periodic       *PeriodicScheduleRequest       `json:"periodic,omitzero"`
	// DeliveryTime is the requested delivery slot, 0 to 3.
	DeliveryTime *int `json:"delivery_time,omitzero"`
}

// PeriodicBillingAddressRequest changes parts of the billing contact.
type PeriodicBillingAddressRequest struct {
	Name               *string `json:"name,omitzero"`
	Phone              *string `json:"phone,omitzero"`
	CountryCallingCode *string `json:"country_calling_code,omitzero"`
	Zip                *string `json:"zip,omitzero"`
	City               *string `json:"city,omitzero"`
	District           *string `json:"district,omitzero"`
	Address1           *string `json:"address1,omitzero"`
}

// PeriodicScheduleRequest changes the cadence; see PeriodicSchedule for the
// meaning of each field per type.
type PeriodicScheduleRequest struct {
	Month *int `json:"month,omitzero"`
	Week  *int `json:"week,omitzero"`
	Day   *int `json:"day,omitzero"`
}

// List returns one page of periodic orders (GET /v1/periodic_orders).
func (s *PeriodicService) List(ctx context.Context, opts *PeriodicOrderListOptions) (*Page[PeriodicOrder], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[PeriodicOrder](ctx, s.client, "v1/periodic_orders", q)
}

// All walks every page of periodic orders (GET /v1/periodic_orders).
func (s *PeriodicService) All(ctx context.Context, opts *PeriodicOrderListOptions) iter.Seq2[PeriodicOrder, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(PeriodicOrder, error) bool) { yield(PeriodicOrder{}, err) }
	}
	return listAll[PeriodicOrder](ctx, s.client, "v1/periodic_orders", q)
}

// Update changes a periodic order (PUT /v1/periodic_orders/{id}).
func (s *PeriodicService) Update(ctx context.Context, id int64, req *PeriodicOrderUpdateRequest) (*PeriodicOrder, *Response, error) {
	var out PeriodicOrder
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/periodic_orders/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// ListEstablishedPreorders returns one page of a periodic order's established
// child orders (GET /v1/periodic_orders/{id}/preorders/established).
func (s *PeriodicService) ListEstablishedPreorders(ctx context.Context, id int64, opts *ListOptions) (*Page[PeriodicPreorder], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[PeriodicPreorder](ctx, s.client, fmt.Sprintf("v1/periodic_orders/%d/preorders/established", id), q)
}

// AllEstablishedPreorders walks every page of a periodic order's established
// child orders (GET /v1/periodic_orders/{id}/preorders/established).
func (s *PeriodicService) AllEstablishedPreorders(ctx context.Context, id int64, opts *ListOptions) iter.Seq2[PeriodicPreorder, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(PeriodicPreorder, error) bool) { yield(PeriodicPreorder{}, err) }
	}
	return listAll[PeriodicPreorder](ctx, s.client, fmt.Sprintf("v1/periodic_orders/%d/preorders/established", id), q)
}
