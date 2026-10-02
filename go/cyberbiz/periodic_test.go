package cyberbiz

import (
	"encoding/json/v2"
	"errors"
	"testing"
)

const periodicSampleBody = `{"customer_id":1315,"number":"P0001","sales_page":{"id":1087,"title":"每月茶葉"},
	"recent_preorder_id":1516,"preorders_count":3,"order_prepare_days":2,
	"periodic":{"type":"monthly","month":1,"week":2,"day":3},
	"start_date":"2026-01-20","next_date":"2026-02-20","last_date":null,"delivery_time":2,"note":"",
	"billing_address":{"zip":"100","city":"台北市","district":"中正區","address1":"忠孝東路一段1號","name":"王小明","phone":"0912345678","country_calling_code":"886"},
	"created_at":"2026-01-18 10:00:00","affiliate_vendor":null,"id":59}`

func TestPeriodicOrdersGolden(t *testing.T) {
	var out []PeriodicOrder
	decodeGolden(t, "v1/GET_v1_periodic_orders.json", &out)
	if out == nil || len(out) != 0 {
		t.Errorf("orders = %#v", out)
	}
	c := goldenServer(t)
	page, err := c.Periodic.List(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 {
		t.Errorf("page = %+v", page)
	}
	_, err = c.Periodic.ListEstablishedPreorders(testCtx, 1, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("preorders of unknown parent: got %v", err)
	}
}

func TestPeriodicOrderDecodesPostmanShape(t *testing.T) {
	c, _ := New("tok")
	var po PeriodicOrder
	if err := c.decode([]byte(periodicSampleBody), &po); err != nil {
		t.Fatal(err)
	}
	if po.ID != 59 || po.Number != "P0001" || po.SalesPage == nil || po.SalesPage.ID != 1087 {
		t.Errorf("order = %+v", po)
	}
	if po.Periodic == nil || po.Periodic.Type != PeriodicTypeMonthly || po.Periodic.Day != 3 {
		t.Errorf("periodic = %+v", po.Periodic)
	}
	if po.StartDate.String() != "2026-01-20" || !po.LastDate.IsZero() || po.CreatedAt.String() != "2026-01-18 10:00:00" {
		t.Errorf("dates = %v %v %v", po.StartDate, po.LastDate, po.CreatedAt)
	}
	if po.BillingAddress == nil || po.BillingAddress.City != "台北市" || po.AffiliateVendor != nil {
		t.Errorf("address = %+v vendor = %+v", po.BillingAddress, po.AffiliateVendor)
	}

	var pre PeriodicPreorder
	body := `{"shipping_address":{"name":"x"},"cancel_at":null,"cancel_reason":null,"fulfilled_times":1,"number":2,
		"delivery_date":"2026-02-22","base_delivery_date":"2026-02-20","order":{"id":9,"order_number":1020},"id":1516}`
	if err := c.decode([]byte(body), &pre); err != nil {
		t.Fatal(err)
	}
	if pre.ID != 1516 || pre.Number != 2 || pre.CancelAt != nil || pre.DeliveryDate.String() != "2026-02-22" {
		t.Errorf("preorder = %+v", pre)
	}
	var order struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(pre.Order, &order); err != nil || order.ID != 9 {
		t.Errorf("order = %s (%v)", pre.Order, err)
	}
}

func TestPeriodicRequests(t *testing.T) {
	c, call := discountsSpyClient(t, 200, `[]`)
	since, _ := ParseTime("2026-01-01 00:00:00")
	if _, err := c.Periodic.List(testCtx, &PeriodicOrderListOptions{ListOptions: ListOptions{PerPage: 15}, PreorderStartAt: since}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/periodic_orders")
	if call.Query.Get("preorder_start_at") != "2026-01-01 00:00:00" || call.Query.Get("per_page") != "15" {
		t.Errorf("query = %v", call.Query)
	}

	if _, err := c.Periodic.ListEstablishedPreorders(testCtx, 59, &ListOptions{Page: 2}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/periodic_orders/59/preorders/established")
	if call.Query.Get("page") != "2" {
		t.Errorf("query = %v", call.Query)
	}

	var n int
	for _, err := range c.Periodic.AllEstablishedPreorders(testCtx, 59, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 0 || call.Query.Get("per_page") != "50" {
		t.Errorf("walked %d, query = %v", n, call.Query)
	}

	u, ucall := discountsSpyClient(t, 200, periodicSampleBody)
	note := ""
	day := 0
	slot := 1
	city := "新北市"
	po, _, err := u.Periodic.Update(testCtx, 59, &PeriodicOrderUpdateRequest{
		Note: &note, DeliveryTime: &slot,
		Periodic:       &PeriodicScheduleRequest{Day: &day},
		BillingAddress: &PeriodicBillingAddressRequest{City: &city},
	})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, ucall, "PUT", "/v1/periodic_orders/59")
	discountsAssertJSONBody(t, ucall, `{"note":"","delivery_time":1,"periodic":{"day":0},"billing_address":{"city":"新北市"}}`)
	if po.ID != 59 || po.Number != "P0001" {
		t.Errorf("order = %+v", po)
	}
}
