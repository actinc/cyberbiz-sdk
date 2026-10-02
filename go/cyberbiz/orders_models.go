package cyberbiz

// Order is one shop order as returned by GET /v1/orders and
// GET /v1/orders/{id}. The list and detail shapes are identical except that
// the list also carries Token.
type Order struct {
	ID            int64 `json:"id"`
	SubtotalPrice Money `json:"subtotal_price"` // line items only, before shipping
	CreatedAt     Time  `json:"created_at"`
	UpdatedAt     Time  `json:"updated_at"`
	// OrderNumber is the shop-facing sequential number; the swagger documents
	// it as a string but the platform sends an integer.
	OrderNumber int64  `json:"order_number"`
	OrderName   string `json:"order_name"` // e.g. "#1101"

	Customer       *OrderCustomer `json:"customer"`
	Buyer          *Buyer         `json:"buyer"`
	Receiver       *Receiver      `json:"receiver"`
	BillingAddress *Address       `json:"billing_address"`
	LineItems      []LineItem     `json:"line_items"`

	ShippingType   string          `json:"shipping_type"`
	ShippingName   string          `json:"shipping_name"`
	ShippingVendor *ShippingVendor `json:"shipping_vendor"`
	LogisticsID    string          `json:"logistics_id"`  // ECPay logistics order id
	DeliveryDate   Date            `json:"delivery_date"` // requested delivery day, custom feature
	DeliveryTime   int             `json:"delivery_time"` // requested delivery slot 0-3, custom feature
	Delegate       string          `json:"delegate"`      // staff email when the order was placed on the customer's behalf
	Fulfillments   []Fulfillment   `json:"fulfillments"`

	PaymentName          string             `json:"payment_name"`
	PaymentMethod        string             `json:"payment_method"`
	PaymentURL           string             `json:"payment_url"`
	MultiplePaymentInfos []OrderPaymentInfo `json:"multiple_payment_infos"` // POS split payments
	Prices               *Prices            `json:"prices"`
	Card4No              string             `json:"card4no"` // last four digits of the card
	TransactionNumber    string             `json:"transaction_number"`
	MerchantTradeNo      string             `json:"merchant_trade_no"`

	Einvoice                 *OrderEinvoice `json:"einvoice"`
	PaperInvoiceNo           string         `json:"paper_invoice_no"`
	PaperCompanyNo           string         `json:"paper_company_no"`
	PrepaymentPaperInvoiceNo string         `json:"prepayment_paper_invoice_no"`
	PrepaymentPaperCompanyNo string         `json:"prepayment_paper_company_no"`

	Statuses        *Statuses       `json:"statuses"`
	Timings         *Timings        `json:"timings"`
	ReturnHistories []ReturnHistory `json:"return_histories"`
	Note            string          `json:"note"`

	BranchStore              *OrderBranchStore `json:"branch_store"` // pickup store
	ReferralCode             string            `json:"referral_code"`
	CheckoutReferralCode     string            `json:"checkout_referral_code"`
	CheckoutReferralUserName string            `json:"checkout_referral_user_name"`
	RegisterReferralCode     string            `json:"register_referral_code"`
	// TotalBonusRedemptionPrice is the total bonus points spent in the bonus
	// mall, in the shop currency.
	TotalBonusRedemptionPrice Money `json:"total_bonus_redemption_price"`

	PosInfo           *PosInfo          `json:"pos_info"`
	ExchangeHistories []ExchangeHistory `json:"exchange_histories"`
	LinkedOrderInfo   *LinkedOrderInfo  `json:"linked_order_info"`
	Tags              []Tag             `json:"tags"`

	ExpressDeliveryBranchStore *OrderBranchStore          `json:"express_delivery_branch_store"`
	ShippingStatus             string                     `json:"shipping_status"`
	ExtraInfo                  string                     `json:"extra_info"`
	FromDevice                 string                     `json:"from_device"` // e.g. 桌機, 手機
	CustomerCancelReasonDetail *OrderCustomerCancelReason `json:"customer_cancel_reason_detail"`
	SerialNumbers              []string                   `json:"serial_numbers"` // campaign serial numbers
	OrderWeight                float64                    `json:"order_weight"`
	WarehouseTypeID            int64                      `json:"warehouse_type_id"`
	UTMTracking                *UTMTracking               `json:"utm_tracking"`
	// Token is the order's public token, present on the list endpoint only.
	Token string `json:"token"`
}

