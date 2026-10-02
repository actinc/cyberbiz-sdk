package gendocs

import (
	"regexp"
)

// correctionRule is one entry of the corrections table: a fix the generator
// applies to the CYBERBIZ source material because live traffic (Golden Files
// and schema_observed.json) or the CYBERBIZ Notion pages contradict it.
// The ID is what the run report counts.
type correctionRule struct {
	ID          string
	Description string
}

// correctionsTable documents every programmatic fix. Keep it in sync with the
// apply* functions below; the tests assert each ID is exercised.
var correctionsTable = []correctionRule{
	{"legacy-auth", "The swagger's HMAC / api.cyberbiz.co authentication section is legacy; the document describes Bearer auth against app-store-api.cyberbiz.io instead."},
	{"strip-html", "HTML in descriptions is converted to plain text (<br> -> line break, <li> -> bullet)."},
	{"tag-descriptions", "Tag descriptions are rewritten in English (the swagger has a 'Translation missing' tag)."},
	{"typo-interger", "The swagger type 'Interger' (third_party_discount) is read as integer."},
	{"observed-type", "A field whose observed JSON type differs from the swagger takes the observed type (order_number integer, weight number, shipping_vendor.name string, concurrently_apply boolean, ...)."},
	{"observed-array", "A field documented as an object but observed as an array becomes an array of that object (exchange_histories, stock items, check_logs, order returns)."},
	{"observed-object", "A response documented as an array but observed as an object becomes that object (vip_groups, customer_groups check_status)."},
	{"observed-null", "A field observed as null in any Golden File is marked nullable (type: [T, 'null'] or anyOf with null)."},
	{"observed-field", "A field present in Golden Files but missing from the swagger is added with the observed type and x-cyberbiz-observed: true."},
	{"observed-format", "Strings observed as 'YYYY-MM-DD HH:MM:SS' reference the shared Timestamp schema; 'YYYY-MM-DD' references Date."},
	{"money", "Every price, amount, discount, cost, balance, threshold, bonus and points field references the shared Money schema (a float) even where the swagger says integer."},
	{"order-number", "order_number is an integer, not a string."},
	{"timestamp-suffix", "String fields ending in _at reference the shared Timestamp schema (Asia/Taipei, no zone)."},
	{"pagination-params", "page / per_page / offset query parameters reference the shared pagination parameter components; list responses declare the X-Page ... X-Prev-Page headers."},
	{"form-to-json", "Swagger formData parameters become a JSON request body (the platform accepts JSON); bracketed names are nested; file uploads stay multipart."},
	{"error-responses", "Every operation declares the shared 401 / 403 / 404 / 422 / 429 responses; 401 also means the feature is not licensed for the shop."},
	{"not-found-null", "Lookups whose Golden File is a bare JSON null document that the platform answers 200 null when the resource does not exist."},
	{"enum-docs", "Status-like fields (order/financial/fulfillment/return status, coupon_status and gift_order_status with their combination rule, invoice status/type, tax type) get the documented enum values and an English description."},
	{"translate", "Descriptions are translated from zh-TW to English through the embedded glossary; untranslated strings pass through unchanged."},
}

var (
	reMoneyKey   = regexp.MustCompile(`(^|_)(price|prices|amount|discount|discounts|cost|balance|balance_after|threshold|subtotal|total|fee|bonus|points|consumption|credit|value)($|_)`)
	reNotMoney   = regexp.MustCompile(`(_id|_ids|_enabled|_name|_type|_code|_status|_days|_count|_percentage|_rate|_quantity|_limit|_position|_url|_management|_policy|_at)$|^(sell_weight|weight|meas|quantity|position|percentage|order_discount_value|usage_limit_value|coupon_value|value)$`)
	reTimeSuffix = regexp.MustCompile(`_at$`)
)

// isMoneyKey reports whether a numeric field named key holds a currency
// amount. usage_limit_value / coupon_value / value are excluded because they
// are counts or percentages depending on context.
func isMoneyKey(key string) bool {
	return reMoneyKey.MatchString(key) && !reNotMoney.MatchString(key)
}

