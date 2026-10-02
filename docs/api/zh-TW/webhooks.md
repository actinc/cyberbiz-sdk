# CYBERBIZ Webhooks

Version: 1.0.1

每當商店發生 app 訂閱的事件，CYBERBIZ 會對 app 註冊的 HTTPS 網址（app manifest 的 `webhook_url`）送出 HTTP 請求。
本文件列出所有事件、CYBERBIZ 送出的 HTTP header、簽章驗證方式，以及每種 payload 的合成範例。

## 傳送方式

- 方法 `POST`，內容 `application/json`，UTF-8。body 即為事件的 payload 物件，沒有外層封裝。
- 每個資源的每個事件各送一個請求。同一資源的多個事件（例如 `orders/create` 與 `orders/paid`，或兩次 `customers/update`）可能在同一秒內以任意順序抵達；請依 `updated_at` 與資源 id 排序，而非抵達時間。
- 請盡快（數秒內）回應 `2xx`，並以非同步方式處理。

### Header

| Header | 範例 | 說明 |
| --- | --- | --- |
| `User-Agent` | `CyberbizAppWebhook/1.0` | App webhook 的固定值（實際觀察）。文件寫的是 `CyberbizWebhook/1.0`，那是舊版商店 webhook。 |
| `X-Cyberbiz-Domain` | `example.cyberbiz.co` | Shop Domain，即商店在 CYBERBIZ 的網址。用它找出商店的 App Secret。 |
| `X-Cyberbiz-Shop-Domain` | `www.example-shop.com` | 商家自訂的前台網址。僅供參考，不是識別碼。 |
| `X-Cyberbiz-Event` | `orders/paid` | 事件，命名為 `resource/action`。 |
| `X-Cyberbiz-Hmac-Sha256` | `3f2a…`（64 個十六進位字元） | 簽章：以 App Secret 為金鑰對原始 request body 計算的 HMAC-SHA256。 |
| `X-Cyberbiz-Domain-Hmac-Sha256` | `9b1c…`（64 個十六進位字元） | 網域簽章：以 App Secret 為金鑰對 `X-Cyberbiz-Domain` 值計算的 HMAC-SHA256。 |

## 驗證簽章

1. 在解析之前先讀取原始 body 位元組。
2. 計算 `HMAC-SHA256(app_secret, body)`。
3. 以常數時間比較摘要的 **hex** 編碼與 `X-Cyberbiz-Hmac-Sha256`。正式環境的 App webhook 送的是 hex（64 字元），雖然 CYBERBIZ 文件寫的是 base64 — 請同時接受 base64 作為備援，以免平台日後變更時驗證失敗。
4. 可選擇以相同方式用 `X-Cyberbiz-Domain` 檢查 `X-Cyberbiz-Domain-Hmac-Sha256`；它將請求綁定到單一商店，但不驗證 payload。
5. 驗證失敗時回應 `401`，不要處理 body。

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

## 重送與冪等

- 重送策略**未有文件說明**：請假設失敗的傳送可能重送也可能不重送，重送可能數分鐘後才到，也可能永遠不到。懷疑有遺漏時請透過 API（`GET /v1/orders/{order_id}` 等）核對。
- 傳送不帶 delivery id。請以 `(X-Cyberbiz-Event, payload id, payload updated_at)` 去重。
- 所有處理程序都應具冪等性：同一事件可能傳送多次，同一資源的事件也可能在同一秒內亂序抵達。

## 用 Postman 測試你的接收端

本檔案旁的 collection `cyberbiz-webhooks.postman_collection.json` 會將下列所有事件重放到你自己的接收端：在 Postman 環境中設定 `webhookUrl`、`shopDomain`、`customDomain` 與 `appSecret`（絕不提交），其 pre-request script 會以 CYBERBIZ 相同的方式為每個請求簽章（`signatureEncoding` 選擇 hex（正式環境形式）或 base64（文件形式））。

## 事件

