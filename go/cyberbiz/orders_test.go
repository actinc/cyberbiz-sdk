package cyberbiz

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

// ordersRecorded captures the last request a recording client received.
type ordersRecorded struct {
	Method string
	Path   string
	Query  url.Values
	Body   string
}

// ordersRecordingClient returns a client whose server records each request and
// answers with reply.
func ordersRecordingClient(t *testing.T, reply string) (*Client, *ordersRecorded) {
	t.Helper()
	rec := &ordersRecorded{}
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		rec.Method, rec.Path, rec.Query, rec.Body = r.Method, r.URL.Path, r.URL.Query(), string(body)
		_, _ = w.Write([]byte(reply))
	})
	return c, rec
}

// ordersRequestCase is one request-shape expectation.
type ordersRequestCase struct {
	name   string
	reply  string
	call   func(c *Client) error
	method string
	path   string
	query  url.Values
	body   string
}

func runOrdersRequestCases(t *testing.T, cases []ordersRequestCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reply := tc.reply
			if reply == "" {
				reply = "{}"
			}
			c, rec := ordersRecordingClient(t, reply)
			if err := tc.call(c); err != nil {
				t.Fatalf("call: %v", err)
			}
			if rec.Method != tc.method || rec.Path != tc.path {
				t.Errorf("got %s %s, want %s %s", rec.Method, rec.Path, tc.method, tc.path)
			}
			want := tc.query
			if want == nil {
				want = url.Values{}
			}
			if !reflect.DeepEqual(rec.Query, want) {
				t.Errorf("query = %v, want %v", rec.Query, want)
			}
			if rec.Body != tc.body {
				t.Errorf("body = %s, want %s", rec.Body, tc.body)
			}
		})
	}
}

func ordersTaipei(y int, m time.Month, d, hh, mm, ss int) Time {
	return NewTime(time.Date(y, m, d, hh, mm, ss, 0, Taipei))
}

func TestOrderGoldenDetail(t *testing.T) {
	var o Order
	decodeGolden(t, "v1/GET_v1_orders_{id}.json", &o)

	if o.ID != 56943817 || o.OrderNumber != 1101 || o.OrderName != "#1101" {
		t.Errorf("identity = %d %d %q", o.ID, o.OrderNumber, o.OrderName)
	}
	if o.SubtotalPrice != MoneyFromInt(9999) {
		t.Errorf("subtotal = %v", o.SubtotalPrice)
	}
	if !o.CreatedAt.Equal(ordersTaipei(2026, 9, 7, 12, 35, 5).Time) {
		t.Errorf("created_at = %v", o.CreatedAt)
	}
	if o.Customer == nil || o.Customer.ID != 42058619 || o.Customer.Status != "enabled" {
		t.Fatalf("customer = %+v", o.Customer)
	}
	if o.Customer.Birthday.String() != "1990-01-01" || o.Customer.BonusRemain != MoneyFromInt(5000) {
		t.Errorf("customer birthday/bonus = %v %v", o.Customer.Birthday, o.Customer.BonusRemain)
	}
	if o.Customer.ConfirmedAt == nil || o.Customer.Address == nil || o.Customer.Address.DetailAddress == nil {
		t.Errorf("customer nested = %+v", o.Customer)
	}
	if o.Buyer == nil || o.Buyer.Email != "redacted@example.com" || o.Buyer.Mobile != "" {
		t.Errorf("buyer = %+v", o.Buyer)
	}
	if o.Receiver == nil || o.Receiver.Phone != "0912345678" || o.Receiver.CVSStoreID != "" {
		t.Errorf("receiver = %+v", o.Receiver)
	}
	if o.BillingAddress == nil || o.BillingAddress.DetailAddress.Address2 != "" {
		t.Errorf("billing = %+v", o.BillingAddress)
	}
	assertOrderLineItems(t, o.LineItems)
	if o.ShippingVendor == nil || o.ShippingVendor.Type != "custom" {
		t.Errorf("shipping vendor = %+v", o.ShippingVendor)
	}
	if !o.DeliveryDate.IsZero() || o.DeliveryTime != 0 || o.Delegate != "redacted@example.com" {
		t.Errorf("delivery = %v %d %q", o.DeliveryDate, o.DeliveryTime, o.Delegate)
	}
	assertOrderPrices(t, o.Prices)
	if o.Einvoice != nil || o.BranchStore != nil || o.UTMTracking != nil || o.CustomerCancelReasonDetail != nil {
		t.Error("null objects should decode to nil")
	}
	if o.Statuses == nil || o.Statuses.OrderStatus != OrderStatusOpen ||
		o.Statuses.FinancialStatus != FinancialStatusPending ||
		o.Statuses.FulfillmentStatus != FulfillmentStatusUnshipped ||
		o.Statuses.ReturnStatus != ReturnStatusNoNeed {
		t.Errorf("statuses = %+v", o.Statuses)
	}
	if o.Timings == nil || o.Timings.ClosedAt != nil || o.Timings.CancelledAt != nil {
		t.Errorf("timings = %+v", o.Timings)
	}
	if o.Timings.ConfirmedAt == nil || !o.Timings.ConfirmedAt.Equal(ordersTaipei(2026, 9, 7, 12, 35, 5).Time) {
		t.Errorf("confirmed_at = %v", o.Timings.ConfirmedAt)
	}
	if o.PosInfo == nil || o.PosInfo.PosID != 0 || o.PosInfo.PosInfo != "" {
		t.Errorf("pos_info = %+v", o.PosInfo)
	}
	if o.LinkedOrderInfo == nil || o.LinkedOrderInfo.Source != "非導購訂單" {
		t.Errorf("linked_order_info = %+v", o.LinkedOrderInfo)
	}
	if len(o.Tags) != 0 || len(o.ReturnHistories) != 0 || len(o.ExchangeHistories) != 0 || len(o.SerialNumbers) != 0 {
		t.Error("empty arrays should decode to empty slices")
	}
	if o.FromDevice != "桌機" || o.WarehouseTypeID != 0 || o.Token != "" {
		t.Errorf("misc = %q %d %q", o.FromDevice, o.WarehouseTypeID, o.Token)
	}
}

