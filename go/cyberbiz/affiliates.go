package cyberbiz

import (
	"context"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// AffiliatesService exposes orders attributed to affiliate vendors
// (/v2/affiliate_vendor_orders). Child orders of periodic orders placed with
// affiliate parameters are included.
type AffiliatesService struct {
	client *Client
}

// AffiliateOrderListOptions filters List and All. Time bounds are inclusive.
type AffiliateOrderListOptions struct {
	ListOptions
	// StartTime and EndTime bound the order creation time.
	StartTime Time `url:"start_time,omitempty"`
	EndTime   Time `url:"end_time,omitempty"`
	// ClosedAtStartTime and ClosedAtEndTime bound the order close time.
	ClosedAtStartTime Time `url:"closed_at_start_time,omitempty"`
	ClosedAtEndTime   Time `url:"closed_at_end_time,omitempty"`
	// Statuses keeps orders in any of the given states.
	Statuses []OrderStatus `url:"statuses,omitempty,comma"`
}

// List returns one page of affiliate orders (GET /v2/affiliate_vendor_orders).
func (s *AffiliatesService) List(ctx context.Context, opts *AffiliateOrderListOptions) (*Page[AffiliateOrder], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[AffiliateOrder](ctx, s.client, "v2/affiliate_vendor_orders", q)
}

// All walks every page of affiliate orders (GET /v2/affiliate_vendor_orders).
func (s *AffiliatesService) All(ctx context.Context, opts *AffiliateOrderListOptions) iter.Seq2[AffiliateOrder, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(AffiliateOrder, error) bool) { yield(AffiliateOrder{}, err) }
	}
	return listAll[AffiliateOrder](ctx, s.client, "v2/affiliate_vendor_orders", q)
}
