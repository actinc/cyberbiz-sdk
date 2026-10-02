package cyberbiz

// VIPGroupsOverview is the reply of GET /v1/vip_groups: the groups in force
// and the draft set that ApplyDraft will publish. The swagger documents an
// array; the platform returns this object.
type VIPGroupsOverview struct {
	Current []VIPGroup `json:"current"`
	Draft   []VIPGroup `json:"draft"`
}

// VIPGroup is a VIP membership group with its ordered levels.
type VIPGroup struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	// DescriptionURL is the public page explaining the group, when set.
	DescriptionURL string `json:"description_url"`
	// CustomerTags are the customer tags that place a customer in the group.
	CustomerTags   []string        `json:"customer_tags"`
	VIPGroupLevels []VIPGroupLevel `json:"vip_group_levels"`
}

// VIPCouponTypeID is the coupon kind of a VIP gift coupon preset.
type VIPCouponTypeID int

// Known VIPCouponTypeID values (coupon_type_id of a VIP coupon preset).
const (
	VIPCouponTypeAmount       VIPCouponTypeID = 1 // fixed amount off
	VIPCouponTypePercent      VIPCouponTypeID = 2 // percentage off
	VIPCouponTypeFreeShipping VIPCouponTypeID = 3 // waives shipping
	VIPCouponTypeGift         VIPCouponTypeID = 4 // grants a gift item
)

// VIPGroupLevel is one tier of a VIP group. No Golden File carries a level
// (the test shop has none); the shape follows the swagger and the Postman
// examples, with the documented request types preferred where the response
// documentation is self-contradictory (enabled flags are booleans,
// order_discount_value is a percentage, presets is an array).
type VIPGroupLevel struct {
	ID       int64  `json:"id"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	// ValidityDays is how long membership in the level lasts.
	ValidityDays int `json:"validity_days"`
	// UpgradeValidityDays is the window over which upgrade spending is summed.
	UpgradeValidityDays int `json:"upgrade_validity_days"`
	// Upgrade and renewal conditions; zero when not set.
	UpgradeConditionTotalSpent               Money `json:"upgrade_condition_total_spent"`
	UpgradeConditionTotalSpentInValidityDays Money `json:"upgrade_condition_total_spent_in_validity_days"`
	RenewalConditionTotalSpent               Money `json:"renewal_condition_total_spent"`
	RenewalConditionTotalSpentInValidityDays Money `json:"renewal_condition_total_spent_in_validity_days"`
	// Bonus point multiplier: BonusPointValue points per BonusPointThreshold
	// spent, expiring after BonusPointExpiryDays.
	BonusPointEnabled    bool  `json:"bonus_point_enabled"`
	BonusPointThreshold  Money `json:"bonus_point_threshold"`
	BonusPointValue      Money `json:"bonus_point_value"`
	BonusPointExpiryDays int   `json:"bonus_point_expiry_days"`
	// Birthday gift.
	BirthGiftEnabled    bool            `json:"birth_gift_enabled"`
	BirthGiftName       string          `json:"birth_gift_name"`
	BirthGiftBeforeDays int             `json:"birth_gift_before_days"`
	BirthGiftSetting    *VIPGiftSetting `json:"birth_gift_setting"`
	// Upgrade gift.
	UpgradeGiftEnabled bool            `json:"upgrade_gift_enabled"`
	UpgradeGiftSetting *VIPGiftSetting `json:"upgrade_gift_setting"`
	// Order discount, as a percentage.
	OrderDiscountEnabled bool                `json:"order_discount_enabled"`
	OrderDiscountValue   float64             `json:"order_discount_value"`
	OrderDiscountSetting *VIPRestrictSetting `json:"order_discount_setting"`
	// Free shipping above FreeShippingThreshold.
	FreeShippingEnabled   bool                `json:"free_shipping_enabled"`
	FreeShippingThreshold Money               `json:"free_shipping_threshold"`
	FreeShippingSetting   *VIPRestrictSetting `json:"free_shipping_setting"`
}

// VIPGiftSetting is the bonus and coupon parts of a birthday or upgrade gift.
type VIPGiftSetting struct {
	Bonus  *VIPBonusSetting  `json:"bonus"`
	Coupon *VIPCouponSetting `json:"coupon"`
}

// VIPBonusSetting grants bonus points as part of a gift.
type VIPBonusSetting struct {
	Enabled bool  `json:"enabled"`
	Value   Money `json:"value"`
	// ExpiryDays is the points' lifetime; 0 means they never expire.
	ExpiryDays int `json:"expiry_days"`
}

// VIPCouponSetting grants coupons as part of a gift.
type VIPCouponSetting struct {
	Enabled bool              `json:"enabled"`
	Presets []VIPCouponPreset `json:"presets"`
}

// VIPCouponPreset describes one coupon a gift grants.
type VIPCouponPreset struct {
	CouponTypeID VIPCouponTypeID `json:"coupon_type_id"`
	// Value is the amount or percentage, depending on CouponTypeID.
	Value               Money  `json:"value"`
	Code                string `json:"code"`
	OrderPriceThreshold Money  `json:"order_price_threshold"`
	UsableDays          int    `json:"usable_days"`
	// UsageLimit is the number of coupons granted.
	UsageLimit        int                `json:"usage_limit"`
	ProductTags       []string           `json:"product_tags"`
	RestrictStrategy  RestrictStrategy   `json:"restrict_strategy"`
	RestrictCampaigns []RestrictCampaign `json:"restrict_campaigns"`
}

// VIPRestrictSetting limits how a VIP benefit combines with other campaigns.
type VIPRestrictSetting struct {
	RestrictStrategy  RestrictStrategy   `json:"restrict_strategy"`
	RestrictCampaigns []RestrictCampaign `json:"restrict_campaigns"`
}