func assertOrderLineItems(t *testing.T, items []LineItem) {
	t.Helper()
	if len(items) != 1 {
		t.Fatalf("line items = %d", len(items))
	}
	li := items[0]
	if li.ID != 126360067 || li.ProductVariantID != 83653695 || li.Title != "哥本組" {
		t.Errorf("line item = %+v", li)
	}
	if li.Price != MoneyFromInt(9999) || li.Cost != 0 || li.Quantity != 1 || li.Weight != 0 {
		t.Errorf("line item amounts = %v %v %d %v", li.Price, li.Cost, li.Quantity, li.Weight)
	}
	if li.ItemType != OrderItemTypeNormal || li.TaxTypeID != TaxTypeInclusive || li.ReturnStatus != "" {
		t.Errorf("line item enums = %q %q %q", li.ItemType, li.TaxTypeID, li.ReturnStatus)
	}
	if len(li.Discounts) != 0 || len(li.CustomFields) != 0 {
		t.Errorf("discounts/custom fields = %v %v", li.Discounts, li.CustomFields)
	}
	if len(li.RelatedItems) != 1 || li.RelatedItems[0].Quantity != 1 || len(li.RelatedItems[0].Items) != 2 {
		t.Fatalf("related items = %+v", li.RelatedItems)
	}
	ri := li.RelatedItems[0].Items[1]
	if ri.SKU != "SKU-d76be1" || ri.Price != MoneyFromInt(14990) ||
		ri.ComboProductPriceDifference != MoneyFromInt(9990) ||
		len(ri.ComboProductPriceDiffDetails) != 1 || ri.ComboProductPriceDiffDetails[0] != MoneyFromInt(9990) {
		t.Errorf("related item = %+v", ri)
	}
}

func assertOrderPrices(t *testing.T, p *Prices) {
	t.Helper()
	if p == nil || p.TotalPrice != MoneyFromInt(9999) || p.ShippingRatePrice != 0 {
		t.Fatalf("prices = %+v", p)
	}
	d := p.Discounts
	if d == nil || d.ShopDiscount != nil || d.CouponDiscount != nil || len(d.CouponDiscounts) != 0 {
		t.Errorf("discounts = %+v", d)
	}
	if d.VIPDiscount != 0 || d.BonusConsumed != 0 || d.ThirdPartyDiscount != 0 {
		t.Errorf("discount amounts = %+v", d)
	}
}