| 事件 | 說明 | Payload |
| --- | --- | --- |
| `customers/create` | 會員註冊。 | [Customer](#customer) |
| `customers/update` | 會員修改。 | [Customer](#customer) |
| `uid_providers/create` | 會員新增第三方登入 UID（參考文件未另記 payload；送出的是會員物件）。 | [Customer](#customer) |
| `uid_providers/update` | 會員的第三方登入 UID 更新（會員物件）。 | [Customer](#customer) |
| `bonus_points/create` | 獲得紅利。 | [Bonus Points](#bonus-points) |
| `bonus_points/update` | 使用紅利。 | [Bonus Points](#bonus-points) |
| `bonus_points/destroy` | 刪除紅利。 | [Bonus Points](#bonus-points) |
| `comment_bonus/create` | 商品評論審核通過發送紅利。 | [Bonus Points](#bonus-points) |
| `orders/create` | 訂單成立。 | [Order](#order) |
| `orders/paid` | 訂單付款。 | [Order](#order) |
| `orders/preparing` | 準備出貨。 | [Order](#order) |
| `orders/fulfilled` | 訂單出貨。 | [Order](#order) |
| `orders/received` | 訂單收貨。 | [Order](#order) |
| `orders/arrived` | 訂單到店。 | [Order](#order) |
| `orders/expired` | 逾期未取。 | [Order](#order) |
| `orders/cancelled` | 訂單取消。 | [Order](#order) |
| `orders/returned` | 訂單退貨。 | [Order](#order) |
| `orders/partial_return` | 訂單部分退貨。 | [Order](#order) |
| `orders/refunded` | 訂單退款。 | [Order](#order) |
| `orders/partial_refunded` | 訂單部分退款。 | [Order](#order) |
| `orders/closed` | 訂單結案。 | [Order](#order) |
| `orders/request_return` | 退貨申請。 | [Order](#order) |
| `orders/opened` | 訂單開啟。 | [Order](#order) |
| `express_delivery_orders/update` | 快速到貨訂單更新。 | [Order](#order) |
| `products/create` | 商品新增。 | [Product](#product) |
| `products/update` | 商品修改。 | [Product](#product) |
| `products/delete` | 商品刪除。 | [Product](#product) |
| `variants/create` | 款式新增。 | [Product Variant](#product-variant) |
| `variants/update` | 款式修改。 | [Product Variant](#product-variant) |
| `variants/delete` | 款式刪除。 | [Product Variant](#product-variant) |
| `coupons/create` | 新增優惠券。 | [Coupon](#coupon) |
| `coupons/update` | 使用優惠券。 | [Coupon](#coupon) |
| `coupons/destroy` | 刪除優惠券。 | [Coupon](#coupon) |
| `customer_vip_level/update` | 會員 VIP 等級更新。 | [Customer VIP Level](#customer-vip-level) |
| `apps/uninstall` | App 解除安裝。 | [App Uninstall](#app-uninstall) |

### 沒有 payload 文件的事件

以下事件出現在 app manifest（`GET /settings` 的 `webhook_events`）中，但參考文件未記載其 payload：

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

## Payload 物件

欄位類型依 CYBERBIZ 文件所載。時間為 Asia/Taipei 的 `YYYY-MM-DD HH:MM:SS`；金額欄位為浮點數；任何欄位都可能為 `null`。

### Customer

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | 會員 ID |
| `name` | `String` | 姓名 |
| `status` | `String` | 狀態：`pending` (帳號未啟用)、`validate` (帳號未驗證)、`enabled` (帳號已啟用)、`disabled` (帳號已禁用)、`invited` (已邀請會員啟用)、`declined` (啟用邀請被拒絕)、`warning` (已列為警示帳號) |
| `email` | `String` | Email |
| `country_calling_code` | `String` | 手機國碼 (ex. +886) |
| `mobile` | `String` | 手機號碼 |
| `gender` | `String` | 性別 |
| `birthday` | `String` | 生日 |
| `enable_cvs_pickup` | `Boolean` | 是否可使用超商取貨 |
| `enable_cvs_cod` | `Boolean` | 是否可使用超商貨到付款 |
| `enable_home_delivery_cod` | `Boolean` | 是否可使用宅配貨到付款 |
| `accepts_marketing` | `Boolean` | 是否接受行銷 |
| `accepts_email_notification` | `Boolean` | 是否接收通知信 |
| `tags` | `[Object]` | 會員標籤陣列 |
| `address` | `Object` | 會員地址 |
| `other_accumulated_consumption` | `Integer` | 其他通路累積金額 |
| `other_accumulated_consumption_expired_at` | `String` | 其他通路累積金額起算日 |
| `note` | `String` | 備註 |
| `custom_fields` | `[Object]` | 自定欄位陣列 |
| `bonus_remain` | `Float` | 紅利點數總計 |
| `uid_providers` | `[Object]` | UID Providers 陣列 |
| `created_at` | `String` | 建立時間 |
| `updated_at` | `String` | 更新時間 |
| `confirmed_at` | `String` | Email 驗證時間 |
| `mobile_sms_confirmed_at` | `String` | 手機驗證時間 |

### Customer Tag

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `name` | `String` | 標籤名稱 |

### Customer Address

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `company` | `String` | 公司名稱 |
| `country_calling_code` | `String` | 電話國碼 |
| `phone` | `String` | 電話 |
| `address` | `String` | 地址 |
| `detail_address` | `Object` | 地址詳細資訊 |

### Detail Address

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `zip` | `String` | 郵遞區號 |
| `country` | `String` | 國家 |
| `province` | `String` | 省 |
| `city` | `String` | 市 |
| `district` | `String` | 區 |
| `address1` | `String` | 地址一 |
| `address2` | `String` | 地址二 |

### Custom Field

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `name` | `String` | 欄位名稱 |
| `label` | `String` | 欄位顯示名稱 |
| `value` | `String` | 欄位內容 |

### UID Providers

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `provider_type` | `String` | Provider 類型 |
| `uid` | `String` | UID |

### Bonus Points

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `title` | `String` | 名稱 |
| `points` | `Float` | 點數 |
| `unused_points` | `Float` | 尚未使用點數 |
| `consumption_price` | `Float` | 產生此紅利的消費金額 |
| `deadline` | `String` | 使用期限 |
| `customer_id` | `Integer` | 關聯顧客 ID |
| `source` | `String` | 來源 |
| `order_id` | `Integer` | 關聯訂單 ID |

### Order

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | 訂單 ID |
| `token` | `String` | Token |
| `order_number` | `String` | 訂單編號 |
| `order_name` | `String` | 訂單名稱 |
| `customer` | `Object` | 顧客資訊 |
| `buyer` | `Object` | 購買會員資訊 |
| `receiver` | `Object` | 收貨人資訊 |
| `billing_address` | `Object` | 購買人資訊 (客製項目) |
| `line_items` | `[Object]` | 訂單商品資訊陣列 |
| `shipping_type` | `String` | 配送方式 |
| `shipping_name` | `String` | 配送名稱 |
| `shipping_vendor` | `Hash` | 物流商 |
| `logistics_id` | `String` | 綠界物流訂單編號 |
| `delivery_date` | `String` | 指定配送日期 |
| `delivery_time` | `Integer` | 指定配送時間 |
| `delegate` | `String` | 代客下單操作人 |
| `fulfillments` | `[Object]` | 配送資訊陣列 |
| `payment_name` | `String` | 付款方式 |
| `payment_method` | `String` | 付款方式名稱 |
| `payment_url` | `String` | 付款網址 |
| `multiple_payment_infos` | `[Object]` | 多付款方式資訊陣列 |
| `prices` | `Object` | 價格資訊 |
| `card4no` | `String` | 付款卡號後四碼 |
| `transaction_number` | `String` | 第三方交易編號 |
| `merchant_trade_no` | `String` | 第三方串接編號 |
| `einvoice` | `Object` | 發票資訊 |
| `paper_invoice_no` | `String` | 紙本發票號碼 |
| `statuses` | `Object` | 狀態資訊 |
| `timings` | `Object` | 時間資訊 |
| `return_histories` | `[Object]` | 退款資訊陣列 |
| `note` | `String` | 訂單備註 |
| `branch_store` | `Object` | 自取門市資訊 |
| `referral_code` | `String` | 推薦分潤代碼 |
| `checkout_referral_code` | `String` | 結帳人分潤代碼 |
| `checkout_referral_user_name` | `String` | 結帳人名稱 |
| `register_referral_code` | `String` | 註冊人分潤代碼 |
| `total_bonus_redemption_price` | `Integer` | 紅利商城紅利總使用量 |
| `pos_info` | `Object` | POS 相關資訊 |
| `exchange_histories` | `Object` | 換貨相關資訊 |
| `linked_order_info` | `Object` | 導購訂單資訊 |
| `tags` | `[Object]` | 訂單標籤陣列 |
| `express_delivery_branch_store` | `Object` | 快速到貨門市資訊 |
| `shipping_status` | `String` | 送貨狀態 |
| `extra_info` | `String` | 額外資訊 |
| `from_device` | `String` | 訂單來源 |
| `customer_cancel_reason_detail` | `Object` | 取消原因詳情 |
| `serial_numbers` | `[String]` | 活動序號 |
| `order_weight` | `Float` | 總重量 |
| `created_at` | `String` | 成立時間 |
| `updated_at` | `String` | 更新時間 |

### Buyer

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `email` | `String` | 購買會員 Email |
| `mobile` | `String` | 購買會員手機 |

### Receiver

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `name` | `String` | 收貨人名稱 |
| `country_calling_code` | `String` | 收貨人電話國碼 |
| `phone` | `String` | 收貨人電話 |
| `address` | `String` | 收貨人地址 |
| `detail_address` | `Object` | 收貨人地址詳細資訊 |
| `cvs_store_id` | `String` | 超取店號 |
| `allpay_logistics_id` | `String` | 綠界物流編號 |

### Line Item

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `product_id` | `Integer` | 商品 ID |
| `product_variant_id` | `Integer` | 款式 ID |
| `title` | `String` | 商品名稱 |
| `variant_title` | `String` | 款式名稱 |
| `sku` | `String` | 款式 SKU |
| `qc` | `String` | 款式廠商編號 |
| `vendor` | `String` | 商品來源廠商 |
| `price` | `Float` | 金額 |
| `cost` | `Float` | 成本 |
| `quantity` | `Integer` | 數量 |
| `item_type` | `String` | 類型 |
| `return_status` | `String` | 退貨狀態 |
| `discount_name` | `String` | 折扣名稱 |
| `discounts` | `[Object]` | 商品折扣拆負項陣列 |
| `total_price_before_discounts` | `Float` | 折扣前總金額 |
| `total_discount` | `Float` | 總折扣金額 |
| `total_price_after_discounts` | `Float` | 折扣後總金額 |
| `tax_type_id` | `String` | 商品課稅類別：`inclusive_tax` (應稅)、`zero_tax` (零稅率)、`exclusive_tax` (免稅) |
| `bonus_redemption_price` | `Integer` | 紅利商城兌換點數，若非紅利兌換則為 null |
| `related_items` | `[Object]` | 組合品子商品資訊 |
| `channel` | `String` | 商品通路 |
| `weight` | `String` | 商品重量 |
| `photo` | `String` | 商品圖片 |
| `created_at` | `Date` | 建立時間 |

### Discount

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `position` | `Integer` | 第 n 個商品 |
| `id` | `Integer` | 折扣類型 ID |
| `code` | `String` | 折扣類型代碼 |
| `name` | `String` | 折扣類型名稱 |
| `discount` | `Integer` | 折扣金額 |

### Shipping Vendor

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `type` | `String` | 物流商代碼 |
| `name` | `String` | 物流商名稱 |

### Fulfillment

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `tracking_company` | `String` | 配送方式 |
| `tracking_number` | `String` | 配送單號 |
| `fulfilled_at` | `String` | 出貨日期 |
| `received_at` | `String` | 收貨日期 |
| `status` | `String` | 配送狀態 |
| `line_items` | `[Object]` | 配送商品陣列 |
| `tracking_url` | `String` | 貨態追蹤網址 (目前只支援 Uber Direct 與 Pandago) |

### Multiple Payment Info

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `name` | `String` | 付款名稱 |
| `amount` | `Float` | 付款金額 |

### Prices

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `total_line_items_price` | `Float` | 商品總金額 |
| `shipping_rate_price` | `Float` | 運費 |
| `discounts` | `Object` | 折扣金額明細 |
| `total_price` | `Float` | 訂單總金額 |

### Discounts Detail

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `special_collection_discount` | `Integer` | 行銷活動折扣 |
| `vip_discount` | `Float` | VIP 折扣 |
| `shop_discount` | `Object` | 全館折扣明細 |
| `coupon_discount` | `Object` | 優惠券明細 (單個) [已棄用] |
| `coupon_discounts` | `[Object]` | 優惠券明細陣列 |
| `bonus_consumed` | `Float` | 紅利折扣 |
| `vip_shipping_discount` | `Float` | VIP 運費折扣 |
| `coupon_shipping_discount` | `Float` | 免運券運費折扣 |
| `price_discount` | `Integer` | 店長折扣 |
| `third_party_discount` | `Integer` | 第三方折扣 |

### Shop Discount

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `name` | `String` | 全館活動名稱 |
| `amount` | `Float` | 全館活動折扣 |

### Coupon Discount

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | 優惠券名稱 |
| `code` | `String` | 優惠券代碼 |
| `amount` | `Float` | 優惠券折扣 |
| `coupon_id` | `Integer` | 優惠券 ID |

### E-Invoice

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `title` | `String` | 發票抬頭 |
| `order_id` | `Integer` | 訂單 ID |
| `company_no` | `String` | 發票統編 |
| `invoice_no` | `String` | 發票號碼 |
| `invoice_status` | `String` | 發票狀態 |
| `invoice_at` | `String` | 發票開立時間 |
| `invalid_at` | `String` | 發票作廢或折讓時間 |
| `random_num` | `String` | 隨機碼 |
| `invoice_type` | `String` | 發票類型 |
| `love_code` | `String` | 捐贈碼 |
| `phone_barcode` | `String` | 手機條碼 |
| `nature_person` | `String` | 自然人憑證 |

### Statuses

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `order_status` | `String` | 訂單狀態 |
| `financial_status` | `String` | 付款狀態 |
| `fulfillment_status` | `String` | 配送狀態 |
| `return_status` | `String` | 退貨狀態 |

### Timings

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `request_return_at` | `String` | 申請退貨時間 |
| `return_at` | `String` | 退貨時間 |
| `refund_at` | `String` | 退款時間 |
| `closed_at` | `String` | 關閉時間 |
| `cancelled_at` | `String` | 取消時間 |
| `expired_at` | `String` | 逾期未取時間 |
| `confirmed_at` | `String` | 訂單認單時間 |

### Return History

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `body` | `String` | 退款內容簡述 |
| `price` | `Float` | 退款金額 |
| `refunded_at` | `String` | 退款時間 |

### Branch Store

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `store_no` | `String` | 門市編號 |
| `name` | `String` | 門市名稱 |
| `phone` | `String` | 門市電話 |
| `county` | `String` | 城市 |
| `district` | `String` | 鄉鎮地區 |
| `address` | `String` | 地址 |
| `zip` | `String` | 郵遞區號 |
| `opening_hours` | `String` | 營業時間 |
| `lat` | `Float` | 緯度座標 |
| `lng` | `Float` | 經度座標 |
| `enabled` | `Boolean` | 是否啟用 |
| `shipping_rates` | `[Object]` | 門市運費設定陣列 |
| `source_type` | `String` | 來源類型：`BranchStore` 或 `PosShop` |
| `source_id` | `Integer` | 來源 ID |

### POS Info

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `pos_user_id` | `Integer` | POS 銷售員 ID |
| `pos_user_email` | `String` | POS 銷售員 Email |
| `pos_shop_id` | `Integer` | POS 銷售員所屬店家 ID |
| `pos_info` | `String` | POS 相關資訊 |
| `pos_id` | `Integer` | 訂單銷售使用 POS 機 ID |
| `pos_name` | `String` | 訂單銷售使用 POS 機名稱 |

### Linked Order Info

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `source` | `String` | 導購來源 |
| `shopdotcom_rid` | `String` | 美安 RID |
| `shopdotcom_click_id` | `String` | 美安 Click_ID |
| `line_shopping_ecid` | `String` | LINE 購物 ecid |
| `line_shopping_affiliate` | `String` | LINE 購物 affiliate |
| `ichannel_gid` | `String` | iChannel gid |

### Customer Cancel Reason Detail

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `source` | `String` | 取消訂單來源 |
| `reason_id` | `Integer` | 取消原因 ID |
| `reason_detail` | `String` | 取消原因詳情 |

### Product

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | 商品 ID |
| `title` | `String` | 商品名稱 |
| `english_title` | `String` | 商品英文名稱 |
| `product_url` | `String` | 商品網址 |
| `published` | `Boolean` | 是否公開 |
| `sell_from` | `String` | 商品販售起始時間 |
| `sell_to` | `String` | 商品販售結束時間 |
| `product_type` | `String` | 商品類型 |
| `product_type_code` | `String` | 商品類型代碼 |
| `slogan` | `String` | 商品標語 |
| `brief` | `String` | 商品簡述 HTML 樣板文字 |
| `brief_text` | `String` | 商品簡述純文字 |
| `brief_includes_html` | `Boolean` | 商品簡述是否使用 HTML 樣板文字 |
| `body_html` | `String` | 商品介紹內容 |
| `vendor` | `String` | 商品來源廠商 |
| `price` | `Float` | 商品最低金額 |
| `sell_weight` | `Integer` | 商品販售數量 |
| `tax_type_id` | `String` | 商品課稅類別：`inclusive_tax` (應稅)、`zero_tax` (零稅率)、`exclusive_tax` (免稅) |
| `custom_collections` | `[Object]` | 關聯自訂群組陣列 |
| `special_collection` | `Object` | 關聯行銷活動 |
| `tags` | `[Object]` | 商品標籤陣列 |
| `product_variants` | `[Object]` | 商品款式陣列 |
| `product_options` | `[Object]` | 商品選項陣列 |
| `pos_shop` | `Object` | 關聯 POS 商店資訊 |
| `photo_urls` | `[String]` | 商品照片網址陣列 |
| `photos` | `[Object]` | 商品照片陣列 |
| `channel` | `Object` | 商品通路 |
| `related_collections` | `[Object]` | 商品關聯群組陣列 |
| `branch_store` | `Object` | 商品所屬門市 |
| `temperature_types` | `[String]` | 溫層 |
| `searchable` | `Boolean` | 商品搜尋功能 |
| `google_product_category_id` | `Integer` | Google 產品類別 ID |
| `product_custom_fields` | `[Hash]` | 商品自訂欄位 |
| `seo_meta_tags` | `Hash` | SEO 標籤 |
| `required_customer_tags` | `[String]` | 限定購買會員標籤 |
| `created_at` | `String` | 建立時間 |
| `updated_at` | `String` | 更新時間 |

### Product Tag

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | 標籤名稱 |
| `category` | `Integer` | 標籤類型 |

### Product Photo

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `url` | `String` | 商品圖片網址 |
| `position` | `Integer` | 商品圖片順序 |

### Channel

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | 通路名稱 |

### Product Option

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | 商品選項名稱 |
| `position` | `Integer` | 商品選項順序 |
| `types` | `String` | 商品選項區別方式 |

### Custom Collection

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `title` | `String` | 自訂群組名稱 |
| `handle` | `String` | 自訂群組網址名 |
| `published` | `Boolean` | 是否公開 |
| `body_html` | `String` | 自訂群組內文 |
| `products_order_name` | `String` | 自訂群組商品排序方式 |
| `position` | `Integer` | 自訂群組順序 |
| `products` | `[Object]` | 群組商品陣列 |

### Special Collection

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `title` | `String` | 行銷活動群組名稱 |
| `handle` | `String` | 行銷活動群組網址名 |
| `published` | `Boolean` | 是否公開 |
| `start_date` | `String` | 群組起始時間 |
| `end_date` | `String` | 群組結束時間 |
| `body_html` | `String` | 行銷活動群組內文 |
| `position` | `Integer` | 行銷活動群組順序 |
| `special_collection_type` | `Object` | 行銷活動群組類型 |
| `type_rules` | `[Object]` | 行銷活動群組規則陣列 |
| `rest_include_discount` | `Boolean` | 剩餘商品是否計入折扣 |
| `products` | `[Object]` | 群組商品陣列 |

### Product Variant

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `product_id` | `Integer` | 關聯商品 ID |
| `name` | `String` | 關聯商品名稱 |
| `position` | `Integer` | 商品款式順序 |
| `price` | `Float` | 商品款式金額 |
| `cost` | `Float` | 商品款式成本 |
| `compare_at_price` | `Float` | 商品款式定價 |
| `meas` | `Float` | 商品款式材積 |
| `max_usable_bonus` | `Float` | 商品款式可用紅利上限 |
| `weight` | `Float` | 商品款式重量 |
| `option1` | `String` | 商品款式類型 1 |
| `option2` | `String` | 商品款式類型 2 |
| `option3` | `String` | 商品款式類型 3 |
| `inventory_management` | `Boolean` | 是否管理庫存 |
| `inventory_quantity` | `Integer` | 商品款式庫存數量 |
| `sold` | `Integer` | 商品款式已銷數量 |
| `safety_inventory_quantity` | `Integer` | 商品款式安全水位 |
| `inventory_policy` | `String` | 無庫存時是否允許購買 |
| `sku` | `String` | 商品款式 SKU |
| `qc` | `String` | 商品款式廠商編號 |
| `requires_shipping` | `Boolean` | 商品款式是否需要收貨地址 |
| `honeycomb_sync` | `Boolean` | 倉庫庫存同步 |
| `vendor` | `String` | 商品來源廠商 |
| `photo_urls` | `[String]` | 商品照片網址陣列 |
| `pim_infos` | `[Object]` | 商品 PIM 資訊陣列 |
| `created_at` | `String` | 建立時間 |
| `updated_at` | `String` | 更新時間 |

### PIM Info

| 欄位 | 類型 | 說明 |
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

CYBERBIZ 記載了此 payload 但未命名對應的事件；未曾觀察到任何傳送，SDK 也未定義購物車事件。此處僅為完整性而列出。

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `cart_add_items` | `[Object]` | 被加入購物車商品款式陣列 |
| `customer` | `Object` | 顧客資訊 |

### Cart Add Item

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `variant` | `Object` | 商品款式 |
| `quantity` | `Integer` | 加入數量 |

### Coupon

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `customer_id` | `Integer` | 關聯顧客 ID (若為全館優惠券則為空值) |
| `title` | `String` | 折價券名稱 |
| `code` | `String` | 折價券代碼 |
| `coupon_type_name` | `String` | 折價券類型：金額、百分比、免運 |
| `coupon_value` | `String` | 折價券折扣金額 |
| `order_price_threshold` | `Integer` | 折價券最低訂單總金額 |
| `start_date` | `String` | 折價券起始時間 |
| `end_date` | `String` | 折價券結束時間 |
| `concurrently_apply` | `String` | 是否與「任選折扣群組」或「全館活動」併用 |
| `usage_limit` | `Integer` | 優惠券使用次數上限 |
| `can_accumulate_bonus` | `Boolean` | 是否可累積紅利點數 |
| `usage_unlimited` | `Boolean` | 是否無使用次數限制 |
| `used_times` | `Integer` | 優惠券已使用次數 |
| `gift_order_id` | `Integer` | 獲得優惠券來源訂單 ID |
| `gift_days` | `Integer` | 獲得優惠券可用日數 |
| `account_usage_limit_enabled` | `Boolean` | 啟用限定各帳號使用次數 |
| `account_usage_limit` | `Integer` | 限定各帳號使用次數 |
| `restrict_strategy` | `String` | 與其他行銷活動併用限制類型：`unrestricted` (所有行銷活動商品皆可使用)、`restrict` (除指定活動商品外可使用)、`forbidden` (訂單包含指定活動商品時不得使用) |
| `restrict_campaigns` | `[String]` | 與其他行銷活動併用限制活動 |
| `tags` | `[String]` | 優惠券綁定商品標籤 |
| `pos_shop_ids` | `[Integer]` | 優惠券綁定門市 ID |
| `coupon_status` | `String` | 優惠券狀態：`no_start_use` (尚未啟用)、`used` (已全部使用完畢)、`has_expire_date` (可使用且有到期日)、`no_expire_date` (可使用且無到期日)、`expired` (已過期) |
| `gift_order_status` | `String` | 獲得優惠券來源訂單狀態：`closed` (已結案)、`open` (進行中)、`cancelled` (已取消) |
| `valid` | `Boolean` | 優惠券是否有效 |
| `customer_used_times` | `Integer` | 顧客已使用全館優惠券的次數 |
| `customer_usable` | `Boolean` | 顧客是否還有剩餘使用次數 |

### Customer VIP Level

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `customer_id` | `Integer` | 顧客 ID |
| `current_group` | `Object` | 當前 VIP 群組 |
| `current_level` | `Object` | 當前 VIP 層級 |
| `next_level` | `Object` | 下一個 VIP 層級 |
| `extra_info` | `Object` | 其他資訊 |

### Current VIP Group

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `name` | `String` | 群組名稱 |
| `position` | `Integer` | 位置 |
| `description_url` | `String` | 說明頁 |
| `customer_tags` | `[String]` | 顧客標籤 |
| `vip_group_levels` | `[Object]` | VIP 階層陣列 |

### VIP Level

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `position` | `Integer` | 位置 |
| `name` | `String` | 等級名稱 |
| `validity_days` | `Integer` | 效期 |
| `upgrade_condition_total_spent` | `Integer` | 升等條件：單筆消費總額 |
| `upgrade_condition_total_spent_in_validity_days` | `Integer` | 升等條件：會員效期內消費總額 |
| `renewal_condition_total_spent` | `Integer` | 續會條件：單筆消費總額 |
| `renewal_condition_total_spent_in_validity_days` | `Integer` | 續會條件：會員效期內消費總額 |
| `bonus_point_enabled` | `Boolean` | 啟用紅利倍數 |
| `bonus_point_threshold` | `Integer` | 紅利倍數：消費門檻 |
| `bonus_point_value` | `Integer` | 紅利倍數：贈送紅利 (點) |
| `bonus_point_expiry_days` | `Integer` | 紅利倍數：有效期限 (天) |
| `birth_gift_enabled` | `Boolean` | 啟用生日禮 |
| `birth_gift_name` | `String` | 生日禮：名稱 |
| `birth_gift_before_days` | `String` | 生日禮：提前贈送天數 |
| `birth_gift_setting` | `Object` | 生日禮：細部設定 |
| `upgrade_gift_enabled` | `Boolean` | 啟用升等禮 |
| `upgrade_gift_setting` | `Object` | 升等禮：細部設定 |
| `order_discount_enabled` | `Boolean` | 啟用訂單折扣 |
| `order_discount_value` | `Integer` | 訂單折扣 (百分比) |
| `order_discount_setting` | `Object` | 訂單折扣設定 |
| `free_shipping_enabled` | `Boolean` | 啟用訂單免運 |
| `free_shipping_setting` | `Object` | 訂單免運設定 |

### Extra Info

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `start_at` | `String` | 階層生效時間 |
| `end_at` | `String` | 階層失效時間 |
| `difference_of_total_spent_in_validity_days_for_renewal` | `Integer` | 續會條件差額（期間內累計金額） |
| `difference_of_total_spent_for_renewal` | `Integer` | 續會條件差額（單筆消費金額） |
| `difference_of_total_spent_in_validity_days_for_upgrade` | `Integer` | 升等條件差額（期間內累計金額） |
| `difference_of_total_spent_for_upgrade` | `Integer` | 升等條件差額（單筆消費金額） |

### Gift Setting

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `bonus` | `Object` | 紅利設定 |
| `coupon` | `Object` | 優惠券設定 |

### Bonus Setting

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `enabled` | `Boolean` | 啟用 |
| `value` | `Integer` | 贈送紅利 (點) |
| `expiry_days` | `Integer` | 有效期限 (天)，天數設為 0 時效期為無限期 |

### Coupon Setting

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `enabled` | `Boolean` | 啟用 |
| `presets` | `Object` | 設定 |

### Coupon Presets

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `coupon_type_id` | `Integer` | 優惠券種類：`1` (金額)、`2` (百分比) |
| `value` | `Integer` | 折扣 (金額/百分比) |
| `code` | `String` | 折價序號 |
| `order_price_threshold` | `Integer` | 最低消費門檻 |
| `usable_days` | `Integer` | 有效使用天數 |
| `usage_limit` | `Integer` | 張數 |
| `product_tags` | `[String]` | 商品標籤 |
| `restrict_strategy` | `String` | 與其他行銷活動併用限制類型 |
| `restrict_campaigns` | `[String]` | 與其他行銷活動併用限制活動 |

### App Uninstall

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `app_uuid` | `String` | App UUID in App Market |
| `app_version_uuid` | `String` | App Version UUID in App Market |
| `app_client_id` | `String` | App Client ID |

### Exchange Histories

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `created_at` | `String` | 換貨時間 |
| `price` | `Float` | 換貨金額 |
| `order_price_before` | `Integer` | 訂單換貨前金額 |
| `order_price_after` | `Integer` | 訂單換貨後金額 |
| `pos_shop_id` | `Integer` | 換貨 POS 店 ID |
| `pos_id` | `Integer` | 換貨 POS 機 ID |
| `payment_name` | `String` | 付款方式 |
| `payment_method` | `String` | 付款方式名稱 |
| `multiple_payment_infos` | `[Object]` | 多付款方式資訊陣列 |
| `line_items` | `[Object]` | 換貨商品資訊陣列 |
| `einvoice` | `Object` | 換貨發票資訊 |

### Exchange Line Item

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `product_variant_id` | `Integer` | 款式 ID |
| `name` | `String` | 名稱 |
| `sku` | `String` | SKU |
| `qc` | `String` | 廠商編號 |
| `price` | `Float` | 金額 |
| `quantity` | `Integer` | 數量 |

### Related Items

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `quantity` | `Integer` | 組合數量 |
| `items` | `[Object]` | 組合內容陣列 |

### Combo Item

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `product_id` | `Integer` | 商品 ID |
| `product_variant_id` | `Integer` | 款式 ID |
| `title` | `String` | 商品名稱 |
| `variant_title` | `String` | 款式名稱 |
| `sku` | `String` | 款式 SKU |
| `qc` | `String` | 款式廠商編號 |
| `vendor` | `String` | 商品來源廠商 |
| `price` | `Float` | 金額 |
| `cost` | `Float` | 成本 |
| `quantity` | `Integer` | 數量 |
| `combo_product_price_difference` | `Float` | 組合品價差 |

### Shipping Rate

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | 門市運費設定 ID |
| `name` | `String` | 快遞/物流名稱 |
| `min_order_subtotal` | `Integer` | 消費金額 |
| `price` | `Integer` | 運費 |
| `payments` | `[Object]` | 允許的付款方式陣列 |

### Payment

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | 付款方式 ID |
| `name` | `String` | 付款方式名稱 |

### Special Collection Type

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `name` | `String` | 類型名稱 |
| `code` | `String` | 類型代碼 |

### Type Rules

| 欄位 | 類型 | 說明 |
| --- | --- | --- |
| `id` | `Integer` | ID |
| `quantity` | `Integer` | 規則商品數量 |
| `price` | `String` | 規則折扣金額 |
| `percentage` | `Integer` | 規則折扣百分比 |

## 範例

以下範例皆為合成資料。簽章 header 是以 App Secret `example-app-secret` 對所示的 JSON 位元組（兩空格縮排、結尾換行的美化輸出）計算，因此可用來測試驗證程式。

### `customers/create`

會員註冊。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

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

會員修改。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

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

會員新增第三方登入 UID（參考文件未另記 payload；送出的是會員物件）。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

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

會員的第三方登入 UID 更新（會員物件）。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `tJX7CKyPFfxr+lo3tbNt4pMLJpuObJG63lrdMFbHjvE=`

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

獲得紅利。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

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

使用紅利。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

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

刪除紅利。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

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

商品評論審核通過發送紅利。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `sv1JAn3l1DIKx2aX+SOBMNLSh0Nfof4XzZtySR8h2mg=`

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

訂單成立。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `LlF157q3tPHdwV0a6FtgYPceCbMZWleHHGMm2wWVkdY=`

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

訂單付款。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `pOF3PZTDy/yr2vchjcETNEVBeozDHwB9LqQfGeogPNY=`

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

準備出貨。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `lyeWCsXSCJ5TgxzV6KR2zd76wA9PoXDafOu/n2kUCqc=`

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

訂單出貨。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `TDfpyVgVlAu9+rllEyDGYs1KvfV6ARx32CioabNd9cc=`

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

訂單收貨。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `dbRvtVoagQHAMTvatJCeLlsZJTTcsZKUNo90HOgCk6c=`

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

訂單到店。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `ptH5gKV6Bn9PaFs9Lm2WO7+UNl8AxGxUXc2gtO30Gkg=`

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

逾期未取。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `1/SWzXaoXQLi5mAYjizIKY1GoeUJdnf5qYV1XHGX/vw=`

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

訂單取消。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `CkuhYrBrU2db3NGeyI4FoIEbrFWMXmj4Xf4T5SrbG3I=`

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

訂單退貨。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `61xT2AFYCd+mdXqwBK5f+osPJHzP1ISRoYTgCHSnLWk=`

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

訂單部分退貨。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `3IkOpC4fk/rkEW4QE84DKcNBvoGR11aUhBEA5jXB2HE=`

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

訂單退款。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `bAzf3gzQVSIM+8m4QS+63PHhWB3VsUiiSG6RgBVTG7E=`

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

訂單部分退款。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `w0cqzb6VUe5bzt2XkaWxXugYk6Phz8C9TVUskvJqtRw=`

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

訂單結案。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `0p+9GZ6KygPO8MXK4slFpI/32CQmsgBHS5jp5VudN/s=`

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

退貨申請。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `Rbi5HFHPLQ6X7tPQZ9aOZe8lw8F35mNesBmpQV4JdOA=`

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

訂單開啟。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `Q+uwtFdQhDc51TMFajZcVznH+ilK7Qlz5UmyDfD55BY=`

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

快速到貨訂單更新。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `Fprm+PGTjtYvzrV/K68a88vUBR0DacEuQDYq5JbePOk=`

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

商品新增。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `n1sti7OYgl1U8qmtS9qas4jBM3zW65apyvtyFH1x0Ik=`

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

商品修改。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `n1sti7OYgl1U8qmtS9qas4jBM3zW65apyvtyFH1x0Ik=`

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

商品刪除。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `n1sti7OYgl1U8qmtS9qas4jBM3zW65apyvtyFH1x0Ik=`

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

款式新增。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `7JeDR9MZBQT1/76led7rIU/tIePOXypee5ePh9YFdX8=`

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

款式修改。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `7JeDR9MZBQT1/76led7rIU/tIePOXypee5ePh9YFdX8=`

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

款式刪除。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `7JeDR9MZBQT1/76led7rIU/tIePOXypee5ePh9YFdX8=`

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

新增優惠券。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `BxAbAWMUhh25eF4H3Ly36Ijjr+z2EFusgqdlskaJeJA=`

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

使用優惠券。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `BxAbAWMUhh25eF4H3Ly36Ijjr+z2EFusgqdlskaJeJA=`

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

刪除優惠券。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `BxAbAWMUhh25eF4H3Ly36Ijjr+z2EFusgqdlskaJeJA=`

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

會員 VIP 等級更新。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `jwLpnseZccxKvzf2iYK58yLuDGXG3r5+2cOjLvoyjXg=`

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

App 解除安裝。

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

同一簽章的 Base64 形式（CYBERBIZ 文件所描述的格式）： `VDNU3zSQWH4JG52fFrS4gSt8qVky0J+qyGkm9vMzukU=`

```json
{
  "app_uuid": "00000000-0000-4000-8000-000000000001",
  "app_version_uuid": "00000000-0000-4000-8000-000000000002",
  "app_client_id": "example-app-client-id"
}
```

## Release Notes

### 1.0.1 (2026-10-02)

- Postman collection 說明不再指向 repository 內部的檔案。

### 1.0.0 (2026-09-08)

- 初版：由 CYBERBIZ v1 swagger、v1/v2 Postman collection、Notion v2 參考頁面與 webhook 參考文件產生。
- 套用實際觀察的修正：可為 null 的欄位、金額為浮點數、order_number 為整數、文件記為物件的陣列、swagger 缺少的欄位、Bearer 驗證。
- 每個操作與事件皆附有由 Golden File 衍生的合成範例。
