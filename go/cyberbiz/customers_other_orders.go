package cyberbiz

import (
	"context"
	"fmt"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CustomerOtherValidOrder is a purchase made through another channel that
// counts toward the customer's spend (GET /v1/customers/{id}/other_valid_orders).
type CustomerOtherValidOrder struct {
	ID         int64  `json:"id"`
	Price      Money  `json:"price"`
	ValidAt    Time   `json:"valid_at"`   // when the order was established
	InvalidAt  *Time  `json:"invalid_at"` // when it was cancelled; nil while valid
	CustomerID int64  `json:"customer_id"`
	UID        string `json:"uid"` // caller-defined identifier
	// ExtraInfos needs the other-valid-orders plugin; nil when absent.
	ExtraInfos *CustomerOtherValidOrderExtraInfo `json:"extra_infos"`
}

// CustomerOtherValidOrderExtraInfo describes an other-channel order in detail.
type CustomerOtherValidOrderExtraInfo struct {
	ChannelName  string                           `json:"channel_name,omitzero"`
	InvoiceNo    string                           `json:"invoice_no,omitzero"`
	Source       string                           `json:"source,omitzero"`
	Note         string                           `json:"note,omitzero"`     // merchant note
	Operator     string                           `json:"operator,omitzero"` // who recorded the order
	CustomerNote string                           `json:"customer_note,omitzero"`
	Products     []CustomerOtherValidOrderProduct `json:"products,omitzero"`
}

// CustomerOtherValidOrderProduct is one line of an other-channel order.
type CustomerOtherValidOrderProduct struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    Money  `json:"price"` // unit price
}

// CustomerOtherValidOrderCreateRequest is the body of
// POST /v1/customers/{id}/other_valid_orders.
type CustomerOtherValidOrderCreateRequest struct {
	Price      Money                             `json:"price"`
	ValidAt    Date                              `json:"valid_at"`
	UID        string                            `json:"uid,omitzero"`
	ExtraInfos *CustomerOtherValidOrderExtraInfo `json:"extra_infos,omitzero"`
}

// CustomerOtherValidOrderUpdateRequest is the body of
// PUT /v1/customers/{id}/other_valid_orders/{other_valid_order_id}.
type CustomerOtherValidOrderUpdateRequest struct {
	Price      *Money                            `json:"price,omitzero"`
	ValidAt    Date                              `json:"valid_at,omitzero"`
	ExtraInfos *CustomerOtherValidOrderExtraInfo `json:"extra_infos,omitzero"`
}

// ListOtherValidOrders returns one page of the customer's other-channel
// orders (GET /v1/customers/{id}/other_valid_orders).
func (s *CustomersService) ListOtherValidOrders(ctx context.Context, id int64, opts *ListOptions) (*Page[CustomerOtherValidOrder], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerOtherValidOrder](ctx, s.client, fmt.Sprintf("v1/customers/%d/other_valid_orders", id), q)
}

// GetOtherValidOrder returns one other-channel order
// (GET /v1/customers/{id}/other_valid_orders/{other_valid_order_id}).
func (s *CustomersService) GetOtherValidOrder(ctx context.Context, customerID, orderID int64) (*CustomerOtherValidOrder, *Response, error) {
	var out CustomerOtherValidOrder
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d/other_valid_orders/%d", customerID, orderID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateOtherValidOrder records an other-channel order for the customer
// (POST /v1/customers/{id}/other_valid_orders).
func (s *CustomersService) CreateOtherValidOrder(ctx context.Context, id int64, req *CustomerOtherValidOrderCreateRequest) (*CustomerOtherValidOrder, *Response, error) {
	var out CustomerOtherValidOrder
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/customers/%d/other_valid_orders", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateOtherValidOrder changes an other-channel order
// (PUT /v1/customers/{id}/other_valid_orders/{other_valid_order_id}).
func (s *CustomersService) UpdateOtherValidOrder(ctx context.Context, customerID, orderID int64, req *CustomerOtherValidOrderUpdateRequest) (*CustomerOtherValidOrder, *Response, error) {
	var out CustomerOtherValidOrder
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/customers/%d/other_valid_orders/%d", customerID, orderID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// DeleteOtherValidOrder removes an other-channel order
// (DELETE /v1/customers/{id}/other_valid_orders/{other_valid_order_id}).
func (s *CustomersService) DeleteOtherValidOrder(ctx context.Context, customerID, orderID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/customers/%d/other_valid_orders/%d", customerID, orderID), nil)
}

// InvalidateOtherValidOrder cancels an other-channel order so it no longer
// counts toward spend; this is a custom feature that needs enabling
// (POST /v1/customers/{id}/other_valid_orders/{other_valid_order_id}/invalid).
func (s *CustomersService) InvalidateOtherValidOrder(ctx context.Context, customerID, orderID int64) (*CustomerOtherValidOrder, *Response, error) {
	var out CustomerOtherValidOrder
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/customers/%d/other_valid_orders/%d/invalid", customerID, orderID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
