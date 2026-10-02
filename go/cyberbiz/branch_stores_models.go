package cyberbiz

// BranchStoreSourceType says what a branch store was created from.
type BranchStoreSourceType string

// Known BranchStoreSourceType values (source_type of a branch store).
const (
	BranchStoreSourceBranchStore BranchStoreSourceType = "BranchStore" // a stand-alone pickup store
	BranchStoreSourcePosShop     BranchStoreSourceType = "PosShop"     // mirrors a POS shop
)

// BranchStore is a physical store customers can pick up orders from (門市).
// The recorded shop has no branch stores, so only the empty list exists as a
// Golden File; the shape follows the swagger and Postman example. The detail
// response omits the id; Get and Update fill it in.
type BranchStore struct {
	ID       int64  `json:"id"`
	StoreNo  string `json:"store_no"` // merchant-assigned store code
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	County   string `json:"county"`
	District string `json:"district"`
	Address  string `json:"address"`
	Zip      string `json:"zip"`
	// OpeningHours is free text shown to customers.
	OpeningHours string `json:"opening_hours"`
	// Lat and Lng are the map coordinates in decimal degrees.
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Enabled bool    `json:"enabled"`
	// ShippingRates are the delivery fees offered from this store.
	ShippingRates []BranchStoreShippingRate `json:"shipping_rates"`
	// SourceType and SourceID say which POS shop the store mirrors, if any.
	SourceType BranchStoreSourceType `json:"source_type"`
	SourceID   int64                 `json:"source_id"`
}

// BranchStoreShippingRate is one delivery fee tier of a branch store. The
// detail response omits the id; GetShippingRate fills it in.
type BranchStoreShippingRate struct {
	ID   int64  `json:"id"`
	Name string `json:"name"` // courier or delivery method name
	// MinOrderSubtotal is the order subtotal from which this tier applies.
	MinOrderSubtotal Money `json:"min_order_subtotal"`
	Price            Money `json:"price"` // the delivery fee
	// Payments are the payment methods allowed with this tier.
	Payments []BranchStorePayment `json:"payments"`
}

// BranchStorePayment is a payment method available for store pickup
// (GET /v1/branch_stores/payments).
type BranchStorePayment struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// BranchStoreUser is a staff account of a branch store. The swagger types
// email as an integer; it is a string.
type BranchStoreUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PrepareDeliveryConfig is a branch store's preparation-time settings, which
// control the delivery date a customer may pick at checkout.
type PrepareDeliveryConfig struct {
	// DeliveryDateEnabled shows the delivery date picker.
	DeliveryDateEnabled bool `json:"delivery_date_enabled"`
	// DeliveryDateRequired makes the picker mandatory.
	DeliveryDateRequired bool `json:"delivery_date_required"`
	// PrepareDeliveryDay is the lead time in days, 1 to 90.
	PrepareDeliveryDay int `json:"prepare_delivery_day"`
	// SelectableRange is how many days after the lead time may be chosen.
	SelectableRange int `json:"selectable_range"`
}
