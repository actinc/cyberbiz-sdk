package cyberbiz

import (
	"testing"
)

const affiliatesSampleBody = `[{"status_code":"create_succeed","uid":"102d7ca1","cid":"102d7ca1","vendor_name":"vendor","order":{
	"customer_id":12,"subtotal_price":300.0,"created_at":"2023-05-17 14:51:12","updated_at":"2023-05-17 14:51:12",
	"order_number":1020,"order_name":"#1020","buyer":{"email":"demo@example.com","mobile":"0987654321"},
	"line_items":[{"id":91,"product_id":21,"product_variant_id":77,"title":"AIR Pods Pro","variant_title":"","sku":null,"qc":null,
		"vendor":null,"price":300.0,"cost":null,"quantity":1,"item_type":"normal","discount_name":"",
		"discounts":[{"position":1,"id":3,"code":"shop","name":"全館","discount":433.0}],
		"total_price_before_discounts":300.0,"total_discount":0,"total_price_after_discounts":300.0,
		"tax_type_id":"inclusive_tax","bonus_redemption_price":null,"related_items":[]}],
	"shipping_type":"cyberbiz","shipping_name":"自取","delivery_date":null,"delivery_time":0,"delegate":null,
	"prices":{"total_line_items_price":300.0,"shipping_rate_price":10.0,"discounts":{"special_collection_discount":0,
		"vip_discount":0.0,"shop_discount":null,"coupon_discount":null,"coupon_discounts":[],"bonus_consumed":0.0,
		"vip_shipping_discount":0,"coupon_shipping_discount":0,"price_discount":0},"total_price":310.0},
	"transaction_number":null,"merchant_trade_no":"S13#1020",
	"statuses":{"order_status":"open","financial_status":"pending","fulfillment_status":"unshipped","return_status":"no_need"},
	"timings":{"request_return_at":null,"return_at":null,"refund_at":null,"closed_at":null,"cancelled_at":null,"expired_at":null,
		"confirmed_at":"2023-05-17 14:51:12"},
	"return_histories":[],"note":"","total_bonus_redemption_price":0.0,"linked_order_info":{"source":"非導購訂單"},
	"shipping_status":null,"extra_info":null,"from_device":"桌機"}}]`

func TestAffiliateOrdersGolden(t *testing.T) {
	var out []AffiliateOrder
	decodeGolden(t, "v2/GET_v2_affiliate_vendor_orders.json", &out)
	if out == nil || len(out) != 0 {
		t.Errorf("orders = %#v", out)
	}
	c := goldenServer(t)
	page, err := c.Affiliates.List(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 || page.Pagination.Page != 1 {
		t.Errorf("page = %+v", page)
	}
}

func TestAffiliateOrderDecodesNotionSample(t *testing.T) {
	c, _ := New("tok")
	var out []AffiliateOrder
	if err := c.decode([]byte(affiliatesSampleBody), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d", len(out))
	}
	a := out[0]
	if a.StatusCode != "create_succeed" || a.VendorName != "vendor" || a.Order == nil {
		t.Fatalf("entry = %+v", a)
	}
	o := a.Order
	if o.OrderNumber != 1020 || o.OrderName != "#1020" || o.SubtotalPrice != MoneyFromInt(300) || o.Buyer.Mobile != "0987654321" {
		t.Errorf("order = %+v", o)
	}
	if len(o.LineItems) != 1 || o.LineItems[0].Price != MoneyFromInt(300) || o.LineItems[0].Cost != 0 || o.LineItems[0].SKU != "" {
		t.Errorf("line items = %+v", o.LineItems)
	}
	if o.LineItems[0].Discounts[0].Discount != MoneyFromInt(433) || o.LineItems[0].RelatedItems == nil {
		t.Errorf("line item = %+v", o.LineItems[0])
	}
	if o.Prices == nil || o.Prices.TotalPrice != MoneyFromInt(310) || o.Prices.Discounts.ShopDiscount != nil || o.Prices.Discounts.CouponDiscounts == nil {
		t.Errorf("prices = %+v", o.Prices)
	}
	if o.Statuses.OrderStatus != OrderStatusOpen || o.Statuses.FinancialStatus != FinancialStatusPending ||
		o.Statuses.FulfillmentStatus != FulfillmentStatusUnshipped || o.Statuses.ReturnStatus != ReturnStatusNoNeed {
		t.Errorf("statuses = %+v", o.Statuses)
	}
	if o.Timings.ClosedAt != nil || o.Timings.ConfirmedAt == nil || o.Timings.ConfirmedAt.String() != "2023-05-17 14:51:12" {
		t.Errorf("timings = %+v", o.Timings)
	}
	if !o.DeliveryDate.IsZero() || o.LinkedOrderInfo.Source != "非導購訂單" || o.FromDevice != "桌機" || string(o.ExtraInfo) != "null" {
		t.Errorf("order = %+v", o)
	}
}

func TestAffiliateRequests(t *testing.T) {
	c, call := discountsSpyClient(t, 200, affiliatesSampleBody)
	start, _ := ParseTime("2023-11-30 00:00:00")
	end, _ := ParseTime("2023-11-30 23:59:59")
	page, err := c.Affiliates.List(testCtx, &AffiliateOrderListOptions{
		ListOptions: ListOptions{Page: 1, PerPage: 50}, StartTime: start, EndTime: end,
		ClosedAtStartTime: start, Statuses: []OrderStatus{OrderStatusOpen, OrderStatusClosed},
	})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v2/affiliate_vendor_orders")
	q := call.Query
	if q.Get("start_time") != "2023-11-30 00:00:00" || q.Get("end_time") != "2023-11-30 23:59:59" ||
		q.Get("closed_at_start_time") != "2023-11-30 00:00:00" || q.Has("closed_at_end_time") ||
		q.Get("statuses") != "open,closed" || q.Get("per_page") != "50" {
		t.Errorf("query = %v", q)
	}
	if len(page.Items) != 1 || page.Items[0].Order.OrderNumber != 1020 {
		t.Errorf("page = %+v", page.Items)
	}

	var n int
	for _, err := range c.Affiliates.All(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 1 || call.Query.Get("page") != "1" {
		t.Errorf("walked %d, query = %v", n, call.Query)
	}
}
