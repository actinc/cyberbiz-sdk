# CYBERBIZ Webhooks

Version: 2.1.0

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
| `X-Cyberbiz-Domain-Hmac-Sha256` | `9b1c…` (64 hex chars) | The Domain Signature: lowercase hex HMAC-SHA256, keyed by the App Secret, of the standard base64 encoding of the `X-Cyberbiz-Domain` value (not of the domain itself). |

## Verifying the signature

1. Read the raw body bytes before parsing them.
2. Compute `HMAC-SHA256(app_secret, body)`.
3. Compare, in constant time, the **hex** encoding of the digest with `X-Cyberbiz-Hmac-Sha256`. Production App webhooks send hex (64 characters), although the CYBERBIZ documentation describes base64 — accept base64 as a fallback so a future change on the platform does not break verification.
4. Optionally check `X-Cyberbiz-Domain-Hmac-Sha256` against the base64 of `X-Cyberbiz-Domain` (hex digest); it binds the request to one shop but does not authenticate the payload.
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
| `orders/update` | An order changed. | [Order](#order) |
| `products/update_options` | A product's options changed. | [Product](#product) |
| `products/update_photos` | A product's photos changed. | [Product](#product) |
| `variants/out_of_stock` | A product variant sold out. | [Product Variant](#product-variant) |
| `cart/add` | A customer added items to the cart. | [Cart](#cart) |
| `customer_tags/update` | A customer's tags changed. | [Customer Tags](#customer-tags) |
| `affiliate_vendor_orders/create` | An affiliate order was created. | [Affiliate Vendor Order](#affiliate-vendor-order) |
| `affiliate_vendor_orders/cancelled` | An affiliate order was cancelled. | [Affiliate Vendor Order](#affiliate-vendor-order) |
| `affiliate_vendor_orders/closed` | An affiliate order was closed. | [Affiliate Vendor Order](#affiliate-vendor-order) |
| `collections/create` | A product collection was created. | [Collection](#collection) |
| `collections/update` | A product collection changed. | [Collection](#collection) |
| `collections/delete` | A product collection was deleted. | [Collection](#collection) |
| `menus/create` | A storefront menu was created. | [Menu](#menu) |
| `menus/update` | A storefront menu changed. | [Menu Update](#menu-update) |
| `menus/destroy` | A storefront menu was deleted. | [Menu Destroy](#menu-destroy) |
| `domains/create` | A shop domain was added. | [Domain](#domain) |
| `domains/update` | A shop domain changed. | [Domain](#domain) |
| `domains/destroy` | A shop domain was removed. | [Domain](#domain) |
| `bulk_operations/finish` | A bulk operation finished. | [Bulk Operation](#bulk-operation) |
| `line_chat_bot/update` | The shop's LINE Messaging API settings changed. | [LINE Chat Bot](#line-chat-bot) |
| `fbe_ad_accounts/reclaim` | A Facebook ad account was reclaimed. | [FBE Ad Account](#fbe-ad-account) |
| `pos_wallet_transactions/topup` | A POS wallet was topped up. | [POS Wallet Topup](#pos-wallet-topup) |
| `pos_wallet_transactions/refund` | A POS wallet top-up was refunded. | [POS Wallet Refund](#pos-wallet-refund) |
| `stock_invoices/created` | A stock invoice (outbound) was created. | [Stock Invoice](#stock-invoice) |
| `stock_invoices/done` | A stock invoice was shipped out. | [Stock Invoice](#stock-invoice) |
| `stock_invoices/cancelled` | A stock invoice was cancelled. | [Stock Invoice](#stock-invoice) |
| `stock_receipts/created` | A stock receipt (inbound) was created. | [Stock Receipt](#stock-receipt) |
| `stock_receipts/done` | A stock receipt was put into stock. | [Stock Receipt](#stock-receipt) |
| `stock_receipts/cancelled` | A stock receipt was cancelled. | [Stock Receipt](#stock-receipt) |
| `stock_requisitions/created` | A stock requisition (transfer) was created. | [Stock Requisition](#stock-requisition) |
| `stock_requisitions/approved` | A stock requisition was approved. | [Stock Requisition](#stock-requisition) |
| `stock_requisitions/cancelled` | A stock requisition was cancelled. | [Stock Requisition](#stock-requisition) |
| `stored_orders/retrieved` | Stored goods were picked up. | [Stored Order Retrieved](#stored-order-retrieved) |
| `stored_orders/refunded` | A stored order was refunded. | [Stored Order Refunded](#stored-order-refunded) |
| `send_edm_finish` | An EDM send job finished (Merchant webhooks only). | [EDM Job](#edm-job) |
| `send_sms_finish` | An SMS send job finished (Merchant webhooks only). | [SMS Job](#sms-job) |
| `send_line_oa_message_finish` | A LINE OA message send job finished (Merchant webhooks only). | [LINE OA Message Job](#line-oa-message-job) |
| `get_customer_ids_finish` | A customer-group member lookup finished (Merchant webhooks only). | [Customer IDs Job](#customer-ids-job) |

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

### Customer Tags

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | Customer ID. |
| `email` | `String` | Email |
| `tags` | `[Object]` | Customer tags. |

### Customer Tags tags

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | 標籤ID |
| `name` | `String` | Tag name. |

### Affiliate Vendor Order

| Field | Type | Description |
| --- | --- | --- |
| `status_code` | `String` | 狀態代碼 |
| `uid` | `String` | 分潤 uid |
| `cid` | `String` | 分潤 cid |
| `vendor_name` | `String` | 分潤訂單廠商 |
| `order` | `Object` | Order information. |

### Affiliate Vendor Order order

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | Order ID. |
| `subtotal_price` | `Float` | Subtotal (excluding shipping). |
| `created_at` | `String` | Created at. |
| `updated_at` | `String` | Updated at. |
| `order_number` | `String` | Order number. |
| `order_name` | `String` | Order name. |
| `buyer` | `Object` | Buyer information. |
| `billing_address` | `[Object]` | Billing information (custom feature; contact your consultant). |
| `line_items` | `[Object]` | Order line items. |
| `shipping_type` | `String` | Shipping method. |
| `shipping_name` | `String` | Shipping method name. |
| `delivery_date` | `String` | Requested delivery date. |
| `delivery_time` | `Integer` | Requested delivery time. |
| `delegate` | `String` | Staff who placed the order on the customer's behalf. |
| `prices` | `[Object]` | Price information. |
| `transaction_number` | `String` | Third-party transaction number. |
| `merchant_trade_no` | `String` | Merchant trade number. |
| `statuses` | `[Object]` | Status information. |
| `timings` | `[Object]` | Timestamps. |
| `return_histories` | `[Object]` | Refund information. |
| `note` | `String` | Order note. |
| `total_bonus_redemption_price` | `Integer` | Total bonus points redeemed in the bonus mall. |
| `linked_order_info` | `Object` | Linked-order (affiliate) information. |
| `shipping_status` | `String` | Shipping status. |
| `extra_info` | `String` | Extra information. |
| `from_device` | `String` | Order source. |
| `customer_cancel_reason_detail` | `Object` | customer cancel reason detail |
| `serial_numbers` | `[String]` | Campaign serial numbers. |
| `order_weight` | `Float` | Total weight. |
| `warehouse_type_id` | `Integer` | Warehouse type ID. |

### Affiliate Vendor Order order buyer

| Field | Type | Description |
| --- | --- | --- |
| `email` | `String` | Buyer e-mail. |
| `mobile` | `String` | Buyer mobile number. |

### Affiliate Vendor Order order billing_address

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Billing name. |
| `country_calling_code` | `String` | Billing phone country calling code. |
| `phone` | `String` | Billing phone. |
| `address` | `String` | Billing address. |
| `detail_address` | `[Object]` | Billing address details. |

### Affiliate Vendor Order order billing_address detail_address

| Field | Type | Description |
| --- | --- | --- |
| `zip` | `String` | Area code. |
| `country` | `String` | Country. |
| `province` | `String` | Province. |
| `city` | `String` | City. |
| `district` | `String` | District. |
| `address1` | `String` | Address line 1. |
| `address2` | `String` | Address line 2. |

### Affiliate Vendor Order order line_items

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
| `tax_type_id` | `Integer` | 商品課稅類別(inclusive_tax: 應稅; zero_tax: 零稅率; exclusive_tax: 免稅) |
| `bonus_redemption_price` | `Integer` | Bonus points redeemed in the bonus mall; null when not a bonus redemption. |
| `related_items` | `[Object]` | Bundle component items. |
| `created_at` | `Date` | Created at. |
| `channel` | `String` | Product channel. |
| `weight` | `String` | Product weight. |
| `photo` | `String` | Product photo. |
| `custom_fields` | `[Object]` | Product custom fields (custom feature). |

### Affiliate Vendor Order order line_items discounts

| Field | Type | Description |
| --- | --- | --- |
| `position` | `Integer` | Item position (nth item). |
| `id` | `Integer` | Discount type ID. |
| `code` | `String` | Discount type code. |
| `name` | `String` | Discount type name. |
| `discount` | `Integer` | Discount amount. |

### Affiliate Vendor Order order line_items related_items

| Field | Type | Description |
| --- | --- | --- |
| `quantity` | `Integer` | Bundle quantity. |
| `items` | `[Object]` | Bundle contents. |

### Affiliate Vendor Order order line_items related_items items

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
| `combo_product_price_difference` | `Float` | Bundle price difference (Enterprise plan only; contact your consultant). |
| `combo_product_price_diff_details` | `[Integer]` | Bundle price difference details (Enterprise plan only; contact your consultant). |

### Affiliate Vendor Order order line_items custom_fields

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Custom field name. |
| `label` | `String` | Custom field label. |
| `value` | `String` | Custom field value. |

### Affiliate Vendor Order order prices

| Field | Type | Description |
| --- | --- | --- |
| `total_line_items_price` | `Float` | Products total. |
| `shipping_rate_price` | `Float` | Shipping fee. |
| `discounts` | `[Object]` | Discount breakdown. |
| `total_price` | `Float` | Order total. |

### Affiliate Vendor Order order prices discounts

| Field | Type | Description |
| --- | --- | --- |
| `special_collection_discount` | `Integer` | Special collection (campaign) discount. |
| `vip_discount` | `Float` | VIP discount. |
| `shop_discount` | `Object` | Shop-wide discount details. |
| `coupon_discount` | `Object` | Coupon detail (single; deprecated, will be removed soon). |
| `coupon_discounts` | `[Object]` | Coupon details. |
| `bonus_consumed` | `Float` | Bonus discount. |
| `vip_shipping_discount` | `Float` | VIP shipping discount. |
| `coupon_shipping_discount` | `Float` | Free-shipping coupon discount. |
| `price_discount` | `Integer` | Manager (manual) discount. |
| `third_party_discount` | `Interger` | third party discount |

### Affiliate Vendor Order order prices discounts shop_discount

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Shop-wide campaign name. |
| `amount` | `Float` | Shop-wide campaign discount. |

### Affiliate Vendor Order order prices discounts coupon_discount

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Coupon name. |
| `code` | `String` | Coupon code. |
| `amount` | `Float` | Coupon discount. |
| `coupon_id` | `Integer` | Coupon ID. |

### Affiliate Vendor Order order prices discounts coupon_discounts

| Field | Type | Description |
| --- | --- | --- |
| `name` | `String` | Coupon name. |
| `code` | `String` | Coupon code. |
| `amount` | `Float` | Coupon discount. |
| `coupon_id` | `Integer` | Coupon ID. |
| `id` | `Integer` | ID |

### Affiliate Vendor Order order statuses

| Field | Type | Description |
| --- | --- | --- |
| `order_status` | `String` | Order status. |
| `financial_status` | `String` | Payment status. |
| `fulfillment_status` | `String` | Shipping status. |
| `return_status` | `String` | Return status. |

### Affiliate Vendor Order order timings

| Field | Type | Description |
| --- | --- | --- |
| `request_return_at` | `String` | Return requested at. |
| `return_at` | `String` | Returned at. |
| `refund_at` | `String` | Refunded at. |
| `closed_at` | `String` | Closed at. |
| `cancelled_at` | `String` | Cancelled at. |
| `expired_at` | `String` | Pickup expired at. |
| `confirmed_at` | `String` | Order confirmed at. |

### Affiliate Vendor Order order return_histories

| Field | Type | Description |
| --- | --- | --- |
| `body` | `String` | Refund summary. |
| `price` | `Float` | Refund amount. |
| `refunded_at` | `String` | Refunded at. |

### Affiliate Vendor Order order linked_order_info

| Field | Type | Description |
| --- | --- | --- |
| `source` | `String` | Affiliate source. |
| `shopdotcom_rid` | `String` | shopdotcom (Market America) RID. |
| `shopdotcom_click_id` | `Integer` | shopdotcom (Market America) click ID. |
| `line_shopping_ecid` | `String` | LINE Shopping ecid. |
| `line_shopping_affiliate` | `String` | LINE Shopping affiliate. |
| `ichannel_gid` | `String` | iChannel gid |

### Affiliate Vendor Order order customer_cancel_reason_detail

| Field | Type | Description |
| --- | --- | --- |
| `source` | `String` | source of cancel order |
| `reason_id` | `Integer` | id of cancel reason detail |
| `reason_detail` | `String` | cancel reason detail |

### Collection

| Field | Type | Description |
| --- | --- | --- |
| `title` | `String` | Custom collection name. |
| `handle` | `String` | Custom collection handle. |
| `published` | `Boolean` | Whether it is published. |
| `body_html` | `String` | Custom collection description. |
| `products_order_name` | `String` | Custom collection product ordering. |
| `position` | `Integer` | Custom collection position. |
| `id` | `Integer` | ID |

### Menu

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | id |
| `title` | `String` | Menu title |
| `system_default` | `Boolean` | Whether Menu is the system default |

### Menu Update

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` |  |
| `items` | `[Object]` |  |
| `system_default` | `Boolean` |  |
| `title` | `String` | Menu title |

### Menu Update items

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` |  |
| `items` | `[String]` |  |
| `position` | `Integer` |  |
| `title` | `String` | item title |
| `url` | `String` | https://example.com |

### Menu Destroy

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` |  |
| `system_default` | `Boolean` |  |
| `title` | `String` | Menu title |

### Domain

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `subdomain` | `String` | 子網域 |
| `domain` | `String` | 網域 |
| `host` | `String` | 主機 |
| `primary` | `Boolean` | 是否為主要網域 |
| `force_domain` | `Boolean` | 是否強制使用網域 |

### Bulk Operation

| Field | Type | Description |
| --- | --- | --- |
| `completed_at` | `String` | 2026-10-08T12:00:00+08:00 |
| `ids` | `[Integer]` |  |
| `status` | `String` | Completed \| Failed |
| `type` | `String` | BonusPoint |

### LINE Chat Bot

| Field | Type | Description |
| --- | --- | --- |
| `channel_id` | `String` | LINE channel id |
| `channel_secret` | `String` | LINE channel secret |
| `channel_token` | `String` | LINE channel access token |

### FBE Ad Account

| Field | Type | Description |
| --- | --- | --- |
| `account_id` | `Integer` | Reclaimed ad account id |

### POS Wallet Topup

| Field | Type | Description |
| --- | --- | --- |
| `topup_amount` | `Float` | Top-up amount |
| `created_at` | `String` | Created at |
| `payment_name` | `String` | Payment method name |
| `einvoice` | `Object` | E-invoice info |
| `paper_invoice_no` | `String` | Paper invoice number |

### POS Wallet Topup einvoice

| Field | Type | Description |
| --- | --- | --- |
| `title` | `String` | Invoice title |
| `company_no` | `String` | Company number |
| `invoice_no` | `String` | Invoice number |
| `invoice_status` | `String` | Invoice status |
| `invoice_at` | `String` | Invoice issued at |
| `invalid_at` | `String` | Invoice invalidated at |
| `random_num` | `String` | Random number |
| `invoice_type` | `String` | Invoice type |
| `love_code` | `String` | Donation love code |
| `phone_barcode` | `String` | Phone barcode carrier |
| `nature_person` | `String` | Nature person carrier |

### POS Wallet Refund

| Field | Type | Description |
| --- | --- | --- |
| `refund_amount` | `Float` | Refund amount (negative) |
| `created_at` | `String` | Created at |
| `invalid_einvoices` | `[Object]` | Invalidated einvoices |
| `allowance_einvoices` | `[Object]` | Allowance einvoices |

### POS Wallet Refund invalid_einvoices

| Field | Type | Description |
| --- | --- | --- |
| `title` | `String` | Invoice title |
| `company_no` | `String` | Company number |
| `invoice_no` | `String` | Invoice number |
| `invoice_status` | `String` | Invoice status |
| `invoice_at` | `String` | Invoice issued at |
| `invalid_at` | `String` | Invoice invalidated at |
| `random_num` | `String` | Random number |
| `invoice_type` | `String` | Invoice type |
| `love_code` | `String` | Donation love code |
| `phone_barcode` | `String` | Phone barcode carrier |
| `nature_person` | `String` | Nature person carrier |

### POS Wallet Refund allowance_einvoices

| Field | Type | Description |
| --- | --- | --- |
| `title` | `String` | Invoice title |
| `company_no` | `String` | Company number |
| `invoice_no` | `String` | Invoice number |
| `invoice_status` | `String` | Invoice status |
| `invoice_at` | `String` | Invoice issued at |
| `invalid_at` | `String` | Invoice invalidated at |
| `random_num` | `String` | Random number |
| `invoice_type` | `String` | Invoice type |
| `love_code` | `String` | Donation love code |
| `phone_barcode` | `String` | Phone barcode carrier |
| `nature_person` | `String` | Nature person carrier |

### Stock Invoice

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `pos_shop_name` | `String` | Outbound POS shop name. |
| `pos_shop_id` | `Integer` | Outbound POS shop ID. |
| `target_pos_shop_name` | `String` | Inbound POS shop name. |
| `target_pos_shop_id` | `Integer` | Inbound POS shop ID. |
| `barcode` | `String` | Barcode. |
| `status` | `String` | Processing status. |
| `stock_receipt_id` | `Integer` | Related stock receipt ID. |
| `items` | `[Object]` | Stock invoice items. |
| `created_at` | `Date` | Created at. |
| `comment` | `String` | Note. |
| `stock_requisition_id` | `Integer` | Source stock requisition ID |
| `updated_at` | `Date` | Updated at |

### Stock Invoice items

| Field | Type | Description |
| --- | --- | --- |
| `sku` | `String` | Variant SKU. |
| `qc` | `String` | Variant vendor item code. |
| `name` | `String` | Variant name. |
| `quantity` | `Integer` | Quantity. |

### Stock Receipt

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `pos_shop_name` | `String` | Inbound POS shop name. |
| `pos_shop_id` | `Integer` | Inbound POS shop ID. |
| `source_pos_shop_name` | `String` | Outbound POS shop name. |
| `source_pos_shop_id` | `Integer` | Outbound POS shop ID. |
| `barcode` | `String` | Barcode. |
| `status` | `String` | Processing status. |
| `stock_invoice_id` | `Integer` | Related stock invoice ID. |
| `items` | `[Object]` | Stock receipt items. |
| `check_logs` | `[Object]` | Check record. |
| `created_at` | `Date` | Created at. |
| `comment` | `String` | Note. |
| `stock_requisition_id` | `Integer` | Source stock requisition ID |
| `updated_at` | `Date` | Updated at |

### Stock Receipt items

| Field | Type | Description |
| --- | --- | --- |
| `sku` | `String` | Variant SKU. |
| `qc` | `String` | Variant vendor item code. |
| `name` | `String` | Variant name. |
| `quantity` | `Integer` | Quantity. |

### Stock Receipt check_logs

| Field | Type | Description |
| --- | --- | --- |
| `items` | `[Object]` | Checked items. |
| `created_at` | `Date` | Checked at. |

### Stock Receipt check_logs items

| Field | Type | Description |
| --- | --- | --- |
| `sku` | `String` | Variant SKU. |
| `qc` | `String` | Variant vendor item code. |
| `name` | `String` | Variant name. |
| `quantity` | `Integer` | Quantity. |

### Stock Requisition

| Field | Type | Description |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `pos_shop_name` | `String` | POS shop name. |
| `source_shop_name` | `String` | Source shop name. |
| `barcode` | `String` | Barcode. |
| `status` | `String` | Processing status. |
| `items` | `[Object]` | Stock requisition items. |
| `created_at` | `Date` | Created at. |
| `comment` | `String` | Note. |
| `pos_shop_id` | `Integer` | Target POS shop ID, 0 means the online store |
| `source_pos_shop_id` | `Integer` | Source POS shop ID, 0 means the online store |
| `updated_at` | `Date` | Updated at |

### Stock Requisition items

| Field | Type | Description |
| --- | --- | --- |
| `sku` | `String` | Variant SKU. |
| `qc` | `String` | Variant vendor item code. |
| `name` | `String` | Variant name. |
| `quantity` | `Integer` | Quantity. |

### Stored Order Retrieved

| Field | Type | Description |
| --- | --- | --- |
| `customer_id` | `Integer` | Customer ID |
| `order_id` | `Integer` | Order ID |
| `stored_order_id` | `Integer` | Stored order ID |
| `pos_shop_id` | `Integer` | POS shop ID |
| `pos_shop_name` | `String` | POS shop name |
| `user_id` | `Integer` | Operator user ID |
| `retrieved_at` | `String` | Retrieved time |
| `items` | `[Object]` | Retrieved items in this operation |

### Stored Order Refunded

| Field | Type | Description |
| --- | --- | --- |
| `customer_id` | `Integer` | Customer ID |
| `order_id` | `Integer` | Order ID |
| `stored_order_id` | `Integer` | Stored order ID |
| `status` | `String` | Stored order status |
| `cancelled_at` | `String` | Cancellation time |
| `refund_amount` | `Integer` | Total refund amount |
| `allowance_amount` | `Integer` | Allowance amount |
| `refund_at` | `String` | Refund time |
| `invoice_no` | `String` | Invoice number |
| `einvoice_status` | `String` | E-invoice status |
| `order_financial_status` | `String` | Order financial status |
| `order_return_status` | `String` | Order return status |
| `pos_shop_id` | `Integer` | POS shop ID |
| `pos_shop_name` | `String` | POS shop name |
| `user_id` | `Integer` | Operator user ID |
| `items` | `[Object]` | Cancelled stored order items |

### EDM Job

| Field | Type | Description |
| --- | --- | --- |
| `job_id` | `String` | sidekiq jid |
| `job_status` | `String` | success \| error |
| `result` | `Object` |  |

### EDM Job result

| Field | Type | Description |
| --- | --- | --- |
| `cost` | `Integer` |  |
| `edm_history_id` | `Integer` |  |
| `sended` | `Integer` |  |

### SMS Job

| Field | Type | Description |
| --- | --- | --- |
| `job_id` | `String` | sidekiq jid |
| `job_status` | `String` | success \| error |
| `result` | `Object` |  |

### SMS Job result

| Field | Type | Description |
| --- | --- | --- |
| `cost` | `Integer` |  |
| `sended` | `Integer` |  |

### LINE OA Message Job

| Field | Type | Description |
| --- | --- | --- |
| `job_id` | `String` | sidekiq jid |
| `job_status` | `String` | success \| error |
| `result` | `Object` |  |

### LINE OA Message Job result

| Field | Type | Description |
| --- | --- | --- |
| `error_customers` | `Object` |  |
| `success_count` | `Integer` |  |

### LINE OA Message Job result error_customers

| Field | Type | Description |
| --- | --- | --- |
| `1` | `[String]` |  |

### Customer IDs Job

| Field | Type | Description |
| --- | --- | --- |
| `job_id` | `String` | sidekiq jid |
| `job_status` | `String` | success \| error |
| `result` | `Object` |  |

### Customer IDs Job result

| Field | Type | Description |
| --- | --- | --- |
| `customer_ids` | `[Integer]` |  |

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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `VDNU3zSQWH4JG52fFrS4gSt8qVky0J+qyGkm9vMzukU=`

```json
{
  "app_uuid": "00000000-0000-4000-8000-000000000001",
  "app_version_uuid": "00000000-0000-4000-8000-000000000002",
  "app_client_id": "example-app-client-id"
}
```

### `orders/update`

An order changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: orders/update
X-Cyberbiz-Hmac-Sha256: d6e3752e548d49996a3b2ef78f62c020563c4a3c5046de13df6f951cee6af542
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `1uN1LlSNSZlqOy73j2LAIFY8SjxQRt4T32+VHO5q9UI=`

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
      "sku": "SKU-020",
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

### `products/update_options`

A product's options changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: products/update_options
X-Cyberbiz-Hmac-Sha256: 9f5b2d8bb398825d54f2a9ad4bda9ab388c1337cd6eb96a9cafb72147d71d089
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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

### `products/update_photos`

A product's photos changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: products/update_photos
X-Cyberbiz-Hmac-Sha256: 9f5b2d8bb398825d54f2a9ad4bda9ab388c1337cd6eb96a9cafb72147d71d089
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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

### `variants/out_of_stock`

A product variant sold out.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: variants/out_of_stock
X-Cyberbiz-Hmac-Sha256: ec978347d3190504f5ffbea579deeb214fed21e3ce5f2a5e7b978f87d605757f
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
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

### `cart/add`

A customer added items to the cart.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: cart/add
X-Cyberbiz-Hmac-Sha256: 45d236846abdaebe65f1ede179a0f9278a35e040a04a90b1157938ad6fc113f8
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `RdI2hGq9rr5l8e3heaD5J4o14ECgSpCxFXk4rW/BE/g=`

```json
{
  "cart_add_items": [
    {
      "variant": {
        "id": 1,
        "product_id": 1,
        "name": "範例名稱",
        "position": 1,
        "price": 100.0,
        "cost": 100.0,
        "compare_at_price": 100.0,
        "meas": 100.0,
        "max_usable_bonus": 100.0,
        "weight": 100.0,
        "option1": "範例",
        "option2": "範例",
        "option3": "範例",
        "inventory_management": true,
        "inventory_quantity": 1,
        "sold": 1,
        "safety_inventory_quantity": 1,
        "inventory_policy": "範例",
        "sku": "SKU-021",
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
        "created_at": "2026-09-01 10:00:00",
        "updated_at": "2026-09-01 10:00:00"
      },
      "quantity": 1
    }
  ],
  "customer": {
    "id": 1,
    "name": "範例名稱",
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
    "tags": [
      {
        "name": "範例名稱"
      }
    ],
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
    "other_accumulated_consumption": 1,
    "other_accumulated_consumption_expired_at": "2026-09-01 10:00:00",
    "note": "範例備註",
    "custom_fields": [
      {
        "name": "範例名稱",
        "label": "範例",
        "value": "範例"
      }
    ],
    "bonus_remain": 100.0,
    "uid_providers": [
      {
        "provider_type": "line",
        "uid": "U0000000000000000000000000000001"
      }
    ],
    "created_at": "2026-09-01 10:00:00",
    "updated_at": "2026-09-01 10:00:00",
    "confirmed_at": "2026-09-01 10:00:00",
    "mobile_sms_confirmed_at": "2026-09-01 10:00:00"
  }
}
```

### `customer_tags/update`

A customer's tags changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: customer_tags/update
X-Cyberbiz-Hmac-Sha256: 18754a836f4c46f27bc67e650d5940dc9e8ad408a6ab97bc5795a641c0924a85
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `GHVKg29MRvJ7xn5lDVlA3J6K1Aimq5e8V5WmQcCSSoU=`

```json
{
  "id": 1,
  "email": "customer@example.com",
  "tags": [
    {
      "id": 1,
      "name": "範例名稱"
    }
  ]
}
```

### `affiliate_vendor_orders/create`

An affiliate order was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: affiliate_vendor_orders/create
X-Cyberbiz-Hmac-Sha256: 5aa26e6d52940092dc3fb0b9cee9defa1c24ca6421b9fa0e41005046425570b7
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `WqJubVKUAJLcP7C5zune+hwkymQhufoOQQBQRkJVcLc=`

```json
{
  "status_code": "範例",
  "uid": "U0000000000000000000000000000001",
  "cid": "範例",
  "vendor_name": "範例",
  "order": {
    "id": 1,
    "subtotal_price": 100.0,
    "created_at": "2026-09-01 10:00:00",
    "updated_at": "2026-09-01 10:00:00",
    "order_number": "範例",
    "order_name": "範例",
    "buyer": {
      "email": "customer@example.com",
      "mobile": "0912345678"
    },
    "billing_address": [
      {
        "name": "範例名稱",
        "country_calling_code": "+886",
        "phone": "0912345678",
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
        "detail_address": [
          {
            "zip": "110",
            "country": "範例",
            "province": "範例",
            "city": "台北市",
            "district": "信義區",
            "address1": "市府路1號",
            "address2": "範例"
          }
        ]
      }
    ],
    "line_items": [
      {
        "id": 1,
        "product_id": 1,
        "product_variant_id": 1,
        "title": "範例標題",
        "variant_title": "範例商品 - 紅色",
        "sku": "SKU-022",
        "qc": "QC-001",
        "vendor": "範例",
        "price": 100.0,
        "cost": 100.0,
        "quantity": 1,
        "item_type": "normal",
        "return_status": "no_need",
        "discount_name": "範例",
        "discounts": [
          {
            "position": 1,
            "id": 1,
            "code": "SAMPLE100",
            "name": "範例名稱",
            "discount": 1
          }
        ],
        "total_price_before_discounts": 100.0,
        "total_discount": 100.0,
        "total_price_after_discounts": 100.0,
        "tax_type_id": 1,
        "bonus_redemption_price": 1,
        "related_items": [
          {
            "quantity": 1,
            "items": [
              {
                "id": 1,
                "product_id": 1,
                "product_variant_id": 1,
                "title": "範例標題",
                "variant_title": "範例商品 - 紅色",
                "sku": "SKU-023",
                "qc": "QC-001",
                "vendor": "範例",
                "price": 100.0,
                "cost": 100.0,
                "quantity": 1,
                "combo_product_price_difference": 100.0,
                "combo_product_price_diff_details": [
                  1
                ]
              }
            ]
          }
        ],
        "created_at": "2026-09-01 10:00:00",
        "channel": {
          "id": 1,
          "name": "範例名稱"
        },
        "weight": "範例",
        "photo": "//example.cyberbiz.co/media/sample-product.jpg",
        "custom_fields": [
          {
            "name": "範例名稱",
            "label": "範例",
            "value": "範例"
          }
        ]
      }
    ],
    "shipping_type": "範例",
    "shipping_name": "範例",
    "delivery_date": "2026-09-01 10:00:00",
    "delivery_time": 1,
    "delegate": "範例",
    "prices": [
      {
        "total_line_items_price": 100.0,
        "shipping_rate_price": 100.0,
        "discounts": [
          {
            "special_collection_discount": 1,
            "vip_discount": 100.0,
            "shop_discount": {
              "name": "範例名稱",
              "amount": 100.0
            },
            "coupon_discount": {
              "name": "範例名稱",
              "code": "SAMPLE100",
              "amount": 100.0,
              "coupon_id": 1
            },
            "coupon_discounts": [
              {
                "name": "範例名稱",
                "code": "SAMPLE100",
                "amount": 100.0,
                "coupon_id": 1,
                "id": 1
              }
            ],
            "bonus_consumed": 100.0,
            "vip_shipping_discount": 100.0,
            "coupon_shipping_discount": 100.0,
            "price_discount": 1,
            "third_party_discount": "範例"
          }
        ],
        "total_price": 100.0
      }
    ],
    "transaction_number": "TXN0000001",
    "merchant_trade_no": "S1#1001",
    "statuses": [
      {
        "order_status": "open",
        "financial_status": "paid",
        "fulfillment_status": "unshipped",
        "return_status": "no_need"
      }
    ],
    "timings": [
      {
        "request_return_at": "2026-09-01 10:00:00",
        "return_at": "2026-09-01 10:00:00",
        "refund_at": "2026-09-01 10:00:00",
        "closed_at": "2026-09-01 10:00:00",
        "cancelled_at": "2026-09-01 10:00:00",
        "expired_at": "2026-09-01 10:00:00",
        "confirmed_at": "2026-09-01 10:00:00"
      }
    ],
    "return_histories": [
      {
        "body": "退款說明",
        "price": 100.0,
        "refunded_at": "2026-09-01 10:00:00"
      }
    ],
    "note": "範例備註",
    "total_bonus_redemption_price": 1,
    "linked_order_info": {
      "source": "shop",
      "shopdotcom_rid": "範例",
      "shopdotcom_click_id": 1,
      "line_shopping_ecid": "範例",
      "line_shopping_affiliate": "範例",
      "ichannel_gid": "範例"
    },
    "shipping_status": "範例",
    "extra_info": "範例",
    "from_device": "範例",
    "customer_cancel_reason_detail": {
      "source": "shop",
      "reason_id": 1,
      "reason_detail": "顧客改變心意"
    },
    "serial_numbers": [],
    "order_weight": 100.0,
    "warehouse_type_id": 1
  }
}
```

### `affiliate_vendor_orders/cancelled`

An affiliate order was cancelled.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: affiliate_vendor_orders/cancelled
X-Cyberbiz-Hmac-Sha256: 9ab4240e940e124925d14876be2e286acb1edc980b43d9f83cab090e25821b33
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `mrQkDpQOEkkl0Uh2vi4oasse3JgLQ9n4PKsJDiWCGzM=`

```json
{
  "status_code": "範例",
  "uid": "U0000000000000000000000000000001",
  "cid": "範例",
  "vendor_name": "範例",
  "order": {
    "id": 1,
    "subtotal_price": 100.0,
    "created_at": "2026-09-01 10:00:00",
    "updated_at": "2026-09-01 10:00:00",
    "order_number": "範例",
    "order_name": "範例",
    "buyer": {
      "email": "customer@example.com",
      "mobile": "0912345678"
    },
    "billing_address": [
      {
        "name": "範例名稱",
        "country_calling_code": "+886",
        "phone": "0912345678",
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
        "detail_address": [
          {
            "zip": "110",
            "country": "範例",
            "province": "範例",
            "city": "台北市",
            "district": "信義區",
            "address1": "市府路1號",
            "address2": "範例"
          }
        ]
      }
    ],
    "line_items": [
      {
        "id": 1,
        "product_id": 1,
        "product_variant_id": 1,
        "title": "範例標題",
        "variant_title": "範例商品 - 紅色",
        "sku": "SKU-024",
        "qc": "QC-001",
        "vendor": "範例",
        "price": 100.0,
        "cost": 100.0,
        "quantity": 1,
        "item_type": "normal",
        "return_status": "no_need",
        "discount_name": "範例",
        "discounts": [
          {
            "position": 1,
            "id": 1,
            "code": "SAMPLE100",
            "name": "範例名稱",
            "discount": 1
          }
        ],
        "total_price_before_discounts": 100.0,
        "total_discount": 100.0,
        "total_price_after_discounts": 100.0,
        "tax_type_id": 1,
        "bonus_redemption_price": 1,
        "related_items": [
          {
            "quantity": 1,
            "items": [
              {
                "id": 1,
                "product_id": 1,
                "product_variant_id": 1,
                "title": "範例標題",
                "variant_title": "範例商品 - 紅色",
                "sku": "SKU-025",
                "qc": "QC-001",
                "vendor": "範例",
                "price": 100.0,
                "cost": 100.0,
                "quantity": 1,
                "combo_product_price_difference": 100.0,
                "combo_product_price_diff_details": [
                  1
                ]
              }
            ]
          }
        ],
        "created_at": "2026-09-01 10:00:00",
        "channel": {
          "id": 1,
          "name": "範例名稱"
        },
        "weight": "範例",
        "photo": "//example.cyberbiz.co/media/sample-product.jpg",
        "custom_fields": [
          {
            "name": "範例名稱",
            "label": "範例",
            "value": "範例"
          }
        ]
      }
    ],
    "shipping_type": "範例",
    "shipping_name": "範例",
    "delivery_date": "2026-09-01 10:00:00",
    "delivery_time": 1,
    "delegate": "範例",
    "prices": [
      {
        "total_line_items_price": 100.0,
        "shipping_rate_price": 100.0,
        "discounts": [
          {
            "special_collection_discount": 1,
            "vip_discount": 100.0,
            "shop_discount": {
              "name": "範例名稱",
              "amount": 100.0
            },
            "coupon_discount": {
              "name": "範例名稱",
              "code": "SAMPLE100",
              "amount": 100.0,
              "coupon_id": 1
            },
            "coupon_discounts": [
              {
                "name": "範例名稱",
                "code": "SAMPLE100",
                "amount": 100.0,
                "coupon_id": 1,
                "id": 1
              }
            ],
            "bonus_consumed": 100.0,
            "vip_shipping_discount": 100.0,
            "coupon_shipping_discount": 100.0,
            "price_discount": 1,
            "third_party_discount": "範例"
          }
        ],
        "total_price": 100.0
      }
    ],
    "transaction_number": "TXN0000001",
    "merchant_trade_no": "S1#1001",
    "statuses": [
      {
        "order_status": "open",
        "financial_status": "paid",
        "fulfillment_status": "unshipped",
        "return_status": "no_need"
      }
    ],
    "timings": [
      {
        "request_return_at": "2026-09-01 10:00:00",
        "return_at": "2026-09-01 10:00:00",
        "refund_at": "2026-09-01 10:00:00",
        "closed_at": "2026-09-01 10:00:00",
        "cancelled_at": "2026-09-01 10:00:00",
        "expired_at": "2026-09-01 10:00:00",
        "confirmed_at": "2026-09-01 10:00:00"
      }
    ],
    "return_histories": [
      {
        "body": "退款說明",
        "price": 100.0,
        "refunded_at": "2026-09-01 10:00:00"
      }
    ],
    "note": "範例備註",
    "total_bonus_redemption_price": 1,
    "linked_order_info": {
      "source": "shop",
      "shopdotcom_rid": "範例",
      "shopdotcom_click_id": 1,
      "line_shopping_ecid": "範例",
      "line_shopping_affiliate": "範例",
      "ichannel_gid": "範例"
    },
    "shipping_status": "範例",
    "extra_info": "範例",
    "from_device": "範例",
    "customer_cancel_reason_detail": {
      "source": "shop",
      "reason_id": 1,
      "reason_detail": "顧客改變心意"
    },
    "serial_numbers": [],
    "order_weight": 100.0,
    "warehouse_type_id": 1
  }
}
```

### `affiliate_vendor_orders/closed`

An affiliate order was closed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: affiliate_vendor_orders/closed
X-Cyberbiz-Hmac-Sha256: 3a05e57c70e487509e721b98b5edc86fe8345655ded7cfbed57bc199e61d76c4
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `OgXlfHDkh1CechuYte3Ib+g0VlXe18++1XvBmeYddsQ=`

```json
{
  "status_code": "範例",
  "uid": "U0000000000000000000000000000001",
  "cid": "範例",
  "vendor_name": "範例",
  "order": {
    "id": 1,
    "subtotal_price": 100.0,
    "created_at": "2026-09-01 10:00:00",
    "updated_at": "2026-09-01 10:00:00",
    "order_number": "範例",
    "order_name": "範例",
    "buyer": {
      "email": "customer@example.com",
      "mobile": "0912345678"
    },
    "billing_address": [
      {
        "name": "範例名稱",
        "country_calling_code": "+886",
        "phone": "0912345678",
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
        "detail_address": [
          {
            "zip": "110",
            "country": "範例",
            "province": "範例",
            "city": "台北市",
            "district": "信義區",
            "address1": "市府路1號",
            "address2": "範例"
          }
        ]
      }
    ],
    "line_items": [
      {
        "id": 1,
        "product_id": 1,
        "product_variant_id": 1,
        "title": "範例標題",
        "variant_title": "範例商品 - 紅色",
        "sku": "SKU-026",
        "qc": "QC-001",
        "vendor": "範例",
        "price": 100.0,
        "cost": 100.0,
        "quantity": 1,
        "item_type": "normal",
        "return_status": "no_need",
        "discount_name": "範例",
        "discounts": [
          {
            "position": 1,
            "id": 1,
            "code": "SAMPLE100",
            "name": "範例名稱",
            "discount": 1
          }
        ],
        "total_price_before_discounts": 100.0,
        "total_discount": 100.0,
        "total_price_after_discounts": 100.0,
        "tax_type_id": 1,
        "bonus_redemption_price": 1,
        "related_items": [
          {
            "quantity": 1,
            "items": [
              {
                "id": 1,
                "product_id": 1,
                "product_variant_id": 1,
                "title": "範例標題",
                "variant_title": "範例商品 - 紅色",
                "sku": "SKU-027",
                "qc": "QC-001",
                "vendor": "範例",
                "price": 100.0,
                "cost": 100.0,
                "quantity": 1,
                "combo_product_price_difference": 100.0,
                "combo_product_price_diff_details": [
                  1
                ]
              }
            ]
          }
        ],
        "created_at": "2026-09-01 10:00:00",
        "channel": {
          "id": 1,
          "name": "範例名稱"
        },
        "weight": "範例",
        "photo": "//example.cyberbiz.co/media/sample-product.jpg",
        "custom_fields": [
          {
            "name": "範例名稱",
            "label": "範例",
            "value": "範例"
          }
        ]
      }
    ],
    "shipping_type": "範例",
    "shipping_name": "範例",
    "delivery_date": "2026-09-01 10:00:00",
    "delivery_time": 1,
    "delegate": "範例",
    "prices": [
      {
        "total_line_items_price": 100.0,
        "shipping_rate_price": 100.0,
        "discounts": [
          {
            "special_collection_discount": 1,
            "vip_discount": 100.0,
            "shop_discount": {
              "name": "範例名稱",
              "amount": 100.0
            },
            "coupon_discount": {
              "name": "範例名稱",
              "code": "SAMPLE100",
              "amount": 100.0,
              "coupon_id": 1
            },
            "coupon_discounts": [
              {
                "name": "範例名稱",
                "code": "SAMPLE100",
                "amount": 100.0,
                "coupon_id": 1,
                "id": 1
              }
            ],
            "bonus_consumed": 100.0,
            "vip_shipping_discount": 100.0,
            "coupon_shipping_discount": 100.0,
            "price_discount": 1,
            "third_party_discount": "範例"
          }
        ],
        "total_price": 100.0
      }
    ],
    "transaction_number": "TXN0000001",
    "merchant_trade_no": "S1#1001",
    "statuses": [
      {
        "order_status": "open",
        "financial_status": "paid",
        "fulfillment_status": "unshipped",
        "return_status": "no_need"
      }
    ],
    "timings": [
      {
        "request_return_at": "2026-09-01 10:00:00",
        "return_at": "2026-09-01 10:00:00",
        "refund_at": "2026-09-01 10:00:00",
        "closed_at": "2026-09-01 10:00:00",
        "cancelled_at": "2026-09-01 10:00:00",
        "expired_at": "2026-09-01 10:00:00",
        "confirmed_at": "2026-09-01 10:00:00"
      }
    ],
    "return_histories": [
      {
        "body": "退款說明",
        "price": 100.0,
        "refunded_at": "2026-09-01 10:00:00"
      }
    ],
    "note": "範例備註",
    "total_bonus_redemption_price": 1,
    "linked_order_info": {
      "source": "shop",
      "shopdotcom_rid": "範例",
      "shopdotcom_click_id": 1,
      "line_shopping_ecid": "範例",
      "line_shopping_affiliate": "範例",
      "ichannel_gid": "範例"
    },
    "shipping_status": "範例",
    "extra_info": "範例",
    "from_device": "範例",
    "customer_cancel_reason_detail": {
      "source": "shop",
      "reason_id": 1,
      "reason_detail": "顧客改變心意"
    },
    "serial_numbers": [],
    "order_weight": 100.0,
    "warehouse_type_id": 1
  }
}
```

### `collections/create`

A product collection was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: collections/create
X-Cyberbiz-Hmac-Sha256: 8428f5ffe4299af712c7176542776d6c67d14aaddf2d844bcefbd2dd83f48a8d
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `hCj1/+QpmvcSxxdlQndtbGfRSq3fLYRLzvvS3YP0io0=`

```json
{
  "title": "範例標題",
  "handle": "範例",
  "published": true,
  "body_html": "<p>範例內容</p>",
  "products_order_name": "範例",
  "position": 1,
  "id": 1
}
```

### `collections/update`

A product collection changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: collections/update
X-Cyberbiz-Hmac-Sha256: 8428f5ffe4299af712c7176542776d6c67d14aaddf2d844bcefbd2dd83f48a8d
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `hCj1/+QpmvcSxxdlQndtbGfRSq3fLYRLzvvS3YP0io0=`

```json
{
  "title": "範例標題",
  "handle": "範例",
  "published": true,
  "body_html": "<p>範例內容</p>",
  "products_order_name": "範例",
  "position": 1,
  "id": 1
}
```

### `collections/delete`

A product collection was deleted.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: collections/delete
X-Cyberbiz-Hmac-Sha256: 8428f5ffe4299af712c7176542776d6c67d14aaddf2d844bcefbd2dd83f48a8d
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `hCj1/+QpmvcSxxdlQndtbGfRSq3fLYRLzvvS3YP0io0=`

```json
{
  "title": "範例標題",
  "handle": "範例",
  "published": true,
  "body_html": "<p>範例內容</p>",
  "products_order_name": "範例",
  "position": 1,
  "id": 1
}
```

### `menus/create`

A storefront menu was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: menus/create
X-Cyberbiz-Hmac-Sha256: b975e270297c57ff8adad11b437a31a598440c4b6f07166ba22f1f043ef8a144
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `uXXicCl8V/+K2tEbQ3oxpZhEDEtvBxZroi8fBD74oUQ=`

```json
{
  "id": 1,
  "title": "範例標題",
  "system_default": true
}
```

### `menus/update`

A storefront menu changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: menus/update
X-Cyberbiz-Hmac-Sha256: 45e687b22a2103cc4612345b94f3986e77fcb56c2be9d39c1a6baa6caea82b96
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `ReaHsiohA8xGEjRblPOYbnf8tWwr6dOcGmuqbK6oK5Y=`

```json
{
  "id": 1,
  "items": [
    {
      "id": 1,
      "items": [
        "範例"
      ],
      "position": 1,
      "title": "範例標題",
      "url": "//example.cyberbiz.co/media/sample-product.jpg"
    }
  ],
  "system_default": true,
  "title": "範例標題"
}
```

### `menus/destroy`

A storefront menu was deleted.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: menus/destroy
X-Cyberbiz-Hmac-Sha256: ea47f8b45eaebba412d239f206442244275b46952be4f71c3abe58c5b7fafa77
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `6kf4tF6uu6QS0jnyBkQiRCdbRpUr5PccOr5Yxbf6+nc=`

```json
{
  "id": 1,
  "system_default": true,
  "title": "範例標題"
}
```

### `domains/create`

A shop domain was added.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: domains/create
X-Cyberbiz-Hmac-Sha256: 1de261e0b94320d0fb84399319cd562ab1f0c46b0cbe1d7c2f875c11354ed37e
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `HeJh4LlDIND7hDmTGc1WKrHwxGsMvh18L4dcETVO034=`

```json
{
  "id": 1,
  "subdomain": "example",
  "domain": "www.example-shop.com",
  "host": "www.example-shop.com",
  "primary": true,
  "force_domain": true
}
```

### `domains/update`

A shop domain changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: domains/update
X-Cyberbiz-Hmac-Sha256: 1de261e0b94320d0fb84399319cd562ab1f0c46b0cbe1d7c2f875c11354ed37e
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `HeJh4LlDIND7hDmTGc1WKrHwxGsMvh18L4dcETVO034=`

```json
{
  "id": 1,
  "subdomain": "example",
  "domain": "www.example-shop.com",
  "host": "www.example-shop.com",
  "primary": true,
  "force_domain": true
}
```

### `domains/destroy`

A shop domain was removed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: domains/destroy
X-Cyberbiz-Hmac-Sha256: 1de261e0b94320d0fb84399319cd562ab1f0c46b0cbe1d7c2f875c11354ed37e
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `HeJh4LlDIND7hDmTGc1WKrHwxGsMvh18L4dcETVO034=`

```json
{
  "id": 1,
  "subdomain": "example",
  "domain": "www.example-shop.com",
  "host": "www.example-shop.com",
  "primary": true,
  "force_domain": true
}
```

### `bulk_operations/finish`

A bulk operation finished.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: bulk_operations/finish
X-Cyberbiz-Hmac-Sha256: 13331630be5f45976b39b2341e9980b42e0342a16a069591e4c5f1bffee00f50
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `EzMWML5fRZdrObI0HpmAtC4DQqFqBpWR5MXxv/7gD1A=`

```json
{
  "completed_at": "2026-09-01 10:00:00",
  "ids": [
    1
  ],
  "status": "enabled",
  "type": "範例"
}
```

### `line_chat_bot/update`

The shop's LINE Messaging API settings changed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: line_chat_bot/update
X-Cyberbiz-Hmac-Sha256: e16582ddd31eba98ac8a2f090e04116c6c02ac11286dd7e69fa6d6473c43f98f
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `4WWC3dMeupisii8JDgQRbGwCrBEobdfmn6bWRzxD+Y8=`

```json
{
  "channel_id": "1234567890",
  "channel_secret": "example-channel-secret",
  "channel_token": "example-channel-access-token"
}
```

### `fbe_ad_accounts/reclaim`

A Facebook ad account was reclaimed.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: fbe_ad_accounts/reclaim
X-Cyberbiz-Hmac-Sha256: 51b83b134b1c8ee567a5fcb34c1fd524090c68a002dd01f20e7798563c06e4a3
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `Ubg7E0scjuVnpfyzTB/VJAkMaKAC3QHyDneYVjwG5KM=`

```json
{
  "account_id": 1
}
```

### `pos_wallet_transactions/topup`

A POS wallet was topped up.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: pos_wallet_transactions/topup
X-Cyberbiz-Hmac-Sha256: 3a2d7b0773ee8d79f6d9cb7900dcbde7b6402b51249e831b4b9b907c59d6a61d
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `Oi17B3PujXn22ct5ANy957ZAK1EknoMbS5uQfFnWph0=`

```json
{
  "topup_amount": 100.0,
  "created_at": "2026-09-01 10:00:00",
  "payment_name": "範例",
  "einvoice": {
    "title": "範例標題",
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
  "paper_invoice_no": "PA12345678"
}
```

### `pos_wallet_transactions/refund`

A POS wallet top-up was refunded.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: pos_wallet_transactions/refund
X-Cyberbiz-Hmac-Sha256: 19ccaf20d88e87050d59991adb943e5325dcaf6941c589a89f240c384b3a980c
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `GcyvINiOhwUNWZka25Q+UyXcr2lBxYmonyQMOEs6mAw=`

```json
{
  "refund_amount": 100.0,
  "created_at": "2026-09-01 10:00:00",
  "invalid_einvoices": [
    {
      "title": "範例標題",
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
    }
  ],
  "allowance_einvoices": [
    {
      "title": "範例標題",
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
    }
  ]
}
```

### `stock_invoices/created`

A stock invoice (outbound) was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_invoices/created
X-Cyberbiz-Hmac-Sha256: 65cafba8bbb31c1640ae012d1d647923130e5ca885c0bd9c1e34365b9bd60dd5
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `Zcr7qLuzHBZArgEtHWR5IxMOXKiFwL2cHjQ2W5vWDdU=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "pos_shop_id": 1,
  "target_pos_shop_name": "範例",
  "target_pos_shop_id": 1,
  "barcode": "SI00000001",
  "status": "enabled",
  "stock_receipt_id": 1,
  "items": [
    {
      "sku": "SKU-028",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "stock_requisition_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_invoices/done`

A stock invoice was shipped out.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_invoices/done
X-Cyberbiz-Hmac-Sha256: 71512a0137c75f9633abbcb9d0281a8a9ca1b3ab3db8ea41d4842c6a1f5488de
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `cVEqATfHX5Yzq7y50Cgaipyhs6s9uOpB1IQsah9UiN4=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "pos_shop_id": 1,
  "target_pos_shop_name": "範例",
  "target_pos_shop_id": 1,
  "barcode": "SI00000001",
  "status": "enabled",
  "stock_receipt_id": 1,
  "items": [
    {
      "sku": "SKU-029",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "stock_requisition_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_invoices/cancelled`

A stock invoice was cancelled.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_invoices/cancelled
X-Cyberbiz-Hmac-Sha256: 6dc9757629248c14d785a116fa89cf026d1d7cf5a64e64d2e7e182752f930c47
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `bcl1dikkjBTXhaEW+onPAm0dfPWmTmTS5+GCdS+TDEc=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "pos_shop_id": 1,
  "target_pos_shop_name": "範例",
  "target_pos_shop_id": 1,
  "barcode": "SI00000001",
  "status": "enabled",
  "stock_receipt_id": 1,
  "items": [
    {
      "sku": "SKU-030",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "stock_requisition_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_receipts/created`

A stock receipt (inbound) was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_receipts/created
X-Cyberbiz-Hmac-Sha256: 462c316477d3c58d0e9b73615a4612662066da78d1e15e4036433c7d7de8c4f9
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `RiwxZHfTxY0Om3NhWkYSZiBm2njR4V5ANkM8fX3oxPk=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "pos_shop_id": 1,
  "source_pos_shop_name": "範例",
  "source_pos_shop_id": 1,
  "barcode": "SI00000001",
  "status": "enabled",
  "stock_invoice_id": 1,
  "items": [
    {
      "sku": "SKU-031",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "check_logs": [
    {
      "items": [
        {
          "sku": "SKU-032",
          "qc": "QC-001",
          "name": "範例名稱",
          "quantity": 1
        }
      ],
      "created_at": "2026-09-01 10:00:00"
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "stock_requisition_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_receipts/done`

A stock receipt was put into stock.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_receipts/done
X-Cyberbiz-Hmac-Sha256: 55d9639fbea12f7a98c6f5c8ac9e1d2a34435671cafa21b28c3a5fe62042fc53
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `Vdljn76hL3qYxvXIrJ4dKjRDVnHK+iGyjDpf5iBC/FM=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "pos_shop_id": 1,
  "source_pos_shop_name": "範例",
  "source_pos_shop_id": 1,
  "barcode": "SI00000001",
  "status": "enabled",
  "stock_invoice_id": 1,
  "items": [
    {
      "sku": "SKU-033",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "check_logs": [
    {
      "items": [
        {
          "sku": "SKU-034",
          "qc": "QC-001",
          "name": "範例名稱",
          "quantity": 1
        }
      ],
      "created_at": "2026-09-01 10:00:00"
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "stock_requisition_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_receipts/cancelled`

A stock receipt was cancelled.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_receipts/cancelled
X-Cyberbiz-Hmac-Sha256: d8e4ef575754b7749a0d7e1cb7678b8ef7e332dbb13fd8bbfed69f25c725e8c1
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `2OTvV1dUt3SaDX4ct2eLjvfjMtuxP9i7/tafJccl6ME=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "pos_shop_id": 1,
  "source_pos_shop_name": "範例",
  "source_pos_shop_id": 1,
  "barcode": "SI00000001",
  "status": "enabled",
  "stock_invoice_id": 1,
  "items": [
    {
      "sku": "SKU-035",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "check_logs": [
    {
      "items": [
        {
          "sku": "SKU-036",
          "qc": "QC-001",
          "name": "範例名稱",
          "quantity": 1
        }
      ],
      "created_at": "2026-09-01 10:00:00"
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "stock_requisition_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_requisitions/created`

A stock requisition (transfer) was created.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_requisitions/created
X-Cyberbiz-Hmac-Sha256: c0d9dadb28deda94e9b4474ba1fe4b4e5ce7d57fd2e27ccb64d69f44191d7be7
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `wNna2yje2pTptEdLof5LTlzn1X/S4nzLZNafRBkde+c=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "source_shop_name": "範例",
  "barcode": "SI00000001",
  "status": "enabled",
  "items": [
    {
      "sku": "SKU-037",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "pos_shop_id": 1,
  "source_pos_shop_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_requisitions/approved`

A stock requisition was approved.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_requisitions/approved
X-Cyberbiz-Hmac-Sha256: 5e347ac013c68430a9117c82abade2544758d6635d9f459d2cd631848cb1b4b7
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `XjR6wBPGhDCpEXyCq63iVEdY1mNdn0WdLNYxhIyxtLc=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "source_shop_name": "範例",
  "barcode": "SI00000001",
  "status": "enabled",
  "items": [
    {
      "sku": "SKU-038",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "pos_shop_id": 1,
  "source_pos_shop_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stock_requisitions/cancelled`

A stock requisition was cancelled.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stock_requisitions/cancelled
X-Cyberbiz-Hmac-Sha256: 7dae4a8d8e6fa54837f55c8c7cb09c48fb5235ac88c38dfaa2c298e18c24bdc4
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `fa5KjY5vpUg39VyMfLCcSPtSNayIw436osKY4YwkvcQ=`

```json
{
  "id": 1,
  "pos_shop_name": "範例",
  "source_shop_name": "範例",
  "barcode": "SI00000001",
  "status": "enabled",
  "items": [
    {
      "sku": "SKU-039",
      "qc": "QC-001",
      "name": "範例名稱",
      "quantity": 1
    }
  ],
  "created_at": "2026-09-01 10:00:00",
  "comment": "範例",
  "pos_shop_id": 1,
  "source_pos_shop_id": 1,
  "updated_at": "2026-09-01 10:00:00"
}
```

### `stored_orders/retrieved`

Stored goods were picked up.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stored_orders/retrieved
X-Cyberbiz-Hmac-Sha256: 21f643580e4fff5bc0a836fb5586b724853112645095e2665280f6ac0a741e3e
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `IfZDWA5P/1vAqDb7VYa3JIUxEmRQleJmUoD2rAp0Hj4=`

```json
{
  "customer_id": 1,
  "order_id": 1,
  "stored_order_id": 1,
  "pos_shop_id": 1,
  "pos_shop_name": "範例",
  "user_id": 1,
  "retrieved_at": "2026-09-01 10:00:00",
  "items": []
}
```

### `stored_orders/refunded`

A stored order was refunded.

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: stored_orders/refunded
X-Cyberbiz-Hmac-Sha256: 77ddaab98c53c103b917ea76b27d3c0790e55bc91165a8388a3d43343be54165
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `d92quYxTwQO5F+p2sn08B5DlW8kRZag4ij1DNDvlQWU=`

```json
{
  "customer_id": 1,
  "order_id": 1,
  "stored_order_id": 1,
  "status": "enabled",
  "cancelled_at": "2026-09-01 10:00:00",
  "refund_amount": 1,
  "allowance_amount": 1,
  "refund_at": "2026-09-01 10:00:00",
  "invoice_no": "AB12345678",
  "einvoice_status": "範例",
  "order_financial_status": "範例",
  "order_return_status": "範例",
  "pos_shop_id": 1,
  "pos_shop_name": "範例",
  "user_id": 1,
  "items": []
}
```

### `send_edm_finish`

An EDM send job finished (Merchant webhooks only).

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: send_edm_finish
X-Cyberbiz-Hmac-Sha256: 06a7ed8a52ccc0c487a6b47f3d36a5bfeb5dcd25fee235367c1318ca03447e7a
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `BqftilLMwMSHprR/PTalv+tdzSX+4jU2fBMYygNEfno=`

```json
{
  "job_id": "5f1a2b3c4d5e6f7a8b9c0d1e",
  "job_status": "success",
  "result": {
    "cost": 1,
    "edm_history_id": 1,
    "sended": 1
  }
}
```

### `send_sms_finish`

An SMS send job finished (Merchant webhooks only).

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: send_sms_finish
X-Cyberbiz-Hmac-Sha256: 49d9800ed825eeeedde78f28470bc78ed76bb50794368ff24ff61a00483f4d03
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `SdmADtgl7u7d548oRwvHjtdrtQeUNo/yT/YaAEg/TQM=`

```json
{
  "job_id": "5f1a2b3c4d5e6f7a8b9c0d1e",
  "job_status": "success",
  "result": {
    "cost": 1,
    "sended": 1
  }
}
```

### `send_line_oa_message_finish`

A LINE OA message send job finished (Merchant webhooks only).

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: send_line_oa_message_finish
X-Cyberbiz-Hmac-Sha256: 6dc6fddd3507985a250486f67d54987a1b542284f517dd7a1e7b484810a640a7
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `bcb93TUHmFolBIb2fVSYehtUIoT1F916HntISBCmQKc=`

```json
{
  "job_id": "5f1a2b3c4d5e6f7a8b9c0d1e",
  "job_status": "success",
  "result": {
    "error_customers": {
      "1": [
        "範例"
      ]
    },
    "success_count": 1
  }
}
```

### `get_customer_ids_finish`

A customer-group member lookup finished (Merchant webhooks only).

```http
POST /webhooks/cyberbiz HTTP/1.1
Host: example.com
Content-Type: application/json
User-Agent: CyberbizAppWebhook/1.0
X-Cyberbiz-Domain: example.cyberbiz.co
X-Cyberbiz-Shop-Domain: www.example-shop.com
X-Cyberbiz-Event: get_customer_ids_finish
X-Cyberbiz-Hmac-Sha256: e9a839c41e460802b277222d15f2a16406fb69dcaa3f76980ce461815a69eb79
X-Cyberbiz-Domain-Hmac-Sha256: e62a39a44930a761b4e8c6155113d484feb3350accc02444598fcfb01dd7891c
```

Base64 form of the same signature (as the CYBERBIZ documentation describes it): `6ag5xB5GCAKydyItFfKhZAb7adyqP3aYDORhgVpp63k=`

```json
{
  "job_id": "5f1a2b3c4d5e6f7a8b9c0d1e",
  "job_status": "success",
  "result": {
    "customer_ids": [
      1
    ]
  }
}
```

## Release Notes

### 2.1.0 (2026-10-14)

- webhooks.md and the webhook Postman collection document 38 more events from the source-derived webhook collection, including the four Merchant-webhook job events.

### 2.0.1 (2026-10-14)

- The webhook Domain Signature is documented and computed as the HMAC-SHA256 of the base64-encoded X-Cyberbiz-Domain, which is what the platform signs.

### 2.0.0 (2026-10-09)

- Removed the /v1/limit_collections endpoints: the platform does not serve them on app-store-api.cyberbiz.io.

### 1.1.0 (2026-10-09)

- Product create and POS batch create document the sku request parameter, which the platform requires for shops with the POS feature.

### 1.0.1 (2026-10-02)

- Postman collection descriptions no longer point to repository-internal files.

### 1.0.0 (2026-09-08)

- Initial generation from the CYBERBIZ v1 swagger, the v1/v2 Postman collections, the Notion v2 reference and the webhook reference.
- Observed corrections applied: nullable fields, money as float, order_number as integer, arrays documented as objects, fields missing from the swagger, Bearer authentication.
- Synthetic samples derived from the Golden Files for every operation and event.
