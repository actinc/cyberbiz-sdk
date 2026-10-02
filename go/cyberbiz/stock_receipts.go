package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// StockReceiptListOptions filter ListReceipts.
type StockReceiptListOptions struct {
	ListOptions
	// PosShopID is the receiving shop; 0 is EC. A pointer so 0 can be sent.
	PosShopID *int64 `url:"pos_shop_id,omitempty"`
	// SourcePosShopID is the shipping shop; 0 is EC, -1 a third party.
	SourcePosShopID *int64      `url:"source_pos_shop_id,omitempty"`
	Status          StockStatus `url:"status,omitempty"`
	// StartDate and EndDate bound the creation date, inclusive.
	StartDate Date `url:"start_date,omitempty"`
	EndDate   Date `url:"end_date,omitempty"`
}

// StockReceiptCreateRequest is the body of CreateReceipt.
type StockReceiptCreateRequest struct {
	// PosShopID is the receiving shop; 0 is EC.
	PosShopID int64                `json:"pos_shop_id"`
	Items     []StockLineItemInput `json:"items"`
	// EstDate is the expected arrival date.
	EstDate Date    `json:"est_date,omitzero"`
	Comment *string `json:"comment,omitzero"`
}

// StockReceiptCheckRequest is the body of CheckReceipt: the quantities
// actually counted on arrival.
type StockReceiptCheckRequest struct {
	Items []StockLineItemInput `json:"items"`
}

// ListReceipts returns one page of stock receipts (GET /v1/stock_receipts).
func (s *StockService) ListReceipts(ctx context.Context, opts *StockReceiptListOptions) (*Page[StockReceipt], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[StockReceipt](ctx, s.client, "v1/stock_receipts", q)
}

// AllReceipts walks every page of stock receipts (GET /v1/stock_receipts).
func (s *StockService) AllReceipts(ctx context.Context, opts *StockReceiptListOptions) iter.Seq2[StockReceipt, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(StockReceipt, error) bool) { yield(StockReceipt{}, err) }
	}
	return listAll[StockReceipt](ctx, s.client, "v1/stock_receipts", q)
}

// GetReceipt returns one stock receipt (GET /v1/stock_receipts/{id}).
func (s *StockService) GetReceipt(ctx context.Context, id int64) (*StockReceipt, *Response, error) {
	var out StockReceipt
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/stock_receipts/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateReceipt creates a stock receipt (POST /v1/stock_receipts).
func (s *StockService) CreateReceipt(ctx context.Context, req *StockReceiptCreateRequest) (*StockReceipt, *Response, error) {
	var out StockReceipt
	resp, err := s.client.post(ctx, "v1/stock_receipts", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ApproveReceipt confirms a pending receipt. The platform declares this
// state change as a GET; a receipt that cannot be approved yields a 403
// (GET /v1/stock_receipts/{id}/approve).
func (s *StockService) ApproveReceipt(ctx context.Context, id int64) (*StockReceipt, *Response, error) {
	return s.receiptAction(ctx, id, "approve")
}

// RejectReceipt rejects a pending receipt. The platform declares this state
// change as a GET (GET /v1/stock_receipts/{id}/reject).
func (s *StockService) RejectReceipt(ctx context.Context, id int64) (*StockReceipt, *Response, error) {
	return s.receiptAction(ctx, id, "reject")
}

// CancelReceipt cancels a receipt. The platform declares this state change
// as a GET (GET /v1/stock_receipts/{id}/cancel).
func (s *StockService) CancelReceipt(ctx context.Context, id int64) (*StockReceipt, *Response, error) {
	return s.receiptAction(ctx, id, "cancel")
}

// ArriveReceipt marks the goods as arrived, moving the receipt to
// processing. The platform declares this state change as a GET
// (GET /v1/stock_receipts/{id}/arrival).
func (s *StockService) ArriveReceipt(ctx context.Context, id int64) (*StockReceipt, *Response, error) {
	return s.receiptAction(ctx, id, "arrival")
}

// CheckReceipt records the counted quantities for a receipt
// (POST /v1/stock_receipts/{id}/check).
func (s *StockService) CheckReceipt(ctx context.Context, id int64, req *StockReceiptCheckRequest) (*StockReceipt, *Response, error) {
	var out StockReceipt
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/stock_receipts/%d/check", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

func (s *StockService) receiptAction(ctx context.Context, id int64, action string) (*StockReceipt, *Response, error) {
	var out StockReceipt
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/stock_receipts/%d/%s", id, action), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
