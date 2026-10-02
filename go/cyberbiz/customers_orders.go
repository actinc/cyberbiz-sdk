package cyberbiz

import (
	"context"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CustomerOrder is an order as listed under a customer
// (GET /v1/customers/{id}/orders). The endpoint returns the full order
// document; this type models the parts most useful when browsing a
// customer's history. Use OrdersService for the complete order.
type CustomerOrder struct {
	ID            int64                  `json:"id"`
	OrderNumber   int64                  `json:"order_number"`
	OrderName     string                 `json:"order_name"` // e.g. "#1068"
	SubtotalPrice Money                  `json:"subtotal_price"`
	CreatedAt     Time                   `json:"created_at"`
	UpdatedAt     Time                   `json:"updated_at"`
	Customer      *Customer              `json:"customer"`
	Buyer         *CustomerOrderBuyer    `json:"buyer"`
	LineItems     []CustomerLineItem     `json:"line_items"`
	ShippingType  string                 `json:"shipping_type"` // e.g. "cyberbiz"
	ShippingName  string                 `json:"shipping_name"`
	PaymentName   string                 `json:"payment_name"`
	PaymentMethod string                 `json:"payment_method"`
	Prices        *CustomerOrderPrices   `json:"prices"`
	Statuses      *CustomerOrderStatuses `json:"statuses"`
	Timings       *CustomerOrderTimings  `json:"timings"`
	Note          string                 `json:"note"`
	// TotalBonusRedemptionPrice is the bonus points spent in the bonus mall.
	TotalBonusRedemptionPrice Money  `json:"total_bonus_redemption_price"`
	MerchantTradeNo           string `json:"merchant_trade_no"`
	FromDevice                string `json:"from_device"` // e.g. "桌機"
	Token                     string `json:"token"`       // the storefront order token
}

// CustomerOrderBuyer is who placed the order.
type CustomerOrderBuyer struct {
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

// CustomerOrderPrices are the order totals.
type CustomerOrderPrices struct {
	TotalLineItemsPrice Money `json:"total_line_items_price"`
	ShippingRatePrice   Money `json:"shipping_rate_price"`
	TotalPrice          Money `json:"total_price"`
}

// CustomerOrderStatuses are the four order state machines.
type CustomerOrderStatuses struct {
	OrderStatus       OrderStatus       `json:"order_status"`
	FinancialStatus   FinancialStatus   `json:"financial_status"`
	FulfillmentStatus FulfillmentStatus `json:"fulfillment_status"`
	ReturnStatus      ReturnStatus      `json:"return_status"`
}

// CustomerOrderTimings are the order's lifecycle timestamps; the pointers
// are nil while the event has not happened.
type CustomerOrderTimings struct {
	ConfirmedAt     Time  `json:"confirmed_at"`
	RequestReturnAt *Time `json:"request_return_at"`
	ReturnAt        *Time `json:"return_at"`
	RefundAt        *Time `json:"refund_at"`
	ClosedAt        *Time `json:"closed_at"`
	CancelledAt     *Time `json:"cancelled_at"`
	ExpiredAt       *Time `json:"expired_at"`
}

// CustomerLineItem is one purchased variant, as embedded in CustomerOrder
// and as returned by RecentPurchases.
type CustomerLineItem struct {
	ID                        int64                      `json:"id"`
	ProductID                 int64                      `json:"product_id"`
	ProductVariantID          int64                      `json:"product_variant_id"`
	Title                     string                     `json:"title"`
	VariantTitle              string                     `json:"variant_title"`
	SKU                       string                     `json:"sku"`
	QC                        string                     `json:"qc"` // vendor code
	Vendor                    string                     `json:"vendor"`
	Price                     Money                      `json:"price"`
	Cost                      Money                      `json:"cost"`
	Quantity                  int                        `json:"quantity"`
	ItemType                  string                     `json:"item_type"` // e.g. "normal"
	ReturnStatus              ReturnStatus               `json:"return_status"`
	DiscountName              string                     `json:"discount_name"`
	Discounts                 []CustomerLineItemDiscount `json:"discounts"`
	TotalPriceBeforeDiscounts Money                      `json:"total_price_before_discounts"`
	TotalDiscount             Money                      `json:"total_discount"`
	TotalPriceAfterDiscounts  Money                      `json:"total_price_after_discounts"`
	TaxTypeID                 string                     `json:"tax_type_id"` // inclusive_tax, zero_tax, exclusive_tax
	// BonusRedemptionPrice is the points paid in the bonus mall; zero otherwise.
	BonusRedemptionPrice Money                        `json:"bonus_redemption_price"`
	RelatedItems         []CustomerLineItemRelatedSet `json:"related_items"` // bundle contents
	CreatedAt            Time                         `json:"created_at"`
	Channel              string                       `json:"channel"`
	Weight               float64                      `json:"weight"`
	Photo                string                       `json:"photo"` // image path or URL
	CustomFields         []CustomerCustomField        `json:"custom_fields"`
}

// CustomerLineItemDiscount is one discount applied to a line item.
type CustomerLineItemDiscount struct {
	Position int    `json:"position"` // which unit of the line the discount hit
	ID       int64  `json:"id"`       // discount type id
	Code     string `json:"code"`     // e.g. "bundle_discount"
	Name     string `json:"name"`
	Discount Money  `json:"discount"`
}

// CustomerLineItemRelatedSet is one set of a bundle's contents.
type CustomerLineItemRelatedSet struct {
	Quantity int                       `json:"quantity"` // number of sets
	Items    []CustomerLineItemRelated `json:"items"`
}

// CustomerLineItemRelated is one variant inside a bundle.
type CustomerLineItemRelated struct {
	ID               int64  `json:"id"`
	ProductID        int64  `json:"product_id"`
	ProductVariantID int64  `json:"product_variant_id"`
	Title            string `json:"title"`
	VariantTitle     string `json:"variant_title"`
	SKU              string `json:"sku"`
	QC               string `json:"qc"`
	Vendor           string `json:"vendor"`
	Price            Money  `json:"price"`
	Cost             Money  `json:"cost"`
	Quantity         int    `json:"quantity"`
	// ComboProductPriceDifference is the bundle price difference (enterprise feature).
	ComboProductPriceDifference  Money   `json:"combo_product_price_difference"`
	ComboProductPriceDiffDetails []Money `json:"combo_product_price_diff_details"`
}

// CustomerRecentPurchasesOptions bound RecentPurchases; every field is
// required by the platform.
type CustomerRecentPurchasesOptions struct {
	StartDate   Date `url:"start_date"`
	EndDate     Date `url:"end_date"`
	MaxProducts int  `url:"max_products"`
}

// ListOrders returns one page of the customer's orders
// (GET /v1/customers/{id}/orders).
func (s *CustomersService) ListOrders(ctx context.Context, id int64, opts *ListOptions) (*Page[CustomerOrder], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerOrder](ctx, s.client, fmt.Sprintf("v1/customers/%d/orders", id), q)
}

// AllOrders walks every page of the customer's orders
// (GET /v1/customers/{id}/orders).
func (s *CustomersService) AllOrders(ctx context.Context, id int64, opts *ListOptions) iter.Seq2[CustomerOrder, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(CustomerOrder, error) bool) { yield(CustomerOrder{}, err) }
	}
	return listAll[CustomerOrder](ctx, s.client, fmt.Sprintf("v1/customers/%d/orders", id), q)
}

// RecentPurchases returns the line items of the customer's recent valid
// orders in a date range (GET /v1/customers/{id}/recent_purchases).
func (s *CustomersService) RecentPurchases(ctx context.Context, id int64, opts *CustomerRecentPurchasesOptions) ([]CustomerLineItem, *Response, error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, nil, err
	}
	out := []CustomerLineItem{}
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/customers/%d/recent_purchases", id), q, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}