// applyCorrections runs every content fix on a converted document.
func applyCorrections(doc *Document, obs *observedDoc, log *report) {
	fixTagDescriptions(doc, log)
	applyObserved(doc, obs, log)
	applyNameRules(doc, log)
	applyEnumDocs(doc, log)
}

// fixTagDescriptions replaces the swagger tag descriptions with English ones.
func fixTagDescriptions(doc *Document, log *report) {
	for _, t := range doc.Tags {
		if d, ok := tagDescriptions[t.Name]; ok {
			t.Description = d
			log.count("tag-descriptions")
		}
	}
}

// applyNameRules applies the field-name based rules (money, order_number,
// _at timestamps) to every schema in the document.
func applyNameRules(doc *Document, log *report) {
	walkDocumentSchemas(doc, func(key string, s *Schema) *Schema {
		if key == "" {
			return s
		}
		base := s.BaseType()
		switch {
		case key == "order_number" && base == "string":
			s.Type = replaceBaseType(s.Type, "integer")
			log.count("order-number")
		case isMoneyKey(key) && (base == "integer" || base == "number"):
			log.count("money")
			return refSchema("Money", s)
		case reTimeSuffix.MatchString(key) && base == "string" && s.Format == "" && s.RefName() == "":
			log.count("timestamp-suffix")
			return refSchema("Timestamp", s)
		}
		return s
	})
}

// refSchema turns s into a reference to a shared schema, keeping the
// description and nullability of s.
func refSchema(name string, s *Schema) *Schema {
	out := &Schema{Ref: schemaRef(name), Description: s.Description}
	if s.IsNullable() {
		return markNullable(out)
	}
	return out
}

// replaceBaseType swaps the non-null type name, preserving a null entry.
func replaceBaseType(t any, base string) any {
	if ts, ok := t.([]string); ok {
		out := make([]string, len(ts))
		for i, e := range ts {
			if e == "null" {
				out[i] = e
			} else {
				out[i] = base
			}
		}
		return out
	}
	return base
}

// markNullable allows null for s (OpenAPI 3.1 style). A $ref becomes an
// anyOf of the ref and null.
func markNullable(s *Schema) *Schema {
	if s.IsNullable() {
		return s
	}
	switch {
	case s.Ref != "":
		return &Schema{
			Description: s.Description,
			AnyOf:       []*Schema{{Ref: s.Ref}, {Type: "null"}},
			XObserved:   s.XObserved,
		}
	case len(s.AnyOf) > 0:
		s.AnyOf = append(s.AnyOf, &Schema{Type: "null"})
	case s.BaseType() != "":
		s.Type = []string{s.BaseType(), "null"}
	}
	return s
}

// walkDocumentSchemas visits every schema (components, parameters, request
// bodies, responses) depth-first. fn receives the property name the schema
// sits under ("" for roots and array items) and may return a replacement.
func walkDocumentSchemas(doc *Document, fn func(key string, s *Schema) *Schema) {
	if doc.Components != nil && doc.Components.Schemas != nil {
		for _, name := range doc.Components.Schemas.Keys() {
			v, _ := doc.Components.Schemas.Get(name)
			doc.Components.Schemas.Set(name, walkSchema("", v.(*Schema), fn))
		}
	}
	for _, path := range doc.Paths.Keys() {
		item, _ := doc.Paths.Get(path)
		for _, m := range httpMethods {
			op := item.(*PathItem).Operation(m)
			if op == nil {
				continue
			}
			walkOperationSchemas(op, fn)
		}
	}
}

func walkOperationSchemas(op *Operation, fn func(key string, s *Schema) *Schema) {
	for _, p := range op.Parameters {
		if p.Schema != nil {
			p.Schema = walkSchema("", p.Schema, fn)
		}
	}
	if op.RequestBody != nil {
		walkContent(op.RequestBody.Content, fn)
	}
	for _, code := range op.Responses.Keys() {
		r, _ := op.Responses.Get(code)
		walkContent(r.(*Response).Content, fn)
	}
}

