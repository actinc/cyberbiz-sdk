# CYBERBIZ Webhooks

Version: 1.0.1

CYBERBIZ posts an HTTP request to the HTTPS URL an app registers (`webhook_url` in the app
manifest) whenever a subscribed Event happens in a Shop. This document lists every Event, the
HTTP headers CYBERBIZ sends, how to verify the Signature, and a synthetic Sample of every payload.

## Delivery

- Method `POST`, body `application/json`, UTF-8. The body is the payload object of the Event, not an envelope.
- One request per Event per resource. Several Events for the same resource (e.g. `orders/create` and `orders/paid`, or two `customers/update`) can arrive within the same second and in any order; order by `updated_at` and the resource id, not by arrival time.
- Answer `2xx` quickly (a few seconds). Do the work asynchronously.

### Headers

| Header | Example | Description |
| --- | --- | --- |
| `User-Agent` | `CyberbizAppWebhook/1.0` | Fixed for App webhooks (observed). The documentation says `CyberbizWebhook/1.0`, which is the older shop-webhook flavour. |
| `X-Cyberbiz-Domain` | `example.cyberbiz.co` | The Shop Domain, i.e. the shop's CYBERBIZ-issued hostname. Use it to find the shop's App Secret. |
| `X-Cyberbiz-Shop-Domain` | `www.example-shop.com` | The merchant's custom storefront hostname. Informational; not an identifier. |
| `X-Cyberbiz-Event` | `orders/paid` | The Event, named `resource/action`. |
| `X-Cyberbiz-Hmac-Sha256` | `3f2a…` (64 hex chars) | The Signature: HMAC-SHA256 of the raw request body keyed by the App Secret. |
| `X-Cyberbiz-Domain-Hmac-Sha256` | `9b1c…` (64 hex chars) | The Domain Signature: HMAC-SHA256 of the `X-Cyberbiz-Domain` value keyed by the App Secret. |

## Verifying the signature

1. Read the raw body bytes before parsing them.
2. Compute `HMAC-SHA256(app_secret, body)`.
3. Compare, in constant time, the **hex** encoding of the digest with `X-Cyberbiz-Hmac-Sha256`. Production App webhooks send hex (64 characters), although the CYBERBIZ documentation describes base64 — accept base64 as a fallback so a future change on the platform does not break verification.
4. Optionally check `X-Cyberbiz-Domain-Hmac-Sha256` the same way against `X-Cyberbiz-Domain`; it binds the request to one shop but does not authenticate the payload.
5. Reject anything that fails with `401` and do not process the body.

```go
func verify(body []byte, header, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sum := mac.Sum(nil)
	hexOK := hmac.Equal([]byte(hex.EncodeToString(sum)), []byte(header))
	b64OK := hmac.Equal([]byte(base64.StdEncoding.EncodeToString(sum)), []byte(header))
	return hexOK || b64OK
}
```

## Retries and idempotency

- The retry policy is **not documented**: assume a failed delivery may or may not be retried, and that a retry may arrive minutes later or never. Reconcile through the API (`GET /v1/orders/{order_id}`, ...) when a gap is suspected.
- Deliveries carry no delivery id. Deduplicate on `(X-Cyberbiz-Event, payload id, payload updated_at)`.
- Treat every handler as idempotent: the same Event can be delivered more than once, and Events for one resource can arrive out of order within the same second.

## Testing your receiver with Postman

The collection `cyberbiz-webhooks.postman_collection.json` next to this file replays every Event below against your own receiver: set `webhookUrl`, `shopDomain`, `customDomain` and `appSecret` in a Postman environment (never commit it), and its pre-request script signs each request exactly as CYBERBIZ does (`signatureEncoding` selects hex, the production form, or base64, the documented form).

## Events

