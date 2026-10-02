package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"strconv"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// OrdersService exposes orders, order e-tickets and the v2 shipping
// endpoints. Fulfillment and label methods live in orders_fulfillments.go.
type OrdersService struct {
	client *Client
}

// OrderListOptions filters GET /v1/orders. Every time range is inclusive and
// expressed in the platform's "YYYY-MM-DD hh:mm:ss" format; multi-value
// filters are sent comma-separated and match any of the values.
type OrderListOptions struct {
	ListOptions
	StartTime            Time `url:"start_time,omitempty"` // created_at lower bound
	EndTime              Time `url:"end_time,omitempty"`   // created_at upper bound
	UpdatedAtStartTime   Time `url:"updated_at_start_time,omitempty"`
	UpdatedAtEndTime     Time `url:"updated_at_end_time,omitempty"`
	ClosedAtStartTime    Time `url:"closed_at_start_time,omitempty"`
	ClosedAtEndTime      Time `url:"closed_at_end_time,omitempty"`
	RefundAtStartTime    Time `url:"refund_at_start_time,omitempty"`
	RefundAtEndTime      Time `url:"refund_at_end_time,omitempty"`
	CancelledAtStartTime Time `url:"cancelled_at_start_time,omitempty"`
	CancelledAtEndTime   Time `url:"cancelled_at_end_time,omitempty"`

	Statuses            []OrderStatus       `url:"statuses,omitempty,comma"`
	FinancialStatuses   []FinancialStatus   `url:"financial_statuses,omitempty,comma"`
	FulfillmentStatuses []FulfillmentStatus `url:"fulfillment_statuses,omitempty,comma"`
	ReturnStatuses      []ReturnStatus      `url:"return_statuses,omitempty,comma"`
	Tags                []string            `url:"tags,omitempty,comma"`
	ExcludedTags        []string            `url:"excluded_tags,omitempty,comma"`
	DataSource          DataSource          `url:"data_source,omitempty"`
	Vendor              string              `url:"vendor,omitempty"` // product vendor
}

// OrderUpdateRequest changes the editable fields of an order. Delivery date
// and time are custom features that must be enabled by CYBERBIZ.
type OrderUpdateRequest struct {
	Note         *string `json:"note,omitzero"`
	DeliveryDate Date    `json:"delivery_date,omitzero"`
	DeliveryTime *int    `json:"delivery_time,omitzero"` // 0-3
}

// OrderCancelRequest cancels an order. Only CancelReason is required.
type OrderCancelRequest struct {
	CancelReason            CancelReason         `json:"cancel_reason"`
	CancelReasonDetail      CustomerCancelReason `json:"cancel_reason_detail,omitzero"`
	OtherCancelReasonDetail string               `json:"other_cancel_reason_detail,omitzero"`
	Email                   *bool                `json:"email,omitzero"`       // notify the customer
	SetWarning              *bool                `json:"set_warning,omitzero"` // flag the customer's account
	ShopdotcomCancel        *bool                `json:"shopdotcom_cancel,omitzero"`
	RefundShopdotcom        *Money               `json:"refund_shopdotcom,omitzero"`
}

// OrderTransactionCreateRequest records a payment against an order.
type OrderTransactionCreateRequest struct {
	Kind     TransactionKind     `json:"kind"`
	PaidType TransactionPaidType `json:"paid_type"`
}

// OrderEticketListOptions filters GET /v1/order_etickets. An empty Q lists
// every e-ticket.
type OrderEticketListOptions struct {
	ListOptions
	SearchColumn EticketSearchColumn `url:"search_column,omitempty"`
	Q            string              `url:"q,omitempty"`
}

// EticketRedeemRequest redeems units of an e-ticket at a branch store.
type EticketRedeemRequest struct {
	TicketNumber   string `json:"ticket_number"`
	RedeemQuantity int    `json:"redeem_quantity"`
	UserID         int64  `json:"user_id"`         // branch store staff
	BranchStoreID  int64  `json:"branch_store_id"` // redeeming store
}

// List returns one page of orders (GET /v1/orders).
func (s *OrdersService) List(ctx context.Context, opts *OrderListOptions) (*Page[Order], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Order](ctx, s.client, "v1/orders", q)
}

// All walks every page of orders (GET /v1/orders).
func (s *OrdersService) All(ctx context.Context, opts *OrderListOptions) iter.Seq2[Order, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Order, error) bool) { yield(Order{}, err) }
	}
	return listAll[Order](ctx, s.client, "v1/orders", q)
}

