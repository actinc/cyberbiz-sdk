package cyberbiz

// Invoice types accepted by PUT /v1/einvoices/{order_id} in addition to the
// shared [InvoiceType] constants.
const (
	InvoiceTypeLoveCode     InvoiceType = "love_code"     // donation code carrier
	InvoiceTypeClassDefault InvoiceType = "class_default" // 二聯式 paper invoice
	InvoiceTypeClassCompany InvoiceType = "class_company" // 三聯式 paper invoice with company number
)

// InvoiceStatusNotIssued is the "nil" status PUT /v1/einvoices/{order_id}
// accepts to mark an invoice as not yet issued.
const InvoiceStatusNotIssued InvoiceStatus = "nil"

// Einvoice is an electronic invoice (GET /v1/einvoices/{id}). No successful
// Golden File exists (the test shop has no invoices); the shape follows the
// swagger and Postman examples.
type Einvoice struct {
	// Title is the invoice header (抬頭).
	Title   string `json:"title"`
	OrderID int64  `json:"order_id"`
	// CompanyNo is the buyer's tax ID (統一編號).
	CompanyNo     string        `json:"company_no"`
	InvoiceNo     string        `json:"invoice_no"`
	InvoiceStatus InvoiceStatus `json:"invoice_status"`
	InvoiceAt     Time          `json:"invoice_at"`
	// InvalidAt is when the invoice was voided or allowed; nil if never.
	InvalidAt *Time `json:"invalid_at"`
	// RandomNum is the four-digit lottery random code.
	RandomNum   string      `json:"random_num"`
	InvoiceType InvoiceType `json:"invoice_type"`
	// Carrier fields, one of which is set depending on InvoiceType.
	LoveCode     string `json:"love_code"`
	PhoneBarcode string `json:"phone_barcode"`
	NaturePerson string `json:"nature_person"`
}

// OfflineEinvoice is an invoice issued outside the shop, in the Ministry of
// Finance e-invoice field naming, submitted so the customer earns bonus
// points for it. Every field is a string because the source format is.
type OfflineEinvoice struct {
	InvNum        string                  `json:"invNum"`
	InvoiceTime   string                  `json:"invoiceTime"`
	InvStatus     string                  `json:"invStatus"`
	SellerName    string                  `json:"sellerName"`
	InvPeriod     string                  `json:"invPeriod"`
	InvDate       string                  `json:"invDate"`
	SellerAddress string                  `json:"sellerAddress"`
	SellerBan     string                  `json:"sellerBan"`
	BuyerBan      string                  `json:"buyerBan,omitzero"`
	Currency      string                  `json:"currency,omitzero"`
	Details       []OfflineEinvoiceDetail `json:"details"`
}

// OfflineEinvoiceDetail is one line of an OfflineEinvoice.
type OfflineEinvoiceDetail struct {
	UnitPrice   string `json:"unitPrice"`
	Amount      string `json:"amount"`
	Quantity    string `json:"quantity"`
	RowNum      string `json:"rowNum"`
	Description string `json:"description"`
	SKU         string `json:"sku"`
}
