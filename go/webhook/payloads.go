package webhook

import (
	"encoding/json/jsontext"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

// Payload types for each Event resource. They follow the field tables in
// CYBERBIZ's webhook documentation, corrected where recorded deliveries
// disagree with it (see webhook_test.go's testdata). Scalars are value
// types, so a JSON null decodes to the zero value; nested objects are
// pointers so a missing object is nil. These types are deliberately
// independent of the cyberbiz package's API models: webhook bodies and API
// responses do not share a schema.

// CustomerPayload is the body of customers/* events and the "customer"
// object inside an order.
type CustomerPayload struct {
	ID int64 `json:"id"`
	// Name is the customer's display name.
	Name string `json:"name"`
	// Status is one of pending, validate, enabled, disabled, invited,
	// declined, warning.
	Status                   string           `json:"status"`
	Email                    string           `json:"email"`
	CountryCallingCode       string           `json:"country_calling_code"`
	Mobile                   string           `json:"mobile"`
	Gender                   string           `json:"gender"`
	Birthday                 cyberbiz.Date    `json:"birthday"`
	EnableCVSPickup          bool             `json:"enable_cvs_pickup"`
	EnableCVSCOD             bool             `json:"enable_cvs_cod"`
	EnableHomeDeliveryCOD    bool             `json:"enable_home_delivery_cod"`
	AcceptsMarketing         bool             `json:"accepts_marketing"`
	AcceptsEmailNotification bool             `json:"accepts_email_notification"`
	Tags                     []NamedTag       `json:"tags"`
	Address                  *CustomerAddress `json:"address"`
	// OtherAccumulatedConsumption is spend accumulated through other
	// channels, counted toward VIP thresholds.
	OtherAccumulatedConsumption          cyberbiz.Money `json:"other_accumulated_consumption"`
	OtherAccumulatedConsumptionExpiredAt cyberbiz.Time  `json:"other_accumulated_consumption_expired_at"`
	Note                                 string         `json:"note"`
	CustomFields                         []CustomField  `json:"custom_fields"`
	// BonusRemain is the customer's current bonus point balance.
	BonusRemain  float64       `json:"bonus_remain"`
	UIDProviders []UIDProvider `json:"uid_providers"`
	CreatedAt    cyberbiz.Time `json:"created_at"`
	UpdatedAt    cyberbiz.Time `json:"updated_at"`
	// ConfirmedAt is when the email was verified; nil if never.
	ConfirmedAt *cyberbiz.Time `json:"confirmed_at"`
	// MobileSMSConfirmedAt is when the mobile number was verified; nil if
	// never.
	MobileSMSConfirmedAt *cyberbiz.Time `json:"mobile_sms_confirmed_at"`
}

// NamedTag is a tag carried as an object with a name, as on customers and
// orders.
type NamedTag struct {
	Name string `json:"name"`
}

// CustomerAddress is a customer's stored address.
type CustomerAddress struct {
	Company            string         `json:"company"`
	CountryCallingCode string         `json:"country_calling_code"`
	Phone              string         `json:"phone"`
	Address            string         `json:"address"`
	DetailAddress      *DetailAddress `json:"detail_address"`
}

// DetailAddress is an address broken into its parts.
type DetailAddress struct {
	Zip      string `json:"zip"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
}

// CustomField is one shop-defined customer field.
type CustomField struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// UIDProvider links a customer to an external identity such as a LINE
// user id.
type UIDProvider struct {
	// ProviderType names the provider, e.g. "line".
	ProviderType string `json:"provider_type"`
	UID          string `json:"uid"`
}

// UIDProviderPayload is the body of uid_providers/* events. The
// documentation lists only these two fields; use [Event.Decode] for any
// others a delivery carries.
type UIDProviderPayload = UIDProvider

// BonusPointPayload is the body of bonus_points/* and comment_bonus/*
// events: one bonus point grant.
type BonusPointPayload struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// Points granted; UnusedPoints is how many remain.
	Points       float64 `json:"points"`
	UnusedPoints float64 `json:"unused_points"`
	// ConsumptionPrice is the order amount that earned the points.
	ConsumptionPrice cyberbiz.Money `json:"consumption_price"`
	// Deadline is when the points expire. Documented as a date, delivered
	// as a full timestamp.
	Deadline   cyberbiz.Time `json:"deadline"`
	CustomerID int64         `json:"customer_id"`
	Source     string        `json:"source"`
	// OrderID is the order that earned the points, 0 when none.
	OrderID int64 `json:"order_id"`
}

// OrderPayload is the body of orders/* and express_delivery_orders/*
// events.
type OrderPayload struct {
	ID    int64  `json:"id"`
	Token string `json:"token"`
	// OrderNumber is documented as a string but delivered as an integer.
	OrderNumber int64  `json:"order_number"`
	OrderName   string `json:"order_name"`
	// SubtotalPrice is undocumented but present in every delivery.
	SubtotalPrice cyberbiz.Money   `json:"subtotal_price"`
	Customer      *CustomerPayload `json:"customer"`
	Buyer         *Buyer           `json:"buyer"`
	Receiver      *Receiver        `json:"receiver"`
	// BillingAddress is a shop-specific customisation; its shape is not
	// documented.
	BillingAddress jsontext.Value  `json:"billing_address"`
	LineItems      []LineItem      `json:"line_items"`
	ShippingType   string          `json:"shipping_type"`
	ShippingName   string          `json:"shipping_name"`
	ShippingVendor *ShippingVendor `json:"shipping_vendor"`
	// LogisticsID is the ECPay logistics order number.
	LogisticsID  string        `json:"logistics_id"`
	DeliveryDate cyberbiz.Date `json:"delivery_date"`
	// DeliveryTime is the requested delivery time slot code.
	DeliveryTime int `json:"delivery_time"`
	// Delegate is the staff member who placed the order on the customer's
	// behalf.
	Delegate     string        `json:"delegate"`
	Fulfillments []Fulfillment `json:"fulfillments"`
	// PaymentName is the payment method code, PaymentMethod its label.
	PaymentName          string        `json:"payment_name"`
	PaymentMethod        string        `json:"payment_method"`
	PaymentURL           string        `json:"payment_url"`
	MultiplePaymentInfos []PaymentInfo `json:"multiple_payment_infos"`
	Prices               *Prices       `json:"prices"`
	// Card4No is the last four digits of the card used.
	Card4No                  string          `json:"card4no"`
	TransactionNumber        string          `json:"transaction_number"`
	MerchantTradeNo          string          `json:"merchant_trade_no"`
	EInvoice                 *EInvoice       `json:"einvoice"`
	PaperInvoiceNo           string          `json:"paper_invoice_no"`
	Statuses                 *OrderStatuses  `json:"statuses"`
	Timings                  *OrderTimings   `json:"timings"`
	ReturnHistories          []ReturnHistory `json:"return_histories"`
	Note                     string          `json:"note"`
	BranchStore              *BranchStore    `json:"branch_store"`
	ReferralCode             string          `json:"referral_code"`
	CheckoutReferralCode     string          `json:"checkout_referral_code"`
	CheckoutReferralUserName string          `json:"checkout_referral_user_name"`
	RegisterReferralCode     string          `json:"register_referral_code"`
	// TotalBonusRedemptionPrice is the bonus points spent in the bonus
	// mall for this order.
	TotalBonusRedemptionPrice cyberbiz.Money `json:"total_bonus_redemption_price"`
	POSInfo                   *POSInfo       `json:"pos_info"`
	// ExchangeHistories is documented as an object but delivered as an
	// array.
	ExchangeHistories          []ExchangeHistory `json:"exchange_histories"`
	LinkedOrderInfo            *LinkedOrderInfo  `json:"linked_order_info"`
	Tags                       []NamedTag        `json:"tags"`
	ExpressDeliveryBranchStore *BranchStore      `json:"express_delivery_branch_store"`
	ShippingStatus             string            `json:"shipping_status"`
	ExtraInfo                  string            `json:"extra_info"`
	// FromDevice is the device the order was placed from, as a label.
	FromDevice                 string              `json:"from_device"`
	CustomerCancelReasonDetail *CancelReasonDetail `json:"customer_cancel_reason_detail"`
	SerialNumbers              []string            `json:"serial_numbers"`
	OrderWeight                float64             `json:"order_weight"`
	CreatedAt                  cyberbiz.Time       `json:"created_at"`
	UpdatedAt                  cyberbiz.Time       `json:"updated_at"`
}

// Buyer is the purchasing member's contact details.
type Buyer struct {
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

// Receiver is the delivery recipient.
type Receiver struct {
	Name               string         `json:"name"`
	CountryCallingCode string         `json:"country_calling_code"`
	Phone              string         `json:"phone"`
	Address            string         `json:"address"`
	DetailAddress      *DetailAddress `json:"detail_address"`
	// CVSStoreID is the convenience store pickup point.
	CVSStoreID        string `json:"cvs_store_id"`
	AllpayLogisticsID string `json:"allpay_logistics_id"`
}

// LineItem is one product line on an order or a fulfillment.
type LineItem struct {
	ID               int64  `json:"id"`
	ProductID        int64  `json:"product_id"`
	ProductVariantID int64  `json:"product_variant_id"`
	Title            string `json:"title"`
	VariantTitle     string `json:"variant_title"`
	SKU              string `json:"sku"`
	// QC is the vendor's own item code.
	QC       string         `json:"qc"`
	Vendor   string         `json:"vendor"`
	Price    cyberbiz.Money `json:"price"`
	Cost     cyberbiz.Money `json:"cost"`
	Quantity int            `json:"quantity"`
	// ItemType is e.g. "normal" or "gift".
	ItemType                  string                `json:"item_type"`
	ReturnStatus              cyberbiz.ReturnStatus `json:"return_status"`
	DiscountName              string                `json:"discount_name"`
	Discounts                 []LineItemDiscount    `json:"discounts"`
	TotalPriceBeforeDiscounts cyberbiz.Money        `json:"total_price_before_discounts"`
	TotalDiscount             cyberbiz.Money        `json:"total_discount"`
	TotalPriceAfterDiscounts  cyberbiz.Money        `json:"total_price_after_discounts"`
	// TaxTypeID is inclusive_tax, zero_tax or exclusive_tax.
	TaxTypeID string `json:"tax_type_id"`
	// BonusRedemptionPrice is the points paid for a bonus mall item, 0
	// otherwise.
	BonusRedemptionPrice cyberbiz.Money `json:"bonus_redemption_price"`
	RelatedItems         []RelatedItems `json:"related_items"`
	Channel              string         `json:"channel"`
	// Weight is documented as a string but delivered as a number.
	Weight    float64       `json:"weight"`
	Photo     string        `json:"photo"`
	CreatedAt cyberbiz.Time `json:"created_at"`
}

// LineItemDiscount is one discount applied to a line item.
type LineItemDiscount struct {
	// Position is the index of the unit the discount applies to.
	Position int            `json:"position"`
	ID       int64          `json:"id"`
	Code     string         `json:"code"`
	Name     string         `json:"name"`
	Discount cyberbiz.Money `json:"discount"`
}

// RelatedItems describes the components of a bundle line item.
type RelatedItems struct {
	Quantity int         `json:"quantity"`
	Items    []ComboItem `json:"items"`
}

// ComboItem is one component of a bundle.
type ComboItem struct {
	ID                          int64          `json:"id"`
	ProductID                   int64          `json:"product_id"`
	ProductVariantID            int64          `json:"product_variant_id"`
	Title                       string         `json:"title"`
	VariantTitle                string         `json:"variant_title"`
	SKU                         string         `json:"sku"`
	QC                          string         `json:"qc"`
	Vendor                      string         `json:"vendor"`
	Price                       cyberbiz.Money `json:"price"`
	Cost                        cyberbiz.Money `json:"cost"`
	Quantity                    int            `json:"quantity"`
	ComboProductPriceDifference cyberbiz.Money `json:"combo_product_price_difference"`
}

// ShippingVendor is the carrier.
type ShippingVendor struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// Fulfillment is one shipment of an order.
type Fulfillment struct {
	ID              int64                      `json:"id"`
	TrackingCompany string                     `json:"tracking_company"`
	TrackingNumber  string                     `json:"tracking_number"`
	FulfilledAt     cyberbiz.Time              `json:"fulfilled_at"`
	ReceivedAt      cyberbiz.Time              `json:"received_at"`
	Status          cyberbiz.FulfillmentStatus `json:"status"`
	LineItems       []LineItem                 `json:"line_items"`
	// TrackingURL is only provided for Uber Direct and Pandago.
	TrackingURL string `json:"tracking_url"`
}

// PaymentInfo is one part of a split payment.
type PaymentInfo struct {
	Name   string         `json:"name"`
	Amount cyberbiz.Money `json:"amount"`
}

// Prices is the order's price breakdown.
type Prices struct {
	TotalLineItemsPrice cyberbiz.Money  `json:"total_line_items_price"`
	ShippingRatePrice   cyberbiz.Money  `json:"shipping_rate_price"`
	Discounts           *DiscountDetail `json:"discounts"`
	TotalPrice          cyberbiz.Money  `json:"total_price"`
}

// DiscountDetail itemises an order's discounts.
type DiscountDetail struct {
	SpecialCollectionDiscount cyberbiz.Money `json:"special_collection_discount"`
	VIPDiscount               cyberbiz.Money `json:"vip_discount"`
	ShopDiscount              *ShopDiscount  `json:"shop_discount"`
	// CouponDiscount is deprecated in favour of CouponDiscounts.
	CouponDiscount         *CouponDiscount  `json:"coupon_discount"`
	CouponDiscounts        []CouponDiscount `json:"coupon_discounts"`
	BonusConsumed          cyberbiz.Money   `json:"bonus_consumed"`
	VIPShippingDiscount    cyberbiz.Money   `json:"vip_shipping_discount"`
	CouponShippingDiscount cyberbiz.Money   `json:"coupon_shipping_discount"`
	// PriceDiscount is a manual discount by the shop.
	PriceDiscount      cyberbiz.Money `json:"price_discount"`
	ThirdPartyDiscount cyberbiz.Money `json:"third_party_discount"`
}

// ShopDiscount is a storewide promotion applied to an order.
type ShopDiscount struct {
	Name   string         `json:"name"`
	Amount cyberbiz.Money `json:"amount"`
}

// CouponDiscount is one coupon applied to an order.
type CouponDiscount struct {
	ID       int64          `json:"id"`
	Name     string         `json:"name"`
	Code     string         `json:"code"`
	Amount   cyberbiz.Money `json:"amount"`
	CouponID int64          `json:"coupon_id"`
}

// EInvoice is the Taiwanese e-invoice issued for an order.
type EInvoice struct {
	Title         string                 `json:"title"`
	OrderID       int64                  `json:"order_id"`
	CompanyNo     string                 `json:"company_no"`
	InvoiceNo     string                 `json:"invoice_no"`
	InvoiceStatus cyberbiz.InvoiceStatus `json:"invoice_status"`
	InvoiceAt     cyberbiz.Time          `json:"invoice_at"`
	// InvalidAt is when the invoice was voided or allowanced; nil if never.
	InvalidAt    *cyberbiz.Time       `json:"invalid_at"`
	RandomNum    string               `json:"random_num"`
	InvoiceType  cyberbiz.InvoiceType `json:"invoice_type"`
	LoveCode     string               `json:"love_code"`
	PhoneBarcode string               `json:"phone_barcode"`
	NaturePerson string               `json:"nature_person"`
}

// OrderStatuses groups an order's four status dimensions.
type OrderStatuses struct {
	OrderStatus       cyberbiz.OrderStatus       `json:"order_status"`
	FinancialStatus   cyberbiz.FinancialStatus   `json:"financial_status"`
	FulfillmentStatus cyberbiz.FulfillmentStatus `json:"fulfillment_status"`
	ReturnStatus      cyberbiz.ReturnStatus      `json:"return_status"`
}

// OrderTimings records when each lifecycle step happened; nil means it has
// not.
type OrderTimings struct {
	RequestReturnAt *cyberbiz.Time `json:"request_return_at"`
	ReturnAt        *cyberbiz.Time `json:"return_at"`
	RefundAt        *cyberbiz.Time `json:"refund_at"`
	ClosedAt        *cyberbiz.Time `json:"closed_at"`
	CancelledAt     *cyberbiz.Time `json:"cancelled_at"`
	ExpiredAt       *cyberbiz.Time `json:"expired_at"`
	ConfirmedAt     *cyberbiz.Time `json:"confirmed_at"`
}

// ReturnHistory is one refund on an order.
type ReturnHistory struct {
	Body       string         `json:"body"`
	Price      cyberbiz.Money `json:"price"`
	RefundedAt cyberbiz.Time  `json:"refunded_at"`
}

// BranchStore is a pickup store or express-delivery store.
type BranchStore struct {
	StoreNo       string         `json:"store_no"`
	Name          string         `json:"name"`
	Phone         string         `json:"phone"`
	County        string         `json:"county"`
	District      string         `json:"district"`
	Address       string         `json:"address"`
	Zip           string         `json:"zip"`
	OpeningHours  string         `json:"opening_hours"`
	Lat           float64        `json:"lat"`
	Lng           float64        `json:"lng"`
	Enabled       bool           `json:"enabled"`
	ShippingRates []ShippingRate `json:"shipping_rates"`
	// SourceType is "BranchStore" or "PosShop".
	SourceType string `json:"source_type"`
	SourceID   int64  `json:"source_id"`
}

// ShippingRate is a store's shipping fee rule.
type ShippingRate struct {
	ID               int64          `json:"id"`
	Name             string         `json:"name"`
	MinOrderSubtotal cyberbiz.Money `json:"min_order_subtotal"`
	Price            cyberbiz.Money `json:"price"`
	Payments         []Payment      `json:"payments"`
}

// Payment is a payment method allowed by a shipping rate.
type Payment struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// POSInfo identifies the point-of-sale terminal and clerk for a POS order.
type POSInfo struct {
	POSUserID    int64  `json:"pos_user_id"`
	POSUserEmail string `json:"pos_user_email"`
	POSShopID    int64  `json:"pos_shop_id"`
	POSInfo      string `json:"pos_info"`
	POSID        int64  `json:"pos_id"`
	POSName      string `json:"pos_name"`
}

// LinkedOrderInfo carries affiliate attribution.
type LinkedOrderInfo struct {
	Source                string `json:"source"`
	ShopdotcomRID         string `json:"shopdotcom_rid"`
	ShopdotcomClickID     string `json:"shopdotcom_click_id"`
	LineShoppingECID      string `json:"line_shopping_ecid"`
	LineShoppingAffiliate string `json:"line_shopping_affiliate"`
	IChannelGID           string `json:"ichannel_gid"`
}

// CancelReasonDetail is why the customer cancelled.
type CancelReasonDetail struct {
	Source       string `json:"source"`
	ReasonID     int64  `json:"reason_id"`
	ReasonDetail string `json:"reason_detail"`
}

// ExchangeHistory is one in-store exchange on a POS order.
type ExchangeHistory struct {
	CreatedAt            cyberbiz.Time      `json:"created_at"`
	Price                cyberbiz.Money     `json:"price"`
	OrderPriceBefore     cyberbiz.Money     `json:"order_price_before"`
	OrderPriceAfter      cyberbiz.Money     `json:"order_price_after"`
	POSShopID            int64              `json:"pos_shop_id"`
	POSID                int64              `json:"pos_id"`
	PaymentName          string             `json:"payment_name"`
	PaymentMethod        string             `json:"payment_method"`
	MultiplePaymentInfos []PaymentInfo      `json:"multiple_payment_infos"`
	LineItems            []ExchangeLineItem `json:"line_items"`
	EInvoice             *EInvoice          `json:"einvoice"`
}

// ExchangeLineItem is one item in an exchange.
type ExchangeLineItem struct {
	ProductVariantID int64          `json:"product_variant_id"`
	Name             string         `json:"name"`
	SKU              string         `json:"sku"`
	QC               string         `json:"qc"`
	Price            cyberbiz.Money `json:"price"`
	Quantity         int            `json:"quantity"`
}

// ProductPayload is the body of products/* events.
type ProductPayload struct {
	ID                int64         `json:"id"`
	Title             string        `json:"title"`
	EnglishTitle      string        `json:"english_title"`
	ProductURL        string        `json:"product_url"`
	Published         bool          `json:"published"`
	SellFrom          cyberbiz.Time `json:"sell_from"`
	SellTo            cyberbiz.Time `json:"sell_to"`
	ProductType       string        `json:"product_type"`
	ProductTypeCode   string        `json:"product_type_code"`
	Slogan            string        `json:"slogan"`
	Brief             string        `json:"brief"`
	BriefText         string        `json:"brief_text"`
	BriefIncludesHTML bool          `json:"brief_includes_html"`
	BodyHTML          string        `json:"body_html"`
	Vendor            string        `json:"vendor"`
	// Price is the lowest variant price.
	Price cyberbiz.Money `json:"price"`
	// SellWeight is the quantity sold.
	SellWeight int `json:"sell_weight"`
	// TaxTypeID is inclusive_tax, zero_tax or exclusive_tax.
	TaxTypeID         string             `json:"tax_type_id"`
	CustomCollections []CustomCollection `json:"custom_collections"`
	SpecialCollection *SpecialCollection `json:"special_collection"`
	Tags              []ProductTag       `json:"tags"`
	ProductVariants   []VariantPayload   `json:"product_variants"`
	ProductOptions    []ProductOption    `json:"product_options"`
	// POSShop's shape is not documented.
	POSShop   jsontext.Value `json:"pos_shop"`
	PhotoURLs []string       `json:"photo_urls"`
	Photos    []ProductPhoto `json:"photos"`
	Channel   *Channel       `json:"channel"`
	// RelatedCollections' item shape is not documented.
	RelatedCollections      jsontext.Value `json:"related_collections"`
	BranchStore             *BranchStore   `json:"branch_store"`
	TemperatureTypes        []string       `json:"temperature_types"`
	Searchable              bool           `json:"searchable"`
	GoogleProductCategoryID int64          `json:"google_product_category_id"`
	// ProductCustomFields and SEOMetaTags are free-form.
	ProductCustomFields  jsontext.Value `json:"product_custom_fields"`
	SEOMetaTags          jsontext.Value `json:"seo_meta_tags"`
	RequiredCustomerTags []string       `json:"required_customer_tags"`
	CreatedAt            cyberbiz.Time  `json:"created_at"`
	UpdatedAt            cyberbiz.Time  `json:"updated_at"`
}

// ProductTag is a tag on a product.
type ProductTag struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category int    `json:"category"`
}

// ProductPhoto is one product image.
type ProductPhoto struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	Position int    `json:"position"`
}

// Channel is a sales channel.
type Channel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ProductOption is one option axis (e.g. colour) of a product.
type ProductOption struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Types    string `json:"types"`
}

// CustomCollection is a manually curated collection containing the
// product. The documented "products" array is omitted; decode it yourself
// if needed.
type CustomCollection struct {
	ID                int64  `json:"id"`
	Title             string `json:"title"`
	Handle            string `json:"handle"`
	Published         bool   `json:"published"`
	BodyHTML          string `json:"body_html"`
	ProductsOrderName string `json:"products_order_name"`
	Position          int    `json:"position"`
}

// SpecialCollection is a promotion collection containing the product.
type SpecialCollection struct {
	ID                    int64                  `json:"id"`
	Title                 string                 `json:"title"`
	Handle                string                 `json:"handle"`
	Published             bool                   `json:"published"`
	StartDate             cyberbiz.Time          `json:"start_date"`
	EndDate               cyberbiz.Time          `json:"end_date"`
	BodyHTML              string                 `json:"body_html"`
	Position              int                    `json:"position"`
	SpecialCollectionType *SpecialCollectionType `json:"special_collection_type"`
	TypeRules             []TypeRule             `json:"type_rules"`
	RestIncludeDiscount   bool                   `json:"rest_include_discount"`
}

// SpecialCollectionType classifies a promotion.
type SpecialCollectionType struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// TypeRule is one pricing rule of a promotion.
type TypeRule struct {
	ID         int64          `json:"id"`
	Quantity   int            `json:"quantity"`
	Price      cyberbiz.Money `json:"price"`
	Percentage int            `json:"percentage"`
}

// VariantPayload is the body of variants/* events and each entry of a
// product's product_variants.
type VariantPayload struct {
	ID             int64          `json:"id"`
	ProductID      int64          `json:"product_id"`
	Name           string         `json:"name"`
	Position       int            `json:"position"`
	Price          cyberbiz.Money `json:"price"`
	Cost           cyberbiz.Money `json:"cost"`
	CompareAtPrice cyberbiz.Money `json:"compare_at_price"`
	// Meas is the volumetric size.
	Meas                    float64 `json:"meas"`
	MaxUsableBonus          float64 `json:"max_usable_bonus"`
	Weight                  float64 `json:"weight"`
	Option1                 string  `json:"option1"`
	Option2                 string  `json:"option2"`
	Option3                 string  `json:"option3"`
	InventoryManagement     bool    `json:"inventory_management"`
	InventoryQuantity       int     `json:"inventory_quantity"`
	Sold                    int     `json:"sold"`
	SafetyInventoryQuantity int     `json:"safety_inventory_quantity"`
	// InventoryPolicy says whether purchases continue at zero stock.
	InventoryPolicy  string        `json:"inventory_policy"`
	SKU              string        `json:"sku"`
	QC               string        `json:"qc"`
	RequiresShipping bool          `json:"requires_shipping"`
	HoneycombSync    bool          `json:"honeycomb_sync"`
	Vendor           string        `json:"vendor"`
	PhotoURLs        []string      `json:"photo_urls"`
	PIMInfos         []PIMInfo     `json:"pim_infos"`
	CreatedAt        cyberbiz.Time `json:"created_at"`
	UpdatedAt        cyberbiz.Time `json:"updated_at"`
}

// PIMInfo links a variant to a product information management channel.
type PIMInfo struct {
	ID               int64  `json:"id"`
	ProductID        int64  `json:"product_id"`
	ProductVariantID int64  `json:"product_variant_id"`
	PIMProductID     int64  `json:"pim_product_id"`
	PIMVariantID     int64  `json:"pim_variant_id"`
	Channel          int    `json:"channel"`
	ChannelShopName  string `json:"channel_shop_name"`
	IsConnected      bool   `json:"is_connected"`
	IsSource         bool   `json:"is_source"`
	ShopID           int64  `json:"shop_id"`
}

// CouponPayload is the body of coupons/* events: one coupon held by a
// customer, or a storewide coupon when CustomerID is 0.
type CouponPayload struct {
	CustomerID int64  `json:"customer_id"`
	Title      string `json:"title"`
	Code       string `json:"code"`
	// CouponTypeName is the localised type label: amount, percentage or
	// free shipping.
	CouponTypeName string `json:"coupon_type_name"`
	// CouponValue is the discount amount, or the percentage for percentage
	// coupons. It is documented as a string; Money decodes both strings and
	// numbers.
	CouponValue         cyberbiz.Money `json:"coupon_value"`
	OrderPriceThreshold cyberbiz.Money `json:"order_price_threshold"`
	StartDate           cyberbiz.Time  `json:"start_date"`
	EndDate             cyberbiz.Time  `json:"end_date"`
	// ConcurrentlyApply says whether the coupon combines with promotions;
	// documented as a string.
	ConcurrentlyApply  string `json:"concurrently_apply"`
	UsageLimit         int    `json:"usage_limit"`
	CanAccumulateBonus bool   `json:"can_accumulate_bonus"`
	UsageUnlimited     bool   `json:"usage_unlimited"`
	UsedTimes          int    `json:"used_times"`
	// GiftOrderID is the order that granted the coupon, 0 when none.
	GiftOrderID              int64 `json:"gift_order_id"`
	GiftDays                 int   `json:"gift_days"`
	AccountUsageLimitEnabled bool  `json:"account_usage_limit_enabled"`
	AccountUsageLimit        int   `json:"account_usage_limit"`
	// RestrictStrategy is unrestricted, restrict or forbidden.
	RestrictStrategy  string   `json:"restrict_strategy"`
	RestrictCampaigns []string `json:"restrict_campaigns"`
	// Tags are the product tags the coupon is limited to.
	Tags              []string                 `json:"tags"`
	POSShopIDs        []int64                  `json:"pos_shop_ids"`
	CouponStatus      cyberbiz.CouponStatus    `json:"coupon_status"`
	GiftOrderStatus   cyberbiz.GiftOrderStatus `json:"gift_order_status"`
	Valid             bool                     `json:"valid"`
	CustomerUsedTimes int                      `json:"customer_used_times"`
	CustomerUsable    bool                     `json:"customer_usable"`
}

// Usable applies the documented combination of CouponStatus and
// GiftOrderStatus; see [cyberbiz.CouponUsable].
func (c *CouponPayload) Usable() bool {
	return cyberbiz.CouponUsable(c.CouponStatus, c.GiftOrderStatus)
}

// VIPLevelPayload is the body of customer_vip_level/* events.
type VIPLevelPayload struct {
	CustomerID   int64         `json:"customer_id"`
	CurrentGroup *VIPGroup     `json:"current_group"`
	CurrentLevel *VIPLevel     `json:"current_level"`
	NextLevel    *VIPLevel     `json:"next_level"`
	ExtraInfo    *VIPExtraInfo `json:"extra_info"`
}

// VIPGroup is a VIP programme with ordered levels.
type VIPGroup struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Position       int        `json:"position"`
	DescriptionURL string     `json:"description_url"`
	CustomerTags   []string   `json:"customer_tags"`
	VIPGroupLevels []VIPLevel `json:"vip_group_levels"`
}

// VIPLevel is one tier of a VIP group and its benefits.
type VIPLevel struct {
	ID                                       int64          `json:"id"`
	Position                                 int            `json:"position"`
	Name                                     string         `json:"name"`
	ValidityDays                             int            `json:"validity_days"`
	UpgradeConditionTotalSpent               cyberbiz.Money `json:"upgrade_condition_total_spent"`
	UpgradeConditionTotalSpentInValidityDays cyberbiz.Money `json:"upgrade_condition_total_spent_in_validity_days"`
	RenewalConditionTotalSpent               cyberbiz.Money `json:"renewal_condition_total_spent"`
	RenewalConditionTotalSpentInValidityDays cyberbiz.Money `json:"renewal_condition_total_spent_in_validity_days"`
	BonusPointEnabled                        bool           `json:"bonus_point_enabled"`
	BonusPointThreshold                      cyberbiz.Money `json:"bonus_point_threshold"`
	BonusPointValue                          int            `json:"bonus_point_value"`
	BonusPointExpiryDays                     int            `json:"bonus_point_expiry_days"`
	BirthGiftEnabled                         bool           `json:"birth_gift_enabled"`
	BirthGiftName                            string         `json:"birth_gift_name"`
	// BirthGiftBeforeDays is documented as a string.
	BirthGiftBeforeDays  string          `json:"birth_gift_before_days"`
	BirthGiftSetting     *VIPGiftSetting `json:"birth_gift_setting"`
	UpgradeGiftEnabled   bool            `json:"upgrade_gift_enabled"`
	UpgradeGiftSetting   *VIPGiftSetting `json:"upgrade_gift_setting"`
	OrderDiscountEnabled bool            `json:"order_discount_enabled"`
	// OrderDiscountValue is a percentage.
	OrderDiscountValue int `json:"order_discount_value"`
	// OrderDiscountSetting and FreeShippingSetting have undocumented shapes.
	OrderDiscountSetting jsontext.Value `json:"order_discount_setting"`
	FreeShippingEnabled  bool           `json:"free_shipping_enabled"`
	FreeShippingSetting  jsontext.Value `json:"free_shipping_setting"`
}

// VIPGiftSetting configures a birthday or upgrade gift.
type VIPGiftSetting struct {
	Bonus  *VIPBonusSetting  `json:"bonus"`
	Coupon *VIPCouponSetting `json:"coupon"`
}

// VIPBonusSetting is the bonus point part of a gift.
type VIPBonusSetting struct {
	Enabled bool `json:"enabled"`
	Value   int  `json:"value"`
	// ExpiryDays of 0 means the points never expire.
	ExpiryDays int `json:"expiry_days"`
}

// VIPCouponSetting is the coupon part of a gift.
type VIPCouponSetting struct {
	Enabled bool              `json:"enabled"`
	Presets *VIPCouponPresets `json:"presets"`
}

// VIPCouponPresets is the template for a gifted coupon.
type VIPCouponPresets struct {
	// CouponTypeID is 1 for an amount, 2 for a percentage.
	CouponTypeID        int            `json:"coupon_type_id"`
	Value               cyberbiz.Money `json:"value"`
	Code                string         `json:"code"`
	OrderPriceThreshold cyberbiz.Money `json:"order_price_threshold"`
	UsableDays          int            `json:"usable_days"`
	UsageLimit          int            `json:"usage_limit"`
	ProductTags         []string       `json:"product_tags"`
	RestrictStrategy    string         `json:"restrict_strategy"`
	RestrictCampaigns   []string       `json:"restrict_campaigns"`
}

// VIPExtraInfo is the customer's standing within their level.
type VIPExtraInfo struct {
	StartAt                                        cyberbiz.Time  `json:"start_at"`
	EndAt                                          cyberbiz.Time  `json:"end_at"`
	DifferenceOfTotalSpentInValidityDaysForRenewal cyberbiz.Money `json:"difference_of_total_spent_in_validity_days_for_renewal"`
	DifferenceOfTotalSpentForRenewal               cyberbiz.Money `json:"difference_of_total_spent_for_renewal"`
	DifferenceOfTotalSpentInValidityDaysForUpgrade cyberbiz.Money `json:"difference_of_total_spent_in_validity_days_for_upgrade"`
	DifferenceOfTotalSpentForUpgrade               cyberbiz.Money `json:"difference_of_total_spent_for_upgrade"`
}

// AppPayload is the body of apps/uninstall.
type AppPayload struct {
	AppUUID        string `json:"app_uuid"`
	AppVersionUUID string `json:"app_version_uuid"`
	AppClientID    string `json:"app_client_id"`
}
