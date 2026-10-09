package cyberbiz

import (
	"encoding/json/jsontext"
	"errors"
	"strconv"
)

// PeriodicType is the cadence of a periodic order.
type PeriodicType string

// Known PeriodicType values (periodic.type of a periodic order).
const (
	PeriodicTypeWeekly     PeriodicType = "weekly"       // every N weeks on a weekday
	PeriodicTypeMonthly    PeriodicType = "monthly"      // every N months on the Nth weekday
	PeriodicTypeDayOfMonth PeriodicType = "day_of_month" // every N months on a calendar day
)

// PeriodicOrder is a subscription parent order (GET /v1/periodic_orders).
// The test shop has none, so the shape follows the swagger and Postman
// examples.
type PeriodicOrder struct {
	ID         int64 `json:"id"`
	CustomerID int64 `json:"customer_id"`
	// Number is the parent order number. The platform stores a sequence
	// integer and sends a JSON number; older examples show a string.
	Number    PeriodicNumber     `json:"number"`
	SalesPage *PeriodicSalesPage `json:"sales_page"`
	// RecentPreorderID is the most recent child pre-order.
	RecentPreorderID int64 `json:"recent_preorder_id"`
	PreordersCount   int   `json:"preorders_count"`
	// OrderPrepareDays is how many days before delivery a child order is created.
	OrderPrepareDays int               `json:"order_prepare_days"`
	Periodic         *PeriodicSchedule `json:"periodic"`
	// StartDate, NextDate and LastDate are when the first, next and final
	// child orders are created.
	StartDate Date `json:"start_date"`
	NextDate  Date `json:"next_date"`
	LastDate  Date `json:"last_date"`
	// DeliveryTime is the requested delivery slot, 0 to 3.
	DeliveryTime    int                      `json:"delivery_time"`
	Note            string                   `json:"note"`
	BillingAddress  *PeriodicBillingAddress  `json:"billing_address"`
	CreatedAt       Time                     `json:"created_at"`
	AffiliateVendor *PeriodicAffiliateVendor `json:"affiliate_vendor"`
}

// PeriodicNumber is the number of a periodic order as text. It accepts a
// JSON number (what the platform sends), a string, or null.
type PeriodicNumber string

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
func (n *PeriodicNumber) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		*n = ""
	case '"', '0':
		*n = PeriodicNumber(tok.String())
	default:
		return errors.New("cyberbiz: PeriodicNumber must be a JSON number, string or null")
	}
	return nil
}

// PeriodicSalesPage is the sales page a periodic order was placed from.
type PeriodicSalesPage struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// PeriodicSchedule is the cadence of a periodic order. Which of Month, Week
// and Day are meaningful depends on Type: monthly uses Month (1-4) and Week
// or Day (0-6), weekly uses Week (1-8) and Day (0-6), day_of_month uses
// Month (1-12) and Day (1-31). The platform stores the schedule as a raw
// hash: values written through the API come back as strings ("3"), others
// as numbers, so each is a PeriodicValue.
type PeriodicSchedule struct {
	Type  PeriodicType  `json:"type"`
	Month PeriodicValue `json:"month"`
	Week  PeriodicValue `json:"week"`
	Day   PeriodicValue `json:"day"`
}

// PeriodicValue is one number of a periodic schedule. It accepts a JSON
// number, a numeric string, or null.
type PeriodicValue int

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
func (v *PeriodicValue) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		*v = 0
		return nil
	case '0', '"':
		if tok.String() == "" {
			*v = 0
			return nil
		}
		n, err := strconv.Atoi(tok.String())
		if err != nil {
			return errors.New("cyberbiz: invalid periodic value " + strconv.Quote(tok.String()))
		}
		*v = PeriodicValue(n)
		return nil
	}
	return errors.New("cyberbiz: PeriodicValue must be a JSON number, string or null")
}

// PeriodicBillingAddress is the billing contact of a periodic order. The
// swagger documents it as an untyped object; the fields are those the
// update endpoint accepts.
type PeriodicBillingAddress struct {
	Name               string `json:"name"`
	Phone              string `json:"phone"`
	CountryCallingCode string `json:"country_calling_code"`
	Zip                string `json:"zip"`
	City               string `json:"city"`
	District           string `json:"district"`
	Address1           string `json:"address1"`
}

// PeriodicAffiliateVendor is the affiliate that referred a periodic order.
type PeriodicAffiliateVendor struct {
	ID         int64  `json:"id"`
	UID        string `json:"uid"`
	CID        string `json:"cid"`
	VendorName string `json:"vendor_name"`
	// ShopAddOnID is the app installation the affiliate belongs to.
	ShopAddOnID int64 `json:"shop_add_on_id"`
}

// PeriodicPreorder is a child order of a periodic order that has been
// established (GET /v1/periodic_orders/{id}/preorders/established).
type PeriodicPreorder struct {
	ID int64 `json:"id"`
	// Number is the child's sequence number within the parent.
	Number int `json:"number"`
	// ShippingAddress is documented as an untyped object.
	ShippingAddress jsontext.Value `json:"shipping_address"`
	// CancelAt is when the child was cancelled; nil if it was not.
	CancelAt       *Time  `json:"cancel_at"`
	CancelReason   string `json:"cancel_reason"`
	FulfilledTimes int    `json:"fulfilled_times"`
	// DeliveryDate is the actual and BaseDeliveryDate the planned ship date.
	DeliveryDate     Date `json:"delivery_date"`
	BaseDeliveryDate Date `json:"base_delivery_date"`
	// Order is the full order record, in the same shape as GET
	// /v1/orders/{id}; decode it into Order.
	Order jsontext.Value `json:"order"`
}