// Get returns one order (GET /v1/orders/{id}).
func (s *OrdersService) Get(ctx context.Context, id int64) (*Order, *Response, error) {
	var out Order
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/orders/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// LookupIDs maps shop-facing order numbers to API ids
// (GET /v1/orders/get_order_id).
func (s *OrdersService) LookupIDs(ctx context.Context, orderNumbers []int64) ([]OrderNumberID, *Response, error) {
	parts := make([]string, len(orderNumbers))
	for i, n := range orderNumbers {
		parts[i] = strconv.FormatInt(n, 10)
	}
	q := map[string][]string{"order_numbers": {strings.Join(parts, ",")}}
	var out []OrderNumberID
	resp, err := s.client.get(ctx, "v1/orders/get_order_id", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// decodeOrder returns the decoded order, or nil when the platform replied
// with an empty or null body.
func decodeOrder(out *Order, resp *Response) *Order {
	if resp.IsNull() || len(resp.Body) == 0 {
		return nil
	}
	return out
}

// putOrder sends a PUT and decodes the updated order.
func (s *OrdersService) putOrder(ctx context.Context, path string, body any) (*Order, *Response, error) {
	var out Order
	resp, err := s.client.put(ctx, path, body, &out)
	if err != nil {
		return nil, resp, err
	}
	return decodeOrder(&out, resp), resp, nil
}

// postOrder sends a POST and decodes the updated order.
func (s *OrdersService) postOrder(ctx context.Context, path string, body any) (*Order, *Response, error) {
	var out Order
	resp, err := s.client.post(ctx, path, body, &out)
	if err != nil {
		return nil, resp, err
	}
	return decodeOrder(&out, resp), resp, nil
}

// Update changes an order's note and delivery preferences (PUT /v1/orders/{id}).
func (s *OrdersService) Update(ctx context.Context, id int64, req *OrderUpdateRequest) (*Order, *Response, error) {
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d", id), req)
}

// UpdateTags replaces an order's tags (PUT /v1/orders/{id}/tags).
func (s *OrdersService) UpdateTags(ctx context.Context, id int64, tags []string) (*Order, *Response, error) {
	body := struct {
		Tags []string `json:"tags"`
	}{Tags: tags}
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/tags", id), body)
}

// UpdateWarehouseNote sets the logistics message shown for an order
// (PUT /v1/orders/{id}/update_warehouse_note).
func (s *OrdersService) UpdateWarehouseNote(ctx context.Context, id int64, note string) (*Order, *Response, error) {
	body := struct {
		WarehouseNote string `json:"warehouse_note"`
	}{WarehouseNote: note}
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/update_warehouse_note", id), body)
}

// UpdateStatus opens or closes an order; only OrderStatusOpen and
// OrderStatusClosed are accepted (PUT /v1/orders/{id}/update_status).
func (s *OrdersService) UpdateStatus(ctx context.Context, id int64, status OrderStatus) (*Order, *Response, error) {
	body := struct {
		Status OrderStatus `json:"status"`
	}{Status: status}
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/update_status", id), body)
}

// UpdateFinancialStatus asks the platform to refresh an order's payment
// state; it takes no parameters (PUT /v1/orders/{id}/update_financial_status).
func (s *OrdersService) UpdateFinancialStatus(ctx context.Context, id int64) (*Order, *Response, error) {
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/update_financial_status", id), nil)
}

// MarkUnshipped sets the fulfillment status back to unshipped
// (PUT /v1/orders/{id}/unshipped).
func (s *OrdersService) MarkUnshipped(ctx context.Context, id int64) (*Order, *Response, error) {
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/unshipped", id), nil)
}

// MarkPreparing sets the fulfillment status to preparing
// (PUT /v1/orders/{id}/preparing).
func (s *OrdersService) MarkPreparing(ctx context.Context, id int64) (*Order, *Response, error) {
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/preparing", id), nil)
}

// Cancel cancels an order (PUT /v1/orders/{id}/cancelled).
func (s *OrdersService) Cancel(ctx context.Context, id int64, req *OrderCancelRequest) (*Order, *Response, error) {
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/cancelled", id), req)
}

// ManualReturn moves an order's return status by hand
// (PUT /v1/orders/{id}/manual_return).
func (s *OrdersService) ManualReturn(ctx context.Context, id int64, op ManualReturnOperation) (*Order, *Response, error) {
	body := struct {
		Operation ManualReturnOperation `json:"operation"`
	}{Operation: op}
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/manual_return", id), body)
}

// ChangeToCustomShipping switches an order to merchant-arranged shipping
// (PUT /v1/orders/{id}/change_to_custom_shipping).
func (s *OrdersService) ChangeToCustomShipping(ctx context.Context, id int64) (*Order, *Response, error) {
	return s.putOrder(ctx, fmt.Sprintf("v1/orders/%d/change_to_custom_shipping", id), nil)
}

// ListTransactions returns one page of an order's payments
// (GET /v1/orders/{id}/transactions).
func (s *OrdersService) ListTransactions(ctx context.Context, id int64, opts *ListOptions) (*Page[OrderTransaction], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[OrderTransaction](ctx, s.client, fmt.Sprintf("v1/orders/%d/transactions", id), q)
}

// CreateTransaction records a manual payment against an order
// (POST /v1/orders/{id}/transactions).
func (s *OrdersService) CreateTransaction(ctx context.Context, id int64, req *OrderTransactionCreateRequest) (*OrderTransaction, *Response, error) {
	var out OrderTransaction
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/orders/%d/transactions", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ListReturns returns an order's return shipments. The endpoint is not
// paginated; the swagger documents a single object but the platform sends
// an array (GET /v1/orders/{id}/returns).
func (s *OrdersService) ListReturns(ctx context.Context, id int64) (*Page[OrderReturn], error) {
	return list[OrderReturn](ctx, s.client, fmt.Sprintf("v1/orders/%d/returns", id), nil)
}

// ListEtickets returns one page of e-tickets. The shop must have the
// e-ticket feature, else the platform answers 403 (GET /v1/order_etickets).
func (s *OrdersService) ListEtickets(ctx context.Context, opts *OrderEticketListOptions) (*Page[OrderEticket], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[OrderEticket](ctx, s.client, "v1/order_etickets", q)
}

// RedeemEticket redeems units of an e-ticket
// (POST /v1/order_etickets/submit_redeem).
func (s *OrdersService) RedeemEticket(ctx context.Context, req *EticketRedeemRequest) (*OrderEticket, *Response, error) {
	var out OrderEticket
	resp, err := s.client.post(ctx, "v1/order_etickets/submit_redeem", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
