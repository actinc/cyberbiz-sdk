package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// StockInvoiceListOptions filter ListInvoices.
type StockInvoiceListOptions struct {
	ListOptions
	// PosShopID is the shipping shop; 0 is EC. A pointer so 0 can be sent.
	PosShopID *int64 `url:"pos_shop_id,omitempty"`
	// TargetPosShopID is the receiving shop; 0 is EC, -1 a third party.
	TargetPosShopID *int64      `url:"target_pos_shop_id,omitempty"`
	Status          StockStatus `url:"status,omitempty"`
	// StartDate and EndDate bound the creation date, inclusive.
	StartDate Date `url:"start_date,omitempty"`
	EndDate   Date `url:"end_date,omitempty"`
}

// StockInvoiceCreateRequest is the body of CreateInvoice.
type StockInvoiceCreateRequest struct {
	// PosShopID is the shipping shop; 0 is EC.
	PosShopID int64 `json:"pos_shop_id"`
	// TargetType says where the goods go; see [StockInvoiceTargetType].
	TargetType StockInvoiceTargetType `json:"target_type"`
	// TargetPosShopID is required when TargetType is pos_shop.
	TargetPosShopID *int64 `json:"target_pos_shop_id,omitzero"`
	// ThirdParty names the recipient when TargetType is third_party.
	ThirdParty *string              `json:"third_party,omitzero"`
	Items      []StockLineItemInput `json:"items"`
	// EstDate is the expected arrival date.
	EstDate Date    `json:"est_date,omitzero"`
	Comment *string `json:"comment,omitzero"`
}

// ListInvoices returns one page of stock invoices (GET /v1/stock_invoices).
func (s *StockService) ListInvoices(ctx context.Context, opts *StockInvoiceListOptions) (*Page[StockInvoice], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[StockInvoice](ctx, s.client, "v1/stock_invoices", q)
}

// AllInvoices walks every page of stock invoices (GET /v1/stock_invoices).
func (s *StockService) AllInvoices(ctx context.Context, opts *StockInvoiceListOptions) iter.Seq2[StockInvoice, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(StockInvoice, error) bool) { yield(StockInvoice{}, err) }
	}
	return listAll[StockInvoice](ctx, s.client, "v1/stock_invoices", q)
}

// GetInvoice returns one stock invoice (GET /v1/stock_invoices/{id}).
func (s *StockService) GetInvoice(ctx context.Context, id int64) (*StockInvoice, *Response, error) {
	var out StockInvoice
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/stock_invoices/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateInvoice creates a stock invoice (POST /v1/stock_invoices).
func (s *StockService) CreateInvoice(ctx context.Context, req *StockInvoiceCreateRequest) (*StockInvoice, *Response, error) {
	var out StockInvoice
	resp, err := s.client.post(ctx, "v1/stock_invoices", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ConfirmInvoice confirms a pending invoice. The platform declares this
// state change as a GET; an invoice that is not ready yields a 403 with
// "Invoice is not ready." (GET /v1/stock_invoices/{id}/confirm).
func (s *StockService) ConfirmInvoice(ctx context.Context, id int64) (*StockInvoice, *Response, error) {
	return s.invoiceAction(ctx, id, "confirm")
}

// CancelInvoice cancels an invoice. The platform declares this state change
// as a GET (GET /v1/stock_invoices/{id}/cancel).
func (s *StockService) CancelInvoice(ctx context.Context, id int64) (*StockInvoice, *Response, error) {
	return s.invoiceAction(ctx, id, "cancel")
}

func (s *StockService) invoiceAction(ctx context.Context, id int64, action string) (*StockInvoice, *Response, error) {
	var out StockInvoice
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/stock_invoices/%d/%s", id, action), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
