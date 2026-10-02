package gendocs

import (
	"fmt"
	"strings"
)

// buildV2 produces docs/api/<locale>/cyberbiz-openapi-v2.yaml: the v2 endpoints and the
// unprefixed app endpoints, from the Go transcription of the Notion
// reference, the v2 Golden Files and the observed schema.
func buildV2(in *inputs, loc *locale, log *report) (*Document, error) {
	doc := &Document{
		OpenAPI: "3.1.0",
		Info: &Info{
			Title:       "CYBERBIZ API v2 and app endpoints",
			Version:     DocVersion,
			Description: infoDescription(v2Intro, loc),
		},
		Paths:      NewOMap(),
		Components: &Components{Schemas: v2Schemas(loc)},
	}
	seenTags := map[string]bool{}
	for _, ep := range v2Endpoints {
		if !seenTags[ep.Tag] {
			seenTags[ep.Tag] = true
			doc.Tags = append(doc.Tags, &Tag{Name: ep.Tag, Description: tagDescriptions[ep.Tag]})
		}
		itemAny, ok := doc.Paths.Get(ep.Path)
		if !ok {
			itemAny = &PathItem{}
			doc.Paths.Set(ep.Path, itemAny)
		}
		itemAny.(*PathItem).SetOperation(ep.Method, v2Operation(ep, in.golden, loc))
	}
	doc.Paths.SortKeys()
	applyObserved(doc, in.observed, log)
	applyNameRules(doc, log)
	applyEnumDocs(doc, log)
	addSharedComponents(doc, log)
	applySharedParameters(doc, log)
	attachExamples(doc, in.golden, postmanIndex{}, loc, log)
	applyErrorResponses(doc, in.golden, log)
	localizeDocument(doc, loc, log)
	return doc, nil
}

const v2Intro = `# CYBERBIZ API v2 and app endpoints

The ` + "`/v2`" + ` operations and the unprefixed app endpoints (` + "`/shop`, `/settings`" + `) of the CYBERBIZ
e-commerce platform. There is no machine-readable v2 specification; this document is transcribed
from the CYBERBIZ Notion reference and corrected against recorded live responses (Golden Files).
v1 and v2 share the host, the token and the conventions below; v2 is not a replacement for v1
(orders, products and customers CRUD remain on v1).`

// v2Operation builds one operation from a table row.
func v2Operation(ep *v2Endpoint, golden *goldenSet, loc *locale) *Operation {
	desc := loc.T(ep.Description)
	if ep.Scope != "" {
		desc += "\n\n" + fmt.Sprintf(loc.T(scopeTemplate), ep.Scope)
	}
	op := &Operation{
		Tags:        []string{ep.Tag},
		Summary:     ep.Summary,
		Description: desc,
		OperationID: operationIDFor(ep.Method, ep.Path),
		Responses:   NewOMap(),
	}
	for _, p := range ep.Params {
		cp := *p
		op.Parameters = append(op.Parameters, &cp)
	}
	if ep.Paginated {
		op.Parameters = append(op.Parameters,
			&Parameter{Name: "page", In: "query"},
			&Parameter{Name: "per_page", In: "query"})
		if strings.HasSuffix(ep.Path, "customers") || strings.HasSuffix(ep.Path, "affiliate_vendor_orders") {
			op.Parameters = append(op.Parameters, &Parameter{Name: "offset", In: "query"})
		}
	}
	if ep.Body != nil {
		op.RequestBody = &RequestBody{Required: true, Content: NewOMap().Set("application/json", &MediaType{
			Schema:  ep.Body.Schema,
			Example: newSynthesizer().Value("", decodeSample(ep.Body.Sample), nil),
		})}
	}
	resp := &Response{Description: ep.RespDesc, Content: NewOMap()}
	switch {
	case ep.Binary:
		resp.Content.Set("application/zip", &MediaType{Schema: &Schema{Type: "string", Format: "binary"}, Example: "(binary zip archive)"})
	default:
		media := &MediaType{Schema: ep.RespSchema}
		if gf := golden.success(strings.ToUpper(ep.Method), ep.Path); !usableGolden(gf) && ep.RespSample != "" {
			media.Example = newSynthesizer().Value("", decodeSample(ep.RespSample), nil)
		}
		resp.Content.Set("application/json", media)
	}
	op.Responses.Set(ep.Code, resp)
	return op
}

// usableGolden reports whether a Golden File carries a non-empty body worth
// using as the shape of a Sample.
func usableGolden(gf *goldenFile) bool {
	if gf == nil {
		return false
	}
	switch b := gf.Body.(type) {
	case nil:
		return true // a bare null is meaningful
	case []any:
		return len(b) > 0
	case *OMap:
		return b.Len() > 0
	}
	return true
}

