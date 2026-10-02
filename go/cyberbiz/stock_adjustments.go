package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// StockAdjustmentListOptions filter ListAdjustments.
type StockAdjustmentListOptions struct {
	ListOptions
	// PosShopID is the adjusted shop; 0 is EC. A pointer so 0 can be sent.
	PosShopID *int64 `url:"pos_shop_id,omitempty"`
}

// StockAdjustmentItemListOptions filter ListAdjustmentItems.
type StockAdjustmentItemListOptions struct {
	ListOptions
	// PosShopID is the adjusted shop; 0 is EC. A pointer so 0 can be sent.
	PosShopID *int64 `url:"pos_shop_id,omitempty"`
	// StartDate and EndDate bound the adjustment date, inclusive.
	StartDate Date `url:"start_date,omitempty"`
	EndDate   Date `url:"end_date,omitempty"`
	// Type selects one adjustment reason; see [StockAdjustmentType].
	Type StockAdjustmentType `url:"type,omitempty"`
}

// StockAdjustmentCreateRequest is the body of CreateAdjustment.
type StockAdjustmentCreateRequest struct {
	// PosShopID is the adjusted shop; 0 is EC.
	PosShopID int64                      `json:"pos_shop_id"`
	Items     []StockAdjustmentItemInput `json:"items"`
}

// ListAdjustments returns one page of adjustment records
// (GET /v1/stock_adjustments).
func (s *StockService) ListAdjustments(ctx context.Context, opts *StockAdjustmentListOptions) (*Page[StockAdjustment], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[StockAdjustment](ctx, s.client, "v1/stock_adjustments", q)
}

// AllAdjustments walks every page of adjustment records
// (GET /v1/stock_adjustments).
func (s *StockService) AllAdjustments(ctx context.Context, opts *StockAdjustmentListOptions) iter.Seq2[StockAdjustment, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(StockAdjustment, error) bool) { yield(StockAdjustment{}, err) }
	}
	return listAll[StockAdjustment](ctx, s.client, "v1/stock_adjustments", q)
}

// GetAdjustment returns one adjustment record
// (GET /v1/stock_adjustments/{id}).
func (s *StockService) GetAdjustment(ctx context.Context, id int64) (*StockAdjustment, *Response, error) {
	var out StockAdjustment
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/stock_adjustments/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateAdjustment records an inventory adjustment
// (POST /v1/stock_adjustments).
func (s *StockService) CreateAdjustment(ctx context.Context, req *StockAdjustmentCreateRequest) (*StockAdjustment, *Response, error) {
	var out StockAdjustment
	resp, err := s.client.post(ctx, "v1/stock_adjustments", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ListAdjustmentItems returns one page of adjustment lines across all
// records, filtered by shop, date and type (GET /v1/stock_adjustments/items).
func (s *StockService) ListAdjustmentItems(ctx context.Context, opts *StockAdjustmentItemListOptions) (*Page[StockAdjustmentItem], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[StockAdjustmentItem](ctx, s.client, "v1/stock_adjustments/items", q)
}

// AllAdjustmentItems walks every page of adjustment lines
// (GET /v1/stock_adjustments/items).
func (s *StockService) AllAdjustmentItems(ctx context.Context, opts *StockAdjustmentItemListOptions) iter.Seq2[StockAdjustmentItem, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(StockAdjustmentItem, error) bool) { yield(StockAdjustmentItem{}, err) }
	}
	return listAll[StockAdjustmentItem](ctx, s.client, "v1/stock_adjustments/items", q)
}
