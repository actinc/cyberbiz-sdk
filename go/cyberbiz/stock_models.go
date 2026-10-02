package cyberbiz

// StockReceipt is an inbound stock document (進倉單) for a POS shop. The
// swagger documents items and check_logs as objects; the platform returns
// arrays.
type StockReceipt struct {
	ID int64 `json:"id"`
	// PosShopName and PosShopID identify the receiving shop; 0 is the web
	// shop (EC).
	PosShopName string `json:"pos_shop_name"`
	PosShopID   int64  `json:"pos_shop_id"`
	// SourcePosShopName and SourcePosShopID identify the shipping shop;
	// 0 is EC and -1 a third party.
	SourcePosShopName string `json:"source_pos_shop_name"`
	SourcePosShopID   int64  `json:"source_pos_shop_id"`
	Barcode           string `json:"barcode"`
	// Status is the processing state; see [StockStatus].
	Status StockStatus `json:"status"`
	// StockInvoiceID is the outbound invoice this receipt was created from,
	// 0 when the receipt was created directly.
	StockInvoiceID int64           `json:"stock_invoice_id"`
	Items          []StockLineItem `json:"items"`
	// CheckLogs are the counts recorded through CheckReceipt.
	CheckLogs []StockCheckLog `json:"check_logs"`
	CreatedAt Time            `json:"created_at"`
	Comment   string          `json:"comment"`
}

// StockCheckLog is one count (點收) recorded against a stock receipt.
type StockCheckLog struct {
	Items     []StockLineItem `json:"items"`
	CreatedAt Time            `json:"created_at"`
}

// StockRequisition is a stock transfer request (調倉單) from one shop to
// another. The API returns shop names only, not ids.
type StockRequisition struct {
	ID int64 `json:"id"`
	// PosShopName is the shop that receives the stock.
	PosShopName string `json:"pos_shop_name"`
	// SourceShopName is the shop that gives up the stock.
	SourceShopName string `json:"source_shop_name"`
	Barcode        string `json:"barcode"`
	// Status is pending, done or canceled; see [StockStatus].
	Status    StockStatus     `json:"status"`
	Items     []StockLineItem `json:"items"`
	CreatedAt Time            `json:"created_at"`
	Comment   string          `json:"comment"`
}

// StockAdjustmentType is why an inventory quantity was adjusted.
type StockAdjustmentType string

// Known StockAdjustmentType values (type of a stock adjustment item).
const (
	StockAdjustmentTypeSurplus StockAdjustmentType = "surplus" // 盤盈, count found more
	StockAdjustmentTypeLoss    StockAdjustmentType = "loss"    // 盤虧, count found less
	StockAdjustmentTypeSold    StockAdjustmentType = "sold"    // 銷貨, sold outside the platform
	StockAdjustmentTypeReturn  StockAdjustmentType = "return"  // 退貨, returned goods restocked
)

// StockAdjustment is one inventory adjustment record (庫存調整) and its
// lines. The API returns nothing but the id and the items.
type StockAdjustment struct {
	ID    int64                 `json:"id"`
	Items []StockAdjustmentItem `json:"items"`
}

// StockAdjustmentItem is one SKU line of an inventory adjustment.
type StockAdjustmentItem struct {
	SKU      string `json:"sku"`
	QC       string `json:"qc"` // vendor product code, usually empty
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Note     string `json:"note"`
	// Type is why the quantity changed; see [StockAdjustmentType].
	Type      StockAdjustmentType `json:"type"`
	CreatedAt Time                `json:"created_at"`
}

// StockAdjustmentItemInput is one SKU line when creating an adjustment.
type StockAdjustmentItemInput struct {
	SKU      string              `json:"sku"`
	Quantity int                 `json:"quantity"`
	Type     StockAdjustmentType `json:"type"`
}

// StockInvoiceTargetType is where an outbound stock invoice ships to.
type StockInvoiceTargetType string

// Known StockInvoiceTargetType values (target_type of a stock invoice).
const (
	StockInvoiceTargetEC         StockInvoiceTargetType = "ec"          // the web shop
	StockInvoiceTargetPosShop    StockInvoiceTargetType = "pos_shop"    // another POS shop
	StockInvoiceTargetThirdParty StockInvoiceTargetType = "third_party" // an outside party
)

// StockInvoice is an outbound stock document (出倉單). The swagger documents
// items as an object; the platform returns an array.
type StockInvoice struct {
	ID int64 `json:"id"`
	// PosShopName and PosShopID identify the shipping shop; 0 is EC.
	PosShopName string `json:"pos_shop_name"`
	PosShopID   int64  `json:"pos_shop_id"`
	// TargetPosShopName and TargetPosShopID identify the receiving shop;
	// 0 is EC and -1 a third party.
	TargetPosShopName string `json:"target_pos_shop_name"`
	TargetPosShopID   int64  `json:"target_pos_shop_id"`
	Barcode           string `json:"barcode"`
	// Status is the processing state; see [StockStatus].
	Status StockStatus `json:"status"`
	// StockReceiptID is the inbound receipt created for the target shop,
	// 0 until the invoice is confirmed or when the target is a third party.
	StockReceiptID int64           `json:"stock_receipt_id"`
	Items          []StockLineItem `json:"items"`
	CreatedAt      Time            `json:"created_at"`
	Comment        string          `json:"comment"`
}

// InventorySyncPolicy is what a sync group does when its shared inventory
// reaches zero.
type InventorySyncPolicy string

// Known InventorySyncPolicy values (inventory_policy of a sync group).
const (
	InventorySyncPolicyContinue InventorySyncPolicy = "continue" // keep selling
	InventorySyncPolicyDeny     InventorySyncPolicy = "deny"     // stop selling
)

// InventorySyncGroup is a set of product variants that share one inventory
// count (庫存同步群組). No Golden File exists for this resource: the shop used
// for recording does not have the feature (422). The shape follows the
// swagger and Postman example.
type InventorySyncGroup struct {
	// ID is absent from the detail response; Get fills it in.
	ID                int64  `json:"id"`
	Title             string `json:"title"`
	InventoryQuantity int    `json:"inventory_quantity"`
	// InventoryPolicy is continue or deny; see [InventorySyncPolicy].
	InventoryPolicy InventorySyncPolicy         `json:"inventory_policy"`
	Variants        []InventorySyncGroupVariant `json:"variants"`
}

// InventorySyncGroupVariant is one product variant's membership in a sync
// group.
type InventorySyncGroupVariant struct {
	ID                   int64 `json:"id"`
	InventorySyncGroupID int64 `json:"inventory_sync_group_id"`
	ProductVariantID     int64 `json:"product_variant_id"`
	CreatedAt            Time  `json:"created_at"`
	UpdatedAt            Time  `json:"updated_at"`
}
