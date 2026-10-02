package cyberbiz

import "encoding/json/jsontext"

// TaxType (tax_type_id) is shared with orders; see orders_models.go.

// TemperatureType is the storage temperature band of a product.
type TemperatureType string

// Known TemperatureType values; the API uses the Traditional Chinese labels
// themselves as codes (temperature_types).
const (
	TemperatureTypeRoom         TemperatureType = "常溫" // room temperature
	TemperatureTypeRefrigerated TemperatureType = "冷藏" // refrigerated
	TemperatureTypeFrozen       TemperatureType = "冷凍" // frozen
)

// InventoryPolicy says what happens to a variant when its stock reaches zero.
type InventoryPolicy string

// Known InventoryPolicy values (inventory_policy of a variant).
const (
	InventoryPolicyContinue InventoryPolicy = "continue" // keep selling
	InventoryPolicyDeny     InventoryPolicy = "deny"     // stop selling
)

// ProductSearchOrderBy is the sort order accepted by the v1 search endpoints.
type ProductSearchOrderBy string

// Known ProductSearchOrderBy values (order_by of GET /v1/products/search).
const (
	ProductSearchOrderByCreatedDate ProductSearchOrderBy = "created_date" // newest first
	ProductSearchOrderBySalesVolume ProductSearchOrderBy = "sales_volume" // best sellers first
)

// ProductDescriptionSetting names one of the fixed description sections a
// product can carry.
type ProductDescriptionSetting string

// Known ProductDescriptionSetting values (product_description_settings keys).
const (
	ProductDescriptionSettingIntro     ProductDescriptionSetting = "product_description_section_intro"      // introduction
	ProductDescriptionSettingSizeTable ProductDescriptionSetting = "product_description_section_size_table" // size chart
	ProductDescriptionSettingTryReport ProductDescriptionSetting = "product_description_section_try_report" // trial report
	ProductDescriptionSettingDress     ProductDescriptionSetting = "product_description_section_dress"      // styling suggestions
	ProductDescriptionSettingSpec      ProductDescriptionSetting = "product_description_section_spec"       // specifications
	ProductDescriptionSettingShipping  ProductDescriptionSetting = "product_description_section_shipping"   // shipping information
)

// ProductPhotoBatchType selects whether a batch photo operation matches
// variants by SKU or by vendor code (QC).
type ProductPhotoBatchType string

// Known ProductPhotoBatchType values (type of the batch photo endpoints).
const (
	ProductPhotoBatchTypeSKU ProductPhotoBatchType = "sku" // match variants by SKU
	ProductPhotoBatchTypeQC  ProductPhotoBatchType = "qc"  // match variants by vendor code
)

// ProductStoreType is the sales channel filter of the v2 product search.
type ProductStoreType int

// Known ProductStoreType values (store_types of the v2 product search).
const (
	ProductStoreTypeEC          ProductStoreType = 1 // web shop
	ProductStoreTypePOS         ProductStoreType = 2 // point of sale
	ProductStoreTypeBranchStore ProductStoreType = 4 // branch store
)

// BranchStoreInventoryJobStatus is the state of an asynchronous branch-store
// inventory update job. The create response reports QUEUED while the status
// poll reports QUEUE for the same state.
type BranchStoreInventoryJobStatus string

// Known BranchStoreInventoryJobStatus values.
const (
	BranchStoreInventoryJobQueue      BranchStoreInventoryJobStatus = "QUEUE"      // waiting to run (status poll)
	BranchStoreInventoryJobQueued     BranchStoreInventoryJobStatus = "QUEUED"     // waiting to run (create response)
	BranchStoreInventoryJobProcessing BranchStoreInventoryJobStatus = "PROCESSING" // running
	BranchStoreInventoryJobSuccess    BranchStoreInventoryJobStatus = "SUCCESS"    // finished successfully
	BranchStoreInventoryJobFailed     BranchStoreInventoryJobStatus = "FAILED"     // finished with errors
)

// IsDone reports whether the job has finished, successfully or not.
func (s BranchStoreInventoryJobStatus) IsDone() bool {
	return s == BranchStoreInventoryJobSuccess || s == BranchStoreInventoryJobFailed
}

