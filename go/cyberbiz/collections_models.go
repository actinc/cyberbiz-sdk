package cyberbiz

import (
	"encoding/json/jsontext"
	"errors"
)

// Response models and enums for every collection kind. Custom and smart
// collections are verified against Golden Files; special, add-buy and
// variant-discount lists were recorded empty on the test shop; VIP
// collections returned 401 "feature not licensed" there, so those models are
// modelled from the swagger and Postman examples only.

// CollectionProductRef is the minimal product record embedded in a
// collection's product list. Position is absent (zero) for special
// collections, whose product list is unordered.
type CollectionProductRef struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Position int    `json:"position"` // 1-based order inside the collection
}

// CollectionVariantRef is the product variant record embedded in a
// variant-discount collection.
type CollectionVariantRef struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	ProductTitle string `json:"product_title"`
	Option1      string `json:"option1"`
	Option2      string `json:"option2"`
	Option3      string `json:"option3"`
	SKU          string `json:"sku"`
}

// ProductsOrder is how a collection orders its products on the storefront.
// Requests take the code; responses (products_order_name) carry the
// Traditional Chinese label instead, e.g. "按標題拼音升序: A-Z".
type ProductsOrder string

// Known ProductsOrder codes accepted by the products_order request field.
const (
	ProductsOrderTitleAsc       ProductsOrder = "title.asc"        // title A-Z
	ProductsOrderTitleDesc      ProductsOrder = "title.desc"       // title Z-A
	ProductsOrderCreatedAtDesc  ProductsOrder = "created_at.desc"  // newest first
	ProductsOrderCreatedAtAsc   ProductsOrder = "created_at.asc"   // oldest first
	ProductsOrderPriceDesc      ProductsOrder = "price.desc"       // highest price first
	ProductsOrderPriceAsc       ProductsOrder = "price.asc"        // lowest price first
	ProductsOrderManual         ProductsOrder = "manual"           // merchant-arranged; not accepted by limit collections
	ProductsOrderSellWeightDesc ProductsOrder = "sell_weight.desc" // best sellers first
)

// CustomCollection is a hand-picked product group
// (GET /v1/custom_collections, GET /v2/custom_collections).
type CustomCollection struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Handle    string `json:"handle"` // storefront path /collections/{handle}
	Published bool   `json:"published"`
	BodyHTML  string `json:"body_html"`
	// ProductsOrderName is the label of the product ordering, not the
	// [ProductsOrder] code.
	ProductsOrderName string `json:"products_order_name"`
	// Position orders collections in storefront navigation; smaller first.
	Position int `json:"position"`
	// Products is only present on the detail endpoint.
	Products []CollectionProductRef `json:"products"`
}

// SmartCollection is a rule-based product group
// (GET /v1/smart_collections, GET /v2/smart_collections).
type SmartCollection struct {
	ID        int64                  `json:"id"`
	Title     string                 `json:"title"`
	Handle    string                 `json:"handle"`
	Published bool                   `json:"published"`
	BodyHTML  string                 `json:"body_html"`
	Position  int                    `json:"position"`
	Rules     []SmartCollectionRule  `json:"rules"`
	Products  []CollectionProductRef `json:"products"` // products matching the rules
}

// SmartCollectionRule is one filter of a smart collection. Filter is the
// platform's rendered description, e.g. "商品價格 小於 1000"; the column,
// relation and condition it was created from are not returned.
type SmartCollectionRule struct {
	ID     int64  `json:"id"`
	Filter string `json:"filter"`
}

// SmartRuleColumn is the product attribute a smart collection rule filters on.
type SmartRuleColumn string

// Known SmartRuleColumn values (column of a smart collection rule).
const (
	SmartRuleColumnTitle                     SmartRuleColumn = "title"                       // product title
	SmartRuleColumnProductType               SmartRuleColumn = "product_type"                // product type
	SmartRuleColumnVendor                    SmartRuleColumn = "vendor"                      // product vendor
	SmartRuleColumnVariantsPrice             SmartRuleColumn = "variants_price"              // variant selling price
	SmartRuleColumnVariantsCompareAtPrice    SmartRuleColumn = "variants_compare_at_price"   // variant list (compare-at) price
	SmartRuleColumnVariantsInventoryQuantity SmartRuleColumn = "variants_inventory_quantity" // variant stock on hand
	SmartRuleColumnVariantsOption1           SmartRuleColumn = "variants_option1"            // first variant option value
	SmartRuleColumnTagsName                  SmartRuleColumn = "tags_name"                   // product tag name
)

// SmartRuleRelation is the comparison a smart collection rule applies.
type SmartRuleRelation string

