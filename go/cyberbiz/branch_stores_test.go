package cyberbiz

import (
	"errors"
	"testing"
)

// branchStoreJSON is a detail reply shaped after the swagger BranchStoreEntity
// (no id), since the recorded shop has no branch stores.
const branchStoreJSON = `{"store_no":"S001","name":"信義門市","phone":"02-12345678","county":"台北市","district":"信義區","address":"松高路1號","zip":"110","opening_hours":"10:00-22:00","lat":25.0396,"lng":121.5677,"enabled":true,"shipping_rates":[{"id":4811,"name":"門市自取","min_order_subtotal":500,"price":0,"payments":[{"id":227426,"name":"門市付款"}]}],"source_type":"PosShop","source_id":1446}`

func TestBranchStoresGolden(t *testing.T) {
	var stores []BranchStore
	decodeGolden(t, "v1/GET_v1_branch_stores.json", &stores)
	if len(stores) != 0 {
		t.Errorf("stores = %+v", stores)
	}
	var payments []BranchStorePayment
	decodeGolden(t, "v1/GET_v1_branch_stores_payments.json", &payments)
	if len(payments) != 1 || payments[0].ID != 227426 || payments[0].Name != "REDACTED" {
		t.Errorf("payments = %+v", payments)
	}
	for _, name := range []string{
		"errors/GET_v1_branch_stores_{id}.json",
		"errors/GET_v1_branch_stores_{id}_shipping_rates.json",
		"errors/GET_v1_branch_stores_{id}_shipping_rates_{id}.json",
		"errors/GET_v1_branch_stores_{id}_users.json",
		"errors/GET_v1_branch_stores_{id}_users_{id}.json",
		"errors/GET_v1_branch_stores_{id}_prepare_delivery_config.json",
	} {
		if msgs := errorMessages(readGolden(t, name)); len(msgs) != 1 || msgs[0] != "無此資源" {
			t.Errorf("%s: messages = %q", name, msgs)
		}
	}
}

func TestBranchStoreDetailDecodes(t *testing.T) {
	c, _ := stockRecordingClient(t, 200, branchStoreJSON)
	store, _, err := c.BranchStores.Get(testCtx, 42)
	if err != nil {
		t.Fatal(err)
	}
	if store.ID != 42 || store.StoreNo != "S001" || !store.Enabled || store.Lat != 25.0396 {
		t.Errorf("store = %+v", store)
	}
	if store.SourceType != BranchStoreSourcePosShop || store.SourceID != 1446 {
		t.Errorf("source = %q %d", store.SourceType, store.SourceID)
	}
	if len(store.ShippingRates) != 1 || store.ShippingRates[0].MinOrderSubtotal != MoneyFromInt(500) ||
		store.ShippingRates[0].Price != 0 || store.ShippingRates[0].Payments[0].ID != 227426 {
		t.Errorf("shipping_rates = %+v", store.ShippingRates)
	}
}

