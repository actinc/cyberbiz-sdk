package cyberbiz

// CustomerStatus is the account state of a customer.
type CustomerStatus string

// Known CustomerStatus values (status of a customer).
const (
	CustomerStatusPending  CustomerStatus = "pending"  // account not yet activated
	CustomerStatusValidate CustomerStatus = "validate" // account not yet verified
	CustomerStatusEnabled  CustomerStatus = "enabled"  // account active
	CustomerStatusDisabled CustomerStatus = "disabled" // account blocked by the merchant
	CustomerStatusInvited  CustomerStatus = "invited"  // activation invitation sent
	CustomerStatusDeclined CustomerStatus = "declined" // activation invitation declined
	CustomerStatusWarning  CustomerStatus = "warning"  // flagged as a warning account
)

// CustomerUIDProviderType is the social login provider a customer's external
// UID belongs to.
type CustomerUIDProviderType string

// Known CustomerUIDProviderType values (provider_type of a customer UID).
const (
	CustomerUIDProviderLine     CustomerUIDProviderType = "line"     // LINE Login
	CustomerUIDProviderLineAt   CustomerUIDProviderType = "line_at"  // LINE Official Account (LINE@)
	CustomerUIDProviderFacebook CustomerUIDProviderType = "facebook" // Facebook Login
)

// Customer is a shop member (GET /v1/customers, GET /v1/customers/{id},
// GET /v2/customers). The v1 detail response omits id; Get fills it from the
// path. VIPInfo is only present on v2 lists requested with
// include_params=vip_info.
type Customer struct {
	ID                 int64          `json:"id"`
	Name               string         `json:"name"`
	Status             CustomerStatus `json:"status"`
	Email              string         `json:"email"`
	CountryCallingCode string         `json:"country_calling_code"` // e.g. "+886"
	Mobile             string         `json:"mobile"`
	// EnableCVSPickup reports whether the customer may use convenience-store pickup.
	EnableCVSPickup bool `json:"enable_cvs_pickup"`
	// EnableCVSCOD reports whether the customer may pay cash on delivery at a convenience store.
	EnableCVSCOD bool `json:"enable_cvs_cod"`
	// EnableHomeDeliveryCOD reports whether the customer may pay cash on home delivery.
	EnableHomeDeliveryCOD       bool             `json:"enable_home_delivery_cod"`
	AcceptsMarketing            bool             `json:"accepts_marketing"`
	AcceptsEmailNotification    bool             `json:"accepts_email_notification"`
	Tags                        []CustomerTag    `json:"tags"`
	Address                     *CustomerAddress `json:"address"`
	Gender                      string           `json:"gender"` // free text; shops define their own options
	Birthday                    Date             `json:"birthday"`
	OtherAccumulatedConsumption Money            `json:"other_accumulated_consumption"` // spend accumulated through other channels
	// OtherAccumulatedConsumptionExpiredAt is when other-channel spend started counting; nil when unset.
	OtherAccumulatedConsumptionExpiredAt *Time                 `json:"other_accumulated_consumption_expired_at"`
	Note                                 string                `json:"note"`
	CustomFields                         []CustomerCustomField `json:"custom_fields"`
	CreatedAt                            Time                  `json:"created_at"`
	UpdatedAt                            Time                  `json:"updated_at"`
	ConfirmedAt                          *Time                 `json:"confirmed_at"`            // email verified at; nil when never verified
	MobileSMSConfirmedAt                 *Time                 `json:"mobile_sms_confirmed_at"` // mobile verified at; nil when never verified
	BonusRemain                          Money                 `json:"bonus_remain"`            // remaining bonus points
	UIDProviders                         []CustomerUIDProvider `json:"uid_providers"`
	VIPInfo                              *CustomerVIPInfo      `json:"vip_info,omitzero"`
}

// CustomerTag is a label attached to a customer.
type CustomerTag struct {
	Name string `json:"name"`
}

// CustomerAddress is a customer's default address.
type CustomerAddress struct {
	Company            string                 `json:"company"`
	CountryCallingCode string                 `json:"country_calling_code"`
	Phone              string                 `json:"phone"`
	Address            string                 `json:"address"` // the full address as one line
	DetailAddress      *CustomerDetailAddress `json:"detail_address"`
}

// CustomerDetailAddress is an address broken into its parts.
type CustomerDetailAddress struct {
	Zip      string `json:"zip"`
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
}

