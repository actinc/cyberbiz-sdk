package cyberbiz

// DiscountType is the kind of shop-wide discount.
type DiscountType string

// Known DiscountType values (discount_type of a shop-wide discount).
const (
	DiscountTypeAmount  DiscountType = "amount"  // a fixed amount off the order
	DiscountTypePercent DiscountType = "percent" // a percentage off the order
	DiscountTypeCoupon  DiscountType = "coupon"  // grants a coupon instead of a direct discount
)

// SalesChannel is the channel a discount applies to.
type SalesChannel int

// Known SalesChannel values (sales_channel_id of a discount).
const (
	SalesChannelAll SalesChannel = 1 // every channel
	SalesChannelEC  SalesChannel = 2 // the web shop only
	SalesChannelPOS SalesChannel = 3 // point of sale only
)

// Discount is a shop-wide discount (全館折扣), as returned by GET /v1/discounts
// and GET /v1/discounts/{id}. The detail response omits id; Get fills it in.
type Discount struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// DiscountTypeName is the human-readable type, e.g. "金額" or "百分比".
	DiscountTypeName string `json:"discount_type_name"`
	// DiscountValue is the formatted discount content, e.g. "5.0元" or "10%".
	DiscountValue string `json:"discount_value"`
	// Threshold is the minimum order total the discount applies from.
	Threshold Money `json:"threshold"`
	StartAt   Time  `json:"start_at"`
	EndAt     Time  `json:"end_at"`
	// Days is how many days a coupon-type discount stays usable.
	Days int `json:"days"`
	// ConcurrentlyApply reports whether the discount stacks with special
	// collection discounts and other shop campaigns.
	ConcurrentlyApply bool         `json:"concurrently_apply"`
	SalesChannelID    SalesChannel `json:"sales_channel_id"`
}

// RegisterCouponRule is the "coupon on registration" rule of the shop
// (PUT /v1/discounts/register_coupon_rule). No Golden File exists for it.
// The swagger documents expire_day as a string, but it is an integer column
// sent as a JSON number.
type RegisterCouponRule struct {
	Enabled bool `json:"enabled"`
	// Value is the coupon amount or percentage, depending on CouponType.
	Value      Money  `json:"value"`
	CouponType string `json:"coupon_type"`
	UsageLimit int    `json:"usage_limit"`
	// OrderPriceThreshold is the minimum order total to use the coupon.
	OrderPriceThreshold Money `json:"order_price_threshold"`
	// ExpireDay is the number of days the coupon stays usable.
	ExpireDay         int  `json:"expire_day"`
	ConcurrentlyApply bool `json:"concurrently_apply"`
}

// CouponType is the kind of shop coupon a caller can create.
type CouponType string

// Known CouponType values (coupon_type of a shop or POS coupon).
const (
	CouponTypeAmount       CouponType = "amount"        // fixed amount off
	CouponTypePercent      CouponType = "percent"       // percentage off
	CouponTypeFreeShipping CouponType = "free_shipping" // waives shipping
	CouponTypeGift         CouponType = "gift"          // grants a gift item
)

// RestrictStrategy is how a coupon or VIP benefit combines with other
// campaigns.
type RestrictStrategy string

const (
	// RestrictStrategyUnrestricted allows the coupon on every campaign product.
	RestrictStrategyUnrestricted RestrictStrategy = "unrestricted"
	// RestrictStrategyRestrict allows the coupon on everything except the
	// products of the listed campaigns.
	RestrictStrategyRestrict RestrictStrategy = "restrict"
	// RestrictStrategyForbidden rejects the coupon when the order contains a
	// product of the listed campaigns.
	RestrictStrategyForbidden RestrictStrategy = "forbidden"
)

// RestrictCampaign names a campaign type in a restrict_campaigns list.
type RestrictCampaign string