// Product is a product as returned by the product endpoints. List responses
// carry PhotoURLs; the detail response carries Photos instead and omits ID,
// which Get fills in from the path.
type Product struct {
	ID                      int64                        `json:"id"`
	Title                   string                       `json:"title"`
	Handle                  string                       `json:"handle"` // only present in v2 search results
	EnglishTitle            string                       `json:"english_title"`
	ProductURL              string                       `json:"product_url"` // scheme-relative storefront URL
	Published               bool                         `json:"published"`
	SellFrom                *Time                        `json:"sell_from"` // nil when no sales period is set
	SellTo                  *Time                        `json:"sell_to"`   // nil when no sales period is set
	ProductType             string                       `json:"product_type"`
	ProductTypeCode         string                       `json:"product_type_code"`
	Slogan                  string                       `json:"slogan"`
	Brief                   string                       `json:"brief"`
	BriefText               string                       `json:"brief_text"`
	BriefIncludesHTML       bool                         `json:"brief_includes_html"`
	BodyHTML                string                       `json:"body_html"`
	Vendor                  string                       `json:"vendor"`
	Price                   Money                        `json:"price"` // lowest variant price
	SellWeight              float64                      `json:"sell_weight"`
	TaxTypeID               TaxType                      `json:"tax_type_id"`
	CustomCollections       []ProductCollectionRef       `json:"custom_collections"`
	SpecialCollection       *ProductSpecialCollectionRef `json:"special_collection"`
	Tags                    []ProductTag                 `json:"tags"`
	ProductVariants         []ProductVariant             `json:"product_variants"`
	ProductOptions          []ProductOption              `json:"product_options"`
	PosShop                 *ProductPosShopRef           `json:"pos_shop"`
	Photos                  []ProductPhoto               `json:"photos"`     // detail response only
	PhotoURLs               []string                     `json:"photo_urls"` // list and search responses only
	Channel                 *ProductChannel              `json:"channel"`
	RelatedCollections      []ProductRelatedCollection   `json:"related_collections"`
	BranchStore             *ProductBranchStoreRef       `json:"branch_store"`
	CreatedAt               Time                         `json:"created_at"`
	UpdatedAt               Time                         `json:"updated_at"`
	TemperatureTypes        []TemperatureType            `json:"temperature_types"`
	Searchable              bool                         `json:"searchable"`
	GoogleProductCategoryID int64                        `json:"google_product_category_id"`
	// ProductCustomFields is documented as an array of app-defined objects
	// and has never been observed; it is kept raw.
	ProductCustomFields  jsontext.Value      `json:"product_custom_fields,omitzero"`
	SEOMetaTags          *ProductSEOMetaTags `json:"seo_meta_tags"`
	RequiredCustomerTags []string            `json:"required_customer_tags"` // custom feature
}

// ProductVariant is one purchasable variant (SKU) of a product. The detail
// response omits ID, which GetVariant fills in from the path.
type ProductVariant struct {
	ID                      int64            `json:"id"`
	ProductID               int64            `json:"product_id"`
	Name                    string           `json:"name"`
	Position                int              `json:"position"`
	Price                   Money            `json:"price"`
	Cost                    Money            `json:"cost"`
	CompareAtPrice          Money            `json:"compare_at_price"` // list price shown struck through
	Meas                    float64          `json:"meas"`             // volumetric size
	MaxUsableBonus          Money            `json:"max_usable_bonus"` // bonus points cap per unit
	Weight                  float64          `json:"weight"`           // kilograms
	Option1                 string           `json:"option1"`
	Option2                 string           `json:"option2"`
	Option3                 string           `json:"option3"`
	InventoryManagement     bool             `json:"inventory_management"`
	InventoryQuantity       int              `json:"inventory_quantity"`
	Sold                    int              `json:"sold"`
	SafetyInventoryQuantity int              `json:"safety_inventory_quantity"`
	InventoryPolicy         InventoryPolicy  `json:"inventory_policy"`
	SKU                     string           `json:"sku"`
	QC                      string           `json:"qc"` // vendor code
	RequiresShipping        bool             `json:"requires_shipping"`
	CreatedAt               Time             `json:"created_at"`
	UpdatedAt               Time             `json:"updated_at"`
	HoneycombSync           bool             `json:"honeycomb_sync"` // warehouse stock sync
	Vendor                  string           `json:"vendor"`
	PhotoURLs               []string         `json:"photo_urls"`
	PimInfos                []ProductPimInfo `json:"pim_infos"`
}

// ProductPimInfo links a variant to a product information management
// channel. It is documented but has not been observed in real traffic.
type ProductPimInfo struct {
	ID               int64  `json:"id"`
	ProductID        int64  `json:"product_id"`
	ProductVariantID int64  `json:"product_variant_id"`
	PimProductID     int64  `json:"pim_product_id"`
	PimVariantID     int64  `json:"pim_variant_id"`
	Channel          int64  `json:"channel"`
	ChannelShopName  string `json:"channel_shop_name"`
	IsConnected      bool   `json:"is_connected"`
	IsSource         bool   `json:"is_source"`
	ShopID           int64  `json:"shop_id"`
}

