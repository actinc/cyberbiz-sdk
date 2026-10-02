package cyberbiz

// Status enumerations shared by orders, coupons and invoices. Every enum is a
// string type whose known values are constants; a value the platform adds
// later still decodes, it just will not match any constant. Resource-specific
// enums live next to their models.

// OrderStatus is the lifecycle state of an order.
type OrderStatus string

// Known OrderStatus values (statuses.order_status).
const (
	OrderStatusOpen      OrderStatus = "open"      // in progress
	OrderStatusClosed    OrderStatus = "closed"    // completed
	OrderStatusCancelled OrderStatus = "cancelled" // cancelled
)

// FinancialStatus is the payment state of an order.
type FinancialStatus string

// Known FinancialStatus values (statuses.financial_status). ExceptAbandonedOrder
// and All are list-filter pseudo-values, not states an order can be in.
const (
	FinancialStatusPaid                 FinancialStatus = "paid"                   // payment received
	FinancialStatusPending              FinancialStatus = "pending"                // awaiting payment
	FinancialStatusCOD                  FinancialStatus = "cod"                    // cash on delivery
	FinancialStatusFailed               FinancialStatus = "failed"                 // payment failed
	FinancialStatusAbandoned            FinancialStatus = "abandoned"              // customer abandoned the order
	FinancialStatusRefunded             FinancialStatus = "refunded"               // fully refunded
	FinancialStatusNoRefunded           FinancialStatus = "no_refunded"            // refund declined
	FinancialStatusExceptAbandonedOrder FinancialStatus = "except_abandoned_order" // filter: everything but abandoned
	FinancialStatusAll                  FinancialStatus = "all_financial_status"   // filter: every status
	FinancialStatusPendingRefund        FinancialStatus = "pending_refund"         // refund requested
	FinancialStatusProcessing           FinancialStatus = "processing"             // handled by customer service
	FinancialStatusRemitted             FinancialStatus = "remitted"               // bank transfer reported, not yet received
	FinancialStatusPendingPartialRefund FinancialStatus = "pending_partial_refund" // partial refund requested
	FinancialStatusPartialRefunded      FinancialStatus = "partial_refunded"       // partially refunded
	FinancialStatusRefunding            FinancialStatus = "refunding"              // refund in progress
	FinancialStatusRefundFailed         FinancialStatus = "refund_failed"          // refund failed
)

// IsPaid reports whether the platform considers the order paid, per the
// CYBERBIZ status documentation: paid, remitted, and the partial-refund states
// all mean money was collected.
func (s FinancialStatus) IsPaid() bool {
	switch s {
	case FinancialStatusPaid, FinancialStatusRemitted,
		FinancialStatusPendingPartialRefund, FinancialStatusPartialRefunded:
		return true
	}
	return false
}

// FulfillmentStatus is the shipping state of an order.
type FulfillmentStatus string

// Known FulfillmentStatus values (statuses.fulfillment_status).
const (
	FulfillmentStatusUnshipped FulfillmentStatus = "unshipped" // not yet shipped
	FulfillmentStatusPreparing FulfillmentStatus = "preparing" // being prepared for shipment
	FulfillmentStatusCancel    FulfillmentStatus = "cancel"    // shipment cancelled
	FulfillmentStatusFulfilled FulfillmentStatus = "fulfilled" // shipped
	FulfillmentStatusPartial   FulfillmentStatus = "partial"   // partially shipped
	FulfillmentStatusArrived   FulfillmentStatus = "arrived"   // arrived at the pickup store
	FulfillmentStatusReceived  FulfillmentStatus = "received"  // received by the customer
	FulfillmentStatusReturned  FulfillmentStatus = "returned"  // returned to the merchant
	FulfillmentStatusExpired   FulfillmentStatus = "expired"   // pickup window expired
	FulfillmentStatusProblem   FulfillmentStatus = "problem"   // delivery exception
	FulfillmentStatusNoNeed    FulfillmentStatus = "no_need"   // nothing to ship
)

// ReturnStatus is the return state of an order or line item.
type ReturnStatus string

