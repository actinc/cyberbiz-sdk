package cyberbiz

// PosShop is a physical POS store (POS 商店). The web shop itself is not
// listed; stock endpoints refer to it as shop id 0.
type PosShop struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	County   string `json:"county"`
	District string `json:"district"`
	Address  string `json:"address"`
	// VATNumber is the company tax id (統一編號).
	VATNumber string `json:"VAT_number"`
	// Deadline is the day the POS licence expires.
	Deadline Date `json:"deadline"`
	// CanFindOthersOrder lets staff look up orders from other shops.
	CanFindOthersOrder bool `json:"can_find_others_order"`
	// CanAccessCustomers lets staff open the customer list.
	CanAccessCustomers bool `json:"can_access_customers"`
}

// PosShopSummary is the id and name of a POS shop as embedded in a product.
type PosShopSummary struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Pos is one POS terminal (POS 機) of a POS shop.
type Pos struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	PettyCash Money  `json:"petty_cash"` // float in the cash drawer at open
	// AdminPasswordEnabled requires the manager password for privileged
	// actions.
	AdminPasswordEnabled bool `json:"admin_password_enabled"`
}

// PosProduct is a product as returned by a POS shop's product list. It
// carries the fields the shop-level product list shares; nested collections,
// options, photos and SEO tags are left out because the POS listing does not
// populate them.
type PosProduct struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	EnglishTitle string `json:"english_title"`
	// ProductURL is protocol-relative, e.g. "//shop.cyberbiz.co/products/x".
	ProductURL string `json:"product_url"`
	Published  bool   `json:"published"`
	// SellFrom and SellTo bound the selling period; zero when unbounded.
	SellFrom          Time   `json:"sell_from"`
	SellTo            Time   `json:"sell_to"`
	ProductType       string `json:"product_type"`
	ProductTypeCode   string `json:"product_type_code"`
	Slogan            string `json:"slogan"`
	Brief             string `json:"brief"`
	BriefText         string `json:"brief_text"`
	BriefIncludesHTML bool   `json:"brief_includes_html"`
	BodyHTML          string `json:"body_html"`
	Vendor            string `json:"vendor"`
	Price             Money  `json:"price"`
	// SellWeight is in grams.
	SellWeight float64 `json:"sell_weight"`
	// TaxTypeID is e.g. "inclusive_tax".
	TaxTypeID       string              `json:"tax_type_id"`
	Tags            []string            `json:"tags"`
	ProductVariants []PosProductVariant `json:"product_variants"`
	// PosShop is the shop the product belongs to.
	PosShop   *PosShopSummary `json:"pos_shop"`
	CreatedAt Time            `json:"created_at"`
	UpdatedAt Time            `json:"updated_at"`
	// TemperatureTypes are shipping temperature labels such as 常溫.
	TemperatureTypes        []string `json:"temperature_types"`
	Searchable              bool     `json:"searchable"`
	GoogleProductCategoryID int64    `json:"google_product_category_id"`
}

// PosProductVariant is one variant as returned by a POS shop's product or
// variant list.
type PosProductVariant struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	Position  int    `json:"position"`
	Price     Money  `json:"price"`
	Cost      Money  `json:"cost"`
	// CompareAtPrice is the crossed-out list price.
	CompareAtPrice Money `json:"compare_at_price"`
	// Meas is the volumetric size used for shipping.
	Meas float64 `json:"meas"`
	// MaxUsableBonus caps the bonus points spendable on this variant.
	MaxUsableBonus Money   `json:"max_usable_bonus"`
	Weight         float64 `json:"weight"`
	Option1        string  `json:"option1"`
	Option2        string  `json:"option2"`
	Option3        string  `json:"option3"`
	// InventoryManagement is whether stock is tracked.
	InventoryManagement bool `json:"inventory_management"`
	InventoryQuantity   int  `json:"inventory_quantity"`
	Sold                int  `json:"sold"`
	// SafetyInventoryQuantity is the low-stock warning level.
	SafetyInventoryQuantity int `json:"safety_inventory_quantity"`
	// InventoryPolicy is "continue" or "deny": whether to keep selling at
	// zero stock.
	InventoryPolicy  string `json:"inventory_policy"`
	SKU              string `json:"sku"`
	QC               string `json:"qc"` // vendor product code
	RequiresShipping bool   `json:"requires_shipping"`
	CreatedAt        Time   `json:"created_at"`
	UpdatedAt        Time   `json:"updated_at"`
	// HoneycombSync is whether the warehouse integration syncs this stock.
	HoneycombSync bool   `json:"honeycomb_sync"`
	Vendor        string `json:"vendor"`
}