// v2Schemas builds the v2 component schemas from the embedded samples and
// the hand-written request bodies.
func v2Schemas(loc *locale) *OMap {
	reg := NewOMap()
	d := loc.Map(v2FieldDescriptions)
	t := loc.T

	feeds := inferSchema(decodeSample(sampleProductFeeds), "", d)
	extractRef(reg, feeds, "product_feeds", "ProductFeed")
	reg.Set("ProductFeedsEnvelope", feeds)
	reg.Get2("ProductFeed").Prop("name").Enum = []any{"facebook", "google", "shopdotcom", "line"}

	aff := inferSchema(decodeSample(sampleAffiliateOrders), "", d).Items
	order := extractRef(reg, aff, "order", "AffiliateOrder")
	extractRef(reg, order, "line_items", "LineItemV2")
	extractRef(reg, order, "buyer", "OrderBuyer")
	extractRef(reg, order, "prices", "OrderPrices")
	extractRef(reg, order, "statuses", "OrderStatuses")
	extractRef(reg, order, "timings", "OrderTimings")
	reg.Set("AffiliateVendorOrder", aff)

	reg.Set("Page", inferSchema(decodeSample(samplePage), "", d))
	reg.Get2("Page").Prop("status").Enum = []any{"published", "unpublished"}
	reg.Set("PageCreate", objectSchema(d, "title", &Schema{Type: "string"}).require("title"))
	reg.Set("PageUpdate", objectSchema(d,
		"title", &Schema{Type: "string"},
		"handle", &Schema{Type: "string"},
		"status", &Schema{Type: "string", Enum: []any{"published", "unpublished"}},
		"section_id", &Schema{Type: "string"},
		"section_content", &Schema{Type: "string"}))

	cvs := inferSchema(decodeSample(sampleCvsFulfillment), "", d)
	extractRef(reg, cvs, "line_items", "LineItemV2")
	reg.Set("CvsFulfillment", cvs)
	reg.Set("CvsShippingCreate", objectSchema(d,
		"measurement", &Schema{Type: "string", Enum: []any{"S60", "S105"}},
		"size", &Schema{Type: "integer", Enum: []any{60, 90, 105}}))
	reg.Set("CvsShippingLabelsRequest", objectSchema(d,
		"shipping_type", &Schema{Type: "string", Enum: []any{"seven", "seven_c2c", "family", "family_c2c", "family_cold", "family_cold_c2c", "hilife", "hilife_cold", "hilife_c2c", "ezcat_cvs", "ezcat_cvs_cold", "ezcat_cvs_refrigerate"}},
		"fulfillment_ids", &Schema{Type: "array", Items: &Schema{Type: "integer"}}).require("shipping_type", "fulfillment_ids"))
	reg.Set("SupportShippingCreate", objectSchema(d,
		"shipping_type", &Schema{Type: "string", Enum: []any{"ezcat", "sf", "pelican", "hct"}},
		"size", &Schema{Type: "integer", Enum: []any{60, 90, 120, 150}},
		"fridge_or_frozen", &Schema{Type: "string", Enum: []any{"none", "fridge", "frozen"}},
		"is_fragile", &Schema{Type: "boolean"},
		"temperature", &Schema{Type: "string", Enum: []any{"normal", "cold"}},
		"use_transfer", &Schema{Type: "boolean"},
		"shipping_orders", &Schema{Type: "array", Items: objectSchema(d,
			"order_id", &Schema{Type: "integer"},
			"line_item_ids", &Schema{Type: "array", Items: &Schema{Type: "integer"}}).require("order_id", "line_item_ids")},
	).require("shipping_type", "shipping_orders"))
	reg.Set("SupportShippingResult", inferSchema(decodeSample(sampleSupportShippingResult), "", d))
	reg.Set("SupportShippingLabelsRequest", objectSchema(d,
		"shipping_type", &Schema{Type: "string", Enum: []any{"ezcat", "sf", "pelican", "hct"}},
		"print_type", &Schema{Type: "string", Enum: []any{"normal", "thermal"}},
		"order_ids", &Schema{Type: "array", Items: &Schema{Type: "integer"}}).require("shipping_type", "order_ids"))

	reg.Set("MenuSummary", inferSchema(decodeSample(sampleMenus), "", d).Items)
	menu := inferSchema(decodeSample(sampleMenu), "", d)
	item := extractRef(reg, menu, "items", "MenuItem")
	item.SetProp("items", &Schema{Type: "array", Description: d["items"], Items: ref("MenuItem")})
	item.Prop("item_type").Description = t(v2ItemTypeNote)
	reg.Set("Menu", menu)

	reg.Set("Category", inferSchema(decodeSample(sampleCategories), "", d).Items)
	reg.Set("CustomCollectionV2", inferSchema(decodeSample(sampleCustomCollectionsV2), "", d).Items)
	smart := inferSchema(decodeSample(sampleSmartCollectionsV2), "", d).Items
	extractRef(reg, smart, "rules", "SmartCollectionRuleV2")
	extractRef(reg, smart, "products", "CollectionProductV2")
	reg.Set("SmartCollectionV2", smart)

	reg.Set("CustomerOAuthRequest", objectSchema(d, "uid", &Schema{Type: "string"}, "provider", &Schema{Type: "string"}))

	cust := inferSchema(decodeSample(sampleCustomerV2Included), "", d)
	compact := decodeSample(sampleCustomersV2).([]any)[0].(*OMap)
	for _, k := range cust.Properties.Keys() {
		if !compact.Has(k) {
			p := cust.Prop(k)
			p.Description = strings.TrimSpace(p.Description + t(includeNote))
		}
	}
	extractRef(reg, cust, "address", "CustomerAddressV2")
	extractRef(reg, cust, "uid_providers", "UidProvider")
	extractRef(reg, cust, "vip_info", "VipInfo")
	reg.Set("CustomerV2", cust)

	reg.Set("ShopEmails", inferSchema(decodeSample(sampleShopEmails), "", d))

	reg.Set("ProductSearchRequest", objectSchema(d,
		"q", &Schema{Type: "string"},
		"limit", &Schema{Type: "integer"},
		"offset", &Schema{Type: "integer"},
		"filter_published", &Schema{Type: "boolean", Default: true},
		"vendors", &Schema{Type: "array", Items: &Schema{Type: "string"}},
		"product_types", &Schema{Type: "array", Items: &Schema{Type: "string"}},
		"tags", &Schema{Type: "array", Items: &Schema{Type: "string"}},
		"store_types", &Schema{Type: "array", Items: &Schema{Type: "integer", Enum: []any{1, 2, 4}}},
		"filter_branch_store", &Schema{Type: "boolean"},
		"filter_on_sell", &Schema{Type: "boolean"},
		"ids", &Schema{Type: "array", Items: &Schema{Type: "integer"}},
		"exclude_tags", &Schema{Type: "array", Items: &Schema{Type: "string"}},
		"skus", &Schema{Type: "array", Items: &Schema{Type: "string"}},
		"branch_stores", &Schema{Type: "array", Items: &Schema{Type: "integer"}},
		"store_nos", &Schema{Type: "array", Items: &Schema{Type: "string"}}))
	search := inferSchema(decodeSample(sampleProductSearchResult), "", d)
	prod := extractRef(reg, search, "products", "ProductV2")
	extractRef(reg, prod, "product_variants", "ProductVariantV2")
	extractRef(reg, prod, "product_options", "ProductOptionV2")
	prod.SetProp("custom_collections", &Schema{Type: "array", Description: d["custom_collections"], Items: ref("CustomCollectionV2")})
	reg.Set("ProductSearchResult", search)

	reg.Set("InventoryBatchUpdate", objectSchema(d,
		"store_number", &Schema{Type: "string"},
		"items", &Schema{Type: "array", Items: objectSchema(d,
			"sku", &Schema{Type: "string"},
			"inventory_quantity", &Schema{Type: "integer"}).require("sku", "inventory_quantity")}).require("store_number", "items"))
	reg.Set("InventoryJobQueued", inferSchema(decodeSample(sampleInventoryQueued), "", d))
	status := inferSchema(decodeSample(sampleInventoryStatus), "", d)
	status.Prop("status").Enum = []any{"QUEUE", "QUEUED", "PROCESSING", "SUCCESS", "FAILED"}
	status.SetProp("created_at", markNullable(&Schema{Type: "string", Description: t(v2JobCreatedNote)}))
	status.SetProp("completed_at", markNullable(&Schema{Type: "string", Description: t(v2JobCompletedNote)}))
	status.SetProp("result", markNullable(status.Prop("result")))
	status.SetProp("error", markNullable(&Schema{Type: "string", Description: t(v2JobErrorNote)}))
	reg.Set("InventoryJobStatus", status)

	reg.Set("WalletBalance", inferSchema(decodeSample(sampleWalletBalance), "", d))
	tx := inferSchema(decodeSample(sampleWalletTransactions), "", d).Items
	// The first sample entry is a refund; merge the top-up-only fields in.
	topup := decodeSample(sampleWalletTransactions).([]any)[1].(*OMap)
	for _, k := range topup.Keys() {
		if tx.Prop(k) == nil {
			v, _ := topup.Get(k)
			tx.SetProp(k, inferSchema(v, k, d))
		}
	}
	extractRef(reg, tx, "invalid_einvoices", "WalletEinvoice")
	tx.SetProp("allowance_einvoices", &Schema{Type: "array", Description: d["allowance_einvoices"], Items: ref("WalletEinvoice")})
	tx.SetProp("einvoice", &Schema{Ref: schemaRef("WalletEinvoice"), Description: d["einvoice"]})
	tx.Prop("type").Enum = []any{"topup", "consumption", "refund", "cancel_order", "bonus"}
	tx.SetProp("financial_status", &Schema{Type: "string", Enum: []any{"paid", "failed", "timeout", "pending"}, Description: t(v2TopupStatusNote)})
	tx.SetProp("payment_name", markNullable(&Schema{Type: "string", Description: d["payment_name"]}))
	tx.SetProp("paper_invoice_no", markNullable(&Schema{Type: "string", Description: d["paper_invoice_no"]}))
	reg.Set("WalletTransaction", tx)

	reg.Set("ShopInfoEnvelope", inferSchema(decodeSample(sampleShop), "", d))
	settings := inferSchema(decodeSample(sampleSettings), "", d)
	addOn := settings.Prop("shop_add_on")
	addOn.SetProp("settings", &Schema{Type: "object", AdditionalProperties: true, Description: d["settings"]})
	addOn.SetProp("start_at", markNullable(ref("Timestamp")))
	addOn.SetProp("end_at", markNullable(ref("Timestamp")))
	reg.Set("ShopAddOnEnvelope", settings)
	reg.Set("ShopAddOnSettingsUpdate", objectSchema(d,
		"settings", &Schema{Type: "array", Items: objectSchema(d,
			"field", &Schema{Type: "string"},
			"data", &Schema{Type: "string"}).require("field", "data")}).require("settings"))
	updated := inferSchema(decodeSample(sampleSettingsUpdated), "", d)
	updated.Prop("shop_add_on").SetProp("settings", &Schema{Type: "object", AdditionalProperties: true, Description: d["settings"]})
	reg.Set("ShopAddOnUpdatedEnvelope", updated)

	reg.SortKeys()
	return reg
}