// CustomerCustomField is one custom field value on a customer. ID is absent
// from the customer detail embedding and present on the custom field endpoints.
type CustomerCustomField struct {
	ID    int64  `json:"id,omitzero"`
	Name  string `json:"name"`  // field key
	Label string `json:"label"` // display name
	Value string `json:"value"`
}

// CustomerUIDProvider links a customer to an external login identity.
type CustomerUIDProvider struct {
	ProviderType CustomerUIDProviderType `json:"provider_type"`
	UID          string                  `json:"uid"`
}

// CustomerUIDLookup is the reply of GET /v1/customers/{id}/uid_providers/{type}.
type CustomerUIDLookup struct {
	CustomerID int64  `json:"customer_id"`
	UID        string `json:"uid"`
	Message    string `json:"message"` // platform status text, e.g. "查詢成功"
}

// CustomerIDMatch is one hit of the email / mobile lookup
// (GET /v1/customers/get_customer_id).
type CustomerIDMatch struct {
	CustomerID     int64  `json:"customer_id"`
	CustomerEmail  string `json:"customer_email"`
	CustomerMobile string `json:"customer_mobile"`
}

// CustomerNameMatch is one hit of the name lookup
// (GET /v1/customers/get_customer_id_by_name).
type CustomerNameMatch struct {
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
}

// CustomerGenderOption is one extra gender option the shop offers at
// registration besides male and female.
type CustomerGenderOption struct {
	DefaultGenderOption string `json:"default_gender_option"`
}

// CustomerActivationURL is the link that activates a customer's account.
type CustomerActivationURL struct {
	AccountActivationURL string `json:"account_activation_url"`
}

// CustomerSpendingOverview summarises a customer's paid and valid orders in a
// date range.
type CustomerSpendingOverview struct {
	PaidAndValidTotalSpent   Money `json:"paid_and_valid_total_spent"`
	PaidAndValidOrdersCount  int   `json:"paid_and_valid_orders_count"`
	PaidAndValidAverageSpent Money `json:"paid_and_valid_average_spent"`
}

// CustomerVIPInfo is a customer's VIP membership state
// (GET /v1/customers/{id}/vip_info). Group and level pointers are nil when the
// customer is not in a VIP programme.
type CustomerVIPInfo struct {
	CustomerID   int64                 `json:"customer_id"`
	CurrentGroup *CustomerVIPGroup     `json:"current_group"`
	CurrentLevel *CustomerVIPLevel     `json:"current_level"`
	NextLevel    *CustomerVIPLevel     `json:"next_level"`
	ExtraInfo    *CustomerVIPExtraInfo `json:"extra_info"`
}

// CustomerVIPGroup is the VIP group a customer belongs to, as embedded in
// CustomerVIPInfo. Level settings live on the VIP groups resource.
type CustomerVIPGroup struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Position       int      `json:"position"`
	DescriptionURL string   `json:"description_url"`
	CustomerTags   []string `json:"customer_tags"`
}

// CustomerVIPLevel is a VIP level as embedded in CustomerVIPInfo.
type CustomerVIPLevel struct {
	ID                  int64  `json:"id"`
	Position            int    `json:"position"`
	Name                string `json:"name"`
	ValidityDays        int    `json:"validity_days"`
	UpgradeValidityDays int    `json:"upgrade_validity_days"`
	// UpgradeConditionTotalSpent is the single-order spend that upgrades to this level.
	UpgradeConditionTotalSpent Money `json:"upgrade_condition_total_spent"`
	// UpgradeConditionTotalSpentInValidityDays is the spend within the validity window that upgrades to this level.
	UpgradeConditionTotalSpentInValidityDays Money  `json:"upgrade_condition_total_spent_in_validity_days"`
	RenewalConditionTotalSpent               Money  `json:"renewal_condition_total_spent"`
	RenewalConditionTotalSpentInValidityDays Money  `json:"renewal_condition_total_spent_in_validity_days"`
	BonusPointEnabled                        bool   `json:"bonus_point_enabled"`
	BonusPointThreshold                      Money  `json:"bonus_point_threshold"`
	BonusPointValue                          Money  `json:"bonus_point_value"`
	BonusPointExpiryDays                     int    `json:"bonus_point_expiry_days"`
	BirthGiftEnabled                         bool   `json:"birth_gift_enabled"`
	BirthGiftName                            string `json:"birth_gift_name"`
	UpgradeGiftEnabled                       bool   `json:"upgrade_gift_enabled"`
	OrderDiscountEnabled                     bool   `json:"order_discount_enabled"`
	FreeShippingEnabled                      bool   `json:"free_shipping_enabled"`
}