// PosShopCouponType is how a POS shop coupon discounts an order.
type PosShopCouponType string

// Known PosShopCouponType values (coupon_type of a POS shop coupon).
const (
	PosShopCouponTypeAmount       PosShopCouponType = "amount"        // fixed amount off
	PosShopCouponTypePercent      PosShopCouponType = "percent"       // percentage off
	PosShopCouponTypeFreeShipping PosShopCouponType = "free_shipping" // waives shipping
	PosShopCouponTypeGift         PosShopCouponType = "gift"          // grants a gift item
)

// PosShopCoupon is a coupon redeemable at POS shops (門市優惠券). The recorded
// shop lacks the pos_shop_coupon plugin (403), so no Golden File exists; the
// shape is the swagger CouponsEntity, which the shop-coupon Golden File
// confirms. The detail response omits the id; GetCoupon fills it in.
type PosShopCoupon struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Code  string `json:"code"`
	// CustomerID is set only on coupons granted to one customer.
	CustomerID int64 `json:"customer_id"`
	// CouponTypeName is the localised type label, e.g. 金額.
	CouponTypeName string `json:"coupon_type_name"`
	// CouponValue is a display string such as "5.0元" or "10%".
	CouponValue         string `json:"coupon_value"`
	OrderPriceThreshold Money  `json:"order_price_threshold"`
	StartDate           Date   `json:"start_date"`
	EndDate             Date   `json:"end_date"`
	// ConcurrentlyApply allows use together with bundle discounts or shop
	// campaigns. The swagger types it as a string; it is a boolean.
	ConcurrentlyApply  bool `json:"concurrently_apply"`
	UsageLimit         int  `json:"usage_limit"`
	CanAccumulateBonus bool `json:"can_accumulate_bonus"`
	UsageUnlimited     bool `json:"usage_unlimited"`
	UsedTimes          int  `json:"used_times"`
	// GiftOrderID and GiftDays describe a coupon granted by an order.
	GiftOrderID              int64 `json:"gift_order_id"`
	GiftDays                 int   `json:"gift_days"`
	AccountUsageLimitEnabled bool  `json:"account_usage_limit_enabled"`
	AccountUsageLimit        int   `json:"account_usage_limit"`
	// RestrictStrategy is "unrestricted" or "restrict"; RestrictCampaigns
	// lists the campaign kinds excluded when it is "restrict".
	RestrictStrategy  string   `json:"restrict_strategy"`
	RestrictCampaigns []string `json:"restrict_campaigns"`
	Tags              []string `json:"tags"`
	// PosShopIDs are the shops that accept the coupon.
	PosShopIDs []int64 `json:"pos_shop_ids"`
	// CouponStatus and GiftOrderStatus must be read together; see
	// [CouponUsable].
	CouponStatus      CouponStatus    `json:"coupon_status"`
	GiftOrderStatus   GiftOrderStatus `json:"gift_order_status"`
	Valid             bool            `json:"valid"`
	CustomerUsedTimes int             `json:"customer_used_times"`
	CustomerUsable    bool            `json:"customer_usable"`
	ProductIDs        []int64         `json:"product_ids"`
}

