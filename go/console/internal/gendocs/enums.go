package gendocs

// enumDoc is the documented value set of a status-like field, from the
// CYBERBIZ Notion pages (order statuses, coupon statuses, wallet types).
type enumDoc struct {
	Values      []any
	Description string
}

// couponRule is the combination rule from the "Coupons API 狀態欄位說明" page.
const couponRule = "Read together with `gift_order_status`: a coupon whose source order was cancelled is invalid even when `coupon_status` alone looks usable (e.g. `coupon_status=no_start_use` with `gift_order_status=cancelled` means the coupon must be treated as invalid)."

// enumDocs is keyed by field name; applied wherever the field appears.
var enumDocs = map[string]enumDoc{
	"order_status": {
		Values:      []any{"open", "closed", "cancelled"},
		Description: "Order status: `open` (in progress), `closed` (completed), `cancelled`.",
	},
	"financial_status": {
		Values:      []any{"paid", "pending", "cod", "failed", "abandoned", "refunded", "no_refunded", "pending_refund", "processing", "remitted", "pending_partial_refund", "partial_refunded", "refunding", "refund_failed"},
		Description: "Payment status: `paid`, `pending` (awaiting payment), `cod` (cash on delivery), `failed`, `abandoned`, `refunded`, `no_refunded`, `pending_refund`, `processing` (customer service), `remitted` (transfer reported, not received), `pending_partial_refund`, `partial_refunded`, `refunding`, `refund_failed`. The filter pseudo-values `except_abandoned_order` and `all_financial_status` are query helpers, not states.",
	},
	"fulfillment_status": {
		Values:      []any{"unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived", "received", "returned", "expired", "problem", "no_need"},
		Description: "Shipping status: `unshipped`, `preparing`, `cancel`, `fulfilled` (shipped), `partial`, `arrived` (at store), `received`, `returned`, `expired` (not picked up), `problem`, `no_need`.",
	},
	"return_status": {
		Values:      []any{"no_need", "request_return", "returning", "checking", "returned", "in_hub", "problem", "processing", "in_origin_cvs", "refused", "partial_return"},
		Description: "Return status: `no_need`, `request_return`, `returning`, `checking`, `returned`, `in_hub`, `problem`, `processing`, `in_origin_cvs`, `refused`, `partial_return`.",
	},
	"coupon_status": {
		Values:      []any{"no_start_use", "used", "has_expire_date", "no_expire_date", "expired"},
		Description: "Coupon status: `no_start_use` (not yet usable), `used` (fully used up), `has_expire_date` (usable, expires), `no_expire_date` (usable, never expires), `expired`. " + couponRule,
	},
	"gift_order_status": {
		Values:      []any{"closed", "open", "cancelled"},
		Description: "Status of the order that granted the coupon: `closed` (completed), `open` (not yet completed), `cancelled`. " + couponRule,
	},
	"invoice_status": {
		Values:      []any{"issue", "issue_invalid", "allowance", "partial_allowance"},
		Description: "E-invoice status: `null` (not issued), `issue` (issued), `issue_invalid` (voided), `allowance`, `partial_allowance`.",
	},
	"invoice_type": {
		Values:      []any{"default", "company", "phone_barcode", "nature_person", "donate", "paper_invoice"},
		Description: "E-invoice type: `default`, `company` (with tax id), `phone_barcode` (mobile barcode carrier), `nature_person` (citizen digital certificate), `donate`, `paper_invoice` (number in `paper_invoice_no`).",
	},
	"tax_type_id": {
		Values:      []any{"inclusive_tax", "zero_tax", "exclusive_tax"},
		Description: "Tax type: `inclusive_tax` (taxable), `zero_tax` (zero-rated), `exclusive_tax` (tax exempt).",
	},
}

// applyEnumDocs adds the documented enum and description to every schema
// property whose name is in enumDocs and whose type is a string.
func applyEnumDocs(doc *Document, log *report) {
	walkDocumentSchemas(doc, func(key string, s *Schema) *Schema {
		ed, ok := enumDocs[key]
		if !ok || s.BaseType() != "string" || len(s.Enum) > 0 {
			return s
		}
		s.Enum = append([]any(nil), ed.Values...)
		if s.IsNullable() {
			s.Enum = append(s.Enum, nil)
		}
		s.Description = ed.Description
		log.count("enum-docs")
		return s
	})
}
