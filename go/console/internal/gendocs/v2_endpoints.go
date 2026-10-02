package gendocs

// v2Endpoint is one row of the v2 / app endpoint table, transcribed from the
// CYBERBIZ Notion reference (docs/references/notes/notion-v2-api.md), which
// has no machine-readable form.
type v2Endpoint struct {
	Method      string
	Path        string
	Tag         string
	Summary     string
	Description string
	Scope       string
	Params      []*Parameter
	Body        *v2Body
	Code        string // success status code
	RespDesc    string
	RespSchema  *Schema
	RespSample  string // embedded JSON, used when no Golden File exists
	Binary      bool   // zip download
	Paginated   bool
}

type v2Body struct {
	Schema *Schema
	Sample string
}

func ref(name string) *Schema { return &Schema{Ref: schemaRef(name)} }

func arrayOf(name string) *Schema {
	return &Schema{Type: "array", Items: ref(name)}
}

func qp(name, typ, desc string, example any) *Parameter {
	return &Parameter{Name: name, In: "query", Description: desc, Schema: &Schema{Type: typ}, Example: example}
}

func pp(name, desc string) *Parameter {
	return &Parameter{Name: name, In: "path", Required: true, Description: desc, Schema: &Schema{Type: "integer"}, Example: 1}
}

const timeParamNote = " (`YYYY-MM-DD hh:mm:ss`, Asia/Taipei)"