// Known SmartRuleRelation values (relation of a smart collection rule).
const (
	SmartRuleRelationEquals      SmartRuleRelation = "eq"    // equals the condition
	SmartRuleRelationGreaterThan SmartRuleRelation = "gt"    // greater than the condition
	SmartRuleRelationLessThan    SmartRuleRelation = "lt"    // less than the condition
	SmartRuleRelationStartsWith  SmartRuleRelation = "start" // starts with the condition
	SmartRuleRelationEndsWith    SmartRuleRelation = "end"   // ends with the condition
	SmartRuleRelationContains    SmartRuleRelation = "cont"  // contains the condition
	SmartRuleRelationExcludes    SmartRuleRelation = "exc"   // does not contain the condition
)

// SpecialCollectionType is the pricing rule of a special ("任選折扣")
// collection.
type SpecialCollectionType string

// Known SpecialCollectionType values (special_collection_type).
const (
	SpecialCollectionTypeAmount      SpecialCollectionType = "amount"       // any N for a fixed price
	SpecialCollectionTypePercentage  SpecialCollectionType = "percentage"   // any N at a percentage
	SpecialCollectionTypeDiscount    SpecialCollectionType = "discount"     // any N minus a fixed amount
	SpecialCollectionTypePerDiscount SpecialCollectionType = "per_discount" // fixed amount off each item
)

// SpecialCollection is a mix-and-match discount group
// (GET /v1/special_collections).
type SpecialCollection struct {
	ID                    int64                      `json:"id"`
	Title                 string                     `json:"title"`
	Handle                string                     `json:"handle"`
	Published             bool                       `json:"published"`
	StartDate             Time                       `json:"start_date"`
	EndDate               Time                       `json:"end_date"`
	BodyHTML              string                     `json:"body_html"`
	Position              int                        `json:"position"`
	SpecialCollectionType *SpecialCollectionTypeInfo `json:"special_collection_type"`
	TypeRules             []SpecialCollectionRule    `json:"type_rules"`
	// RestIncludeDiscount reports whether items beyond the rule quantity
	// are still discounted.
	RestIncludeDiscount bool                   `json:"rest_include_discount"`
	Products            []CollectionProductRef `json:"products"`
}

// SpecialCollectionTypeInfo is the pricing rule of a special collection,
// as a code plus its display name.
type SpecialCollectionTypeInfo struct {
	Name string                `json:"name"`
	Code SpecialCollectionType `json:"code"`
}

// SpecialCollectionRule is one quantity tier of a special collection.
// Price applies to the amount, discount and per_discount types; Percentage
// applies to the percentage type.
type SpecialCollectionRule struct {
	ID         int64   `json:"id"`
	Quantity   int     `json:"quantity"`
	Price      Money   `json:"price"`
	Percentage float64 `json:"percentage"`
}

// AddBuyCollection is an add-on purchase ("加價購") group
// (GET /v1/add_buy_collections).
type AddBuyCollection struct {
	ID                int64  `json:"id"`
	Title             string `json:"title"`
	Handle            string `json:"handle"`
	Published         bool   `json:"published"`
	ProductsOrderName string `json:"products_order_name"` // label, see [ProductsOrder]
	Position          int    `json:"position"`
	StartDate         Time   `json:"start_date"`
	EndDate           Time   `json:"end_date"`
	// Price is the minimum order subtotal that unlocks the add-on offer.
	Price Money `json:"price"`
	// ItemLimit is the maximum number of add-on items per order.
	ItemLimit int                    `json:"item_limit"`
	Products  []CollectionProductRef `json:"products"`
}

// VariantDiscountType is the pricing rule of a variant-discount collection.
type VariantDiscountType string

// Known VariantDiscountType values (variant_discount_collection_type_code).
const (
	VariantDiscountTypeAmount         VariantDiscountType = "amount"          // fixed unit price
	VariantDiscountTypePercentage     VariantDiscountType = "percentage"      // percentage off each unit
	VariantDiscountTypeDiscountAmount VariantDiscountType = "discount_amount" // fixed amount off
)

// VariantDiscountCollection is a per-variant discount group
// (GET /v1/variant_discount_collections). Only the empty list was recorded
// on the test shop; the detail shape follows the swagger and Postman.
type VariantDiscountCollection struct {
	ID int64 `json:"id"`
	// ShopID is documented as a string but is an id like every other; it is
	// decoded as int64.
	ShopID    int64  `json:"shop_id"`
	Title     string `json:"title"`
	Published bool   `json:"published"`
	// Amount is the fixed unit price for the amount type.
	Amount Money `json:"amount"`
	// DiscountAmount is the fixed amount off for the discount_amount type.
	DiscountAmount Money `json:"discount_amount"`
	// Percentage is the percentage off each unit for the percentage type.
	Percentage                        float64                `json:"percentage"`
	VariantDiscountCollectionTypeID   int64                  `json:"variant_discount_collection_type_id"`
	VariantDiscountCollectionTypeCode VariantDiscountType    `json:"variant_discount_collection_type_code"`
	StartDate                         Time                   `json:"start_date"`
	EndDate                           Time                   `json:"end_date"`
	Variants                          []CollectionVariantRef `json:"variants"`
}