func TestBranchStoresGoldenRoundTrip(t *testing.T) {
	c := goldenServer(t)
	page, err := c.BranchStores.List(testCtx, nil)
	if err != nil || len(page.Items) != 0 || page.Pagination.Page != 1 {
		t.Fatalf("List: %v %+v", err, page)
	}
	payments, _, err := c.BranchStores.ListPayments(testCtx)
	if err != nil || len(payments) != 1 || payments[0].ID != 227426 {
		t.Errorf("ListPayments: %v %+v", err, payments)
	}
	if _, _, err := c.BranchStores.Get(testCtx, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get unknown: %v", err)
	}
	n := 0
	for _, err := range c.BranchStores.All(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 0 {
		t.Errorf("All yielded %d", n)
	}
}

func TestBranchStoresRequestShapes(t *testing.T) {
	rate := `{"name":"宅配","min_order_subtotal":1000,"price":80,"payments":[]}`
	user := `{"name":"小明","email":"ming@example.com"}`
	cfg := `{"delivery_date_enabled":true,"delivery_date_required":false,"prepare_delivery_day":3,"selectable_range":14}`
	stockRunShapes(t, []stockShapeCase{
		{name: "List", method: "GET", path: "/v1/branch_stores", reply: "[]", query: "page=3",
			call: func(c *Client) error { _, err := c.BranchStores.List(testCtx, &ListOptions{Page: 3}); return err }},
		{name: "Get", method: "GET", path: "/v1/branch_stores/7", reply: branchStoreJSON,
			call: func(c *Client) error { _, _, err := c.BranchStores.Get(testCtx, 7); return err }},
		{name: "Create", method: "POST", path: "/v1/branch_stores", reply: branchStoreJSON,
			body: map[string]any{"name": "信義門市", "source_type": "PosShop", "source_id": 1446.0, "enabled": false,
				"payment_ids": "1,2", "lat": 25.5,
				"shipping_rates": []any{map[string]any{"name": "宅配", "min_order_subtotal": 1000.0, "price": 80.5}}},
			absent: []string{"phone", "zip", "county", "lng", "store_no"},
			call: func(c *Client) error {
				_, _, err := c.BranchStores.Create(testCtx, &BranchStoreCreateRequest{
					Name: "信義門市", SourceType: BranchStoreSourcePosShop, SourceID: stockPtr[int64](1446),
					Enabled: stockPtr(false), PaymentIDs: "1,2", Lat: stockPtr(25.5),
					ShippingRates: []BranchStoreShippingRateCreateRequest{
						{Name: "宅配", MinOrderSubtotal: MoneyFromInt(1000), Price: MoneyFromFloat(80.5)}}})
				return err
			}},
		{name: "Update", method: "PUT", path: "/v1/branch_stores/7", reply: branchStoreJSON,
			body:   map[string]any{"phone": "", "enabled": true},
			absent: []string{"name", "payment_ids", "lat", "lng", "source_type"},
			call: func(c *Client) error {
				s, _, err := c.BranchStores.Update(testCtx, 7, &BranchStoreUpdateRequest{Phone: stockPtr(""), Enabled: stockPtr(true)})
				if err == nil && s.ID != 7 {
					t.Errorf("id = %d", s.ID)
				}
				return err
			}},
		{name: "Delete", method: "DELETE", path: "/v1/branch_stores/7", reply: branchStoreJSON,
			call: func(c *Client) error { _, err := c.BranchStores.Delete(testCtx, 7); return err }},
		{name: "ListPayments", method: "GET", path: "/v1/branch_stores/payments", reply: "[]",
			call: func(c *Client) error { _, _, err := c.BranchStores.ListPayments(testCtx); return err }},
		{name: "SetExpressDelivery", method: "PUT", path: "/v1/branch_stores/7/express_delivery_setting", reply: "[]",
			body: map[string]any{"express_delivery_status": false},
			call: func(c *Client) error { _, err := c.BranchStores.SetExpressDelivery(testCtx, 7, false); return err }},
		{name: "ListShippingRates", method: "GET", path: "/v1/branch_stores/7/shipping_rates", reply: "[" + rate + "]",
			call: func(c *Client) error {
				rates, _, err := c.BranchStores.ListShippingRates(testCtx, 7)
				if err == nil && (len(rates) != 1 || rates[0].Price != MoneyFromInt(80)) {
					t.Errorf("rates = %+v", rates)
				}
				return err
			}},
		{name: "GetShippingRate fills id", method: "GET", path: "/v1/branch_stores/7/shipping_rates/9", reply: rate,
			call: func(c *Client) error {
				r, _, err := c.BranchStores.GetShippingRate(testCtx, 7, 9)
				if err == nil && (r.ID != 9 || r.MinOrderSubtotal != MoneyFromInt(1000)) {
					t.Errorf("rate = %+v", r)
				}
				return err
			}},
		{name: "CreateShippingRate", method: "POST", path: "/v1/branch_stores/7/shipping_rates", reply: rate,
			body: map[string]any{"name": "宅配", "min_order_subtotal": 0.0, "price": 80.0},
			call: func(c *Client) error {
				_, _, err := c.BranchStores.CreateShippingRate(testCtx, 7, &BranchStoreShippingRateCreateRequest{Name: "宅配", Price: MoneyFromInt(80)})
				return err
			}},
		{name: "UpdateShippingRate", method: "PUT", path: "/v1/branch_stores/7/shipping_rates/9", reply: rate,
			body: map[string]any{"price": 0.0}, absent: []string{"name", "min_order_subtotal"},
			call: func(c *Client) error {
				r, _, err := c.BranchStores.UpdateShippingRate(testCtx, 7, 9, &BranchStoreShippingRateUpdateRequest{Price: stockPtr(Money(0))})
				if err == nil && r.ID != 9 {
					t.Errorf("id = %d", r.ID)
				}
				return err
			}},
		{name: "DeleteShippingRate", method: "DELETE", path: "/v1/branch_stores/7/shipping_rates/9", reply: rate,
			call: func(c *Client) error { _, err := c.BranchStores.DeleteShippingRate(testCtx, 7, 9); return err }},
		{name: "ListUsers", method: "GET", path: "/v1/branch_stores/7/users", reply: "[" + user + "]", query: "per_page=50",
			call: func(c *Client) error {
				p, err := c.BranchStores.ListUsers(testCtx, 7, &ListOptions{PerPage: 50})
				if err == nil && (len(p.Items) != 1 || p.Items[0].Email != "ming@example.com") {
					t.Errorf("users = %+v", p.Items)
				}
				return err
			}},
		{name: "GetUser fills id", method: "GET", path: "/v1/branch_stores/7/users/3", reply: user,
			call: func(c *Client) error {
				u, _, err := c.BranchStores.GetUser(testCtx, 7, 3)
				if err == nil && (u.ID != 3 || u.Name != "小明") {
					t.Errorf("user = %+v", u)
				}
				return err
			}},
		{name: "CreateUser", method: "POST", path: "/v1/branch_stores/7/users", reply: user,
			body: map[string]any{"name": "小明", "email": "ming@example.com"},
			call: func(c *Client) error {
				_, _, err := c.BranchStores.CreateUser(testCtx, 7, &BranchStoreUserCreateRequest{Name: "小明", Email: "ming@example.com"})
				return err
			}},
		{name: "UpdateUser", method: "PUT", path: "/v1/branch_stores/7/users/3", reply: user,
			body: map[string]any{"email": "new@example.com"}, absent: []string{"name"},
			call: func(c *Client) error {
				u, _, err := c.BranchStores.UpdateUser(testCtx, 7, 3, &BranchStoreUserUpdateRequest{Email: stockPtr("new@example.com")})
				if err == nil && u.ID != 3 {
					t.Errorf("id = %d", u.ID)
				}
				return err
			}},
		{name: "DeleteUser", method: "DELETE", path: "/v1/branch_stores/7/users/3", reply: user,
			call: func(c *Client) error { _, err := c.BranchStores.DeleteUser(testCtx, 7, 3); return err }},
		{name: "GetPrepareDeliveryConfig", method: "GET", path: "/v1/branch_stores/7/prepare_delivery_config", reply: cfg,
			call: func(c *Client) error {
				got, _, err := c.BranchStores.GetPrepareDeliveryConfig(testCtx, 7)
				if err == nil && (!got.DeliveryDateEnabled || got.PrepareDeliveryDay != 3 || got.SelectableRange != 14) {
					t.Errorf("config = %+v", got)
				}
				return err
			}},
		{name: "UpdatePrepareDeliveryConfig", method: "PUT", path: "/v1/branch_stores/7/prepare_delivery_config", reply: cfg,
			body:   map[string]any{"delivery_date_required": false, "prepare_delivery_day": 5.0},
			absent: []string{"delivery_date_enabled", "selectable_range"},
			call: func(c *Client) error {
				_, _, err := c.BranchStores.UpdatePrepareDeliveryConfig(testCtx, 7, &PrepareDeliveryConfigUpdateRequest{
					DeliveryDateRequired: stockPtr(false), PrepareDeliveryDay: stockPtr(5)})
				return err
			}},
	})
}