func TestOrderGoldenList(t *testing.T) {
	var orders []Order
	decodeGolden(t, "v1/GET_v1_orders.json", &orders)
	if len(orders) != 2 {
		t.Fatalf("orders = %d", len(orders))
	}
	if orders[0].Token != "REDACTED" || orders[1].OrderNumber != 1102 || orders[1].ID != 56944000 {
		t.Errorf("list items = %+v", orders)
	}
	if orders[1].Customer == nil || orders[1].Customer.MobileSMSConfirmedAt == nil {
		t.Errorf("customer = %+v", orders[1].Customer)
	}
}

func TestOrderRelatedGolden(t *testing.T) {
	var returns []OrderReturn
	decodeGolden(t, "v1/GET_v1_orders_{id}_returns.json", &returns)
	if len(returns) != 0 {
		t.Errorf("returns = %+v", returns)
	}
	var txs []OrderTransaction
	decodeGolden(t, "v1/GET_v1_orders_{id}_transactions.json", &txs)
	if len(txs) != 0 {
		t.Errorf("transactions = %+v", txs)
	}
	var ids []OrderNumberID
	decodeGolden(t, "v1/GET_v1_orders_get_order_id.json", &ids)
	if len(ids) != 2 || ids[0].OrderNumber != 1102 || ids[0].OrderID != 56944000 || ids[1].OrderID != 56943817 {
		t.Errorf("order ids = %+v", ids)
	}
}

func TestOrderErrorGolden(t *testing.T) {
	cases := []struct {
		golden string
		status int
		want   error
		msg    string
		call   func(c *Client) error
	}{
		{"errors/GET_v1_orders_{id}.json", 404, ErrNotFound, "無此資源", func(c *Client) error {
			_, _, err := c.Orders.Get(testCtx, 1)
			return err
		}},
		{"errors/GET_v1_orders_{id}_fulfillments_{id}.json", 404, ErrNotFound, "無此資源", func(c *Client) error {
			_, _, err := c.Orders.GetFulfillment(testCtx, 1, 2)
			return err
		}},
		{"errors/GET_v1_order_etickets.json", 403, ErrForbidden, "無「電子票券」功能，請聯絡您的開店顧問", func(c *Client) error {
			_, err := c.Orders.ListEtickets(testCtx, nil)
			return err
		}},
	}
	for _, tc := range cases {
		body := readGolden(t, tc.golden)
		c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write(body)
		})
		err := tc.call(c)
		var apiErr *APIError
		if !errors.Is(err, tc.want) || !errors.As(err, &apiErr) {
			t.Fatalf("%s: got %v", tc.golden, err)
		}
		if len(apiErr.Messages) != 1 || apiErr.Messages[0] != tc.msg {
			t.Errorf("%s: messages = %q", tc.golden, apiErr.Messages)
		}
	}
}