// VIPRuleType is the condition that qualifies a customer for a VIP level.
type VIPRuleType string

// Known VIPRuleType values (rule.type of a VIP collection).
const (
	VIPRuleOrderCount   VIPRuleType = "order_count"   // number of orders
	VIPRuleTotalPrice   VIPRuleType = "total_price"   // accumulated spend
	VIPRuleCustomersTag VIPRuleType = "customers_tag" // customer carries a tag
)

// VIPPromotionType is the benefit a VIP level grants.
type VIPPromotionType string

// Known VIPPromotionType values (promotion.type of a VIP collection).
const (
	VIPPromotionDiscount     VIPPromotionType = "discount"      // order discount for members of the level
	VIPPromotionFreeShipping VIPPromotionType = "free_shipping" // free shipping for members of the level
)

// vipRuleLabels maps the Chinese labels the platform sends in responses to
// the codes requests use.
var vipRuleLabels = map[string]VIPRuleType{
	"訂單數量": VIPRuleOrderCount,
	"金額累積": VIPRuleTotalPrice,
	"顧客標籤": VIPRuleCustomersTag,
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
// Responses carry the Chinese label (訂單數量) rather than the code; known
// labels are mapped to their code so the constants compare equal. Unknown
// values are kept as sent.
func (t *VIPRuleType) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	s, err := decodeLabel(dec, "VIPRuleType")
	if err != nil {
		return err
	}
	if code, ok := vipRuleLabels[s]; ok {
		*t = code
		return nil
	}
	*t = VIPRuleType(s)
	return nil
}

// vipPromotionLabels maps the Chinese labels the platform sends in responses
// to the codes requests use.
var vipPromotionLabels = map[string]VIPPromotionType{
	"享優惠":   VIPPromotionDiscount,
	"訂單免運費": VIPPromotionFreeShipping,
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
// Responses carry the Chinese label (享優惠) rather than the code; known
// labels are mapped to their code. Unknown values are kept as sent.
func (t *VIPPromotionType) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	s, err := decodeLabel(dec, "VIPPromotionType")
	if err != nil {
		return err
	}
	if code, ok := vipPromotionLabels[s]; ok {
		*t = code
		return nil
	}
	*t = VIPPromotionType(s)
	return nil
}

// decodeLabel reads a JSON string or null for an enum named name.
func decodeLabel(dec *jsontext.Decoder, name string) (string, error) {
	tok, err := dec.ReadToken()
	if err != nil {
		return "", err
	}
	switch tok.Kind() {
	case 'n':
		return "", nil
	case '"':
		return tok.String(), nil
	}
	return "", errors.New("cyberbiz: " + name + " must be a JSON string or null")
}

// VIPCollection is a VIP level definition (GET /v1/vip_collections). Not
// verified against a Golden File: the test shop has not licensed the
// feature (401 "無權使用該 API"); the shape follows the swagger and Postman.
type VIPCollection struct {
	ID        int64                   `json:"id"`
	Title     string                  `json:"title"`
	Position  int                     `json:"position"` // evaluation priority
	Rule      *VIPCollectionRule      `json:"rule"`
	Promotion *VIPCollectionPromotion `json:"promotion"`
}

// VIPCollectionRule is the qualifying condition of a VIP level. Only the
// fields relevant to RuleType are meaningful.
type VIPCollectionRule struct {
	RuleType VIPRuleType `json:"rule_type"`
	// OrderStart and OrderEnd bound the orders counted toward the rule.
	OrderStart      Date   `json:"order_start"`
	OrderEnd        Date   `json:"order_end"`
	TotalOrderCount int    `json:"total_order_count"`
	TotalPrice      Money  `json:"total_price"`
	CustomersTag    string `json:"customers_tag"`
	// VIPExpiredateStart and VIPExpiredateEnd bound the period the level is
	// held for.
	VIPExpiredateStart Date `json:"vip_expiredate_start"`
	VIPExpiredateEnd   Date `json:"vip_expiredate_end"`
}

// VIPCollectionPromotion is the benefit of a VIP level.
type VIPCollectionPromotion struct {
	PromotionType VIPPromotionType `json:"promotion_type"`
	// Discount is the order discount rate in Taiwanese "折" units (9 means
	// 10% off), only meaningful for the discount type.
	Discount float64 `json:"discount"`
	// ConcurrentlyApply reports whether the benefit stacks with special
	// collections and shop-wide promotions.
	ConcurrentlyApply bool `json:"concurrently_apply"`
}