// CustomerVIPExtraInfo carries the validity window of the current level and
// how far the customer is from renewing or upgrading. Differences are zero
// when there is no such condition.
type CustomerVIPExtraInfo struct {
	StartAt                                        Time  `json:"start_at"`
	EndAt                                          Time  `json:"end_at"`
	DifferenceOfTotalSpentInValidityDaysForRenewal Money `json:"difference_of_total_spent_in_validity_days_for_renewal"`
	DifferenceOfTotalSpentForRenewal               Money `json:"difference_of_total_spent_for_renewal"`
	DifferenceOfTotalSpentInValidityDaysForUpgrade Money `json:"difference_of_total_spent_in_validity_days_for_upgrade"`
	DifferenceOfTotalSpentForUpgrade               Money `json:"difference_of_total_spent_for_upgrade"`
}

// CustomerMessageStatus is the reply state of a customer service thread.
type CustomerMessageStatus string

// Known CustomerMessageStatus values (status of a message post).
const (
	CustomerMessageStatusReplyYet CustomerMessageStatus = "reply_yet" // awaiting a merchant reply
	CustomerMessageStatusReplied  CustomerMessageStatus = "replied"   // merchant has replied
)

// CustomerMessagePost is one customer service thread
// (GET /v1/customers/{id}/message_posts).
type CustomerMessagePost struct {
	ID        int64                    `json:"id"`
	Title     string                   `json:"title"`
	Category  string                   `json:"category"`
	Order     *CustomerMessageOrder    `json:"order"` // the order the thread is about, nil if none
	Status    CustomerMessageStatus    `json:"status"`
	Comments  []CustomerMessageComment `json:"comments"`
	CreatedAt Time                     `json:"created_at"`
}

// CustomerMessageOrder is the order a customer service thread refers to.
type CustomerMessageOrder struct {
	ID   int64  `json:"id"`
	Type string `json:"type"` // the kind of order the thread is linked to
	Name string `json:"name"` // the order name shown to the customer
}

// CustomerMessageComment is one message in a customer service thread.
type CustomerMessageComment struct {
	ID        int64                 `json:"id"`
	Role      string                `json:"role"`  // who replied
	Admin     *CustomerMessageAdmin `json:"admin"` // the replying admin when Role is admin, else nil
	Content   string                `json:"content"`
	CreatedAt Time                  `json:"created_at"`
}

// CustomerMessageAdmin is the shop admin who wrote a reply.
type CustomerMessageAdmin struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CustomerCartItem is a product variant sitting unpaid in the customer's
// web-shop cart (GET /v1/customers/{id}/customer_cart_items). It carries the
// variant fields; the PIM channel-bridge block is not modelled.
type CustomerCartItem struct {
	ID                      int64    `json:"id"`
	ProductID               int64    `json:"product_id"`
	Name                    string   `json:"name"` // product name
	Position                int      `json:"position"`
	Price                   Money    `json:"price"`
	Cost                    Money    `json:"cost"`
	CompareAtPrice          Money    `json:"compare_at_price"`
	Meas                    float64  `json:"meas"` // volumetric size
	MaxUsableBonus          Money    `json:"max_usable_bonus"`
	Weight                  float64  `json:"weight"`
	Option1                 string   `json:"option1"`
	Option2                 string   `json:"option2"`
	Option3                 string   `json:"option3"`
	InventoryManagement     bool     `json:"inventory_management"`
	InventoryQuantity       int      `json:"inventory_quantity"`
	Sold                    int      `json:"sold"`
	SafetyInventoryQuantity int      `json:"safety_inventory_quantity"`
	InventoryPolicy         string   `json:"inventory_policy"` // whether purchase is allowed when out of stock
	SKU                     string   `json:"sku"`
	QC                      string   `json:"qc"` // vendor code
	RequiresShipping        bool     `json:"requires_shipping"`
	CreatedAt               Time     `json:"created_at"`
	UpdatedAt               Time     `json:"updated_at"`
	HoneycombSync           bool     `json:"honeycomb_sync"` // warehouse stock sync
	Vendor                  string   `json:"vendor"`
	PhotoURLs               []string `json:"photo_urls"`
}
