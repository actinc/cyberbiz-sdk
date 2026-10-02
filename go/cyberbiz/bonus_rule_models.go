package cyberbiz

// BonusUsageLimitType is how the per-order bonus redemption cap is expressed.
type BonusUsageLimitType string

// Known BonusUsageLimitType values (usage_limit_type of GET /v1/bonus_rule).
const (
	BonusUsageLimitNone       BonusUsageLimitType = "no_limit"   // points can cover the whole order
	BonusUsageLimitPercentage BonusUsageLimitType = "percentage" // cap is a percentage of the order total
	BonusUsageLimitAmount     BonusUsageLimitType = "amount"     // cap is a fixed amount per order
)

// BonusRule is the shop's bonus point rule (GET /v1/bonus_rule).
type BonusRule struct {
	Enabled bool `json:"enabled"`
	// SignupBonus is the number of points granted on registration.
	SignupBonus Money `json:"signup_bonus"`
	// ConsumptionBonus points are granted per ConsumptionPrice spent.
	ConsumptionPrice Money `json:"consumption_price"`
	ConsumptionBonus Money `json:"consumption_bonus"`
	// ExpireDay is the lifetime of each grant in days; 0 means no expiry.
	ExpireDay int `json:"expire_day"`
	// ThresholdOfTotalPrice is the minimum order total to redeem points.
	ThresholdOfTotalPrice Money               `json:"threshold_of_total_price"`
	UsageLimitType        BonusUsageLimitType `json:"usage_limit_type"`
	// UsageLimitValue is the per-order cap: an amount, or a percentage when
	// UsageLimitType is percentage.
	UsageLimitValue Money `json:"usage_limit_value"`
	// ConversionBonus is the redemption rate: 1 means one point is worth 1 TWD.
	ConversionBonus float64 `json:"conversion_bonus"`
}
