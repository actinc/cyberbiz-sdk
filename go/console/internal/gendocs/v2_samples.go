package gendocs

// Synthetic v2 samples. Their shape follows the CYBERBIZ Notion reference
// (docs/references/notes/notion-v2-api.md); the values are fake. They are
// used where no Golden File exists or the Golden File is empty.

const sampleProductFeeds = `{
  "product_feeds": [
    {"name": "facebook", "url": "https://example.cyberbiz.co/products/fbcatalog/0123456789abcdef0123456789abcdef", "mime_type": "text/csv"},
    {"name": "google", "url": "https://example.cyberbiz.co/products/google_product_feed/0123456789abcdef0123456789abcdef", "mime_type": "text/csv"},
    {"name": "shopdotcom", "url": "https://example.cyberbiz.co/shopdotcom.xml", "mime_type": "text/xml"},
    {"name": "line", "url": "https://example.cyberbiz.co/line_shopping/0123456789abcdef0123456789abcdef/product_full.json", "mime_type": "application/json"}
  ]
}`

const sampleAffiliateOrders = `[
  {
    "status_code": "create_succeed",
    "uid": "0123456789abcdef0123456789abcdef",
    "cid": "0123456789abcdef0123456789abcdef",
    "vendor_name": "example-vendor",
    "order": {
      "customer_id": 1,
      "subtotal_price": 300.0,
      "created_at": "2026-09-01 10:00:00",
      "updated_at": "2026-09-01 10:00:00",
      "order_number": 1001,
      "order_name": "#1001",
      "buyer": {"email": "customer@example.com", "mobile": "0912345678"},
      "line_items": [
        {
          "id": 2, "product_id": 3, "product_variant_id": 4, "title": "範例商品", "variant_title": "",
          "sku": "SKU-001", "qc": null, "vendor": null, "price": 300.0, "cost": null, "quantity": 1,
          "item_type": "normal", "discount_name": "", "discounts": [],
          "total_price_before_discounts": 300.0, "total_discount": 0.0, "total_price_after_discounts": 300.0,
          "tax_type_id": "inclusive_tax", "bonus_redemption_price": null, "related_items": [],
          "created_at": "2026-09-01 10:00:00"
        }
      ],
      "shipping_type": "cyberbiz",
      "shipping_name": "自取",
      "delivery_date": null,
      "delivery_time": 0,
      "delegate": null,
      "prices": {
        "total_line_items_price": 300.0,
        "shipping_rate_price": 10.0,
        "discounts": {
          "special_collection_discount": 0.0, "vip_discount": 0.0, "shop_discount": null, "coupon_discount": null,
          "coupon_discounts": [], "bonus_consumed": 0.0, "vip_shipping_discount": 0.0, "coupon_shipping_discount": 0.0, "price_discount": 0.0
        },
        "total_price": 310.0
      },
      "transaction_number": null,
      "merchant_trade_no": "S1#1001",
      "statuses": {"order_status": "open", "financial_status": "pending", "fulfillment_status": "unshipped", "return_status": "no_need"},
      "timings": {"request_return_at": null, "return_at": null, "refund_at": null, "closed_at": null, "cancelled_at": null, "expired_at": null, "confirmed_at": "2026-09-01 10:00:00"},
      "return_histories": [],
      "note": "",
      "total_bonus_redemption_price": 0.0,
      "linked_order_info": {"source": "非導購訂單"},
      "shipping_status": null,
      "extra_info": null,
      "from_device": "桌機"
    }
  }
]`

const samplePage = `{
  "id": 1,
  "title": "Demo Page",
  "handle": "demo-page",
  "status": "published",
  "html_infos": "{\"1714492800_app_store_api\":\"\\u003ch1\\u003eDemo content\\u003c/h1\\u003e\"}",
  "created_at": "2026-09-01 10:00:00",
  "updated_at": "2026-09-01 10:00:00"
}`

const samplePageCreate = `{"title": "Demo Page"}`

const samplePageUpdate = `{
  "title": "New Title",
  "handle": "new-handle",
  "status": "published",
  "section_id": "1714492800_app_store_api",
  "section_content": "\"\\u003ch1\\u003eNew content\\u003c/h1\\u003e\""
}`

const sampleCvsShippingRequest = `{"measurement": "S60"}`