// v2Endpoints lists every v2 and app endpoint in Notion page order.
var v2Endpoints = []*v2Endpoint{
	{
		Method: "get", Path: "/shop", Tag: "app", Summary: "Get the shop",
		Description: "The shop the token belongs to: primary domain, name, id, contact e-mail, Open Graph image and LINE settings.",
		Code:        "200", RespDesc: "The shop.", RespSchema: ref("ShopInfoEnvelope"), RespSample: sampleShop,
	},
	{
		Method: "get", Path: "/settings", Tag: "app", Summary: "Get the app's settings on this shop",
		Description: "The app installation (`shop_add_on`) on this shop: its setting values, vendor type, token, validity window and the installed app version with its manifest and granted scopes.",
		Code:        "200", RespDesc: "The app installation.", RespSchema: ref("ShopAddOnEnvelope"), RespSample: sampleSettings,
	},
	{
		Method: "put", Path: "/settings", Tag: "app", Summary: "Update the app's settings on this shop",
		Description: "Updates the values of the `setting_fields` declared in the app manifest. Each `field` must equal a field name declared at application time.",
		Body:        &v2Body{Schema: ref("ShopAddOnSettingsUpdate"), Sample: sampleSettingsUpdate},
		Code:        "200", RespDesc: "The updated app installation.", RespSchema: ref("ShopAddOnUpdatedEnvelope"), RespSample: sampleSettingsUpdated,
	},
	{
		Method: "get", Path: "/v2/product_feeds", Tag: "product_feeds", Summary: "List product feed URLs", Scope: "public",
		Description: "The product feed URLs of the shop for facebook, google, shopdotcom and line. No parameters.",
		Code:        "200", RespDesc: "The feeds.", RespSchema: ref("ProductFeedsEnvelope"), RespSample: sampleProductFeeds,
	},
	{
		Method: "get", Path: "/v2/affiliate_vendor_orders", Tag: "affiliates", Summary: "List affiliate vendor orders", Scope: "read_affiliates",
		Description: "Orders attributed to the affiliate (revenue-share) vendor. Child orders of a periodic order created with the affiliate query parameters (`affiliate`, `cid`, `uid`) are included. The Notion reference shows an `affiliate_vendor_orders` envelope; the live API answers a bare array (Golden File), which is what this document describes.",
		Params: []*Parameter{
			qp("start_time", "string", "Order creation start"+timeParamNote, "2026-09-01 00:00:00"),
			qp("end_time", "string", "Order creation end"+timeParamNote, "2026-09-30 23:59:59"),
			qp("closed_at_start_time", "string", "Order `closed_at` start"+timeParamNote, "2026-09-01 00:00:00"),
			qp("closed_at_end_time", "string", "Order `closed_at` end"+timeParamNote, "2026-09-30 23:59:59"),
			qp("statuses", "string", "Order statuses, comma-separated: `open`, `closed`, `cancelled`.", "open,closed"),
		},
		Paginated: true,
		Code:      "200", RespDesc: "The affiliate orders.", RespSchema: arrayOf("AffiliateVendorOrder"), RespSample: sampleAffiliateOrders,
	},
	{
		Method: "post", Path: "/v2/pages", Tag: "pages", Summary: "Create a custom page", Scope: "write_content",
		Description: "Creates a custom page containing one custom HTML section. Pages created here cannot be edited in the admin backend (preview only); change them through `PUT /v2/pages/{page_id}`.",
		Body:        &v2Body{Schema: ref("PageCreate"), Sample: samplePageCreate},
		Code:        "201", RespDesc: "The created page.", RespSchema: ref("Page"), RespSample: samplePage,
	},
	{
		Method: "put", Path: "/v2/pages/{page_id}", Tag: "pages", Summary: "Update a custom page", Scope: "write_content",
		Description: "Updates a page created through `POST /v2/pages`. To change the section content send both `section_id` and `section_content`; `section_content` is a JSON string (unicode-escape the HTML, then JSON-encode). Only the custom HTML section created by `POST /v2/pages` can be modified. The reference says `page_id` is passed in the query string; the Postman collection sends it in both the path and the query.",
		Params: []*Parameter{
			pp("page_id", "Custom page id."),
			qp("page_id", "integer", "Custom page id (the reference documents it as a query parameter).", 1),
		},
		Body: &v2Body{Schema: ref("PageUpdate"), Sample: samplePageUpdate},
		Code: "200", RespDesc: "The updated page.", RespSchema: ref("Page"), RespSample: samplePage,
	},
	{
		Method: "post", Path: "/v2/orders/{order_id}/cvs_shipping", Tag: "shipping", Summary: "Create a convenience-store shipping label", Scope: "write_orders",
		Description: "Creates a CVS shipping label for one order. Supported carriers: FamilyMart (incl. store-to-store and cold chain), Hi-Life (incl. store-to-store and cold chain), 7-ELEVEN (incl. store-to-store) and T-cat quick store delivery (ambient, frozen, refrigerated). Only orders whose reply has a non-null `cvs_shipping_type` can then print a label.",
		Params:      []*Parameter{pp("order_id", "Order id.")},
		Body:        &v2Body{Schema: ref("CvsShippingCreate"), Sample: sampleCvsShippingRequest},
		Code:        "201", RespDesc: "The created fulfillment.", RespSchema: ref("CvsFulfillment"), RespSample: sampleCvsFulfillment,
	},
	{
		Method: "post", Path: "/v2/orders/cvs_shipping_labels", Tag: "shipping", Summary: "Print convenience-store shipping labels", Scope: "write_orders",
		Description: "Returns a zip file containing one PDF with all requested labels (`seven_c2c` labels are delivered as HTML). `Content-Disposition: attachment; filename=cvs_<shipping_type>_shipping_labels_<timestamp>.zip`.",
		Body:        &v2Body{Schema: ref("CvsShippingLabelsRequest"), Sample: sampleCvsLabelsRequest},
		Code:        "200", RespDesc: "A zip archive with the labels.", Binary: true,
	},
	{
		Method: "post", Path: "/v2/orders/fulfillments/support_shippings", Tag: "shipping", Summary: "Create home-delivery shipping labels", Scope: "write_orders",
		Description: "Creates shipping labels for T-cat (`ezcat`), SF Express (`sf`), Pelican (`pelican`) or HCT (`hct`) for several orders at once. `fridge_or_frozen`, `is_fragile` and `use_transfer` are not supported by `sf` and `hct`. Orders that fail are listed in `failed_orders`; the rest get a fulfillment.",
		Body:        &v2Body{Schema: ref("SupportShippingCreate"), Sample: sampleSupportShippingRequest},
		Code:        "200", RespDesc: "Created fulfillments and failed orders.", RespSchema: ref("SupportShippingResult"), RespSample: sampleSupportShippingResult,
	},
	{
		Method: "post", Path: "/v2/orders/fulfillments/support_shipping_labels", Tag: "shipping", Summary: "Print home-delivery shipping labels", Scope: "write_orders",
		Description: "Returns a zip file containing one PDF with all labels for T-cat, SF Express, Pelican or HCT. `print_type` applies to HCT only. `Content-Disposition: attachment; filename=support_<shipping_type>_shipping_labels_<timestamp>.zip`.",
		Body:        &v2Body{Schema: ref("SupportShippingLabelsRequest"), Sample: sampleSupportLabelsRequest},
		Code:        "200", RespDesc: "A zip archive with the labels.", Binary: true,
	},
	{
		Method: "get", Path: "/v2/menus", Tag: "menus", Summary: "List menus", Scope: "read_content",
		Description: "All menus (link lists) of the storefront. The Notion sample is not valid JSON; the live API answers a bare array (Golden File).",
		Paginated:   true,
		Code:        "200", RespDesc: "The menus.", RespSchema: arrayOf("MenuSummary"), RespSample: sampleMenus,
	},
	{
		Method: "get", Path: "/v2/menus/{menu_id}", Tag: "menus", Summary: "Get a menu with its items", Scope: "read_content",
		Description: "One menu with its nested items; nesting is recursive through `items`.",
		Params:      []*Parameter{pp("menu_id", "Menu id.")},
		Code:        "200", RespDesc: "The menu.", RespSchema: ref("Menu"), RespSample: sampleMenu,
	},
	{
		Method: "get", Path: "/v2/categories", Tag: "categories", Summary: "List categories", Scope: "read_products",
		Description: "Multi-level product categories. `full_handle` encodes the hierarchy (`clothing/mens/shirts`) and builds the storefront URL `https://<domain>/categories/<full_handle>`; `handle` is the category's own segment.",
		Params:      []*Parameter{qp("q", "string", "Category title to search (string match).", "服飾")},
		Paginated:   true,
		Code:        "200", RespDesc: "The categories.", RespSchema: arrayOf("Category"), RespSample: sampleCategories,
	},
	{
		Method: "get", Path: "/v2/custom_collections", Tag: "collections", Summary: "Search custom collections", Scope: "read_products",
		Description: "Custom collections, searchable by title or handle. `products_order_name` carries the display label of the ordering (e.g. `按標題拼音升序: A-Z`), not the code; codes are `title.asc`, `title.desc`, `created_at.desc`, `created_at.asc`, `price.desc`, `price.asc`, `manual`, `sell_weight.desc`.",
		Params:      []*Parameter{qp("q", "string", "Title or handle to search (string match).", "首頁")},
		Paginated:   true,
		Code:        "200", RespDesc: "The collections.", RespSchema: arrayOf("CustomCollectionV2"), RespSample: sampleCustomCollectionsV2,
	},
	{
		Method: "get", Path: "/v2/smart_collections", Tag: "collections", Summary: "Search smart collections", Scope: "read_products",
		Description: "Smart (rule-based) collections with their rules and matching products, searchable by title or handle.",
		Params:      []*Parameter{qp("q", "string", "Title or handle to search.", "千元")},
		Paginated:   true,
		Code:        "200", RespDesc: "The collections.", RespSchema: arrayOf("SmartCollectionV2"), RespSample: sampleSmartCollectionsV2,
	},
	{
		Method: "post", Path: "/v2/customer_oauth", Tag: "customer_oauth", Summary: "Customer OAuth", Scope: "customer_oauth",
		Description: "Documentation defect: the Notion section for this endpoint is a verbatim copy of `GET /v2/categories`, so no request/response contract is published. The request body below comes from the CYBERBIZ Postman collection; the response is undocumented. Consult CYBERBIZ support before relying on it.",
		Body:        &v2Body{Schema: ref("CustomerOAuthRequest"), Sample: sampleCustomerOAuthRequest},
		Code:        "200", RespDesc: "Undocumented.", RespSchema: &Schema{Description: "Undocumented response body."},
	},
	{
		Method: "get", Path: "/v2/customers", Tag: "customers", Summary: "List customers", Scope: "read_customers",
		Description: "Customers as a bare array. The live API returns the compact shape shown in the sample; `include_params` adds optional sections (`vip_info` is documented; the Postman collection also sends `uid_providers,tags`). When `ids` is given pagination is disabled.",
		Params: []*Parameter{
			qp("ids", "string", "Comma-separated customer ids (at most 100). Disables pagination.", "1,2,3"),
			qp("include_params", "string", "Extra sections to include, comma-separated. Documented: `vip_info`.", "vip_info"),
		},
		Paginated: true,
		Code:      "200", RespDesc: "The customers.", RespSchema: arrayOf("CustomerV2"), RespSample: sampleCustomersV2,
	},
	{
		Method: "get", Path: "/v2/customers/by_uid_provider", Tag: "customers", Summary: "Get a customer by social-login uid", Scope: "read_customers",
		Description: "Looks up one customer by the uid of a social-login provider. Answers `null` when no customer matches (Golden File). The parameters are `uid` and `provider` (verified against the live API); the Notion reference's `provider_type` is a documentation error.",
		Params: []*Parameter{
			qp("uid", "string", "The provider's user id.", "U0000000000000000000000000000001"),
			qp("provider", "string", "Provider: `line`, `line_at`, `facebook`, ...", "line"),
		},
		Code: "200", RespDesc: "The customer, or `null` when not found.", RespSchema: markNullable(ref("CustomerV2")), RespSample: sampleCustomerV2Included,
	},
	{
		Method: "get", Path: "/v2/shop_emails/shop_emails", Tag: "shop_emails", Summary: "Get the shop e-mail and subscribers", Scope: "read_shop_emails",
		Description: "The shop's e-mail address and the addresses subscribed to order notifications. No parameters.",
		Code:        "200", RespDesc: "The addresses.", RespSchema: ref("ShopEmails"), RespSample: sampleShopEmails,
	},
	{
		Method: "post", Path: "/v2/products/search", Tag: "products", Summary: "Search products", Scope: "read_products",
		Description: "Product search with a JSON body. Unlike other list endpoints it pages with `limit` / `offset` in the body and returns `total_count` / `total_pages` in the response instead of `X-*` headers. `skus`, `branch_stores` and `store_nos` are custom features; contact your consultant.",
		Body:        &v2Body{Schema: ref("ProductSearchRequest"), Sample: sampleProductSearchRequest},
		Code:        "200", RespDesc: "Matching products.", RespSchema: ref("ProductSearchResult"), RespSample: sampleProductSearchResult,
	},
	{
		Method: "put", Path: "/v2/products/batch_update_branch_store_inventory", Tag: "products", Summary: "Batch-update branch-store inventory", Scope: "write_products",
		Description: "Queues an asynchronous job that sets branch-store inventory by store number and SKU; at most 1000 items per call. Requires the 'batch update branch store inventory by store number and sku' plugin (otherwise `422`). Poll the job with `GET /v2/products/branch_store_inventory_update_status`.",
		Body:        &v2Body{Schema: ref("InventoryBatchUpdate"), Sample: sampleInventoryRequest},
		Code:        "202", RespDesc: "The queued job.", RespSchema: ref("InventoryJobQueued"), RespSample: sampleInventoryQueued,
	},
	{
		Method: "get", Path: "/v2/products/branch_store_inventory_update_status", Tag: "products", Summary: "Poll a batch inventory job", Scope: "read_products",
		Description: "Status of a job created by `PUT /v2/products/batch_update_branch_store_inventory`. `status` is `QUEUE` (poll) / `QUEUED` (create), `PROCESSING`, `SUCCESS` or `FAILED`; an unknown job answers `{\"error\": [\"job_id '...' 不存在\"]}`. Timestamps here carry a `+0800` suffix.",
		Params:      []*Parameter{{Name: "job_id", In: "query", Required: true, Description: "Job id from the batch update response.", Schema: &Schema{Type: "string"}, Example: "5f1a2b3c4d5e6f7a8b9c0d1e"}},
		Code:        "200", RespDesc: "The job status.", RespSchema: ref("InventoryJobStatus"), RespSample: sampleInventoryStatus,
	},
	{
		Method: "get", Path: "/v2/pos_wallets/{customer_id}/balance", Tag: "pos_wallets", Summary: "Get a customer's POS wallet balance", Scope: "read_pos",
		Description: "Stored-value balance of a customer. Only for shops with the POS stored-value feature enabled.",
		Params:      []*Parameter{pp("customer_id", "Customer id.")},
		Code:        "200", RespDesc: "The balance.", RespSchema: ref("WalletBalance"), RespSample: sampleWalletBalance,
	},
	{
		Method: "get", Path: "/v2/pos_wallets/{customer_id}/transactions", Tag: "pos_wallets", Summary: "List a customer's POS wallet transactions", Scope: "read_pos",
		Description: "Wallet transaction history, newest first, capped at 1000 records (no pagination). `topup` entries carry `einvoice`, `paper_invoice_no` and `financial_status`; `refund` entries carry `invalid_einvoices` and `allowance_einvoices`; `consumption`, `cancel_order` and `bonus` carry only the common fields. Timestamps are `YYYY-MM-DD HH:MM+0800`.",
		Params:      []*Parameter{pp("customer_id", "Customer id.")},
		Code:        "200", RespDesc: "The transactions.", RespSchema: arrayOf("WalletTransaction"), RespSample: sampleWalletTransactions,
	},
}