| Event | Description | Payload |
| --- | --- | --- |
| `customers/create` | A customer registered. | [Customer](#customer) |
| `customers/update` | A customer's profile changed. | [Customer](#customer) |
| `uid_providers/create` | A social-login uid was linked to a customer (the reference documents no separate payload; the Customer object is posted). | [Customer](#customer) |
| `uid_providers/update` | A social-login uid of a customer changed (Customer object). | [Customer](#customer) |
| `bonus_points/create` | Bonus points were granted. | [Bonus Points](#bonus-points) |
| `bonus_points/update` | Bonus points were used. | [Bonus Points](#bonus-points) |
| `bonus_points/destroy` | A bonus point record was deleted. | [Bonus Points](#bonus-points) |
| `comment_bonus/create` | Bonus points were granted for an approved product review. | [Bonus Points](#bonus-points) |
| `orders/create` | An order was placed. | [Order](#order) |
| `orders/paid` | An order was paid. | [Order](#order) |
| `orders/preparing` | An order is being prepared for shipment. | [Order](#order) |
| `orders/fulfilled` | An order was shipped. | [Order](#order) |
| `orders/received` | An order was received by the customer. | [Order](#order) |
| `orders/arrived` | An order arrived at the pickup store. | [Order](#order) |
| `orders/expired` | An order was not picked up in time. | [Order](#order) |
| `orders/cancelled` | An order was cancelled. | [Order](#order) |
| `orders/returned` | An order was returned. | [Order](#order) |
| `orders/partial_return` | Part of an order was returned. | [Order](#order) |
| `orders/refunded` | An order was refunded. | [Order](#order) |
| `orders/partial_refunded` | Part of an order was refunded. | [Order](#order) |
| `orders/closed` | An order was closed. | [Order](#order) |
| `orders/request_return` | The customer requested a return. | [Order](#order) |
| `orders/opened` | An order was (re)opened. | [Order](#order) |
| `express_delivery_orders/update` | An express-delivery order changed. | [Order](#order) |
| `products/create` | A product was created. | [Product](#product) |
| `products/update` | A product changed. | [Product](#product) |
| `products/delete` | A product was deleted. | [Product](#product) |
| `variants/create` | A product variant was created. | [Product Variant](#product-variant) |
| `variants/update` | A product variant changed. | [Product Variant](#product-variant) |
| `variants/delete` | A product variant was deleted. | [Product Variant](#product-variant) |
| `coupons/create` | A coupon was created. | [Coupon](#coupon) |
| `coupons/update` | A coupon was used. | [Coupon](#coupon) |
| `coupons/destroy` | A coupon was deleted. | [Coupon](#coupon) |
| `customer_vip_level/update` | A customer's VIP level changed. | [Customer VIP Level](#customer-vip-level) |
| `apps/uninstall` | The app was uninstalled from the shop. | [App Uninstall](#app-uninstall) |

### Events without documented payloads

These events appear in app manifests (`webhook_events` of `GET /settings`) but the reference documents no payload for them:

- `customer_tags/update`
- `affiliate_vendor_orders/create`
- `affiliate_vendor_orders/closed`
- `affiliate_vendor_orders/cancelled`
- `orders/update`
- `products/update_photos`
- `products/update_options`
- `domains/create`
- `domains/update`
- `domains/destroy`
- `collections/create`
- `collections/update`
- `collections/delete`
- `bulk_operations/finish`
- `menus/create`
- `menus/update`
- `menus/destroy`

## Payload objects

Field types are as documented by CYBERBIZ. Timestamps are `YYYY-MM-DD HH:MM:SS` in Asia/Taipei; money fields are floats; any field may be `null`.

### Customer

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | Customer ID. |
| `name` | `String` | Name. |
| `status` | `String` | Status: `pending` (not activated), `validate` (not verified), `enabled` (activated), `disabled` (disabled), `invited` (activation invitation sent), `declined` (invitation declined), `warning` (flagged account). |
| `email` | `String` | Email |
| `country_calling_code` | `String` | Mobile country calling code (e.g. +886). |
| `mobile` | `String` | Mobile number. |
| `gender` | `String` | Gender. |
| `birthday` | `String` | Birthday. |
| `enable_cvs_pickup` | `Boolean` | Convenience-store pickup allowed. |
| `enable_cvs_cod` | `Boolean` | Convenience-store cash on delivery allowed. |
| `enable_home_delivery_cod` | `Boolean` | Home-delivery cash on delivery allowed. |
| `accepts_marketing` | `Boolean` | Accepts marketing. |
| `accepts_email_notification` | `Boolean` | Accepts notification e-mails. |
| `tags` | `[Object]` | Customer tags. |
| `address` | `Object` | Customer address. |
| `other_accumulated_consumption` | `Integer` | Accumulated spend from other channels. |
| `other_accumulated_consumption_expired_at` | `String` | Start date of the other-channel accumulated spend. |
| `note` | `String` | Note. |
| `custom_fields` | `[Object]` | Custom fields. |
| `bonus_remain` | `Float` | Total bonus points. |
| `uid_providers` | `[Object]` | UID providers. |
| `created_at` | `String` | Created at. |
| `updated_at` | `String` | Updated at. |
| `confirmed_at` | `String` | E-mail verified at. |
| `mobile_sms_confirmed_at` | `String` | Mobile verified at. |

### Customer Tag

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Tag name. |

### Customer Address

| Field | Type | Description |
| --- | --- | --- |
| `company` | `String` | Company name. |
| `country_calling_code` | `String` | Phone country calling code. |
| `phone` | `String` | Phone. |
| `address` | `String` | Address. |
| `detail_address` | `Object` | Address details. |

### Detail Address

| Field | Type | Description |
| --- | --- | --- |
| `zip` | `String` | Postal code. |
| `country` | `String` | Country. |
| `province` | `String` | Province. |
| `city` | `String` | City. |
| `district` | `String` | District. |
| `address1` | `String` | Address line 1. |
| `address2` | `String` | Address line 2. |

### Custom Field

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Field name. |
| `label` | `String` | Field label. |
| `value` | `String` | Field value. |

### UID Providers

| Field | Type | Description |
| --- | --- | --- |
| `provider_type` | `String` | Provider type. |
| `uid` | `String` | UID |

### Bonus Points

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `title` | `String` | Name. |
| `points` | `Float` | Points. |
| `unused_points` | `Float` | Unused points. |
| `consumption_price` | `Float` | Spend that produced the bonus. |
| `deadline` | `String` | Validity. |
| `customer_id` | `Integer` | Related customer ID. |
| `source` | `String` | Source. |
| `order_id` | `Integer` | Related order ID. |

### Order

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | Order ID. |
| `token` | `String` | Token |
| `order_number` | `String` | Order number. |
| `order_name` | `String` | Order name. |
| `customer` | `Object` | Customer information. |
| `buyer` | `Object` | Buyer information. |
| `receiver` | `Object` | Recipient information. |
| `billing_address` | `Object` | Billing information (custom feature). |
| `line_items` | `[Object]` | Order line items. |
| `shipping_type` | `String` | Shipping method. |
| `shipping_name` | `String` | Shipping method name. |
| `shipping_vendor` | `Hash` | Carrier. |
| `logistics_id` | `String` | ECPay logistics order number. |
| `delivery_date` | `String` | Requested delivery date. |
| `delivery_time` | `Integer` | Requested delivery time. |
| `delegate` | `String` | Staff who placed the order on the customer's behalf. |
| `fulfillments` | `[Object]` | Fulfillments. |
| `payment_name` | `String` | Payment method. |
| `payment_method` | `String` | Payment method name. |
| `payment_url` | `String` | Payment URL. |
| `multiple_payment_infos` | `[Object]` | Multiple payment information. |
| `prices` | `Object` | Price information. |
| `card4no` | `String` | Last four digits of the payment card. |
| `transaction_number` | `String` | Third-party transaction number. |
| `merchant_trade_no` | `String` | Merchant trade number. |
| `einvoice` | `Object` | Invoice information. |
| `paper_invoice_no` | `String` | Paper invoice number. |
| `statuses` | `Object` | Status information. |
| `timings` | `Object` | Timestamps. |
| `return_histories` | `[Object]` | Refund history. |
| `note` | `String` | Order note. |
| `branch_store` | `Object` | Pickup store information. |
| `referral_code` | `String` | Referral (affiliate) code. |
| `checkout_referral_code` | `String` | Checkout affiliate code. |
| `checkout_referral_user_name` | `String` | Checkout user name. |
| `register_referral_code` | `String` | Registrant affiliate code. |
| `total_bonus_redemption_price` | `Integer` | Total bonus points redeemed in the bonus mall. |
| `pos_info` | `Object` | POS information. |
| `exchange_histories` | `Object` | Exchange history. |
| `linked_order_info` | `Object` | Linked-order (affiliate) information. |
| `tags` | `[Object]` | Order tags. |
| `express_delivery_branch_store` | `Object` | Express-delivery store information. |
| `shipping_status` | `String` | Shipping status. |
| `extra_info` | `String` | Extra information. |
| `from_device` | `String` | Order source. |
| `customer_cancel_reason_detail` | `Object` | Cancellation reason details. |
| `serial_numbers` | `[String]` | Campaign serial numbers. |
| `order_weight` | `Float` | Total weight. |
| `created_at` | `String` | Created at. |
| `updated_at` | `String` | Updated at. |

### Buyer

| Field | Type | Description |
| --- | --- | --- |
| `email` | `String` | Buyer e-mail. |
| `mobile` | `String` | Buyer mobile number. |

### Receiver

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Recipient name. |
| `country_calling_code` | `String` | Recipient phone country calling code. |
| `phone` | `String` | Recipient phone. |
| `address` | `String` | Recipient address. |
| `detail_address` | `Object` | Recipient address details. |
| `cvs_store_id` | `String` | Convenience-store pickup store number. |
| `allpay_logistics_id` | `String` | ECPay logistics number. |

### Line Item

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `product_id` | `Integer` | Product ID. |
| `product_variant_id` | `Integer` | Variant ID. |
| `title` | `String` | Product name. |
| `variant_title` | `String` | Variant name. |
| `sku` | `String` | Variant SKU. |
| `qc` | `String` | Variant vendor item code. |
| `vendor` | `String` | Vendor. |
| `price` | `Float` | Amount. |
| `cost` | `Float` | Cost. |
| `quantity` | `Integer` | Quantity. |
| `item_type` | `String` | Type. |
| `return_status` | `String` | Return status. |
| `discount_name` | `String` | Discount name. |
| `discounts` | `[Object]` | Itemised product discounts. |
| `total_price_before_discounts` | `Float` | Total before discounts. |
| `total_discount` | `Float` | Total discount. |
| `total_price_after_discounts` | `Float` | Total after discounts. |
| `tax_type_id` | `String` | Tax type: `inclusive_tax` (taxable), `zero_tax` (zero-rated), `exclusive_tax` (tax exempt). |
| `bonus_redemption_price` | `Integer` | Bonus points redeemed in the bonus mall; null when not a bonus redemption. |
| `related_items` | `[Object]` | Bundle component items. |
| `channel` | `String` | Product channel. |
| `weight` | `String` | Product weight. |
| `photo` | `String` | Product photo. |
| `created_at` | `Date` | Created at. |

### Discount

| Field | Type | Description |
| --- | --- | --- |
| `position` | `Integer` | Item position (nth item). |
| `id` | `Integer` | Discount type ID. |
| `code` | `String` | Discount type code. |
| `name` | `String` | Discount type name. |
| `discount` | `Integer` | Discount amount. |

### Shipping Vendor

| Field | Type | Description |
| --- | --- | --- |
| `type` | `String` | Carrier code. |
| `name` | `String` | Carrier name. |

### Fulfillment

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `tracking_company` | `String` | Shipping method. |
| `tracking_number` | `String` | Tracking number. |
| `fulfilled_at` | `String` | Shipped at. |
| `received_at` | `String` | Received at. |
| `status` | `String` | Shipping status. |
| `line_items` | `[Object]` | Shipped items. |
| `tracking_url` | `String` | Tracking URL (Uber Direct and Pandago only). |

### Multiple Payment Info

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Payment name. |
| `amount` | `Float` | Payment amount. |

### Prices

| Field | Type | Description |
| --- | --- | --- |
| `total_line_items_price` | `Float` | Products total. |
| `shipping_rate_price` | `Float` | Shipping fee. |
| `discounts` | `Object` | Discount breakdown. |
| `total_price` | `Float` | Order total. |

### Discounts Detail

| Field | Type | Description |
| --- | --- | --- |
| `special_collection_discount` | `Integer` | Special collection (campaign) discount. |
| `vip_discount` | `Float` | VIP discount. |
| `shop_discount` | `Object` | Shop-wide discount details. |
| `coupon_discount` | `Object` | Coupon detail (single; deprecated). |
| `coupon_discounts` | `[Object]` | Coupon details. |
| `bonus_consumed` | `Float` | Bonus discount. |
| `vip_shipping_discount` | `Float` | VIP shipping discount. |
| `coupon_shipping_discount` | `Float` | Free-shipping coupon discount. |
| `price_discount` | `Integer` | Manager (manual) discount. |
| `third_party_discount` | `Integer` | Third-party discount. |

### Shop Discount

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Shop-wide campaign name. |
| `amount` | `Float` | Shop-wide campaign discount. |

### Coupon Discount

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | Coupon name. |
| `code` | `String` | Coupon code. |
| `amount` | `Float` | Coupon discount. |
| `coupon_id` | `Integer` | Coupon ID. |

### E-Invoice

| Field | Type | Description |
| --- | --- | --- |
| `title` | `String` | Invoice title. |
| `order_id` | `Integer` | Order ID. |
| `company_no` | `String` | Invoice tax ID. |
| `invoice_no` | `String` | Invoice number. |
| `invoice_status` | `String` | Invoice status. |
| `invoice_at` | `String` | Invoice issued at. |
| `invalid_at` | `String` | Invoice voided or allowance time. |
| `random_num` | `String` | Random code. |
| `invoice_type` | `String` | Invoice type. |
| `love_code` | `String` | Donation code. |
| `phone_barcode` | `String` | Mobile barcode carrier. |
| `nature_person` | `String` | Citizen digital certificate carrier. |

### Statuses

| Field | Type | Description |
| --- | --- | --- |
| `order_status` | `String` | Order status. |
| `financial_status` | `String` | Payment status. |
| `fulfillment_status` | `String` | Shipping status. |
| `return_status` | `String` | Return status. |

### Timings

| Field | Type | Description |
| --- | --- | --- |
| `request_return_at` | `String` | Return requested at. |
| `return_at` | `String` | Returned at. |
| `refund_at` | `String` | Refunded at. |
| `closed_at` | `String` | Closed at. |
| `cancelled_at` | `String` | Cancelled at. |
| `expired_at` | `String` | Pickup expired at. |
| `confirmed_at` | `String` | Order confirmed at. |

### Return History

| Field | Type | Description |
| --- | --- | --- |
| `body` | `String` | Refund summary. |
| `price` | `Float` | Refund amount. |
| `refunded_at` | `String` | Refunded at. |

### Branch Store

| Field | Type | Description |
| --- | --- | --- |
| `store_no` | `String` | Store number. |
| `name` | `String` | Store name. |
| `phone` | `String` | Store phone. |
| `county` | `String` | City. |
| `district` | `String` | District / township. |
| `address` | `String` | Address. |
| `zip` | `String` | Postal code. |
| `opening_hours` | `String` | Opening hours. |
| `lat` | `Float` | Latitude. |
| `lng` | `Float` | Longitude. |
| `enabled` | `Boolean` | Enabled. |
| `shipping_rates` | `[Object]` | Store shipping rates. |
| `source_type` | `String` | Source type: `BranchStore` or `PosShop`. |
| `source_id` | `Integer` | Source ID. |

### POS Info

| Field | Type | Description |
| --- | --- | --- |
| `pos_user_id` | `Integer` | POS salesperson ID. |
| `pos_user_email` | `String` | POS salesperson e-mail. |
| `pos_shop_id` | `Integer` | POS shop ID of the salesperson. |
| `pos_info` | `String` | POS information. |
| `pos_id` | `Integer` | POS terminal ID the order was sold on. |
| `pos_name` | `String` | POS terminal name the order was sold on. |

### Linked Order Info

| Field | Type | Description |
| --- | --- | --- |
| `source` | `String` | Affiliate source. |
| `shopdotcom_rid` | `String` | shopdotcom (Market America) RID. |
| `shopdotcom_click_id` | `String` | shopdotcom (Market America) click ID. |
| `line_shopping_ecid` | `String` | LINE Shopping ecid. |
| `line_shopping_affiliate` | `String` | LINE Shopping affiliate. |
| `ichannel_gid` | `String` | iChannel gid |

### Customer Cancel Reason Detail

| Field | Type | Description |
| --- | --- | --- |
| `source` | `String` | Cancellation source. |
| `reason_id` | `Integer` | Cancellation reason ID. |
| `reason_detail` | `String` | Cancellation reason details. |

### Product

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | Product ID. |
| `title` | `String` | Product name. |
| `english_title` | `String` | Product English name. |
| `product_url` | `String` | Product URL. |
| `published` | `Boolean` | Whether it is published. |
| `sell_from` | `String` | Sell-from time. |
| `sell_to` | `String` | Sell-to time. |
| `product_type` | `String` | Product type. |
| `product_type_code` | `String` | Product type code. |
| `slogan` | `String` | Product slogan. |
| `brief` | `String` | Product brief HTML template text. |
| `brief_text` | `String` | Product brief as plain text. |
| `brief_includes_html` | `Boolean` | Whether the product brief uses the HTML template text. |
| `body_html` | `String` | Product description content. |
| `vendor` | `String` | Vendor. |
| `price` | `Float` | Lowest product price. |
| `sell_weight` | `Integer` | Quantity sold. |
| `tax_type_id` | `String` | Tax type: `inclusive_tax` (taxable), `zero_tax` (zero-rated), `exclusive_tax` (tax exempt). |
| `custom_collections` | `[Object]` | Related custom collections. |
| `special_collection` | `Object` | Related special collection. |
| `tags` | `[Object]` | Product tags. |
| `product_variants` | `[Object]` | Product variants. |
| `product_options` | `[Object]` | Product options. |
| `pos_shop` | `Object` | Related POS shop. |
| `photo_urls` | `[String]` | Product photo URLs. |
| `photos` | `[Object]` | Product photos. |
| `channel` | `Object` | Product channel. |
| `related_collections` | `[Object]` | Related collections. |
| `branch_store` | `Object` | Branch store the product belongs to. |
| `temperature_types` | `[String]` | Temperature zone. |
| `searchable` | `Boolean` | Whether the product is searchable. |
| `google_product_category_id` | `Integer` | Google product category ID. |
| `product_custom_fields` | `[Hash]` | Product custom fields. |
| `seo_meta_tags` | `Hash` | SEO meta tags. |
| `required_customer_tags` | `[String]` | Customer tags allowed to buy. |
| `created_at` | `String` | Created at. |
| `updated_at` | `String` | Updated at. |

### Product Tag

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | Tag name. |
| `category` | `Integer` | Tag category. |

### Product Photo

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `url` | `String` | Product photo URL. |
| `position` | `Integer` | Product photo position. |

### Channel

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | Channel name. |

### Product Option

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | Option name. |
| `position` | `Integer` | Option position. |
| `types` | `String` | Option values. |

### Custom Collection

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `title` | `String` | Custom collection name. |
| `handle` | `String` | Custom collection handle. |
| `published` | `Boolean` | Whether it is published. |
| `body_html` | `String` | Custom collection description. |
| `products_order_name` | `String` | Custom collection product ordering. |
| `position` | `Integer` | Custom collection position. |
| `products` | `[Object]` | Products in the collection. |

### Special Collection

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `title` | `String` | Special collection name. |
| `handle` | `String` | Special collection handle. |
| `published` | `Boolean` | Whether it is published. |
| `start_date` | `String` | Collection start time. |
| `end_date` | `String` | Collection end time. |
| `body_html` | `String` | Special collection description. |
| `position` | `Integer` | Special collection position. |
| `special_collection_type` | `Object` | Special collection type. |
| `type_rules` | `[Object]` | Special collection rules. |
| `rest_include_discount` | `Boolean` | Whether remaining items count toward the discount. |
| `products` | `[Object]` | Products in the collection. |

### Product Variant

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `product_id` | `Integer` | Related product ID. |
| `name` | `String` | Related product name. |
| `position` | `Integer` | Variant position. |
| `price` | `Float` | Variant price. |
| `cost` | `Float` | Variant cost. |
| `compare_at_price` | `Float` | Variant compare-at price. |
| `meas` | `Float` | Variant volume (dimensional size). |
| `max_usable_bonus` | `Float` | Maximum bonus usable on the variant. |
| `weight` | `Float` | Variant weight. |
| `option1` | `String` | Variant option 1. |
| `option2` | `String` | Variant option 2. |
| `option3` | `String` | Variant option 3. |
| `inventory_management` | `Boolean` | Whether inventory is managed. |
| `inventory_quantity` | `Integer` | Variant inventory quantity. |
| `sold` | `Integer` | Variant quantity sold. |
| `safety_inventory_quantity` | `Integer` | Variant safety inventory level. |
| `inventory_policy` | `String` | Whether purchase is allowed when out of stock. |
| `sku` | `String` | Variant SKU. |
| `qc` | `String` | Variant vendor item code. |
| `requires_shipping` | `Boolean` | Whether the variant requires shipping. |
| `honeycomb_sync` | `Boolean` | Warehouse inventory sync. |
| `vendor` | `String` | Vendor. |
| `photo_urls` | `[String]` | Product photo URLs. |
| `pim_infos` | `[Object]` | Product PIM information. |
| `created_at` | `String` | Created at. |
| `updated_at` | `String` | Updated at. |

### PIM Info

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `product_id` | `Integer` | Product ID |
| `product_variant_id` | `Integer` | Product Variant ID |
| `pim_product_id` | `Integer` | Product ID in PIM |
| `pim_variant_id` | `Integer` | Product Variant ID in PIM |
| `channel` | `Integer` | Product Channel |
| `channel_shop_name` | `String` | Name of Channel Shop |
| `is_connected` | `Boolean` | Is Connected |
| `is_source` | `Boolean` | Is Source |
| `shop_id` | `Integer` | Shop ID |

### Cart

CYBERBIZ documents this payload without naming an event for it; no delivery has been observed and the SDK defines no cart Event. It is listed for completeness only.

| Field | Type | Description |
| --- | --- | --- |
| `cart_add_items` | `[Object]` | Variants added to the cart. |
| `customer` | `Object` | Customer information. |

### Cart Add Item

| Field | Type | Description |
| --- | --- | --- |
| `variant` | `Object` | Product variants. |
| `quantity` | `Integer` | Quantity added. |

### Coupon

| Field | Type | Description |
| --- | --- | --- |
| `customer_id` | `Integer` | Related customer ID (null for shop-wide coupons). |
| `title` | `String` | Coupon name. |
| `code` | `String` | Coupon code. |
| `coupon_type_name` | `String` | Coupon type: amount, percent or free shipping. |
| `coupon_value` | `String` | Coupon discount amount. |
| `order_price_threshold` | `Integer` | Coupon minimum order total. |
| `start_date` | `String` | Coupon start time. |
| `end_date` | `String` | Coupon end time. |
| `concurrently_apply` | `String` | Whether it can be combined with special collections or shop-wide campaigns. |
| `usage_limit` | `Integer` | Coupon usage limit. |
| `can_accumulate_bonus` | `Boolean` | Whether bonus points accumulate. |
| `usage_unlimited` | `Boolean` | Whether usage is unlimited. |
| `used_times` | `Integer` | Times the coupon was used. |
| `gift_order_id` | `Integer` | ID of the order that granted the coupon. |
| `gift_days` | `Integer` | Days the granted coupon stays usable. |
| `account_usage_limit_enabled` | `Boolean` | Per-account usage limit enabled. |
| `account_usage_limit` | `Integer` | Usage limit per account. |
| `restrict_strategy` | `String` | Restriction type: `unrestricted` (usable on every campaign product), `restrict` (usable except on the listed campaigns' products), `forbidden` (not usable when the order contains a listed campaign's product). |
| `restrict_campaigns` | `[String]` | Campaigns the restriction applies to. |
| `tags` | `[String]` | Product tags the coupon is bound to. |
| `pos_shop_ids` | `[Integer]` | POS shop IDs the coupon is bound to. |
| `coupon_status` | `String` | Coupon status: `no_start_use` (not yet usable), `used` (fully used up), `has_expire_date` (usable, expires), `no_expire_date` (usable, never expires), `expired`. Read together with `gift_order_status`: a coupon whose source order was cancelled is invalid. |
| `gift_order_status` | `String` | Status of the order that granted the coupon: `closed` (completed), `open` (in progress), `cancelled`. Read together with `coupon_status`. |
| `valid` | `Boolean` | Whether the coupon is valid. |
| `customer_used_times` | `Integer` | Times the customer used the shop coupon. |
| `customer_usable` | `Boolean` | Whether the customer still has uses left. |

### Customer VIP Level

| Field | Type | Description |
| --- | --- | --- |
| `customer_id` | `Integer` | Customer ID. |
| `current_group` | `Object` | Current VIP group. |
| `current_level` | `Object` | Current VIP level. |
| `next_level` | `Object` | Next VIP level. |
| `extra_info` | `Object` | Extra information. |

### Current VIP Group

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | Collection name. |
| `position` | `Integer` | Position. |
| `description_url` | `String` | Description page. |
| `customer_tags` | `[String]` | Customer tags. |
| `vip_group_levels` | `[Object]` | VIP levels. |

### VIP Level

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `position` | `Integer` | Position. |
| `name` | `String` | Level name. |
| `validity_days` | `Integer` | Validity. |
| `upgrade_condition_total_spent` | `Integer` | Upgrade condition: single-order spend. |
| `upgrade_condition_total_spent_in_validity_days` | `Integer` | Upgrade condition: spend within the validity period. |
| `renewal_condition_total_spent` | `Integer` | Renewal condition: single-order spend. |
| `renewal_condition_total_spent_in_validity_days` | `Integer` | Renewal condition: spend within the validity period. |
| `bonus_point_enabled` | `Boolean` | Bonus multiplier enabled. |
| `bonus_point_threshold` | `Integer` | Bonus multiplier: spend threshold. |
| `bonus_point_value` | `Integer` | Bonus multiplier: points granted. |
| `bonus_point_expiry_days` | `Integer` | Bonus multiplier: validity (days). |
| `birth_gift_enabled` | `Boolean` | Birthday gift enabled. |
| `birth_gift_name` | `String` | Birthday gift: name. |
| `birth_gift_before_days` | `String` | Birthday gift: days in advance. |
| `birth_gift_setting` | `Object` | Birthday gift: settings. |
| `upgrade_gift_enabled` | `Boolean` | Upgrade gift enabled. |
| `upgrade_gift_setting` | `Object` | Upgrade gift: settings. |
| `order_discount_enabled` | `Boolean` | Order discount enabled. |
| `order_discount_value` | `Integer` | Order discount (percent). |
| `order_discount_setting` | `Object` | Order discount settings. |
| `free_shipping_enabled` | `Boolean` | Free shipping enabled. |
| `free_shipping_setting` | `Object` | Order free-shipping settings. |

### Extra Info

| Field | Type | Description |
| --- | --- | --- |
| `start_at` | `String` | Level valid from. |
| `end_at` | `String` | Level valid until. |
| `difference_of_total_spent_in_validity_days_for_renewal` | `Integer` | Spend within the period still needed to renew. |
| `difference_of_total_spent_for_renewal` | `Integer` | Single-order spend still needed to renew. |
| `difference_of_total_spent_in_validity_days_for_upgrade` | `Integer` | Spend within the period still needed to upgrade. |
| `difference_of_total_spent_for_upgrade` | `Integer` | Single-order spend still needed to upgrade. |

### Gift Setting

| Field | Type | Description |
| --- | --- | --- |
| `bonus` | `Object` | Bonus settings. |
| `coupon` | `Object` | Coupon settings. |

### Bonus Setting

| Field | Type | Description |
| --- | --- | --- |
| `enabled` | `Boolean` | Enabled. |
| `value` | `Integer` | Bonus points granted. |
| `expiry_days` | `Integer` | Validity in days (0: never expires). |

### Coupon Setting

| Field | Type | Description |
| --- | --- | --- |
| `enabled` | `Boolean` | Enabled. |
| `presets` | `Object` | Settings. |

### Coupon Presets

| Field | Type | Description |
| --- | --- | --- |
| `coupon_type_id` | `Integer` | Coupon type: `1` (amount), `2` (percent). |
| `value` | `Integer` | Discount (amount or percent). |
| `code` | `String` | Coupon code. |
| `order_price_threshold` | `Integer` | Minimum spend. |
| `usable_days` | `Integer` | Usable days. |
| `usage_limit` | `Integer` | Number of coupons. |
| `product_tags` | `[String]` | Product tags. |
| `restrict_strategy` | `String` | Restriction type for combining with other campaigns. |
| `restrict_campaigns` | `[String]` | Campaigns the restriction applies to. |

### App Uninstall

| Field | Type | Description |
| --- | --- | --- |
| `app_uuid` | `String` | App UUID in App Market |
| `app_version_uuid` | `String` | App Version UUID in App Market |
| `app_client_id` | `String` | App Client ID |

### Exchange Histories

| Field | Type | Description |
| --- | --- | --- |
| `created_at` | `String` | Exchanged at. |
| `price` | `Float` | Exchange amount. |
| `order_price_before` | `Integer` | Order amount before the exchange. |
| `order_price_after` | `Integer` | Order amount after the exchange. |
| `pos_shop_id` | `Integer` | Exchange POS shop ID. |
| `pos_id` | `Integer` | Exchange POS terminal ID. |
| `payment_name` | `String` | Payment method. |
| `payment_method` | `String` | Payment method name. |
| `multiple_payment_infos` | `[Object]` | Multiple payment information. |
| `line_items` | `[Object]` | Exchanged items. |
| `einvoice` | `Object` | Exchange invoice. |

### Exchange Line Item

| Field | Type | Description |
| --- | --- | --- |
| `product_variant_id` | `Integer` | Variant ID. |
| `name` | `String` | Name. |
| `sku` | `String` | SKU |
| `qc` | `String` | Vendor item code. |
| `price` | `Float` | Amount. |
| `quantity` | `Integer` | Quantity. |

### Related Items

| Field | Type | Description |
| --- | --- | --- |
| `quantity` | `Integer` | Bundle quantity. |
| `items` | `[Object]` | Bundle contents. |

### Combo Item

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `product_id` | `Integer` | Product ID. |
| `product_variant_id` | `Integer` | Variant ID. |
| `title` | `String` | Product name. |
| `variant_title` | `String` | Variant name. |
| `sku` | `String` | Variant SKU. |
| `qc` | `String` | Variant vendor item code. |
| `vendor` | `String` | Vendor. |
| `price` | `Float` | Amount. |
| `cost` | `Float` | Cost. |
| `quantity` | `Integer` | Quantity. |
| `combo_product_price_difference` | `Float` | Bundle price difference. |

### Shipping Rate

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | Shipping rate ID. |
| `name` | `String` | Courier / carrier name. |
| `min_order_subtotal` | `Integer` | Spend amount. |
| `price` | `Integer` | Shipping fee. |
| `payments` | `[Object]` | Allowed payment methods. |

### Payment

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | Payment method ID. |
| `name` | `String` | Payment method name. |

### Special Collection Type

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Type name. |
| `code` | `String` | Type code. |

### Type Rules

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `quantity` | `Integer` | Rule product quantity. |
| `price` | `String` | Rule discount amount. |
| `percentage` | `Integer` | Rule discount percentage. |

## Samples

Every sample below is synthetic. The signature headers are computed with the App Secret `example-app-secret` over the exact JSON bytes shown (pretty-printed, two-space indent, trailing newline), so the samples can be used to test a verifier.

### `customers/create`

A customer registered.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: customers/create
X-Cyberbiz-Hmac-Sha256: b495fb08ac8f15fc6bfa5a37b5b36de2930b269b8e6c91bade5add3056c78ef1
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

```json
{
  "id": 1,
  "name": "王小明",
  "status": "enabled",
  "email": "customer@example.com",
  "country_calling_code": "+886",
  "mobile": "0912345678",
  "gender": "female",
  "birthday": "1990-01-01",
  "enable_cvs_pickup": true,
  "enable_cvs_cod": true,
  "enable_home_delivery_cod": true,
  "accepts_marketing": true,
  "accepts_email_notification": true,
  "tags": [],
  "address": {
    "company": "範例有限公司",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    }
  },
  "other_accumulated_consumption": 0,
  "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
  "note": "範例備註",
  "custom_fields": [],
  "bonus_remain": 10012.0,
  "uid_providers": [
    {
      "provider_type": "line",
      "uid": "U0000000000000000000000000000001"
    }
  ],
  "created_at": "2023-10-03 12:04:02",
  "updated_at": "2026-03-26 18:30:18",
  "confirmed_at": "2023-11-30 22:30:58",
  "mobile_sms_confirmed_at": "2023-11-30 22:30:58"
}
```

### `customers/update`

A customer's profile changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: customers/update
X-Cyberbiz-Hmac-Sha256: b495fb08ac8f15fc6bfa5a37b5b36de2930b269b8e6c91bade5add3056c78ef1
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

```json
{
  "id": 1,
  "name": "王小明",
  "status": "enabled",
  "email": "customer@example.com",
  "country_calling_code": "+886",
  "mobile": "0912345678",
  "gender": "female",
  "birthday": "1990-01-01",
  "enable_cvs_pickup": true,
  "enable_cvs_cod": true,
  "enable_home_delivery_cod": true,
  "accepts_marketing": true,
  "accepts_email_notification": true,
  "tags": [],
  "address": {
    "company": "範例有限公司",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    }
  },
  "other_accumulated_consumption": 0,
  "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
  "note": "範例備註",
  "custom_fields": [],
  "bonus_remain": 10012.0,
  "uid_providers": [
    {
      "provider_type": "line",
      "uid": "U0000000000000000000000000000001"
    }
  ],
  "created_at": "2023-10-03 12:04:02",
  "updated_at": "2026-03-26 18:30:18",
  "confirmed_at": "2023-11-30 22:30:58",
  "mobile_sms_confirmed_at": "2023-11-30 22:30:58"
}
```

### `uid_providers/create`

A social-login uid was linked to a customer (the reference documents no separate payload; the Customer object is posted).

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: uid_providers/create
X-Cyberbiz-Hmac-Sha256: b495fb08ac8f15fc6bfa5a37b5b36de2930b269b8e6c91bade5add3056c78ef1
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

```json
{
  "id": 1,
  "name": "王小明",
  "status": "enabled",
  "email": "customer@example.com",
  "country_calling_code": "+886",
  "mobile": "0912345678",
  "gender": "female",
  "birthday": "1990-01-01",
  "enable_cvs_pickup": true,
  "enable_cvs_cod": true,
  "enable_home_delivery_cod": true,
  "accepts_marketing": true,
  "accepts_email_notification": true,
  "tags": [],
  "address": {
    "company": "範例有限公司",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    }
  },
  "other_accumulated_consumption": 0,
  "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
  "note": "範例備註",
  "custom_fields": [],
  "bonus_remain": 10012.0,
  "uid_providers": [
    {
      "provider_type": "line",
      "uid": "U0000000000000000000000000000001"
    }
  ],
  "created_at": "2023-10-03 12:04:02",
  "updated_at": "2026-03-26 18:30:18",
  "confirmed_at": "2023-11-30 22:30:58",
  "mobile_sms_confirmed_at": "2023-11-30 22:30:58"
}
```

### `uid_providers/update`

A social-login uid of a customer changed (Customer object).

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: uid_providers/update
X-Cyberbiz-Hmac-Sha256: b495fb08ac8f15fc6bfa5a37b5b36de2930b269b8e6c91bade5add3056c78ef1
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

```json
{
  "id": 1,
  "name": "王小明",
  "status": "enabled",
  "email": "customer@example.com",
  "country_calling_code": "+886",
  "mobile": "0912345678",
  "gender": "female",
  "birthday": "1990-01-01",
  "enable_cvs_pickup": true,
  "enable_cvs_cod": true,
  "enable_home_delivery_cod": true,
  "accepts_marketing": true,
  "accepts_email_notification": true,
  "tags": [],
  "address": {
    "company": "範例有限公司",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    }
  },
  "other_accumulated_consumption": 0,
  "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
  "note": "範例備註",
  "custom_fields": [],
  "bonus_remain": 10012.0,
  "uid_providers": [
    {
      "provider_type": "line",
      "uid": "U0000000000000000000000000000001"
    }
  ],
  "created_at": "2023-10-03 12:04:02",
  "updated_at": "2026-03-26 18:30:18",
  "confirmed_at": "2023-11-30 22:30:58",
  "mobile_sms_confirmed_at": "2023-11-30 22:30:58"
}
```

### `bonus_points/create`

Bonus points were granted.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: bonus_points/create
X-Cyberbiz-Hmac-Sha256: b2fd49027de5d4320ac76697f9238130d2d287435fa1fe17cd9b72491f21da68
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

```json
{
  "id": 1,
  "title": "範例標題",
  "points": 100.0,
  "unused_points": 100.0,
  "consumption_price": 100.0,
  "deadline": "2026-09-01 10:00:00",
  "customer_id": 1,
  "source": "shop",
  "order_id": 1
}
```

### `bonus_points/update`

Bonus points were used.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: bonus_points/update
X-Cyberbiz-Hmac-Sha256: b2fd49027de5d4320ac76697f9238130d2d287435fa1fe17cd9b72491f21da68
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

```json
{
  "id": 1,
  "title": "範例標題",
  "points": 100.0,
  "unused_points": 100.0,
  "consumption_price": 100.0,
  "deadline": "2026-09-01 10:00:00",
  "customer_id": 1,
  "source": "shop",
  "order_id": 1
}
```

### `bonus_points/destroy`

A bonus point record was deleted.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: bonus_points/destroy
X-Cyberbiz-Hmac-Sha256: b2fd49027de5d4320ac76697f9238130d2d287435fa1fe17cd9b72491f21da68
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

```json
{
  "id": 1,
  "title": "範例標題",
  "points": 100.0,
  "unused_points": 100.0,
  "consumption_price": 100.0,
  "deadline": "2026-09-01 10:00:00",
  "customer_id": 1,
  "source": "shop",
  "order_id": 1
}
```

### `comment_bonus/create`

Bonus points were granted for an approved product review.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: comment_bonus/create
X-Cyberbiz-Hmac-Sha256: b2fd49027de5d4320ac76697f9238130d2d287435fa1fe17cd9b72491f21da68
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

```json
{
  "id": 1,
  "title": "範例標題",
  "points": 100.0,
  "unused_points": 100.0,
  "consumption_price": 100.0,
  "deadline": "2026-09-01 10:00:00",
  "customer_id": 1,
  "source": "shop",
  "order_id": 1
}
```

### `orders/create`

An order was placed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/create
X-Cyberbiz-Hmac-Sha256: 2e5175e7bab7b4f1ddc15d1ae85b6060f71e09b3195a57871c6326db059591d6
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `LlF157q3tPHdwV0a6FtgYPceCbMZWleHHGMm2wWVkdY=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-004",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/paid`

An order was paid.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/paid
X-Cyberbiz-Hmac-Sha256: a4e1773d94c3cbfcabdaf7218dc1133445417a8cc31f007d2ea41f19ea203cd6
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `pOF3PZTDy/yr2vchjcETNEVBeozDHwB9LqQfGeogPNY=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-005",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/preparing`

An order is being prepared for shipment.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/preparing
X-Cyberbiz-Hmac-Sha256: 9727960ac5d2089e53831cd5e8a476cddefac00f4fa170da7cebbf9f69140aa7
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `lyeWCsXSCJ5TgxzV6KR2zd76wA9PoXDafOu/n2kUCqc=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-006",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/fulfilled`

An order was shipped.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/fulfilled
X-Cyberbiz-Hmac-Sha256: 4c37e9c95815940bbdfab9651320c662cd4abdf57a011c77d828a869b35df5c7
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `TDfpyVgVlAu9+rllEyDGYs1KvfV6ARx32CioabNd9cc=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-007",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/received`

An order was received by the customer.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/received
X-Cyberbiz-Hmac-Sha256: 75b46fb55a1a8101c0313bdab4909e2e5b192534dcb19294368f741ce80293a7
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `dbRvtVoagQHAMTvatJCeLlsZJTTcsZKUNo90HOgCk6c=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-008",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/arrived`

An order arrived at the pickup store.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/arrived
X-Cyberbiz-Hmac-Sha256: a6d1f980a57a067f4f685b3d2e6d963bbf94365f00c46c545dcda0b4edf41a48
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `ptH5gKV6Bn9PaFs9Lm2WO7+UNl8AxGxUXc2gtO30Gkg=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-009",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/expired`

An order was not picked up in time.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/expired
X-Cyberbiz-Hmac-Sha256: d7f496cd76a85d02e2e660188e2cc8298d46a1e5097677f9a985755c7197fefc
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `1/SWzXaoXQLi5mAYjizIKY1GoeUJdnf5qYV1XHGX/vw=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-010",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/cancelled`

An order was cancelled.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/cancelled
X-Cyberbiz-Hmac-Sha256: 0a4ba162b06b53675bdcd19ec88e05a0811bac558c5e68f85dfe13e52adb1b72
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `CkuhYrBrU2db3NGeyI4FoIEbrFWMXmj4Xf4T5SrbG3I=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-011",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/returned`

An order was returned.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/returned
X-Cyberbiz-Hmac-Sha256: eb5c53d8015809dfa6757ab004ae5ffa8b0f247ccfd48491a184e00874a72d69
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `61xT2AFYCd+mdXqwBK5f+osPJHzP1ISRoYTgCHSnLWk=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-012",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/partial_return`

Part of an order was returned.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/partial_return
X-Cyberbiz-Hmac-Sha256: dc890ea42e1f93fae4116e1013ce0329c341be8191d75694841100e635c1d871
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `3IkOpC4fk/rkEW4QE84DKcNBvoGR11aUhBEA5jXB2HE=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-013",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/refunded`

An order was refunded.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/refunded
X-Cyberbiz-Hmac-Sha256: 6c0cdfde0cd055220cfbc9b8412fbadcf1e1581dd5b148a2486e918015531bb1
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `bAzf3gzQVSIM+8m4QS+63PHhWB3VsUiiSG6RgBVTG7E=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-014",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/partial_refunded`

Part of an order was refunded.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/partial_refunded
X-Cyberbiz-Hmac-Sha256: c3472acdbe9551ee5bcedd9791a5b15ee81893a3e1cfc0bd4d552c92f26ab51c
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `w0cqzb6VUe5bzt2XkaWxXugYk6Phz8C9TVUskvJqtRw=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-015",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/closed`

An order was closed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/closed
X-Cyberbiz-Hmac-Sha256: d29fbd199e8aca03cef0c5cae2c945a48ff7d82426b200474b98e9e55b9d37fb
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `0p+9GZ6KygPO8MXK4slFpI/32CQmsgBHS5jp5VudN/s=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-016",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/request_return`

The customer requested a return.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/request_return
X-Cyberbiz-Hmac-Sha256: 45b8b91c51cf2d0e97eed3d067d68e65ef25c3c177e6635eb019a9415e0974e0
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `Rbi5HFHPLQ6X7tPQZ9aOZe8lw8F35mNesBmpQV4JdOA=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-017",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `orders/opened`

An order was (re)opened.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/opened
X-Cyberbiz-Hmac-Sha256: 43ebb0b45750843739d533056a365c5739c7fa294aed0973e549b20df0f9e416
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `Q+uwtFdQhDc51TMFajZcVznH+ilK7Qlz5UmyDfD55BY=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-018",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `express_delivery_orders/update`

An express-delivery order changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: express_delivery_orders/update
X-Cyberbiz-Hmac-Sha256: 169ae6f8f1938ed62fceb57f2baf1af3cbd4051d0369c12e40362ae496de3ce9
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `Fprm+PGTjtYvzrV/K68a88vUBR0DacEuQDYq5JbePOk=`

```json
{
  "id": 1,
  "token": "synthetic-token-do-not-use",
  "order_number": 1001,
  "order_name": "#1001",
  "customer": {
    "id": 2,
    "name": "王小明",
    "status": "enabled",
    "email": "customer@example.com",
    "country_calling_code": "+886",
    "mobile": "0912345678",
    "gender": "女",
    "birthday": "1990-01-01",
    "enable_cvs_pickup": true,
    "enable_cvs_cod": true,
    "enable_home_delivery_cod": true,
    "accepts_marketing": true,
    "accepts_email_notification": true,
    "tags": [],
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "other_accumulated_consumption": 0,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [],
    "bonus_remain": 5000.0,
    "uid_providers": [],
    "created_at": "2026-03-16 19:51:21",
    "updated_at": "2026-09-07 12:39:06",
    "confirmed_at": "2026-03-16 19:51:21",
    "mobile_sms_confirmed_at": "2026-03-16 19:51:21"
  },
  "buyer": {
    "email": "customer@example.com",
    "mobile": "0912345678"
  },
  "receiver": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "範例",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "billing_address": {
    "name": "王小明",
    "country_calling_code": "+886",
    "phone": "0912345678",
    "address": "台北市信義區市府路1號",
    "detail_address": {
      "zip": "110",
      "country": "範例",
      "province": "",
      "city": "台北市",
      "district": "信義區",
      "address1": "市府路1號",
      "address2": "範例"
    },
    "cvs_store_id": "範例",
    "allpay_logistics_id": "TXN0000001"
  },
  "line_items": [
    {
      "id": 3,
      "product_id": 4,
      "product_variant_id": 5,
      "title": "範例商品",
      "variant_title": "",
      "sku": "SKU-019",
      "qc": "QC-001",
      "vendor": "範例",
      "price": 9999.0,
      "cost": 100.0,
      "quantity": 1,
      "item_type": "normal",
      "return_status": "no_need",
      "discount_name": "",
      "discounts": [],
      "total_price_before_discounts": 9999.0,
      "total_discount": 0,
      "total_price_after_discounts": 9999.0,
      "tax_type_id": "inclusive_tax",
      "bonus_redemption_price": 1,
      "related_items": [
        {
          "quantity": 1,
          "items": [
            {
              "id": 6,
              "product_id": 7,
              "product_variant_id": 6,
              "title": "範例商品",
              "variant_title": "範例商品 - 紅色",
              "sku": "SKU-001",
              "qc": "QC-001",
              "vendor": "範例",
              "price": 14990.0,
              "cost": 100.0,
              "quantity": 1,
              "combo_product_price_difference": 9991,
              "combo_product_price_diff_details": [
                9991
              ]
            }
          ]
        }
      ],
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "weight": 0.0,
      "photo": "//example.cyberbiz.co/media/sample-product.jpg",
      "created_at": "2026-09-07 12:35:06",
      "custom_fields": []
    }
  ],
  "shipping_type": "cyberbiz",
  "shipping_name": "門市取貨",
  "shipping_vendor": {
    "type": "custom",
    "name": "範例名稱"
  },
  "logistics_id": "TXN0000001",
  "delivery_date": "2026-09-01 10:00:00",
  "delivery_time": 0,
  "delegate": "staff@example.com",
  "fulfillments": [],
  "payment_name": "銀行轉帳",
  "payment_method": "銀行轉帳",
  "payment_url": "https://example.com/pay/1001",
  "multiple_payment_infos": [],
  "prices": {
    "total_line_items_price": 9999.0,
    "shipping_rate_price": 0.0,
    "discounts": {
      "special_collection_discount": 0,
      "vip_discount": 0.0,
      "shop_discount": {
        "name": "範例名稱",
        "amount": 100.0
      },
      "coupon_discount": {
        "id": 1,
        "name": "範例名稱",
        "code": "SAMPLE100",
        "amount": 100.0,
        "coupon_id": 1
      },
      "coupon_discounts": [],
      "bonus_consumed": 0.0,
      "vip_shipping_discount": 0,
      "coupon_shipping_discount": 0,
      "price_discount": 0,
      "third_party_discount": 0
    },
    "total_price": 9999.0
  },
  "card4no": "4242",
  "transaction_number": "TXN0000001",
  "merchant_trade_no": "S1#1001",
  "einvoice": {
    "title": "範例標題",
    "order_id": 1,
    "company_no": "12345678",
    "invoice_no": "AB12345678",
    "invoice_status": "issue",
    "invoice_at": "2026-09-01 10:00:00",
    "invalid_at": "2026-09-01 10:00:00",
    "random_num": "1234",
    "invoice_type": "default",
    "love_code": "範例",
    "phone_barcode": "範例",
    "nature_person": "範例"
  },
  "paper_invoice_no": "PA12345678",
  "statuses": {
    "order_status": "open",
    "financial_status": "pending",
    "fulfillment_status": "unshipped",
    "return_status": "no_need"
  },
  "timings": {
    "request_return_at": "2026-09-01 10:00:00",
    "return_at": "2026-09-01 10:00:00",
    "refund_at": "2026-09-01 10:00:00",
    "closed_at": "2026-09-01 10:00:00",
    "cancelled_at": "2026-09-01 10:00:00",
    "expired_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-07 12:35:05"
  },
  "return_histories": [],
  "note": "",
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "referral_code": "SAMPLE100",
  "checkout_referral_code": "SAMPLE100",
  "checkout_referral_user_name": "範例",
  "register_referral_code": "SAMPLE100",
  "total_bonus_redemption_price": 0.0,
  "pos_info": {
    "pos_user_id": 1,
    "pos_user_email": "staff@example.com",
    "pos_shop_id": 1,
    "pos_info": "",
    "pos_id": 1,
    "pos_name": "POS 1"
  },
  "exchange_histories": [],
  "linked_order_info": {
    "source": "非導購訂單",
    "shopdotcom_rid": "範例",
    "shopdotcom_click_id": "範例",
    "line_shopping_ecid": "範例",
    "line_shopping_affiliate": "範例",
    "ichannel_gid": "範例"
  },
  "tags": [],
  "express_delivery_branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "shipping_status": "範例",
  "extra_info": "範例",
  "from_device": "桌機",
  "customer_cancel_reason_detail": {
    "source": "shop",
    "reason_id": 1,
    "reason_detail": "顧客改變心意"
  },
  "serial_numbers": [],
  "order_weight": 100.0,
  "created_at": "2026-09-07 12:35:05",
  "updated_at": "2026-09-07 12:35:06",
  "subtotal_price": 9999.0,
  "paper_company_no": null,
  "warehouse_type_id": null,
  "utm_tracking": null
}
```

### `products/create`

A product was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: products/create
X-Cyberbiz-Hmac-Sha256: 9f5b2d8bb398825d54f2a9ad4bda9ab388c1337cd6eb96a9cafb72147d71d089
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `n1sti7OYgl1U8qmtS9qas4jBM3zW65apyvtyFH1x0Ik=`

```json
{
  "id": 1,
  "title": "範例商品",
  "english_title": "Sample Product",
  "product_url": "//example.cyberbiz.co/products/sample-product",
  "published": false,
  "sell_from": "2026-09-01 10:00:00",
  "sell_to": "2026-09-01 10:00:00",
  "product_type": "範例",
  "product_type_code": "範例",
  "slogan": "",
  "brief": "",
  "brief_text": "範例",
  "brief_includes_html": false,
  "body_html": "",
  "vendor": "範例",
  "price": 200.0,
  "sell_weight": 0,
  "tax_type_id": "inclusive_tax",
  "custom_collections": [],
  "special_collection": {
    "id": 1,
    "title": "範例標題",
    "handle": "範例",
    "published": true,
    "start_date": "2026-09-01 10:00:00",
    "end_date": "2026-09-01 10:00:00",
    "body_html": "<p>範例內容</p>",
    "position": 1,
    "special_collection_type": {
      "name": "範例名稱",
      "code": "SAMPLE100"
    },
    "type_rules": [
      {
        "id": 1,
        "quantity": 1,
        "price": "範例",
        "percentage": 1
      }
    ],
    "rest_include_discount": true,
    "products": []
  },
  "tags": [],
  "product_variants": [
    {
      "id": 11,
      "product_id": 10,
      "name": "範例款式",
      "position": 1,
      "price": 200.0,
      "cost": 190.0,
      "compare_at_price": 200.0,
      "meas": 0.0,
      "max_usable_bonus": 0.0,
      "weight": 0.0,
      "option1": "範例",
      "option2": "範例",
      "option3": "範例",
      "inventory_management": true,
      "inventory_quantity": 0,
      "sold": 0,
      "safety_inventory_quantity": 1,
      "inventory_policy": "deny",
      "sku": "SKU-003",
      "qc": "QC-001",
      "requires_shipping": true,
      "honeycomb_sync": true,
      "vendor": "範例",
      "photo_urls": [
        "範例"
      ],
      "pim_infos": [
        {
          "id": 1,
          "product_id": 1,
          "product_variant_id": 1,
          "pim_product_id": 1,
          "pim_variant_id": 1,
          "channel": {
            "id": 1,
            "name": "範例名稱"
          },
          "channel_shop_name": "範例通路商店",
          "is_connected": true,
          "is_source": true,
          "shop_id": 1
        }
      ],
      "created_at": "2026-07-10 20:22:05",
      "updated_at": "2026-07-10 20:22:07"
    }
  ],
  "product_options": [],
  "pos_shop": {},
  "photo_urls": [
    "範例"
  ],
  "photos": [],
  "channel": {
    "id": 1,
    "name": "範例名稱"
  },
  "related_collections": [],
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "temperature_types": [
    "常溫"
  ],
  "searchable": true,
  "google_product_category_id": 1,
  "product_custom_fields": [],
  "seo_meta_tags": {
    "title": null,
    "description": null,
    "keywords": null
  },
  "required_customer_tags": [],
  "created_at": "2026-07-10 20:22:05",
  "updated_at": "2026-07-29 09:08:40"
}
```

### `products/update`

A product changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: products/update
X-Cyberbiz-Hmac-Sha256: 9f5b2d8bb398825d54f2a9ad4bda9ab388c1337cd6eb96a9cafb72147d71d089
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `n1sti7OYgl1U8qmtS9qas4jBM3zW65apyvtyFH1x0Ik=`

```json
{
  "id": 1,
  "title": "範例商品",
  "english_title": "Sample Product",
  "product_url": "//example.cyberbiz.co/products/sample-product",
  "published": false,
  "sell_from": "2026-09-01 10:00:00",
  "sell_to": "2026-09-01 10:00:00",
  "product_type": "範例",
  "product_type_code": "範例",
  "slogan": "",
  "brief": "",
  "brief_text": "範例",
  "brief_includes_html": false,
  "body_html": "",
  "vendor": "範例",
  "price": 200.0,
  "sell_weight": 0,
  "tax_type_id": "inclusive_tax",
  "custom_collections": [],
  "special_collection": {
    "id": 1,
    "title": "範例標題",
    "handle": "範例",
    "published": true,
    "start_date": "2026-09-01 10:00:00",
    "end_date": "2026-09-01 10:00:00",
    "body_html": "<p>範例內容</p>",
    "position": 1,
    "special_collection_type": {
      "name": "範例名稱",
      "code": "SAMPLE100"
    },
    "type_rules": [
      {
        "id": 1,
        "quantity": 1,
        "price": "範例",
        "percentage": 1
      }
    ],
    "rest_include_discount": true,
    "products": []
  },
  "tags": [],
  "product_variants": [
    {
      "id": 11,
      "product_id": 10,
      "name": "範例款式",
      "position": 1,
      "price": 200.0,
      "cost": 190.0,
      "compare_at_price": 200.0,
      "meas": 0.0,
      "max_usable_bonus": 0.0,
      "weight": 0.0,
      "option1": "範例",
      "option2": "範例",
      "option3": "範例",
      "inventory_management": true,
      "inventory_quantity": 0,
      "sold": 0,
      "safety_inventory_quantity": 1,
      "inventory_policy": "deny",
      "sku": "SKU-003",
      "qc": "QC-001",
      "requires_shipping": true,
      "honeycomb_sync": true,
      "vendor": "範例",
      "photo_urls": [
        "範例"
      ],
      "pim_infos": [
        {
          "id": 1,
          "product_id": 1,
          "product_variant_id": 1,
          "pim_product_id": 1,
          "pim_variant_id": 1,
          "channel": {
            "id": 1,
            "name": "範例名稱"
          },
          "channel_shop_name": "範例通路商店",
          "is_connected": true,
          "is_source": true,
          "shop_id": 1
        }
      ],
      "created_at": "2026-07-10 20:22:05",
      "updated_at": "2026-07-10 20:22:07"
    }
  ],
  "product_options": [],
  "pos_shop": {},
  "photo_urls": [
    "範例"
  ],
  "photos": [],
  "channel": {
    "id": 1,
    "name": "範例名稱"
  },
  "related_collections": [],
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "temperature_types": [
    "常溫"
  ],
  "searchable": true,
  "google_product_category_id": 1,
  "product_custom_fields": [],
  "seo_meta_tags": {
    "title": null,
    "description": null,
    "keywords": null
  },
  "required_customer_tags": [],
  "created_at": "2026-07-10 20:22:05",
  "updated_at": "2026-07-29 09:08:40"
}
```

### `products/delete`

A product was deleted.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: products/delete
X-Cyberbiz-Hmac-Sha256: 9f5b2d8bb398825d54f2a9ad4bda9ab388c1337cd6eb96a9cafb72147d71d089
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `n1sti7OYgl1U8qmtS9qas4jBM3zW65apyvtyFH1x0Ik=`

```json
{
  "id": 1,
  "title": "範例商品",
  "english_title": "Sample Product",
  "product_url": "//example.cyberbiz.co/products/sample-product",
  "published": false,
  "sell_from": "2026-09-01 10:00:00",
  "sell_to": "2026-09-01 10:00:00",
  "product_type": "範例",
  "product_type_code": "範例",
  "slogan": "",
  "brief": "",
  "brief_text": "範例",
  "brief_includes_html": false,
  "body_html": "",
  "vendor": "範例",
  "price": 200.0,
  "sell_weight": 0,
  "tax_type_id": "inclusive_tax",
  "custom_collections": [],
  "special_collection": {
    "id": 1,
    "title": "範例標題",
    "handle": "範例",
    "published": true,
    "start_date": "2026-09-01 10:00:00",
    "end_date": "2026-09-01 10:00:00",
    "body_html": "<p>範例內容</p>",
    "position": 1,
    "special_collection_type": {
      "name": "範例名稱",
      "code": "SAMPLE100"
    },
    "type_rules": [
      {
        "id": 1,
        "quantity": 1,
        "price": "範例",
        "percentage": 1
      }
    ],
    "rest_include_discount": true,
    "products": []
  },
  "tags": [],
  "product_variants": [
    {
      "id": 11,
      "product_id": 10,
      "name": "範例款式",
      "position": 1,
      "price": 200.0,
      "cost": 190.0,
      "compare_at_price": 200.0,
      "meas": 0.0,
      "max_usable_bonus": 0.0,
      "weight": 0.0,
      "option1": "範例",
      "option2": "範例",
      "option3": "範例",
      "inventory_management": true,
      "inventory_quantity": 0,
      "sold": 0,
      "safety_inventory_quantity": 1,
      "inventory_policy": "deny",
      "sku": "SKU-003",
      "qc": "QC-001",
      "requires_shipping": true,
      "honeycomb_sync": true,
      "vendor": "範例",
      "photo_urls": [
        "範例"
      ],
      "pim_infos": [
        {
          "id": 1,
          "product_id": 1,
          "product_variant_id": 1,
          "pim_product_id": 1,
          "pim_variant_id": 1,
          "channel": {
            "id": 1,
            "name": "範例名稱"
          },
          "channel_shop_name": "範例通路商店",
          "is_connected": true,
          "is_source": true,
          "shop_id": 1
        }
      ],
      "created_at": "2026-07-10 20:22:05",
      "updated_at": "2026-07-10 20:22:07"
    }
  ],
  "product_options": [],
  "pos_shop": {},
  "photo_urls": [
    "範例"
  ],
  "photos": [],
  "channel": {
    "id": 1,
    "name": "範例名稱"
  },
  "related_collections": [],
  "branch_store": {
    "store_no": "S001",
    "name": "範例名稱",
    "phone": "0912345678",
    "county": "台北市",
    "district": "信義區",
    "address": {
      "company": "範例有限公司",
      "country_calling_code": "+886",
      "phone": "0912345678",
      "address": "台北市信義區市府路1號",
      "detail_address": {
        "zip": "110",
        "country": "範例",
        "province": "範例",
        "city": "台北市",
        "district": "信義區",
        "address1": "市府路1號",
        "address2": "範例"
      }
    },
    "zip": "110",
    "opening_hours": "09:00-21:00",
    "lat": 100.0,
    "lng": 100.0,
    "enabled": true,
    "shipping_rates": [
      {
        "id": 1,
        "name": "範例名稱",
        "min_order_subtotal": 1,
        "price": 1,
        "payments": [
          {
            "id": 1,
            "name": "範例名稱"
          }
        ]
      }
    ],
    "source_type": "BranchStore",
    "source_id": 1
  },
  "temperature_types": [
    "常溫"
  ],
  "searchable": true,
  "google_product_category_id": 1,
  "product_custom_fields": [],
  "seo_meta_tags": {
    "title": null,
    "description": null,
    "keywords": null
  },
  "required_customer_tags": [],
  "created_at": "2026-07-10 20:22:05",
  "updated_at": "2026-07-29 09:08:40"
}
```

### `variants/create`

A product variant was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: variants/create
X-Cyberbiz-Hmac-Sha256: ec978347d3190504f5ffbea579deeb214fed21e3ce5f2a5e7b978f87d605757f
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `7JeDR9MZBQT1/76led7rIU/tIePOXypee5ePh9YFdX8=`

```json
{
  "id": 11,
  "product_id": 10,
  "name": "範例款式",
  "position": 1,
  "price": 200.0,
  "cost": 190.0,
  "compare_at_price": 200.0,
  "meas": 0.0,
  "max_usable_bonus": 0.0,
  "weight": 0.0,
  "option1": "範例",
  "option2": "範例",
  "option3": "範例",
  "inventory_management": true,
  "inventory_quantity": 0,
  "sold": 0,
  "safety_inventory_quantity": 1,
  "inventory_policy": "deny",
  "sku": "SKU-003",
  "qc": "QC-001",
  "requires_shipping": true,
  "honeycomb_sync": true,
  "vendor": "範例",
  "photo_urls": [
    "範例"
  ],
  "pim_infos": [
    {
      "id": 1,
      "product_id": 1,
      "product_variant_id": 1,
      "pim_product_id": 1,
      "pim_variant_id": 1,
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "channel_shop_name": "範例通路商店",
      "is_connected": true,
      "is_source": true,
      "shop_id": 1
    }
  ],
  "created_at": "2026-07-10 20:22:05",
  "updated_at": "2026-07-10 20:22:07"
}
```

### `variants/update`

A product variant changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: variants/update
X-Cyberbiz-Hmac-Sha256: ec978347d3190504f5ffbea579deeb214fed21e3ce5f2a5e7b978f87d605757f
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `7JeDR9MZBQT1/76led7rIU/tIePOXypee5ePh9YFdX8=`

```json
{
  "id": 11,
  "product_id": 10,
  "name": "範例款式",
  "position": 1,
  "price": 200.0,
  "cost": 190.0,
  "compare_at_price": 200.0,
  "meas": 0.0,
  "max_usable_bonus": 0.0,
  "weight": 0.0,
  "option1": "範例",
  "option2": "範例",
  "option3": "範例",
  "inventory_management": true,
  "inventory_quantity": 0,
  "sold": 0,
  "safety_inventory_quantity": 1,
  "inventory_policy": "deny",
  "sku": "SKU-003",
  "qc": "QC-001",
  "requires_shipping": true,
  "honeycomb_sync": true,
  "vendor": "範例",
  "photo_urls": [
    "範例"
  ],
  "pim_infos": [
    {
      "id": 1,
      "product_id": 1,
      "product_variant_id": 1,
      "pim_product_id": 1,
      "pim_variant_id": 1,
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "channel_shop_name": "範例通路商店",
      "is_connected": true,
      "is_source": true,
      "shop_id": 1
    }
  ],
  "created_at": "2026-07-10 20:22:05",
  "updated_at": "2026-07-10 20:22:07"
}
```

### `variants/delete`

A product variant was deleted.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: variants/delete
X-Cyberbiz-Hmac-Sha256: ec978347d3190504f5ffbea579deeb214fed21e3ce5f2a5e7b978f87d605757f
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `7JeDR9MZBQT1/76led7rIU/tIePOXypee5ePh9YFdX8=`

```json
{
  "id": 11,
  "product_id": 10,
  "name": "範例款式",
  "position": 1,
  "price": 200.0,
  "cost": 190.0,
  "compare_at_price": 200.0,
  "meas": 0.0,
  "max_usable_bonus": 0.0,
  "weight": 0.0,
  "option1": "範例",
  "option2": "範例",
  "option3": "範例",
  "inventory_management": true,
  "inventory_quantity": 0,
  "sold": 0,
  "safety_inventory_quantity": 1,
  "inventory_policy": "deny",
  "sku": "SKU-003",
  "qc": "QC-001",
  "requires_shipping": true,
  "honeycomb_sync": true,
  "vendor": "範例",
  "photo_urls": [
    "範例"
  ],
  "pim_infos": [
    {
      "id": 1,
      "product_id": 1,
      "product_variant_id": 1,
      "pim_product_id": 1,
      "pim_variant_id": 1,
      "channel": {
        "id": 1,
        "name": "範例名稱"
      },
      "channel_shop_name": "範例通路商店",
      "is_connected": true,
      "is_source": true,
      "shop_id": 1
    }
  ],
  "created_at": "2026-07-10 20:22:05",
  "updated_at": "2026-07-10 20:22:07"
}
```

### `coupons/create`

A coupon was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: coupons/create
X-Cyberbiz-Hmac-Sha256: 07101b016314861db9785e07dcbcb7e888e3afecf6105bac82a765b246897890
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `BxAbAWMUhh25eF4H3Ly36Ijjr+z2EFusgqdlskaJeJA=`

```json
{
  "customer_id": 1,
  "title": "範例標題",
  "code": "SAMPLE100",
  "coupon_type_name": "金額",
  "coupon_value": "100",
  "order_price_threshold": 1,
  "start_date": "2026-09-01 10:00:00",
  "end_date": "2026-09-01 10:00:00",
  "concurrently_apply": "true",
  "usage_limit": 1,
  "can_accumulate_bonus": true,
  "usage_unlimited": true,
  "used_times": 1,
  "gift_order_id": 1,
  "gift_days": 1,
  "account_usage_limit_enabled": true,
  "account_usage_limit": 1,
  "restrict_strategy": "unrestricted",
  "restrict_campaigns": [
    "shop_discount"
  ],
  "tags": [
    "VIP"
  ],
  "pos_shop_ids": [
    1
  ],
  "coupon_status": "no_start_use",
  "gift_order_status": "closed",
  "valid": true,
  "customer_used_times": 1,
  "customer_usable": true
}
```

### `coupons/update`

A coupon was used.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: coupons/update
X-Cyberbiz-Hmac-Sha256: 07101b016314861db9785e07dcbcb7e888e3afecf6105bac82a765b246897890
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `BxAbAWMUhh25eF4H3Ly36Ijjr+z2EFusgqdlskaJeJA=`

```json
{
  "customer_id": 1,
  "title": "範例標題",
  "code": "SAMPLE100",
  "coupon_type_name": "金額",
  "coupon_value": "100",
  "order_price_threshold": 1,
  "start_date": "2026-09-01 10:00:00",
  "end_date": "2026-09-01 10:00:00",
  "concurrently_apply": "true",
  "usage_limit": 1,
  "can_accumulate_bonus": true,
  "usage_unlimited": true,
  "used_times": 1,
  "gift_order_id": 1,
  "gift_days": 1,
  "account_usage_limit_enabled": true,
  "account_usage_limit": 1,
  "restrict_strategy": "unrestricted",
  "restrict_campaigns": [
    "shop_discount"
  ],
  "tags": [
    "VIP"
  ],
  "pos_shop_ids": [
    1
  ],
  "coupon_status": "no_start_use",
  "gift_order_status": "closed",
  "valid": true,
  "customer_used_times": 1,
  "customer_usable": true
}
```

### `coupons/destroy`

A coupon was deleted.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: coupons/destroy
X-Cyberbiz-Hmac-Sha256: 07101b016314861db9785e07dcbcb7e888e3afecf6105bac82a765b246897890
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `BxAbAWMUhh25eF4H3Ly36Ijjr+z2EFusgqdlskaJeJA=`

```json
{
  "customer_id": 1,
  "title": "範例標題",
  "code": "SAMPLE100",
  "coupon_type_name": "金額",
  "coupon_value": "100",
  "order_price_threshold": 1,
  "start_date": "2026-09-01 10:00:00",
  "end_date": "2026-09-01 10:00:00",
  "concurrently_apply": "true",
  "usage_limit": 1,
  "can_accumulate_bonus": true,
  "usage_unlimited": true,
  "used_times": 1,
  "gift_order_id": 1,
  "gift_days": 1,
  "account_usage_limit_enabled": true,
  "account_usage_limit": 1,
  "restrict_strategy": "unrestricted",
  "restrict_campaigns": [
    "shop_discount"
  ],
  "tags": [
    "VIP"
  ],
  "pos_shop_ids": [
    1
  ],
  "coupon_status": "no_start_use",
  "gift_order_status": "closed",
  "valid": true,
  "customer_used_times": 1,
  "customer_usable": true
}
```

### `customer_vip_level/update`

A customer's VIP level changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: customer_vip_level/update
X-Cyberbiz-Hmac-Sha256: 8f02e99ec79971cc4abf37f68982b9f322ee0c65c6debe7ed9c3a32efa328d78
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `jwLpnseZccxKvzf2iYK58yLuDGXG3r5+2cOjLvoyjXg=`

```json
{
  "customer_id": 1,
  "current_group": {
    "id": 1,
    "name": "範例名稱",
    "position": 1,
    "description_url": "https://example.cyberbiz.co/pages/vip",
    "customer_tags": [
      "VIP"
    ],
    "vip_group_levels": [
      {
        "id": 1,
        "position": 1,
        "name": "範例名稱",
        "validity_days": 1,
        "upgrade_condition_total_spent": 1,
        "upgrade_condition_total_spent_in_validity_days": 1,
        "renewal_condition_total_spent": 1,
        "renewal_condition_total_spent_in_validity_days": 1,
        "bonus_point_enabled": true,
        "bonus_point_threshold": 1,
        "bonus_point_value": 1,
        "bonus_point_expiry_days": 1,
        "birth_gift_enabled": true,
        "birth_gift_name": "範例",
        "birth_gift_before_days": "7",
        "birth_gift_setting": {
          "bonus": {
            "enabled": true,
            "value": 1,
            "expiry_days": 1
          },
          "coupon": {
            "enabled": true,
            "presets": {
              "coupon_type_id": 1,
              "value": 1,
              "code": "SAMPLE100",
              "order_price_threshold": 1,
              "usable_days": 1,
              "usage_limit": 1,
              "product_tags": [
                "VIP"
              ],
              "restrict_strategy": "unrestricted",
              "restrict_campaigns": [
                "shop_discount"
              ]
            }
          }
        },
        "upgrade_gift_enabled": true,
        "upgrade_gift_setting": {
          "bonus": {
            "enabled": true,
            "value": 1,
            "expiry_days": 1
          },
          "coupon": {
            "enabled": true,
            "presets": {
              "coupon_type_id": 1,
              "value": 1,
              "code": "SAMPLE100",
              "order_price_threshold": 1,
              "usable_days": 1,
              "usage_limit": 1,
              "product_tags": [
                "VIP"
              ],
              "restrict_strategy": "unrestricted",
              "restrict_campaigns": [
                "shop_discount"
              ]
            }
          }
        },
        "order_discount_enabled": true,
        "order_discount_value": 1,
        "order_discount_setting": {},
        "free_shipping_enabled": true,
        "free_shipping_setting": {}
      }
    ]
  },
  "current_level": {
    "id": 1,
    "position": 1,
    "name": "範例名稱",
    "validity_days": 1,
    "upgrade_condition_total_spent": 1,
    "upgrade_condition_total_spent_in_validity_days": 1,
    "renewal_condition_total_spent": 1,
    "renewal_condition_total_spent_in_validity_days": 1,
    "bonus_point_enabled": true,
    "bonus_point_threshold": 1,
    "bonus_point_value": 1,
    "bonus_point_expiry_days": 1,
    "birth_gift_enabled": true,
    "birth_gift_name": "範例",
    "birth_gift_before_days": "7",
    "birth_gift_setting": {
      "bonus": {
        "enabled": true,
        "value": 1,
        "expiry_days": 1
      },
      "coupon": {
        "enabled": true,
        "presets": {
          "coupon_type_id": 1,
          "value": 1,
          "code": "SAMPLE100",
          "order_price_threshold": 1,
          "usable_days": 1,
          "usage_limit": 1,
          "product_tags": [
            "VIP"
          ],
          "restrict_strategy": "unrestricted",
          "restrict_campaigns": [
            "shop_discount"
          ]
        }
      }
    },
    "upgrade_gift_enabled": true,
    "upgrade_gift_setting": {
      "bonus": {
        "enabled": true,
        "value": 1,
        "expiry_days": 1
      },
      "coupon": {
        "enabled": true,
        "presets": {
          "coupon_type_id": 1,
          "value": 1,
          "code": "SAMPLE100",
          "order_price_threshold": 1,
          "usable_days": 1,
          "usage_limit": 1,
          "product_tags": [
            "VIP"
          ],
          "restrict_strategy": "unrestricted",
          "restrict_campaigns": [
            "shop_discount"
          ]
        }
      }
    },
    "order_discount_enabled": true,
    "order_discount_value": 1,
    "order_discount_setting": {},
    "free_shipping_enabled": true,
    "free_shipping_setting": {}
  },
  "next_level": {
    "id": 1,
    "position": 1,
    "name": "範例名稱",
    "validity_days": 1,
    "upgrade_condition_total_spent": 1,
    "upgrade_condition_total_spent_in_validity_days": 1,
    "renewal_condition_total_spent": 1,
    "renewal_condition_total_spent_in_validity_days": 1,
    "bonus_point_enabled": true,
    "bonus_point_threshold": 1,
    "bonus_point_value": 1,
    "bonus_point_expiry_days": 1,
    "birth_gift_enabled": true,
    "birth_gift_name": "範例",
    "birth_gift_before_days": "7",
    "birth_gift_setting": {
      "bonus": {
        "enabled": true,
        "value": 1,
        "expiry_days": 1
      },
      "coupon": {
        "enabled": true,
        "presets": {
          "coupon_type_id": 1,
          "value": 1,
          "code": "SAMPLE100",
          "order_price_threshold": 1,
          "usable_days": 1,
          "usage_limit": 1,
          "product_tags": [
            "VIP"
          ],
          "restrict_strategy": "unrestricted",
          "restrict_campaigns": [
            "shop_discount"
          ]
        }
      }
    },
    "upgrade_gift_enabled": true,
    "upgrade_gift_setting": {
      "bonus": {
        "enabled": true,
        "value": 1,
        "expiry_days": 1
      },
      "coupon": {
        "enabled": true,
        "presets": {
          "coupon_type_id": 1,
          "value": 1,
          "code": "SAMPLE100",
          "order_price_threshold": 1,
          "usable_days": 1,
          "usage_limit": 1,
          "product_tags": [
            "VIP"
          ],
          "restrict_strategy": "unrestricted",
          "restrict_campaigns": [
            "shop_discount"
          ]
        }
      }
    },
    "order_discount_enabled": true,
    "order_discount_value": 1,
    "order_discount_setting": {},
    "free_shipping_enabled": true,
    "free_shipping_setting": {}
  },
  "extra_info": {
    "start_at": "2026-09-01 10:00:00",
    "end_at": "2026-09-01 10:00:00",
    "difference_of_total_spent_in_validity_days_for_renewal": 1,
    "difference_of_total_spent_for_renewal": 1,
    "difference_of_total_spent_in_validity_days_for_upgrade": 1,
    "difference_of_total_spent_for_upgrade": 1
  }
}
```

### `apps/uninstall`

The app was uninstalled from the shop.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: apps/uninstall
X-Cyberbiz-Hmac-Sha256: 543354df3490587e091b9d9f16b4b8812b7ca95932d09faac86926f6f333ba45
X-Cyberbiz-Domain-Hmac-Sha256: 97694009b6e4ade734f9f3303bfa0f544077f2cb5630856d10abc88d3008d959
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `VDNU3zSQWH4JG52fFrS4gSt8qVky0J+qyGkm9vMzukU=`

```json
{
  "app_uuid": "00000000-0000-4000-8000-000000000001",
  "app_version_uuid": "00000000-0000-4000-8000-000000000002",
  "app_client_id": "example-app-client-id"
}
```

## Release Notes

### 1.0.1 (2026-10-02)

- Postman collection descriptions no longer point to repository-internal files.

### 1.0.0 (2026-09-08)

- Initial generation from the CYBERBIZ v1 swagger, the v1/v2 Postman collections, the Notion v2 reference and the webhook reference.
- Observed corrections applied: nullable fields, money as float, order_number as integer, arrays documented as objects, fields missing from the swagger, Bearer authentication.
- Synthetic samples derived from the Golden Files for every operation and event.
