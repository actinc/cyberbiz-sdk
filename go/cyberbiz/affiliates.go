package cyberbiz

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"iter"
	"net/url"
	"strconv"

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
	return s.list(ctx, q)
}

// All walks every page of affiliate orders (GET /v2/affiliate_vendor_orders).
func (s *AffiliatesService) All(ctx context.Context, opts *AffiliateOrderListOptions) iter.Seq2[AffiliateOrder, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(AffiliateOrder, error) bool) { yield(AffiliateOrder{}, err) }
	}
	start, _ := strconv.Atoi(q.Get("page"))
	return allPages(ctx, start, func(ctx context.Context, page int) (*Page[AffiliateOrder], error) {
		return s.list(ctx, withPage(q, page))
	})
}

func (s *AffiliatesService) list(ctx context.Context, q url.Values) (*Page[AffiliateOrder], error) {
	var items affiliateOrderPage
	resp, err := s.client.get(ctx, "v2/affiliate_vendor_orders", q, &items)
	if err != nil {
		return nil, err
	}
	return &Page[AffiliateOrder]{Items: items, Pagination: resp.Pagination, Response: resp}, nil
}

// affiliateOrderPage decodes one page of affiliate orders. A page with
// orders is wrapped as {"affiliate_vendor_orders": [...]}; a page past the
// end is a bare [] array. Both are accepted.
type affiliateOrderPage []AffiliateOrder

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
func (p *affiliateOrderPage) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() != '{' {
		var items []AffiliateOrder
		if err := json.UnmarshalDecode(dec, &items); err != nil {
			return err
		}
		*p = items
		return nil
	}
	var env struct {
		Orders []AffiliateOrder `json:"affiliate_vendor_orders"`
	}
	if err := json.UnmarshalDecode(dec, &env); err != nil {
		return err
	}
	*p = env.Orders
	return nil
}
