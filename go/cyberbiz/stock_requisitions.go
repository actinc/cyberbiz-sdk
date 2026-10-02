package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// StockRequisitionListOptions filter ListRequisitions.
type StockRequisitionListOptions struct {
	ListOptions
	// PosShopID is the receiving shop; 0 is EC. A pointer so 0 can be sent.
	PosShopID *int64 `url:"pos_shop_id,omitempty"`
	// SourcePosShopID is the shop giving up the stock; 0 is EC.
	SourcePosShopID *int64 `url:"source_pos_shop_id,omitempty"`
	// Status is pending, done or canceled.
	Status StockStatus `url:"status,omitempty"`
	// StartDate and EndDate bound the creation date, inclusive.
	StartDate Date `url:"start_date,omitempty"`
	EndDate   Date `url:"end_date,omitempty"`
}

// StockRequisitionCreateRequest is the body of CreateRequisition.
type StockRequisitionCreateRequest struct {
	// PosShopID is the receiving shop; 0 is EC.
	PosShopID int64 `json:"pos_shop_id"`
	// SourcePosShopID is the shop giving up the stock; 0 is EC.
	SourcePosShopID int64                `json:"source_pos_shop_id"`
	Items           []StockLineItemInput `json:"items"`
	Comment         *string              `json:"comment,omitzero"`
}

// ListRequisitions returns one page of stock requisitions
// (GET /v1/stock_requisitions).
func (s *StockService) ListRequisitions(ctx context.Context, opts *StockRequisitionListOptions) (*Page[StockRequisition], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[StockRequisition](ctx, s.client, "v1/stock_requisitions", q)
}

// AllRequisitions walks every page of stock requisitions
// (GET /v1/stock_requisitions).
func (s *StockService) AllRequisitions(ctx context.Context, opts *StockRequisitionListOptions) iter.Seq2[StockRequisition, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(StockRequisition, error) bool) { yield(StockRequisition{}, err) }
	}
	return listAll[StockRequisition](ctx, s.client, "v1/stock_requisitions", q)
}

// GetRequisition returns one stock requisition
// (GET /v1/stock_requisitions/{id}).
func (s *StockService) GetRequisition(ctx context.Context, id int64) (*StockRequisition, *Response, error) {
	var out StockRequisition
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/stock_requisitions/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateRequisition creates a stock requisition
// (POST /v1/stock_requisitions).
func (s *StockService) CreateRequisition(ctx context.Context, req *StockRequisitionCreateRequest) (*StockRequisition, *Response, error) {
	var out StockRequisition
	resp, err := s.client.post(ctx, "v1/stock_requisitions", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ApproveRequisition confirms a pending requisition. The platform declares
// this state change as a GET; a requisition that is not pending yields a
// 403 (GET /v1/stock_requisitions/{id}/approve).
func (s *StockService) ApproveRequisition(ctx context.Context, id int64) (*StockRequisition, *Response, error) {
	return s.requisitionAction(ctx, id, "approve")
}

// RejectRequisition rejects a pending requisition. The platform declares
// this state change as a GET (GET /v1/stock_requisitions/{id}/reject).
func (s *StockService) RejectRequisition(ctx context.Context, id int64) (*StockRequisition, *Response, error) {
	return s.requisitionAction(ctx, id, "reject")
}

func (s *StockService) requisitionAction(ctx context.Context, id int64, action string) (*StockRequisition, *Response, error) {
	var out StockRequisition
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/stock_requisitions/%d/%s", id, action), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