func walkContent(content *OMap, fn func(key string, s *Schema) *Schema) {
	if content == nil {
		return
	}
	for _, mt := range content.Keys() {
		v, _ := content.Get(mt)
		media := v.(*MediaType)
		if media.Schema != nil {
			media.Schema = walkSchema("", media.Schema, fn)
		}
	}
}

func walkSchema(key string, s *Schema, fn func(key string, s *Schema) *Schema) *Schema {
	if s == nil {
		return nil
	}
	if s.Properties != nil {
		for _, k := range s.Properties.Keys() {
			v, _ := s.Properties.Get(k)
			s.Properties.Set(k, walkSchema(k, v.(*Schema), fn))
		}
	}
	if s.Items != nil {
		s.Items = walkSchema("", s.Items, fn)
	}
	for i, a := range s.AnyOf {
		s.AnyOf[i] = walkSchema("", a, fn)
	}
	if ap, ok := s.AdditionalProperties.(*Schema); ok {
		s.AdditionalProperties = walkSchema("", ap, fn)
	}
	return fn(key, s)
}

// tagDescriptions is the English text for every v1 tag.
var tagDescriptions = map[string]string{
	"assets":                       "CKEditor image library (pictures used in rich-text content).",
	"customers":                    "Customers: profile, tags, custom fields, bonus points, coupons, orders, VIP information.",
	"orders":                       "Orders: listing, detail, status transitions, fulfillments, transactions, tags and notes.",
	"order_returns":                "Order returns.",
	"products":                     "Products, variants, options, photos, descriptions, tags and bind-shipping settings.",
	"smart_collections":            "Smart (rule-based) collections and their rules.",
	"custom_collections":           "Custom (manually curated) collections and their products.",
	"special_collections":          "Special collections (mix-and-match discount campaigns) and their rules.",
	"add_buy_collections":          "Add-on purchase (add-buy) collections.",
	"vip_collections":              "VIP settings (legacy VIP collections).",
	"vip_groups":                   "VIP groups and levels (the newer VIP model).",
	"shop_coupons":                 "Shop-wide coupons.",
	"pos_shop_coupons":             "POS shop coupons (requires the pos_shop_coupon plugin).",
	"discounts":                    "Shop-wide discount campaigns.",
	"einvoices":                    "Electronic invoices (e-invoices) of orders.",
	"offline_einvoices":            "Offline (manually issued) e-invoices.",
	"custom_field_types":           "Customer custom-field type definitions.",
	"pos_shops":                    "POS shops, POS terminals and their products (POS only).",
	"stock_invoices":               "Stock invoices / outbound stock documents (POS only).",
	"stock_receipts":               "Stock receipts / inbound stock documents (POS only).",
	"stock_requisitions":           "Stock requisitions / transfers between warehouses (POS only).",
	"stock_adjustments":            "Stock adjustments (POS only).",
	"branch_stores":                "Branch stores: pickup stores, shipping rates, staff users, express-delivery settings.",
	"inventory_sync_groups":        "Inventory sync groups (paid add-on; contact your CYBERBIZ consultant to enable).",
	"order_etickets":               "Electronic tickets (e-tickets) attached to orders.",
	"periodic_orders":              "Periodic (subscription) orders.",
	"bonus_rule":                   "Bonus point earning rules.",
	"limit_collections":            "Purchase-limit collections.",
	"variant_discount_collections": "Per-variant discount collections.",
	"custom_fields":                "Custom fields (newer model).",
	"blogs":                        "Blogs, articles, article tags and SEO meta tags.",
	"app":                          "App-level endpoints without a version prefix: the shop and the app's settings on that shop.",
	"product_feeds":                "Product feed URLs for advertising platforms.",
	"affiliates":                   "Affiliate (revenue-share) vendor orders.",
	"pages":                        "Custom storefront pages with an HTML section.",
	"shipping":                     "Convenience-store and home-delivery shipping labels.",
	"menus":                        "Storefront menus (link lists).",
	"categories":                   "Multi-level product categories.",
	"collections":                  "Custom and smart collections (v2 search endpoints).",
	"customer_oauth":               "Customer OAuth.",
	"shop_emails":                  "Shop e-mail address and subscribers.",
	"pos_wallets":                  "POS stored-value wallets (requires the POS stored-value feature).",
}