const sampleCvsFulfillment = `{
  "id": 1,
  "tracking_company": "FAMILY_C2C",
  "cvs_shipping_type": "family_c2c",
  "tracking_number": "TRK000000001",
  "fulfilled_at": "2026-09-01 10:00:00",
  "received_at": null,
  "status": "fulfilled",
  "line_items": [
    {
      "id": 2, "product_id": 3, "product_variant_id": 4, "title": "範例商品", "variant_title": "",
      "sku": "SKU-001", "qc": null, "vendor": null, "price": 1000.0, "cost": 500.0, "quantity": 1,
      "item_type": "normal", "discount_name": "", "discounts": [],
      "total_price_before_discounts": 1000.0, "total_discount": 0.0, "total_price_after_discounts": 1000.0,
      "tax_type_id": "inclusive_tax", "bonus_redemption_price": null, "related_items": [],
      "created_at": "2026-09-01 10:00:00"
    }
  ]
}`

const sampleCvsLabelsRequest = `{"shipping_type": "seven", "fulfillment_ids": [1, 2, 3]}`

const sampleSupportShippingRequest = `{
  "shipping_type": "ezcat",
  "size": 60,
  "temperature": "normal",
  "fridge_or_frozen": "none",
  "use_transfer": false,
  "is_fragile": true,
  "shipping_orders": [
    {"order_id": 1, "line_item_ids": [2]},
    {"order_id": 3, "line_item_ids": [4]}
  ]
}`

const sampleSupportShippingResult = `{
  "failed_orders": [
    {"order_id": 3, "message": "訂單 id = 3 無此商品 line_item_ids = [4]"}
  ],
  "fulfillments": [
    {"id": 5, "order_id": 1, "tracking_number": "TRK000000001", "tracking_company": "ezcat"}
  ]
}`

const sampleSupportLabelsRequest = `{"shipping_type": "ezcat", "print_type": "normal", "order_ids": [1, 2]}`

const sampleMenus = `[
  {"id": 1, "title": "主選單", "system_default": true},
  {"id": 2, "title": "頁腳", "system_default": true}
]`

const sampleMenu = `{
  "id": 1,
  "title": "主選單",
  "system_default": true,
  "items": [
    {"id": 10, "title": "首頁", "subject_handle": null, "url": "/", "position": 1, "subtitle": null, "icon_image_url": null, "item_type": "frontpage", "level": 1, "items": []},
    {"id": 11, "title": "商品列表", "subject_handle": null, "url": "/collections/all", "position": 2, "subtitle": null, "icon_image_url": null, "item_type": "collections_all", "level": 1,
      "items": [
        {"id": 12, "title": "關於我們", "subject_handle": "about-us", "url": "/pages/about-us", "position": 1, "subtitle": null, "icon_image_url": null, "item_type": "page", "level": 2, "items": []}
      ]
    }
  ]
}`

const sampleCategories = `[
  {"title": "服飾", "handle": "clothing", "full_handle": "clothing", "id": 1},
  {"title": "男裝", "handle": "mens", "full_handle": "clothing/mens", "id": 2}
]`

const sampleCustomCollectionsV2 = `[
  {"title": "首頁商品", "handle": "frontpage", "published": true, "body_html": null, "products_order_name": "按標題拼音升序: A-Z", "position": 1, "id": 1},
  {"title": "範例群組", "handle": "sample", "published": true, "body_html": "<p>範例內容</p>", "products_order_name": "手動排序", "position": 2, "id": 2}
]`

const sampleSmartCollectionsV2 = `[
  {
    "title": "千元以內", "handle": "within_a_thousand", "published": true, "body_html": null, "position": 1,
    "rules": [{"id": 1, "filter": "商品價格 小於 1000"}],
    "products": [{"id": 1, "title": "範例商品", "position": 1}, {"id": 2, "title": "範例商品 B", "position": 2}],
    "id": 1
  }
]`

const sampleCustomerOAuthRequest = `{"uid": "U0000000000000000000000000000001", "provider": "line"}`

const sampleCustomersV2 = `[
  {
    "id": 1,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": null,
    "mobile": "0912345678",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": null,
    "accepts_marketing": true,
    "gender": "female",
    "birthday": "1990-01-01",
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": null,
    "note": null,
    "created_at": "2026-03-01 12:04:02",
    "updated_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-03-01 12:04:02",
    "mobile_sms_confirmed_at": "2026-03-01 12:04:02"
  }
]`

