package cyberbiz

import "encoding/json/jsontext"

// AffiliateOrder is one entry of GET /v2/affiliate_vendor_orders: the
// affiliate attribution plus the order it earned. The test shop has none,
// so the shape follows the Notion reference; the order part mirrors the v1
// order but is modelled separately with Affiliate-prefixed types.
type AffiliateOrder struct {
	// StatusCode is the attribution state, e.g. "create_succeed".
	StatusCode string `json:"status_code"`
	// UID and CID are the affiliate and campaign identifiers passed on the
	// storefront URL.
	UID        string                `json:"uid"`
	CID        string                `json:"cid"`
	VendorName string                `json:"vendor_name"`
	Order      *AffiliateOrderDetail `json:"order"`
}

// AffiliateOrderDetail is the order attributed to an affiliate.
type AffiliateOrderDetail struct {
	CustomerID    int64 `json:"customer_id"`
	SubtotalPrice Money `json:"subtotal_price"`
	CreatedAt     Time  `json:"created_at"`
	UpdatedAt     Time  `json:"updated_at"`
	OrderNumber   int64 `json:"order_number"`
	// OrderName is the display name, e.g. "#1020".
	OrderName string              `json:"order_name"`
	Buyer     *AffiliateBuyer     `json:"buyer"`
	LineItems []AffiliateLineItem `json:"line_items"`
	// ShippingType is the shipping provider kind, e.g. "cyberbiz";
	// ShippingName is the shipping method's display name.
	ShippingType string `json:"shipping_type"`
	ShippingName string `json:"shipping_name"`
	DeliveryDate Date   `json:"delivery_date"`
	// DeliveryTime is the requested delivery slot, 0 to 3.
	DeliveryTime      int                `json:"delivery_time"`
	Prices            *AffiliatePrices   `json:"prices"`
	TransactionNumber string             `json:"transaction_number"`
	MerchantTradeNo   string             `json:"merchant_trade_no"`
	Statuses          *AffiliateStatuses `json:"statuses"`
	Timings           *AffiliateTimings  `json:"timings"`
	Note              string             `json:"note"`
	// TotalBonusRedemptionPrice is the amount paid with bonus points.
	TotalBonusRedemptionPrice Money                     `json:"total_bonus_redemption_price"`
	LinkedOrderInfo           *AffiliateLinkedOrderInfo `json:"linked_order_info"`
	ShippingStatus            string                    `json:"shipping_status"`
	// ExtraInfo is free-form data the platform attaches to some orders.
	ExtraInfo jsontext.Value `json:"extra_info"`
	// FromDevice is the device class the order was placed from, e.g. "桌機".
	FromDevice string `json:"from_device"`
}

// AffiliateBuyer is the contact of an affiliate order.
type AffiliateBuyer struct {
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

// AffiliateLineItem is one line of an affiliate order.
type AffiliateLineItem struct {
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
	// ItemType is the line kind, e.g. "normal", "gift", "add_buy".
	ItemType                  string                      `json:"item_type"`
	DiscountName              string                      `json:"discount_name"`
	Discounts                 []AffiliateLineItemDiscount `json:"discounts"`
	TotalPriceBeforeDiscounts Money                       `json:"total_price_before_discounts"`
	TotalDiscount             Money                       `json:"total_discount"`
	TotalPriceAfterDiscounts  Money                       `json:"total_price_after_discounts"`
	// TaxTypeID is the tax treatment, e.g. "inclusive_tax".
	TaxTypeID            string              `json:"tax_type_id"`
	BonusRedemptionPrice Money               `json:"bonus_redemption_price"`
	RelatedItems         []AffiliateLineItem `json:"related_items"`
	CreatedAt            Time                `json:"created_at"`
}

// AffiliateLineItemDiscount is one discount applied to a line item.
type AffiliateLineItemDiscount struct {
	// Position is the 1-based index of the item the discount applies to.
	Position int    `json:"position"`
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Discount Money  `json:"discount"`
}

// AffiliatePrices is the price breakdown of an affiliate order.
type AffiliatePrices struct {
	TotalLineItemsPrice Money               `json:"total_line_items_price"`
	ShippingRatePrice   Money               `json:"shipping_rate_price"`
	Discounts           *AffiliateDiscounts `json:"discounts"`
	TotalPrice          Money               `json:"total_price"`
}

// AffiliateDiscounts itemises the order-level discounts.
type AffiliateDiscounts struct {
	SpecialCollectionDiscount Money                     `json:"special_collection_discount"`
	VIPDiscount               Money                     `json:"vip_discount"`
	ShopDiscount              *AffiliateShopDiscount    `json:"shop_discount"`
	CouponDiscount            *AffiliateCouponDiscount  `json:"coupon_discount"` // deprecated single coupon
	CouponDiscounts           []AffiliateCouponDiscount `json:"coupon_discounts"`
	BonusConsumed             Money                     `json:"bonus_consumed"`
	VIPShippingDiscount       Money                     `json:"vip_shipping_discount"`
	CouponShippingDiscount    Money                     `json:"coupon_shipping_discount"`
	PriceDiscount             Money                     `json:"price_discount"`
	ThirdPartyDiscount        Money                     `json:"third_party_discount"`
}

// AffiliateShopDiscount is the shop-wide campaign discount on an order.
type AffiliateShopDiscount struct {
	Name   string `json:"name"`
	Amount Money  `json:"amount"`
}

// AffiliateCouponDiscount is one coupon applied to an order.
type AffiliateCouponDiscount struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Amount   Money  `json:"amount"`
	CouponID int64  `json:"coupon_id"`
}

// AffiliateStatuses are the four status axes of an order.
type AffiliateStatuses struct {
	OrderStatus       OrderStatus       `json:"order_status"`
	FinancialStatus   FinancialStatus   `json:"financial_status"`
	FulfillmentStatus FulfillmentStatus `json:"fulfillment_status"`
	ReturnStatus      ReturnStatus      `json:"return_status"`
}

// AffiliateTimings are the lifecycle timestamps of an order; each is nil
// until the event happens.
type AffiliateTimings struct {
	RequestReturnAt *Time `json:"request_return_at"`
	ReturnAt        *Time `json:"return_at"`
	RefundAt        *Time `json:"refund_at"`
	ClosedAt        *Time `json:"closed_at"`
	CancelledAt     *Time `json:"cancelled_at"`
	ExpiredAt       *Time `json:"expired_at"`
	ConfirmedAt     *Time `json:"confirmed_at"`
}

// AffiliateLinkedOrderInfo describes the order's referral source.
type AffiliateLinkedOrderInfo struct {
	Source string `json:"source"`
}