// Known RestrictCampaign values accepted in restrict_campaigns.
const (
	RestrictCampaignVIPCustomPriceDiscount    RestrictCampaign = "vip_custom_price_discount"   // VIP member custom prices
	RestrictCampaignVariantDiscount           RestrictCampaign = "variant_discount"            // variant discount collections
	RestrictCampaignBundleDiscount            RestrictCampaign = "bundle_discount"             // bundle (add-buy set) discounts
	RestrictCampaignSpecialCollectionDiscount RestrictCampaign = "special_collection_discount" // mix-and-match special collections
	RestrictCampaignCategoryCampaignDiscount  RestrictCampaign = "category_campaign_discount"  // category campaigns
	RestrictCampaignShopDiscount              RestrictCampaign = "shop_discount"               // shop-wide discounts
	RestrictCampaignVIPDiscount               RestrictCampaign = "vip_discount"                // VIP level discounts
	RestrictCampaignReferralDiscount          RestrictCampaign = "referral_discount"           // referral discounts
	RestrictCampaignAddBuyDiscount            RestrictCampaign = "add_buy_discount"            // add-on purchase discounts
)

// ShopCoupon is a shop-wide coupon (全館優惠券), as returned by
// GET /v1/shop_coupons and GET /v1/shop_coupons/{id}. The detail response
// omits id; Get fills it in.
type ShopCoupon struct {
	ID int64 `json:"id"`
	// CustomerID is set only for coupons bound to one customer; it is never
	// present on shop-wide coupons.
	CustomerID int64  `json:"customer_id"`
	Title      string `json:"title"`
	Code       string `json:"code"`
	// CouponTypeName is the human-readable type, e.g. "金額" or "免運".
	CouponTypeName string `json:"coupon_type_name"`
	// CouponValue is the formatted discount, e.g. "5.0元".
	CouponValue string `json:"coupon_value"`
	// OrderPriceThreshold is the minimum order total to use the coupon.
	OrderPriceThreshold Money `json:"order_price_threshold"`
	StartDate           Time  `json:"start_date"`
	EndDate             Time  `json:"end_date"`
	// ConcurrentlyApply reports whether the coupon stacks with special
	// collection discounts and other shop campaigns.
	ConcurrentlyApply  bool `json:"concurrently_apply"`
	UsageLimit         int  `json:"usage_limit"`
	CanAccumulateBonus bool `json:"can_accumulate_bonus"`
	UsageUnlimited     bool `json:"usage_unlimited"`
	UsedTimes          int  `json:"used_times"`
	// GiftOrderID is the order that granted the coupon, when any.
	GiftOrderID int64 `json:"gift_order_id"`
	// GiftDays is how many days a granted coupon stays usable.
	GiftDays                 int                `json:"gift_days"`
	AccountUsageLimitEnabled bool               `json:"account_usage_limit_enabled"`
	AccountUsageLimit        int                `json:"account_usage_limit"`
	RestrictStrategy         RestrictStrategy   `json:"restrict_strategy"`
	RestrictCampaigns        []RestrictCampaign `json:"restrict_campaigns"`
	// Tags are the product tags the coupon is bound to.
	Tags       []string `json:"tags"`
	PosShopIDs []int64  `json:"pos_shop_ids"`
	// CouponStatus and GiftOrderStatus must be read together; see
	// [CouponUsable].
	CouponStatus    CouponStatus    `json:"coupon_status"`
	GiftOrderStatus GiftOrderStatus `json:"gift_order_status"`
	Valid           bool            `json:"valid"`
	// CustomerUsedTimes and CustomerUsable describe the calling customer's
	// remaining uses of a shop-wide coupon.
	CustomerUsedTimes int     `json:"customer_used_times"`
	CustomerUsable    bool    `json:"customer_usable"`
	ProductIDs        []int64 `json:"product_ids"`
}

// Usable reports whether the coupon can be used right now, per the
// documented combination of CouponStatus and GiftOrderStatus.
func (c *ShopCoupon) Usable() bool {
	return CouponUsable(c.CouponStatus, c.GiftOrderStatus)
}
