package cyberbiz

// StockService exposes the POS stock documents and inventory sync groups:
//
//   - stock receipts (進倉單), stock_receipts.go
//   - stock requisitions (調倉單), stock_requisitions.go
//   - stock adjustments (庫存調整), stock_adjustments.go
//   - stock invoices (出倉單), stock_invoices.go
//   - inventory sync groups (庫存同步群組), stock_inventory_sync.go
//
// Throughout, a POS shop id of 0 means the web shop (EC) and -1 a third
// party; filters that take a shop id are pointers so that 0 can be sent.
type StockService struct {
	client *Client
}

// StockStatus is the processing state of a stock receipt, invoice or
// requisition. Requisitions only use pending, done and canceled.
type StockStatus string

// Known StockStatus values (status of a stock document).
const (
	StockStatusPending    StockStatus = "pending"    // awaiting confirmation
	StockStatusReady      StockStatus = "ready"      // approved, stock not yet moved
	StockStatusProcessing StockStatus = "processing" // stock movement in progress
	StockStatusDone       StockStatus = "done"       // stock moved
	StockStatusCanceled   StockStatus = "canceled"   // cancelled before completion
)

// StockLineItem is one SKU line on a stock document.
type StockLineItem struct {
	SKU      string `json:"sku"`
	QC       string `json:"qc"` // vendor product code, usually empty
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// StockLineItemInput is one SKU line when creating or checking a stock
// document.
type StockLineItemInput struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}
