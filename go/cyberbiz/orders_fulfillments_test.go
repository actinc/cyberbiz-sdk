package cyberbiz

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestFulfillmentGolden(t *testing.T) {
	var list []Fulfillment
	decodeGolden(t, "v1/GET_v1_orders_{id}_fulfillments.json", &list)
	if len(list) != 0 {
		t.Errorf("fulfillments = %+v", list)
	}

	var f Fulfillment
	decodeGolden(t, "v1/GET_v1_orders_{id}_fulfillments_{id}.json", &f)
	if f.ID != 23064624 || f.TrackingCompany != TrackingCompanyOther || f.TrackingNumber != "TEST-TRK-1086" {
		t.Errorf("fulfillment = %+v", f)
	}
	if f.Status != FulfillmentStatusFulfilled || f.CVSShippingType != "" || f.TrackingURL != "" {
		t.Errorf("fulfillment enums = %q %q %q", f.Status, f.CVSShippingType, f.TrackingURL)
	}
	if !f.FulfilledAt.Equal(ordersTaipei(2026, 6, 3, 21, 39, 47).Time) || !f.ReceivedAt.IsZero() {
		t.Errorf("timestamps = %v %v", f.FulfilledAt, f.ReceivedAt)
	}
	if len(f.LineItems) != 1 {
		t.Fatalf("line items = %d", len(f.LineItems))
	}
	li := f.LineItems[0]
	if li.SKU != "SKU-9c86d0" || li.Price != MoneyFromInt(888) || li.TotalPriceAfterDiscounts != MoneyFromInt(888) {
		t.Errorf("line item = %+v", li)
	}
	if len(li.RelatedItems) != 1 || len(li.RelatedItems[0].Items) != 3 {
		t.Fatalf("related items = %+v", li.RelatedItems)
	}
	ri := li.RelatedItems[0].Items[0]
	if ri.Cost != MoneyFromInt(100) || ri.ComboProductPriceDifference != MoneyFromInt(94) || ri.VariantTitle != "女休閒長T - " {
		t.Errorf("related item = %+v", ri)
	}
}

func TestFulfillmentsGoldenServerRoundTrip(t *testing.T) {
	c := goldenServer(t)
	f, _, err := c.Orders.GetFulfillment(testCtx, 56943817, 23064624)
	if err != nil || f.ID != 23064624 {
		t.Fatalf("get fulfillment = %+v %v", f, err)
	}
	page, err := c.Orders.ListFulfillments(testCtx, 56943817, nil)
	if err != nil || len(page.Items) != 0 || page.Pagination.Total != 1 {
		t.Errorf("list fulfillments = %+v %v", page, err)
	}
	returns, err := c.Orders.ListReturns(testCtx, 56943817)
	if err != nil || len(returns.Items) != 0 {
		t.Errorf("list returns = %+v %v", returns, err)
	}
	txs, err := c.Orders.ListTransactions(testCtx, 56943817, nil)
	if err != nil || len(txs.Items) != 0 {
		t.Errorf("list transactions = %+v %v", txs, err)
	}
}

func TestIDList(t *testing.T) {
	c, rec := ordersRecordingClient(t, "{}")
	_, _, err := c.Orders.CreateExpressDeliveryShipping(testCtx, 1, []int64{3, 14, 15})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Body != `{"line_item_ids":"3,14,15"}` {
		t.Errorf("body = %s", rec.Body)
	}
	c, rec = ordersRecordingClient(t, "{}")
	if _, _, err := c.Orders.CreateExpressDeliveryShipping(testCtx, 1, nil); err != nil {
		t.Fatal(err)
	}
	if rec.Body != `{"line_item_ids":""}` {
		t.Errorf("empty body = %s", rec.Body)
	}
}