// sampleCustomerV2Included is the fuller shape the Notion reference shows
// when include_params asks for the optional sections.
const sampleCustomerV2Included = `{
  "id": 1,
  "name": "王小明",
  "status": "enabled",
  "email": "customer@example.com",
  "country_calling_code": null,
  "mobile": "0912345678",
  "enable_cvs_pickup": true,
  "enable_cvs_cod": true,
  "enable_home_delivery_cod": null,
  "accepts_marketing": true,
  "accepts_email_notification": true,
  "tags": [],
  "address": {
    "company": null,
    "country_calling_code": null,
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {"zip": "110", "country": null, "province": null, "city": "台北市", "district": "信義區", "address1": "市府路1號", "address2": null}
  },
  "gender": "female",
  "birthday": "1990-01-01",
  "other_accumulated_consumption": 0,
  "other_accumulated_consumption_expired_at": null,
  "note": null,
  "custom_fields": [],
  "created_at": "2026-03-01 12:04:02",
  "updated_at": "2026-09-01 10:00:00",
  "confirmed_at": "2026-03-01 12:04:02",
  "mobile_sms_confirmed_at": "2026-03-01 12:04:02",
  "bonus_remain": 30.0,
  "uid_providers": [{"provider_type": "line", "uid": "U0000000000000000000000000000001"}],
  "vip_info": {
    "customer_id": 1,
    "current_group": null,
    "current_level": null,
    "next_level": null,
    "extra_info": {
      "start_at": null,
      "end_at": null,
      "difference_of_total_spent_in_validity_days_for_renewal": null,
      "difference_of_total_spent_for_renewal": null,
      "difference_of_total_spent_in_validity_days_for_upgrade": null,
      "difference_of_total_spent_for_upgrade": null
    }
  }
}`

const sampleShopEmails = `{
  "shop_email": "shop@example.com",
  "subscribes": ["shop@example.com", "owner@example.com"]
}`

const sampleProductSearchRequest = `{
  "q": "範例",
  "limit": 50,
  "offset": 0,
  "filter_published": true,
  "skus": ["SKU-001"],
  "store_types": [1]
}`

const sampleProductSearchResult = `{
  "products": [
    {
      "id": 1,
      "title": "範例商品",
      "handle": "sample-product",
      "english_title": null,
      "product_url": "//example.cyberbiz.co/products/sample-product",
      "published": true,
      "sell_from": null,
      "sell_to": null,
      "product_type": null,
      "product_type_code": null,
      "slogan": "",
      "brief": "",
      "brief_text": null,
      "brief_includes_html": false,
      "body_html": "",
      "vendor": null,
      "price": 1000.0,
      "sell_weight": 0,
      "tax_type_id": "inclusive_tax",
      "custom_collections": [
        {"title": "首頁商品", "handle": "frontpage", "published": true, "body_html": null, "products_order_name": "按標題拼音升序: A-Z", "position": 0, "id": 1}
      ],
      "special_collection": null,
      "tags": [],
      "product_variants": [
        {
          "product_id": 1, "name": "範例商品 - 大 / 紅色", "position": 1, "price": 1000.0, "cost": null, "compare_at_price": 1500.0,
          "meas": 0.0, "max_usable_bonus": 0.0, "weight": 0.0, "option1": "大", "option2": "紅色", "option3": null,
          "inventory_management": false, "inventory_quantity": null, "sold": 0, "safety_inventory_quantity": null,
          "inventory_policy": "deny", "sku": "SKU-001", "qc": null, "requires_shipping": true,
          "created_at": "2026-09-01 10:00:00", "updated_at": "2026-09-01 10:00:00", "honeycomb_sync": true, "vendor": null, "id": 2,
          "photo_urls": ["//example.cyberbiz.co/media/sample-product.jpg"]
        }
      ],
      "product_options": [{"name": "大小", "position": 1, "types": "大,中,小", "id": 3}],
      "pos_shop": null,
      "photo_urls": ["//example.cyberbiz.co/media/sample-product.jpg"],
      "related_collections": [],
      "created_at": "2026-09-01 10:00:00",
      "updated_at": "2026-09-01 10:00:00",
      "temperature_types": ["常溫"],
      "searchable": true,
      "google_product_category_id": null,
      "seo_meta_tags": {"title": null, "description": null, "keywords": null}
    }
  ],
  "total_count": 1,
  "total_pages": 1
}`