// ProductOption is one option axis of a product, e.g. "size" with the
// comma-separated Types "S,M,L". The detail response omits ID, which
// GetOption fills in from the path.
type ProductOption struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Types    string `json:"types"` // comma-separated option values
}

// ProductPhoto is one product image.
type ProductPhoto struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"` // scheme-relative CDN URL
	Position int    `json:"position"`
}

// ProductTag is a tag attached to a product.
type ProductTag struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category int    `json:"category"`
}

// ProductDescription is one description section of a product. The detail
// response omits ID, which GetDescription fills in from the path.
type ProductDescription struct {
	ID          int64                     `json:"id"`
	SettingName ProductDescriptionSetting `json:"setting_name"`
	BodyHTML    string                    `json:"body_html"`
}

// ProductDescriptionSettingName is a description section the shop supports.
type ProductDescriptionSettingName struct {
	SettingName ProductDescriptionSetting `json:"setting_name"`
	Title       string                    `json:"title"`
}

// ProductCollectionRef is the summary of a custom collection embedded in a
// product.
type ProductCollectionRef struct {
	ID                int64  `json:"id"`
	Title             string `json:"title"`
	Handle            string `json:"handle"`
	Published         bool   `json:"published"`
	BodyHTML          string `json:"body_html"`
	ProductsOrderName string `json:"products_order_name"`
	Position          int    `json:"position"`
}

// ProductSpecialCollectionRef is the summary of the special (pick-and-mix
// discount) collection a product belongs to.
type ProductSpecialCollectionRef struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Handle    string `json:"handle"`
	Published bool   `json:"published"`
	Position  int    `json:"position"`
}

// ProductRelatedCollection links a product to a custom or smart collection.
type ProductRelatedCollection struct {
	ID                      int64  `json:"id"`
	RelatableCollectionID   int64  `json:"relatable_collection_id"`
	RelatableCollectionType string `json:"relatable_collection_type"` // "SmartCollection" or "CustomCollection"
}

// ProductPosShopRef is the POS shop a product is sold in.
type ProductPosShopRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ProductChannel is the sales channel a product belongs to.
type ProductChannel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ProductBranchStoreRef is the branch store a product belongs to. It has
// only been observed as null; the fields follow the branch store document.
type ProductBranchStoreRef struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	StoreNo string `json:"store_no"`
	Enabled bool   `json:"enabled"`
}

// ProductSEOMetaTags are the SEO fields of a product.
type ProductSEOMetaTags struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
}

// ProductShippingNames is the envelope of the bind_shippings endpoints.
type ProductShippingNames struct {
	ShippingNames []string `json:"shipping_names"`
}

// ProductSearchResult is the body of the v2 product search, which paginates
// with limit/offset and reports totals in the body instead of headers.
type ProductSearchResult struct {
	Products   []Product `json:"products"`
	TotalCount int       `json:"total_count"`
	TotalPages int       `json:"total_pages"`
}

// BranchStoreInventoryJob is the acknowledgement of a queued batch
// inventory update (PUT /v2/products/batch_update_branch_store_inventory).
type BranchStoreInventoryJob struct {
	JobID       string                        `json:"job_id"`
	Status      BranchStoreInventoryJobStatus `json:"status"`
	Message     string                        `json:"message"`
	StoreNumber string                        `json:"store_number"`
	TotalItems  int                           `json:"total_items"`
}

// BranchStoreInventoryJobState is the polled state of a batch inventory
// update job (GET /v2/products/branch_store_inventory_update_status).
type BranchStoreInventoryJobState struct {
	JobID       string                        `json:"job_id"`
	Status      BranchStoreInventoryJobStatus `json:"status"`
	Message     string                        `json:"message"`
	CreatedAt   Time                          `json:"created_at"`   // zero while the job is still queued
	CompletedAt *Time                         `json:"completed_at"` // nil until the job finishes
	Result      *BranchStoreInventoryResult   `json:"result"`       // nil unless the job succeeded
	Error       string                        `json:"error"`        // set when the job failed
}

// BranchStoreInventoryResult is the outcome of a finished inventory job.
type BranchStoreInventoryResult struct {
	Total       int                              `json:"total"`
	Succeeded   int                              `json:"succeeded"`
	Failed      int                              `json:"failed"`
	FailedItems []BranchStoreInventoryFailedItem `json:"failed_items"`
}

// BranchStoreInventoryFailedItem is one SKU an inventory job could not update.
type BranchStoreInventoryFailedItem struct {
	SKU   string `json:"sku"`
	Error string `json:"error"`
}