// Get2 returns a registered schema by name (panics on a missing name; the
// registry is built by code, not data).
func (m *OMap) Get2(name string) *Schema {
	v, ok := m.Get(name)
	if !ok {
		panic("gendocs: unknown v2 schema " + name)
	}
	return v.(*Schema)
}

// extractRef moves the schema of prop (or its array items) into the registry
// under name and replaces it with a $ref. Returns the extracted schema.
func extractRef(reg *OMap, parent *Schema, prop, name string) *Schema {
	child := parent.Prop(prop)
	if child == nil {
		panic("gendocs: extractRef: no property " + prop)
	}
	desc := child.Description
	target := child
	isArray := child.BaseType() == "array"
	if isArray {
		target = child.Items
	}
	if existing, ok := reg.Get(name); ok {
		target = existing.(*Schema)
	} else {
		target.Description = ""
		reg.Set(name, target)
	}
	var replacement *Schema
	if isArray {
		replacement = &Schema{Type: "array", Description: desc, Items: ref(name)}
	} else {
		replacement = &Schema{Ref: schemaRef(name), Description: desc}
		if child.IsNullable() {
			replacement = markNullable(replacement)
		}
	}
	parent.SetProp(prop, replacement)
	return target
}

// objectSchema builds an object schema from name/schema pairs, filling
// descriptions from descs.
func objectSchema(descs map[string]string, pairs ...any) *Schema {
	s := &Schema{Type: "object", Properties: NewOMap()}
	for i := 0; i+1 < len(pairs); i += 2 {
		name := pairs[i].(string)
		p := pairs[i+1].(*Schema)
		if p.Description == "" {
			p.Description = descs[name]
		}
		s.SetProp(name, p)
	}
	return s
}

func (s *Schema) require(names ...string) *Schema {
	s.Required = append(s.Required, names...)
	return s
}

// Strings authored inline by v2Schemas; listed so the locale tables cover them.
const (
	v2ItemTypeNote     = "Item type; values seen: `frontpage`, `collections_all`, `http`, `product`, `contact`, `page`."
	v2JobCreatedNote   = "Created at. Format `YYYY-MM-DD HH:MM:SS +0800`."
	v2JobCompletedNote = "Completion time, `YYYY-MM-DD HH:MM:SS +0800`."
	v2JobErrorNote     = "Failure reason when `status` is `FAILED`."
	v2TopupStatusNote  = "Top-up payment status: `paid`, `failed`, `timeout` (processing), `pending`."
)

var v2InlineStrings = []string{v2ItemTypeNote, v2JobCreatedNote, v2JobCompletedNote, v2JobErrorNote, v2TopupStatusNote}