const sampleInventoryRequest = `{
  "store_number": "S001",
  "items": [
    {"sku": "SKU-001", "inventory_quantity": 10},
    {"sku": "SKU-002", "inventory_quantity": 5}
  ]
}`

const sampleInventoryQueued = `{
  "job_id": "5f1a2b3c4d5e6f7a8b9c0d1e",
  "status": "QUEUED",
  "message": "Inventory update job has been queued",
  "store_number": "S001",
  "total_items": 2
}`

const sampleInventoryStatus = `{
  "job_id": "5f1a2b3c4d5e6f7a8b9c0d1e",
  "status": "SUCCESS",
  "message": "Inventory update completed successfully",
  "created_at": "2026-09-01 10:00:00 +0800",
  "completed_at": "2026-09-01 10:00:17 +0800",
  "result": {
    "total": 2,
    "succeeded": 1,
    "failed": 1,
    "failed_items": [{"sku": "SKU-002", "error": "sku not found"}]
  },
  "error": null
}`

const sampleWalletBalance = `{"balance": 1500.0, "currency": "TWD"}`

const sampleWalletTransactions = `[
  {
    "type": "refund",
    "amount": -500.0,
    "balance_after": 1000.0,
    "created_at": "2026-09-01 12:04+0800",
    "payment_name": "現金",
    "invalid_einvoices": [
      {"title": "", "company_no": "", "invoice_no": "AB12345678", "invoice_status": "issue_invalid", "invoice_at": "2026-09-01 12:01+0800", "invalid_at": "2026-09-01 12:04+0800", "random_num": "1234", "invoice_type": "default", "love_code": "", "phone_barcode": "", "nature_person": ""}
    ],
    "allowance_einvoices": []
  },
  {
    "type": "topup",
    "amount": 1500.0,
    "balance_after": 1500.0,
    "created_at": "2026-09-01 12:02+0800",
    "payment_name": "現金",
    "financial_status": "paid",
    "einvoice": {"title": "", "company_no": "", "invoice_no": "AB12345679", "invoice_status": "issue", "invoice_at": "2026-09-01 12:02+0800", "invalid_at": null, "random_num": "5678", "invoice_type": "default", "love_code": "", "phone_barcode": "", "nature_person": ""},
    "paper_invoice_no": null
  },
  {"type": "consumption", "amount": -200.0, "balance_after": 1300.0, "created_at": "2026-08-12 15:42+0800", "payment_name": null}
]`

const sampleShop = `{
  "shop_info": {
    "primary_domain": "example.cyberbiz.co",
    "name": "範例商店",
    "id": 1,
    "email": "shop@example.com",
    "og_image_url": "//example.cyberbiz.co/media/og_image.jpg",
    "shop_line_chat_bot": null,
    "shop_line": {"login_enable": true, "liff_id": null, "liff_enable": false},
    "language": "zh-TW",
    "currency": "TWD",
    "sms_prefix": "【範例商店】",
    "merchant_location": "Taiwan",
    "market_location": "Taiwan"
  }
}`

const sampleSettings = `{
  "shop_add_on": {
    "id": 1,
    "settings": {"your_field_name": "value"},
    "vendor_type": "example-app",
    "token": "synthetic-token-do-not-use",
    "start_at": null,
    "end_at": null,
    "add_on_version": {
      "webhook_url": "https://example.com/webhooks/cyberbiz",
      "manifest": {
        "name": "Example App",
        "version": "0.1.0",
        "scopes": "public",
        "manifest_version": 1,
        "type": "other",
        "webhook_events": ["customers/create", "orders/paid"]
      },
      "embedded": true,
      "status": "init",
      "scopes": "public read_products read_customers read_orders"
    }
  }
}`

const sampleSettingsUpdate = `{"settings": [{"field": "your_field_name", "data": "value"}]}`

const sampleSettingsUpdated = `{
  "shop_add_on": {
    "id": 1,
    "settings": {"your_field_name": "value"},
    "vendor_type": "example-app",
    "token": "synthetic-token-do-not-use",
    "start_at": null,
    "end_at": null
  }
}`