// PosWalletBalance is a customer's stored-value balance
// (GET /v2/pos_wallets/{customer_id}/balance).
type PosWalletBalance struct {
	Balance  Money  `json:"balance"`
	Currency string `json:"currency"` // always "TWD" today
}

// PosWalletTransactionType is what a stored-value transaction did.
type PosWalletTransactionType string

// Known PosWalletTransactionType values (type of a wallet transaction).
const (
	PosWalletTransactionTopup       PosWalletTransactionType = "topup"        // 儲值
	PosWalletTransactionConsumption PosWalletTransactionType = "consumption"  // spent on an order
	PosWalletTransactionRefund      PosWalletTransactionType = "refund"       // balance refunded to the customer
	PosWalletTransactionCancelOrder PosWalletTransactionType = "cancel_order" // order cancelled, balance returned
	PosWalletTransactionBonus       PosWalletTransactionType = "bonus"        // bonus credit
)

// PosWalletFinancialStatus is the payment state of a top-up.
type PosWalletFinancialStatus string

// Known PosWalletFinancialStatus values (financial_status of a top-up).
const (
	PosWalletFinancialStatusPaid    PosWalletFinancialStatus = "paid"    // top-up paid
	PosWalletFinancialStatusFailed  PosWalletFinancialStatus = "failed"  // payment failed
	PosWalletFinancialStatusTimeout PosWalletFinancialStatus = "timeout" // still processing
	PosWalletFinancialStatusPending PosWalletFinancialStatus = "pending" // awaiting payment
)

// PosWalletTransaction is one stored-value movement. Which optional fields
// are present depends on Type: a topup carries FinancialStatus, Einvoice and
// PaperInvoiceNo; a refund carries InvalidEinvoices and AllowanceEinvoices;
// the other types carry only the common fields. No Golden File exists; the
// shape follows the CYBERBIZ v2 reference notes.
type PosWalletTransaction struct {
	Type PosWalletTransactionType `json:"type"`
	// Amount is positive when the balance grows and negative when it shrinks.
	Amount       Money `json:"amount"`
	BalanceAfter Money `json:"balance_after"`
	// CreatedAt arrives as "2006-01-02 15:04-0700".
	CreatedAt Time `json:"created_at"`
	// PaymentName is the payment method of a topup or refund, e.g. 現金.
	PaymentName     string                   `json:"payment_name"`
	FinancialStatus PosWalletFinancialStatus `json:"financial_status"`
	// Einvoice is the invoice issued for a topup, nil for other types.
	Einvoice *PosWalletEinvoice `json:"einvoice"`
	// PaperInvoiceNo is set when the topup got a paper invoice instead.
	PaperInvoiceNo     string              `json:"paper_invoice_no"`
	InvalidEinvoices   []PosWalletEinvoice `json:"invalid_einvoices"`
	AllowanceEinvoices []PosWalletEinvoice `json:"allowance_einvoices"`
}

// PosWalletEinvoice is the e-invoice attached to a wallet transaction.
type PosWalletEinvoice struct {
	Title     string `json:"title"`      // invoice title (抬頭)
	CompanyNo string `json:"company_no"` // buyer tax id
	// InvoiceNo is empty for a paper invoice or when not yet issued.
	InvoiceNo string `json:"invoice_no"`
	// InvoiceStatus is empty when not issued; see [InvoiceStatus].
	InvoiceStatus InvoiceStatus `json:"invoice_status"`
	InvoiceAt     Time          `json:"invoice_at"`
	// InvalidAt is nil unless the invoice was voided.
	InvalidAt *Time `json:"invalid_at"`
	// RandomNum is the four-digit lottery check code.
	RandomNum   string      `json:"random_num"`
	InvoiceType InvoiceType `json:"invoice_type"`
	// LoveCode, PhoneBarcode and NaturePerson are the carrier or donation
	// identifiers, one of which is set according to InvoiceType.
	LoveCode     string `json:"love_code"`
	PhoneBarcode string `json:"phone_barcode"`
	NaturePerson string `json:"nature_person"`
}