// OrderCustomer is the customer record embedded in an order. It is the same
// shape as the customers resource; see the Customers service for the full
// model.
type OrderCustomer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Status is the account state: pending, validate, enabled, disabled,
	// invited, declined or warning.
	Status                               string                     `json:"status"`
	Email                                string                     `json:"email"`
	CountryCallingCode                   string                     `json:"country_calling_code"`
	Mobile                               string                     `json:"mobile"`
	EnableCVSPickup                      bool                       `json:"enable_cvs_pickup"`
	EnableCVSCOD                         bool                       `json:"enable_cvs_cod"`
	EnableHomeDeliveryCOD                bool                       `json:"enable_home_delivery_cod"`
	AcceptsMarketing                     bool                       `json:"accepts_marketing"`
	AcceptsEmailNotification             bool                       `json:"accepts_email_notification"`
	Tags                                 []Tag                      `json:"tags"`
	Address                              *Address                   `json:"address"`
	Gender                               string                     `json:"gender"`
	Birthday                             Date                       `json:"birthday"`
	OtherAccumulatedConsumption          Money                      `json:"other_accumulated_consumption"`
	OtherAccumulatedConsumptionExpiredAt Time                       `json:"other_accumulated_consumption_expired_at"`
	Note                                 string                     `json:"note"`
	CustomFields                         []OrderCustomField         `json:"custom_fields"`
	CreatedAt                            Time                       `json:"created_at"`
	UpdatedAt                            Time                       `json:"updated_at"`
	ConfirmedAt                          *Time                      `json:"confirmed_at"`            // email verified
	MobileSMSConfirmedAt                 *Time                      `json:"mobile_sms_confirmed_at"` // mobile verified
	BonusRemain                          Money                      `json:"bonus_remain"`
	UIDProviders                         []OrderCustomerUIDProvider `json:"uid_providers"`
}

// OrderCustomerUIDProvider is a social login identity linked to a customer.
type OrderCustomerUIDProvider struct {
	ProviderType string `json:"provider_type"` // e.g. line, facebook
	UID          string `json:"uid"`
}