// Known ReturnStatus values (statuses.return_status).
const (
	ReturnStatusNoNeed        ReturnStatus = "no_need"        // no return involved
	ReturnStatusRequestReturn ReturnStatus = "request_return" // return requested
	ReturnStatusReturning     ReturnStatus = "returning"      // goods on their way back
	ReturnStatusChecking      ReturnStatus = "checking"       // returned goods under inspection
	ReturnStatusReturned      ReturnStatus = "returned"       // return completed
	ReturnStatusInHub         ReturnStatus = "in_hub"         // at the carrier's transfer hub
	ReturnStatusProblem       ReturnStatus = "problem"        // delivery exception
	ReturnStatusProcessing    ReturnStatus = "processing"     // handled by customer service
	ReturnStatusInOriginCVS   ReturnStatus = "in_origin_cvs"  // at the originating convenience store
	ReturnStatusRefused       ReturnStatus = "refused"        // return rejected
	ReturnStatusPartialReturn ReturnStatus = "partial_return" // partially returned
)

// CouponStatus is the usability state of a customer's coupon. It must be
// read together with [GiftOrderStatus]; see [CouponUsable].
type CouponStatus string

// Known CouponStatus values (status of a customer's coupon).
const (
	CouponStatusNoStartUse    CouponStatus = "no_start_use"    // not yet activated
	CouponStatusUsed          CouponStatus = "used"            // fully used up
	CouponStatusHasExpireDate CouponStatus = "has_expire_date" // usable, has an expiry date
	CouponStatusNoExpireDate  CouponStatus = "no_expire_date"  // usable, no expiry date
	CouponStatusExpired       CouponStatus = "expired"         // past its expiry date
)

// GiftOrderStatus is the state of the order that granted a coupon.
type GiftOrderStatus string

// Known GiftOrderStatus values (gift_order_status of a customer's coupon).
const (
	GiftOrderStatusClosed    GiftOrderStatus = "closed"    // granting order completed
	GiftOrderStatusOpen      GiftOrderStatus = "open"      // granting order in progress
	GiftOrderStatusCancelled GiftOrderStatus = "cancelled" // granting order cancelled; coupon unusable
)

// CouponUsable applies the documented combination rule: a coupon is usable
// only when its own status says so and the order that granted it was not
// cancelled. An empty gift order status means the coupon was not granted by
// an order and only the coupon status counts.
func CouponUsable(status CouponStatus, giftOrder GiftOrderStatus) bool {
	if giftOrder == GiftOrderStatusCancelled {
		return false
	}
	return status == CouponStatusHasExpireDate || status == CouponStatusNoExpireDate
}

// InvoiceStatus is the state of an e-invoice.
type InvoiceStatus string

// Known InvoiceStatus values (invoice_status of an e-invoice); an empty
// value means no invoice has been issued.
const (
	InvoiceStatusIssue            InvoiceStatus = "issue"             // issued
	InvoiceStatusIssueInvalid     InvoiceStatus = "issue_invalid"     // voided
	InvoiceStatusAllowance        InvoiceStatus = "allowance"         // fully credited (allowance)
	InvoiceStatusPartialAllowance InvoiceStatus = "partial_allowance" // partially credited
)

// InvoiceType is how an e-invoice is issued or carried.
type InvoiceType string

// Known InvoiceType values (invoice_type of an e-invoice).
const (
	InvoiceTypeDefault      InvoiceType = "default"       // personal e-invoice
	InvoiceTypeCompany      InvoiceType = "company"       // company e-invoice with tax ID
	InvoiceTypePhoneBarcode InvoiceType = "phone_barcode" // mobile barcode carrier
	InvoiceTypeNaturePerson InvoiceType = "nature_person" // citizen digital certificate carrier
	InvoiceTypeDonate       InvoiceType = "donate"        // donated to a charity code
	InvoiceTypePaperInvoice InvoiceType = "paper_invoice" // paper invoice
)

// DataSource distinguishes web-shop orders from point-of-sale orders.
type DataSource string

// Known DataSource values (data_source of an order).
const (
	DataSourceEC  DataSource = "ec"  // placed on the online shop
	DataSourcePOS DataSource = "pos" // placed at a point-of-sale terminal
)