func TestFulfillmentRequestShapes(t *testing.T) {
	b := func(v bool) *bool { return &v }
	charge := MoneyFromInt(120)
	runOrdersRequestCases(t, []ordersRequestCase{
		{name: "list fulfillments", reply: "[]", call: func(c *Client) error {
			_, err := c.Orders.ListFulfillments(testCtx, 42, &ListOptions{PerPage: 5})
			return err
		}, method: "GET", path: "/v1/orders/42/fulfillments", query: url.Values{"per_page": {"5"}}},
		{name: "get fulfillment", call: func(c *Client) error {
			_, _, err := c.Orders.GetFulfillment(testCtx, 42, 7)
			return err
		}, method: "GET", path: "/v1/orders/42/fulfillments/7"},
		{name: "custom shipping", call: func(c *Client) error {
			_, _, err := c.Orders.CreateCustomShipping(testCtx, 42, &CustomShippingRequest{
				LineItemIDs: []int64{1, 2}, TrackingNumber: "TRK-1", TrackingCompany: TrackingCompanyEzcat})
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/custom_shipping",
			body: `{"line_item_ids":"1,2","tracking_number":"TRK-1","tracking_company":"ezcat","notify_customer":false}`},
		{name: "support shipping", call: func(c *Client) error {
			_, _, err := c.Orders.CreateSupportShipping(testCtx, 42, &SupportShippingRequest{
				Email: "ops@example.com", LineItemIDs: []int64{1}, Source: SupportShippingEzcat, Size: 60,
				Temperature: ShippingTemperatureCold, FridgeOrFrozen: FridgeOrFrozenFrozen, UseTransfer: b(false)})
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/support_shipping",
			body: `{"email":"ops@example.com","line_item_ids":"1","source":"ezcat","size":60,"temperature":"cold","fridge_or_frozen":"frozen","is_fragile":false,"use_transfer":false}`},
		{name: "cvs shipping v1 empty", call: func(c *Client) error {
			_, _, err := c.Orders.CreateCVSShippingV1(testCtx, 42, &CVSShippingRequest{})
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/cvs_shipping", body: `{}`},
		{name: "cvs shipping v1 measurement", call: func(c *Client) error {
			_, _, err := c.Orders.CreateCVSShippingV1(testCtx, 42, &CVSShippingRequest{Measurement: CVSMeasurementS105})
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/cvs_shipping", body: `{"measurement":"S105"}`},
		{name: "partial cvs shipping", call: func(c *Client) error {
			_, _, err := c.Orders.CreatePartialCVSShipping(testCtx, 42, &PartialCVSShippingRequest{
				Items: []PartialCVSShippingItem{{LineItemID: 1, Quantity: 2}}, Charge: &charge})
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/partial_cvs_shipping",
			body: `{"items":[{"line_item_id":1,"quantity":2}],"charge":120}`},
		{name: "conclude partial cvs shipping", call: func(c *Client) error {
			_, _, err := c.Orders.ConcludePartialCVSShipping(testCtx, 42)
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/partial_cvs_shipping/conclude"},
		{name: "mark arrived", call: func(c *Client) error {
			_, _, err := c.Orders.MarkArrived(testCtx, 42)
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/arrived"},
		{name: "mark received", call: func(c *Client) error {
			_, _, err := c.Orders.MarkReceived(testCtx, 42)
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/received"},
		{name: "mark expired", call: func(c *Client) error {
			_, _, err := c.Orders.MarkExpired(testCtx, 42)
			return err
		}, method: "POST", path: "/v1/orders/42/fulfillments/expired"},
		{name: "fulfillment infos", reply: "[]", call: func(c *Client) error {
			_, _, err := c.Orders.GetFulfillmentInfos(testCtx, []int64{5, 6})
			return err
		}, method: "POST", path: "/v1/orders/fulfillment_infos", body: `{"fulfillment_ids":[5,6]}`},
		{name: "support shipping batch v1", call: func(c *Client) error {
			_, _, err := c.Orders.CreateSupportShippingBatch(testCtx, &SupportShippingBatchRequest{
				SupportShippings: []SupportShippingBatchItem{{
					OrderID: 42, LineItemIDs: []int64{1, 2}, Source: SupportShippingHCT, Size: 90,
					Temperature: ShippingTemperatureNormal, FridgeOrFrozen: FridgeOrFrozenNone, IsFragile: true}},
				Email: "ops@example.com"})
			return err
		}, method: "POST", path: "/v1/orders/fulfillments/support_shipping",
			body: `{"support_shippings":[{"order_id":42,"line_item_ids":"1,2","source":"hct","size":90,"temperature":"normal","fridge_or_frozen":"none","is_fragile":true}],"email":"ops@example.com"}`},
		{name: "cvs shipping v2", call: func(c *Client) error {
			_, _, err := c.Orders.CreateCVSShipping(testCtx, 42, &CVSShippingRequest{Size: 60})
			return err
		}, method: "POST", path: "/v2/orders/42/cvs_shipping", body: `{"size":60}`},
		{name: "support shippings v2", call: func(c *Client) error {
			_, _, err := c.Orders.CreateSupportShippings(testCtx, &SupportShippingsRequest{
				ShippingType: SupportShippingEzcat, Size: 60, Temperature: ShippingTemperatureNormal,
				FridgeOrFrozen: FridgeOrFrozenNone, IsFragile: true, UseTransfer: b(false),
				ShippingOrders: []ShippingOrder{{OrderID: 286, LineItemIDs: []int64{427}}}})
			return err
		}, method: "POST", path: "/v2/orders/fulfillments/support_shippings",
			body: `{"shipping_type":"ezcat","size":60,"temperature":"normal","fridge_or_frozen":"none","is_fragile":true,"use_transfer":false,"shipping_orders":[{"order_id":286,"line_item_ids":[427]}]}`},
	})
}

func TestSupportShippingAcceptedYieldsNilFulfillment(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	f, resp, err := c.Orders.CreateSupportShipping(testCtx, 42, &SupportShippingRequest{})
	if err != nil || f != nil || resp.StatusCode != http.StatusAccepted {
		t.Errorf("got %v %v %v", f, resp, err)
	}
}

func TestSupportShippingsDecodesResult(t *testing.T) {
	c, _ := ordersRecordingClient(t, `{"failed_orders":[{"order_id":286,"message":"無此商品"}],`+
		`"fulfillments":[{"id":98,"order_id":285,"tracking_number":null,"tracking_company":"ezcat"}]}`)
	out, _, err := c.Orders.CreateSupportShippings(testCtx, &SupportShippingsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.FailedOrders) != 1 || out.FailedOrders[0].OrderID != 286 {
		t.Errorf("failed orders = %+v", out.FailedOrders)
	}
	if len(out.Fulfillments) != 1 || out.Fulfillments[0].TrackingNumber != "" || out.Fulfillments[0].TrackingCompany != TrackingCompanyEzcat {
		t.Errorf("fulfillments = %+v", out.Fulfillments)
	}
}

func TestPrintLabelsReturnZipBytes(t *testing.T) {
	zip := []byte("PK\x03\x04fake-zip")
	cases := []struct {
		name string
		call func(c *Client) (*Response, error)
		path string
		body string
	}{
		{"cvs", func(c *Client) (*Response, error) {
			return c.Orders.PrintCVSShippingLabels(testCtx, &CVSShippingLabelsRequest{
				ShippingType: CVSShippingSevenC2C, FulfillmentIDs: []int64{1, 2}})
		}, "/v2/orders/cvs_shipping_labels", `{"shipping_type":"seven_c2c","fulfillment_ids":[1,2]}`},
		{"support", func(c *Client) (*Response, error) {
			return c.Orders.PrintSupportShippingLabels(testCtx, &SupportShippingLabelsRequest{
				ShippingType: SupportShippingHCT, PrintType: SupportShippingPrintThermal, OrderIDs: []int64{9}})
		}, "/v2/orders/fulfillments/support_shipping_labels", `{"shipping_type":"hct","print_type":"thermal","order_ids":[9]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &ordersRecorded{}
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				rec.Method, rec.Path, rec.Body = r.Method, r.URL.Path, string(body)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("Content-Disposition", `attachment; filename=labels.zip`)
				_, _ = w.Write(zip)
			})
			resp, err := tc.call(c)
			if err != nil {
				t.Fatal(err)
			}
			if rec.Method != "POST" || rec.Path != tc.path || rec.Body != tc.body {
				t.Errorf("request = %s %s %s", rec.Method, rec.Path, rec.Body)
			}
			if !bytes.Equal(resp.Body, zip) || resp.Header.Get("Content-Type") != "application/octet-stream" {
				t.Errorf("response = %q %q", resp.Body, resp.Header.Get("Content-Type"))
			}
		})
	}
}