func TestOrdersGoldenServerRoundTrip(t *testing.T) {
	c := goldenServer(t)

	o, resp, err := c.Orders.Get(testCtx, 56943817)
	if err != nil {
		t.Fatal(err)
	}
	if o.ID != 56943817 || resp.StatusCode != 200 {
		t.Errorf("get = %d %d", o.ID, resp.StatusCode)
	}

	page, err := c.Orders.List(testCtx, &OrderListOptions{ListOptions: ListOptions{PerPage: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[0].OrderNumber != 1101 {
		t.Errorf("list = %+v", page.Items)
	}
	want := Pagination{Page: 1, Total: 1, TotalPages: 1}
	if page.Pagination != want {
		t.Errorf("pagination = %+v, want %+v", page.Pagination, want)
	}

	var count int
	for _, err := range c.Orders.All(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("all = %d", count)
	}

	ids, _, err := c.Orders.LookupIDs(testCtx, []int64{1101, 1102})
	if err != nil || len(ids) != 2 {
		t.Errorf("lookup ids = %v %v", ids, err)
	}
}

func TestOrdersRequestShapes(t *testing.T) {
	str := func(s string) *string { return &s }
	i := func(n int) *int { return &n }
	b := func(v bool) *bool { return &v }
	runOrdersRequestCases(t, []ordersRequestCase{
		{name: "list filters", reply: "[]", call: func(c *Client) error {
			_, err := c.Orders.List(testCtx, &OrderListOptions{
				ListOptions:         ListOptions{Page: 2, PerPage: 10},
				StartTime:           ordersTaipei(2026, 1, 1, 0, 0, 0),
				UpdatedAtEndTime:    ordersTaipei(2026, 1, 31, 23, 59, 59),
				Statuses:            []OrderStatus{OrderStatusOpen, OrderStatusClosed},
				FinancialStatuses:   []FinancialStatus{FinancialStatusPaid},
				FulfillmentStatuses: []FulfillmentStatus{FulfillmentStatusUnshipped, FulfillmentStatusPreparing},
				ReturnStatuses:      []ReturnStatus{ReturnStatusNoNeed},
				Tags:                []string{"vip", "rush"},
				ExcludedTags:        []string{"test"},
				DataSource:          DataSourceEC,
				Vendor:              "acme",
			})
			return err
		}, method: "GET", path: "/v1/orders", query: url.Values{
			"page": {"2"}, "per_page": {"10"},
			"start_time": {"2026-01-01 00:00:00"}, "updated_at_end_time": {"2026-01-31 23:59:59"},
			"statuses": {"open,closed"}, "financial_statuses": {"paid"},
			"fulfillment_statuses": {"unshipped,preparing"}, "return_statuses": {"no_need"},
			"tags": {"vip,rush"}, "excluded_tags": {"test"}, "data_source": {"ec"}, "vendor": {"acme"},
		}},
		{name: "list nil options", reply: "[]", call: func(c *Client) error {
			_, err := c.Orders.List(testCtx, nil)
			return err
		}, method: "GET", path: "/v1/orders"},
		{name: "get", call: func(c *Client) error {
			_, _, err := c.Orders.Get(testCtx, 42)
			return err
		}, method: "GET", path: "/v1/orders/42"},
		{name: "lookup ids", reply: "[]", call: func(c *Client) error {
			_, _, err := c.Orders.LookupIDs(testCtx, []int64{1101, 1102})
			return err
		}, method: "GET", path: "/v1/orders/get_order_id", query: url.Values{"order_numbers": {"1101,1102"}}},
		{name: "update omits unset", call: func(c *Client) error {
			_, _, err := c.Orders.Update(testCtx, 42, &OrderUpdateRequest{})
			return err
		}, method: "PUT", path: "/v1/orders/42", body: `{}`},
		{name: "update sends explicit zeros", call: func(c *Client) error {
			_, _, err := c.Orders.Update(testCtx, 42, &OrderUpdateRequest{
				Note: str(""), DeliveryDate: NewDate(2026, 9, 10), DeliveryTime: i(0)})
			return err
		}, method: "PUT", path: "/v1/orders/42", body: `{"note":"","delivery_date":"2026-09-10","delivery_time":0}`},
		{name: "update tags", call: func(c *Client) error {
			_, _, err := c.Orders.UpdateTags(testCtx, 42, []string{"a", "b"})
			return err
		}, method: "PUT", path: "/v1/orders/42/tags", body: `{"tags":["a","b"]}`},
		{name: "update warehouse note", call: func(c *Client) error {
			_, _, err := c.Orders.UpdateWarehouseNote(testCtx, 42, "已交寄")
			return err
		}, method: "PUT", path: "/v1/orders/42/update_warehouse_note", body: `{"warehouse_note":"已交寄"}`},
		{name: "update status", call: func(c *Client) error {
			_, _, err := c.Orders.UpdateStatus(testCtx, 42, OrderStatusClosed)
			return err
		}, method: "PUT", path: "/v1/orders/42/update_status", body: `{"status":"closed"}`},
		{name: "update financial status", call: func(c *Client) error {
			_, _, err := c.Orders.UpdateFinancialStatus(testCtx, 42)
			return err
		}, method: "PUT", path: "/v1/orders/42/update_financial_status"},
		{name: "mark unshipped", call: func(c *Client) error {
			_, _, err := c.Orders.MarkUnshipped(testCtx, 42)
			return err
		}, method: "PUT", path: "/v1/orders/42/unshipped"},
		{name: "mark preparing", call: func(c *Client) error {
			_, _, err := c.Orders.MarkPreparing(testCtx, 42)
			return err
		}, method: "PUT", path: "/v1/orders/42/preparing"},
		{name: "cancel minimal", call: func(c *Client) error {
			_, _, err := c.Orders.Cancel(testCtx, 42, &OrderCancelRequest{CancelReason: CancelReasonCustomer})
			return err
		}, method: "PUT", path: "/v1/orders/42/cancelled", body: `{"cancel_reason":"customer"}`},
		{name: "cancel full", call: func(c *Client) error {
			_, _, err := c.Orders.Cancel(testCtx, 42, &OrderCancelRequest{
				CancelReason: CancelReasonOther, CancelReasonDetail: CustomerCancelReasonOther,
				OtherCancelReasonDetail: "改天再買", Email: b(false), SetWarning: b(true),
				RefundShopdotcom: new(Money)})
			return err
		}, method: "PUT", path: "/v1/orders/42/cancelled",
			body: `{"cancel_reason":"other","cancel_reason_detail":"other","other_cancel_reason_detail":"改天再買","email":false,"set_warning":true,"refund_shopdotcom":0}`},
		{name: "manual return", call: func(c *Client) error {
			_, _, err := c.Orders.ManualReturn(testCtx, 42, ManualReturnDone)
			return err
		}, method: "PUT", path: "/v1/orders/42/manual_return", body: `{"operation":"manual_return_done"}`},
		{name: "change to custom shipping", call: func(c *Client) error {
			_, _, err := c.Orders.ChangeToCustomShipping(testCtx, 42)
			return err
		}, method: "PUT", path: "/v1/orders/42/change_to_custom_shipping"},
		{name: "list transactions", reply: "[]", call: func(c *Client) error {
			_, err := c.Orders.ListTransactions(testCtx, 42, &ListOptions{Page: 3})
			return err
		}, method: "GET", path: "/v1/orders/42/transactions", query: url.Values{"page": {"3"}}},
		{name: "create transaction", call: func(c *Client) error {
			_, _, err := c.Orders.CreateTransaction(testCtx, 42, &OrderTransactionCreateRequest{
				Kind: TransactionKindCapture, PaidType: TransactionPaidTypeManual})
			return err
		}, method: "POST", path: "/v1/orders/42/transactions", body: `{"kind":"capture","paid_type":"manual"}`},
		{name: "list returns", reply: "[]", call: func(c *Client) error {
			_, err := c.Orders.ListReturns(testCtx, 42)
			return err
		}, method: "GET", path: "/v1/orders/42/returns"},
		{name: "list etickets", reply: "[]", call: func(c *Client) error {
			_, err := c.Orders.ListEtickets(testCtx, &OrderEticketListOptions{
				ListOptions: ListOptions{PerPage: 20}, SearchColumn: EticketSearchTicketNumber, Q: "ABC"})
			return err
		}, method: "GET", path: "/v1/order_etickets",
			query: url.Values{"per_page": {"20"}, "search_column": {"ticket_number"}, "q": {"ABC"}}},
		{name: "redeem eticket", call: func(c *Client) error {
			_, _, err := c.Orders.RedeemEticket(testCtx, &EticketRedeemRequest{
				TicketNumber: "TK-1", RedeemQuantity: 2, UserID: 7, BranchStoreID: 9})
			return err
		}, method: "POST", path: "/v1/order_etickets/submit_redeem",
			body: `{"ticket_number":"TK-1","redeem_quantity":2,"user_id":7,"branch_store_id":9}`},
	})
}

func TestOrdersAllStartsAtFullPages(t *testing.T) {
	c, rec := ordersRecordingClient(t, "[]")
	for range c.Orders.All(testCtx, &OrderListOptions{Statuses: []OrderStatus{OrderStatusOpen}}) {
	}
	want := url.Values{"page": {"1"}, "per_page": {"50"}, "statuses": {"open"}}
	if !reflect.DeepEqual(rec.Query, want) {
		t.Errorf("query = %v, want %v", rec.Query, want)
	}
}

func TestOrdersNullWriteReplyYieldsNilOrder(t *testing.T) {
	c, _ := ordersRecordingClient(t, "null")
	o, resp, err := c.Orders.MarkPreparing(testCtx, 42)
	if err != nil || o != nil || resp == nil {
		t.Errorf("got %v %v %v", o, resp, err)
	}
}