// OrderCustomField is a shop-defined field value on a customer or line item.
type OrderCustomField struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// Buyer is the member account that placed the order.
type Buyer struct {
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

// Address is a postal address with contact details, used for the billing
// address and the customer's address. Company is only set on the latter.
type Address struct {
	Name               string         `json:"name"`
	Company            string         `json:"company"`
	CountryCallingCode string         `json:"country_calling_code"`
	Phone              string         `json:"phone"`
	Address            string         `json:"address"` // single-line full address
	DetailAddress      *DetailAddress `json:"detail_address"`
}

// DetailAddress is an address split into its components.
type DetailAddress struct {
	Zip      string `json:"zip"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
}

// Receiver is the shipping recipient of an order.
type Receiver struct {
	Name               string         `json:"name"`
	CountryCallingCode string         `json:"country_calling_code"`
	Phone              string         `json:"phone"`
	Address            string         `json:"address"`
	DetailAddress      *DetailAddress `json:"detail_address"`
	CVSStoreID         string         `json:"cvs_store_id"`        // convenience store number for pickup
	AllpayLogisticsID  string         `json:"allpay_logistics_id"` // ECPay logistics id
}

// ShippingVendor is the carrier assigned to an order.
type ShippingVendor struct {
	Type string `json:"type"` // carrier code, e.g. custom, ezcat
	Name string `json:"name"`
}

// OrderItemType classifies a line item.
type OrderItemType string

// Known OrderItemType values (item_type of a line item).
const (
	// OrderItemTypeNormal is a regular purchased product.
	OrderItemTypeNormal OrderItemType = "normal"
	// OrderItemTypeNo marks an item that is not a normal purchase, as seen in
	// the CYBERBIZ v2 samples.
	OrderItemTypeNo OrderItemType = "no"
)

// TaxType is the tax category of a line item.
type TaxType string

// Known TaxType values (tax_type of a line item or product).
const (
	TaxTypeInclusive TaxType = "inclusive_tax" // taxable
	TaxTypeZero      TaxType = "zero_tax"      // zero-rated
	TaxTypeExclusive TaxType = "exclusive_tax" // tax-exempt
)

// LineItem is one product variant on an order, fulfillment or return.
type LineItem struct {
	ID                        int64              `json:"id"`
	ProductID                 int64              `json:"product_id"`
	ProductVariantID          int64              `json:"product_variant_id"`
	Title                     string             `json:"title"`
	VariantTitle              string             `json:"variant_title"`
	SKU                       string             `json:"sku"`
	QC                        string             `json:"qc"` // vendor's own item code
	Vendor                    string             `json:"vendor"`
	Price                     Money              `json:"price"` // unit price
	Cost                      Money              `json:"cost"`  // unit cost
	Quantity                  int                `json:"quantity"`
	ItemType                  OrderItemType      `json:"item_type"`
	ReturnStatus              ReturnStatus       `json:"return_status"`
	DiscountName              string             `json:"discount_name"`
	Discounts                 []LineItemDiscount `json:"discounts"`
	TotalPriceBeforeDiscounts Money              `json:"total_price_before_discounts"`
	TotalDiscount             Money              `json:"total_discount"`
	TotalPriceAfterDiscounts  Money              `json:"total_price_after_discounts"`
	TaxTypeID                 TaxType            `json:"tax_type_id"`
	// BonusRedemptionPrice is the bonus points redeemed for this item; zero
	// when the item was not a bonus mall redemption.
	BonusRedemptionPrice Money              `json:"bonus_redemption_price"`
	RelatedItems         []RelatedItems     `json:"related_items"` // components of a combo product
	CreatedAt            Time               `json:"created_at"`
	Channel              string             `json:"channel"`
	Weight               float64            `json:"weight"`
	Photo                string             `json:"photo"` // thumbnail path
	CustomFields         []OrderCustomField `json:"custom_fields"`
}

// LineItemDiscount is one discount applied to a line item.
type LineItemDiscount struct {
	Position int    `json:"position"`
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Discount Money  `json:"discount"`
}

// RelatedItems is one set of components of a combo product line item.
type RelatedItems struct {
	Quantity int           `json:"quantity"` // number of combo sets
	Items    []RelatedItem `json:"items"`
}

// RelatedItem is one component variant inside a combo product.
type RelatedItem struct {
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
	// ComboProductPriceDifference is the enterprise-only price difference
	// between the component and its share of the combo price.
	ComboProductPriceDifference  Money   `json:"combo_product_price_difference"`
	ComboProductPriceDiffDetails []Money `json:"combo_product_price_diff_details"`
}

// Prices is the price breakdown of an order.
type Prices struct {
	TotalLineItemsPrice Money           `json:"total_line_items_price"`
	ShippingRatePrice   Money           `json:"shipping_rate_price"`
	Discounts           *OrderDiscounts `json:"discounts"`
	TotalPrice          Money           `json:"total_price"`
}

// OrderDiscounts itemises every discount applied to an order.
type OrderDiscounts struct {
	SpecialCollectionDiscount Money         `json:"special_collection_discount"` // campaign discount
	VIPDiscount               Money         `json:"vip_discount"`
	ShopDiscount              *ShopDiscount `json:"shop_discount"` // shop-wide campaign
	// CouponDiscount is the single-coupon form, deprecated by CYBERBIZ in
	// favour of CouponDiscounts.
	CouponDiscount         *CouponDiscount  `json:"coupon_discount"`
	CouponDiscounts        []CouponDiscount `json:"coupon_discounts"`
	BonusConsumed          Money            `json:"bonus_consumed"`
	VIPShippingDiscount    Money            `json:"vip_shipping_discount"`
	CouponShippingDiscount Money            `json:"coupon_shipping_discount"`
	PriceDiscount          Money            `json:"price_discount"` // manual discount by staff
	ThirdPartyDiscount     Money            `json:"third_party_discount"`
}

// ShopDiscount is a shop-wide campaign discount on an order.
type ShopDiscount struct {
	Name   string `json:"name"`
	Amount Money  `json:"amount"`
}

// CouponDiscount is one coupon applied to an order. ID is only present in
// the coupon_discounts list.
type CouponDiscount struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Amount   Money  `json:"amount"`
	CouponID int64  `json:"coupon_id"`
}

// OrderPaymentInfo is one payment of a split (multi-method) POS payment.
type OrderPaymentInfo struct {
	Name   string `json:"name"`
	Amount Money  `json:"amount"`
}

// OrderEinvoice is the electronic invoice attached to an order or exchange.
type OrderEinvoice struct {
	Title         string        `json:"title"` // invoice title (buyer name)
	OrderID       int64         `json:"order_id"`
	CompanyNo     string        `json:"company_no"` // buyer's tax id
	InvoiceNo     string        `json:"invoice_no"`
	InvoiceStatus InvoiceStatus `json:"invoice_status"`
	InvoiceAt     Time          `json:"invoice_at"`
	InvalidAt     *Time         `json:"invalid_at"` // voided or allowance time, nil while valid
	RandomNum     string        `json:"random_num"`
	InvoiceType   InvoiceType   `json:"invoice_type"`
	LoveCode      string        `json:"love_code"` // donation code
	PhoneBarcode  string        `json:"phone_barcode"`
	NaturePerson  string        `json:"nature_person"` // citizen digital certificate
}

// Statuses groups the four state machines of an order.
type Statuses struct {
	OrderStatus       OrderStatus       `json:"order_status"`
	FinancialStatus   FinancialStatus   `json:"financial_status"`
	FulfillmentStatus FulfillmentStatus `json:"fulfillment_status"`
	ReturnStatus      ReturnStatus      `json:"return_status"`
}

// Timings records when each lifecycle event happened; nil means it has not.
type Timings struct {
	RequestReturnAt *Time `json:"request_return_at"`
	ReturnAt        *Time `json:"return_at"`
	RefundAt        *Time `json:"refund_at"`
	ClosedAt        *Time `json:"closed_at"`
	CancelledAt     *Time `json:"cancelled_at"`
	ExpiredAt       *Time `json:"expired_at"` // CVS pickup deadline passed
	ConfirmedAt     *Time `json:"confirmed_at"`
}

// ReturnHistory is one refund made against an order.
type ReturnHistory struct {
	Body       string `json:"body"` // short description
	Price      Money  `json:"price"`
	RefundedAt Time   `json:"refunded_at"`
}

// OrderBranchStore is the branch store attached to an order for pickup or
// express delivery. See the BranchStores service for the full model.
type OrderBranchStore struct {
	StoreNo      string  `json:"store_no"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	County       string  `json:"county"`
	District     string  `json:"district"`
	Address      string  `json:"address"`
	Zip          string  `json:"zip"`
	OpeningHours string  `json:"opening_hours"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	Enabled      bool    `json:"enabled"`
	SourceType   string  `json:"source_type"` // BranchStore or PosShop
	SourceID     int64   `json:"source_id"`
}

// PosInfo identifies the POS terminal and salesperson of a POS order.
type PosInfo struct {
	PosUserID    int64  `json:"pos_user_id"`
	PosUserEmail string `json:"pos_user_email"`
	PosShopID    int64  `json:"pos_shop_id"`
	PosInfo      string `json:"pos_info"`
	PosID        int64  `json:"pos_id"`
	PosName      string `json:"pos_name"`
}

// ExchangeHistory is one POS exchange performed on an order.
type ExchangeHistory struct {
	CreatedAt            Time               `json:"created_at"`
	Price                Money              `json:"price"`
	OrderPriceBefore     Money              `json:"order_price_before"`
	OrderPriceAfter      Money              `json:"order_price_after"`
	PosShopID            int64              `json:"pos_shop_id"`
	PosID                int64              `json:"pos_id"`
	PaymentName          string             `json:"payment_name"`
	PaymentMethod        string             `json:"payment_method"`
	MultiplePaymentInfos []OrderPaymentInfo `json:"multiple_payment_infos"`
	LineItems            []ExchangeLineItem `json:"line_items"`
	Einvoice             *OrderEinvoice     `json:"einvoice"`
	PaperInvoiceNo       string             `json:"paper_invoice_no"`
	PaperCompanyNo       string             `json:"paper_company_no"`
}

// ExchangeLineItem is one variant exchanged in an ExchangeHistory.
type ExchangeLineItem struct {
	ProductVariantID int64  `json:"product_variant_id"`
	Name             string `json:"name"`
	SKU              string `json:"sku"`
	QC               string `json:"qc"`
	Price            Money  `json:"price"`
	Quantity         int    `json:"quantity"`
}

// LinkedOrderInfo describes the affiliate or marketplace that referred an
// order. Source is "非導購訂單" when there was none.
type LinkedOrderInfo struct {
	Source                string `json:"source"`
	ShopdotcomRID         string `json:"shopdotcom_rid"`
	ShopdotcomClickID     string `json:"shopdotcom_click_id"`
	LineShoppingECID      string `json:"line_shopping_ecid"`
	LineShoppingAffiliate string `json:"line_shopping_affiliate"`
	IChannelGID           string `json:"ichannel_gid"`
}

// Tag is a label on an order or customer.
type Tag struct {
	Name string `json:"name"`
}

// OrderCustomerCancelReason is the reason a customer gave when cancelling.
type OrderCustomerCancelReason struct {
	Source       string `json:"source"`
	ReasonID     int64  `json:"reason_id"`
	ReasonDetail string `json:"reason_detail"`
}

// UTMTracking is the UTM attribution captured at checkout (custom feature).
type UTMTracking struct {
	UTMSource    string `json:"utm_source"`
	UTMMedium    string `json:"utm_medium"`
	UTMCampaign  string `json:"utm_campaign"`
	UTMContent   string `json:"utm_content"`
	UTMTerm      string `json:"utm_term"`
	UTMClickTime Time   `json:"utm_click_time"`
}

// OrderNumberID maps a shop-facing order number to its API id
// (GET /v1/orders/get_order_id).
type OrderNumberID struct {
	OrderNumber int64 `json:"order_number"`
	OrderID     int64 `json:"order_id"`
}

// OrderTransaction is one payment recorded against an order.
type OrderTransaction struct {
	ID           int64  `json:"id"`
	Amount       Money  `json:"amount"`
	KindName     string `json:"kind_name"`      // e.g. 已收款
	PaidTypeName string `json:"paid_type_name"` // e.g. 手動
}

// OrderReturn is one return shipment of an order (GET /v1/orders/{id}/returns).
type OrderReturn struct {
	ID              int64           `json:"id"`
	CreatedAt       Time            `json:"created_at"`
	TrackingNumber  string          `json:"tracking_number"`
	TrackingCompany TrackingCompany `json:"tracking_company"`
	LineItems       []LineItem      `json:"line_items"`
	ReturnAddress   string          `json:"return_address"`
	ReturnSuda5     string          `json:"return_suda5"` // return address zip code
	ReturnReason    string          `json:"return_reason"`
	ReturnInfo      string          `json:"return_info"`
}

// OrderEticket is an electronic ticket sold on an order.
type OrderEticket struct {
	Title             string               `json:"title"`
	TicketNumber      string               `json:"ticket_number"` // redemption code
	AvailableQuantity int                  `json:"available_quantity"`
	UsedQuantity      int                  `json:"used_quantity"`
	Enabled           bool                 `json:"enabled"`
	OrderID           int64                `json:"order_id"`
	ProductID         int64                `json:"product_id"`
	Separate          bool                 `json:"separate"` // split into one code per unit
	Transactions      []EticketTransaction `json:"transactions"`
}

// EticketTransaction is one redemption code of a split e-ticket.
type EticketTransaction struct {
	TicketNumber string `json:"ticket_number"`
	UsedAt       Time   `json:"used_at"`
}

// CancelReason is the merchant-side reason for cancelling an order.
type CancelReason string

// Known CancelReason values (cancel_reason of PUT /v1/orders/{id}/cancelled).
const (
	CancelReasonCustomer      CancelReason = "customer"        // customer changed their mind
	CancelReasonDuplicate     CancelReason = "duplicate"       // duplicate order
	CancelReasonNotPay        CancelReason = "not_pay"         // not paid in time
	CancelReasonFraud         CancelReason = "fraud"           // fraudulent order
	CancelReasonInventory     CancelReason = "inventory"       // out of stock
	CancelReasonForgot        CancelReason = "forgot"          // CVS pickup expired
	CancelReasonNotShippedYet CancelReason = "not_shipped_yet" // cancelled before shipment
	CancelReasonPosSPReturn   CancelReason = "pos_sp_return"   // returned through a POS shop
	CancelReasonCardPaidFail  CancelReason = "card_paid_fail"  // card payment failed
	CancelReasonOther         CancelReason = "other"           // any other reason
)

// CustomerCancelReason is the customer-side reason for cancelling an order.
type CustomerCancelReason string

// Known CustomerCancelReason values (cancel_reason_detail of
// PUT /v1/orders/{id}/cancelled).
const (
	CustomerCancelReasonWaitTooLong             CustomerCancelReason = "wait_too_long"              // did not want to wait
	CustomerCancelReasonWantToUseOtherDiscount  CustomerCancelReason = "want_to_use_other_discount" // reordering with another promotion
	CustomerCancelReasonModifyShipmentLocation  CustomerCancelReason = "modify_shipment_location"   // changing the delivery location
	CustomerCancelReasonHaveConcernAboutProduct CustomerCancelReason = "have_concern_about_product" // doubts about the product
	CustomerCancelReasonPriceTooHigh            CustomerCancelReason = "price_too_high"             // found a better price elsewhere
	CustomerCancelReasonOperationMistake        CustomerCancelReason = "operation_mistake"          // ordered by mistake
	CustomerCancelReasonOther                   CustomerCancelReason = "other"                      // any other reason
)

// ManualReturnOperation is a manual transition of an order's return status.
type ManualReturnOperation string

// Known ManualReturnOperation values (operation of
// PUT /v1/orders/{id}/manual_return), each named for the target [ReturnStatus].
const (
	ManualReturnReturning  ManualReturnOperation = "manual_returning"     // to returning
	ManualReturnCheckGoods ManualReturnOperation = "manual_check_goods"   // to checking
	ManualReturnRefuse     ManualReturnOperation = "manual_return_refuse" // to refused
	ManualReturnDone       ManualReturnOperation = "manual_return_done"   // to returned
)

// TransactionKind is the type of a manually recorded payment.
type TransactionKind string

// TransactionKindCapture records that payment was received.
const TransactionKindCapture TransactionKind = "capture"

// TransactionPaidType is how a manually recorded payment was made.
type TransactionPaidType string

// TransactionPaidTypeManual is the only documented paid type.
const TransactionPaidTypeManual TransactionPaidType = "manual"

// EticketSearchColumn selects which field GET /v1/order_etickets searches.
type EticketSearchColumn string

// Known EticketSearchColumn values (search_column of GET /v1/order_etickets).
const (
	EticketSearchPhone        EticketSearchColumn = "phone"         // customer phone number
	EticketSearchTicketNumber EticketSearchColumn = "ticket_number" // e-ticket number
	EticketSearchTitle        EticketSearchColumn = "title"         // e-ticket product title
	EticketSearchName         EticketSearchColumn = "name"          // customer name
	EticketSearchOrderName    EticketSearchColumn = "order_name"    // order number
)
