package cyberbiz

import (
	"errors"
	"testing"
)

// Golden decode tests: each proves a model matches a recorded response.

func TestCustomersGoldenList(t *testing.T) {
	var out []Customer
	decodeGolden(t, "v1/GET_v1_customers.json", &out)
	if len(out) < 2 {
		t.Fatalf("len = %d", len(out))
	}
	c := out[0]
	if c.ID != 23300813 || c.Status != CustomerStatusEnabled || c.Email != "redacted@example.com" {
		t.Errorf("customer = %+v", c)
	}
	if c.BonusRemain != MoneyFromInt(10012) {
		t.Errorf("bonus_remain = %v", c.BonusRemain)
	}
	if c.CreatedAt.String() != "2023-10-03 12:04:02" {
		t.Errorf("created_at = %v", c.CreatedAt)
	}
	if c.ConfirmedAt == nil || c.ConfirmedAt.String() != "2023-11-30 22:30:58" {
		t.Errorf("confirmed_at = %v", c.ConfirmedAt)
	}
	if !c.Birthday.IsZero() || c.OtherAccumulatedConsumptionExpiredAt != nil {
		t.Errorf("nullables not zero: %v %v", c.Birthday, c.OtherAccumulatedConsumptionExpiredAt)
	}
	if c.Address == nil || c.Address.Phone != "0912345678" || c.Address.DetailAddress == nil || c.Address.DetailAddress.Province != "" {
		t.Errorf("address = %+v", c.Address)
	}
	if len(c.UIDProviders) != 1 || c.UIDProviders[0].ProviderType != CustomerUIDProviderLine {
		t.Errorf("uid_providers = %+v", c.UIDProviders)
	}
	if c.Tags == nil || c.CustomFields == nil || c.EnableHomeDeliveryCOD {
		t.Errorf("tags/custom_fields/cod = %v %v %v", c.Tags, c.CustomFields, c.EnableHomeDeliveryCOD)
	}
	if out[1].Address != nil || out[1].Gender != "男" {
		t.Errorf("second customer = %+v", out[1])
	}
}

func TestCustomersGoldenDetail(t *testing.T) {
	var out Customer
	decodeGolden(t, "v1/GET_v1_customers_{id}.json", &out)
	if out.ID != 0 {
		t.Errorf("detail should carry no id, got %d", out.ID)
	}
	if out.Name != "REDACTED" || out.MobileSMSConfirmedAt == nil || out.BonusRemain != MoneyFromInt(10012) {
		t.Errorf("customer = %+v", out)
	}
}

func TestCustomersGoldenV2List(t *testing.T) {
	var out []Customer
	decodeGolden(t, "v2/GET_v2_customers.json", &out)
	if len(out) != 2 || out[0].ID != 23300813 || out[1].ID != 24077804 {
		t.Fatalf("ids = %+v", out)
	}
	if out[0].VIPInfo != nil || out[0].UIDProviders != nil {
		t.Errorf("plain v2 list should not carry includes: %+v", out[0])
	}
	var withInclude []Customer
	decodeGolden(t, "v2/GET_v2_customers_include.json", &withInclude)
	if len(withInclude) != 2 || withInclude[0].ConfirmedAt == nil {
		t.Errorf("include list = %+v", withInclude)
	}
}

func TestCustomersGoldenBonusPoints(t *testing.T) {
	var out []CustomerBonusPoint
	decodeGolden(t, "v1/GET_v1_customers_{id}_bonus_points.json", &out)
	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	bp := out[0]
	if bp.ID != 72070294 || bp.Points != MoneyFromInt(13) || bp.UnusedPoints != MoneyFromInt(12) {
		t.Errorf("bonus point = %+v", bp)
	}
	if bp.Deadline.String() != "2125-12-31 14:14:30" || bp.CustomerID != 23300813 || bp.OrderID != 0 {
		t.Errorf("bonus point = %+v", bp)
	}
	if out[1].Points != MoneyFromInt(10000) || out[1].Source != "一般紅利積點(購物金)" {
		t.Errorf("second = %+v", out[1])
	}
}

func TestCustomersGoldenCoupons(t *testing.T) {
	var out []CustomerCoupon
	decodeGolden(t, "v1/GET_v1_customers_{id}_coupons.json", &out)
	if out == nil || len(out) != 0 {
		t.Errorf("coupons = %v", out)
	}
}

func TestCustomersGoldenShopCoupons(t *testing.T) {
	var out []CustomerShopCoupon
	decodeGolden(t, "v1/GET_v1_customers_{id}_shop_coupons.json", &out)
	if len(out) != 1 {
		t.Fatalf("len = %d", len(out))
	}
	c := out[0]
	if c.ID != 345196115 || c.Code != "SHMHHDNL" || c.OrderPriceThreshold != 0 || !c.ConcurrentlyApply {
		t.Errorf("coupon = %+v", c)
	}
	if c.CouponStatus != CouponStatusNoExpireDate || c.GiftOrderStatus != "" || !c.Usable() {
		t.Errorf("status = %q/%q usable=%v", c.CouponStatus, c.GiftOrderStatus, c.Usable())
	}
	if c.RestrictStrategy != CustomerCouponRestrictUnrestricted || !c.StartDate.IsZero() || c.GiftOrderID != 0 {
		t.Errorf("coupon = %+v", c)
	}
	if c.RestrictCampaigns == nil || c.PosShopIDs == nil || c.UsageLimit != 1000000000 {
		t.Errorf("coupon = %+v", c)
	}

	var one CustomerShopCoupon
	decodeGolden(t, "v1/GET_v1_customers_{id}_shop_coupons_{id}.json", &one)
	if one.ID != 0 || one.Title != "測試優惠碼" || one.UsedTimes != 1 || !one.CustomerUsable {
		t.Errorf("shop coupon detail = %+v", one)
	}
}

func TestCustomersGoldenEmptyLists(t *testing.T) {
	var fields []CustomerCustomField
	decodeGolden(t, "v1/GET_v1_customers_{id}_custom_fields.json", &fields)
	var cart []CustomerCartItem
	decodeGolden(t, "v1/GET_v1_customers_{id}_customer_cart_items.json", &cart)
	var posts []CustomerMessagePost
	decodeGolden(t, "v1/GET_v1_customers_{id}_message_posts.json", &posts)
	var others []CustomerOtherValidOrder
	decodeGolden(t, "v1/GET_v1_customers_{id}_other_valid_orders.json", &others)
	var ids []CustomerIDMatch
	decodeGolden(t, "v1/GET_v1_customers_get_customer_id.json", &ids)
	for name, n := range map[string]int{"custom_fields": len(fields), "cart": len(cart),
		"message_posts": len(posts), "other_valid_orders": len(others), "get_customer_id": len(ids)} {
		if n != 0 {
			t.Errorf("%s: len = %d", name, n)
		}
	}
}

func TestCustomersGoldenOrders(t *testing.T) {
	var out []CustomerOrder
	decodeGolden(t, "v1/GET_v1_customers_{id}_orders.json", &out)
	if len(out) != 2 {
		t.Fatalf("len = %d", len(out))
	}
	o := out[0]
	if o.ID != 49492440 || o.OrderNumber != 1068 || o.OrderName != "#1068" || o.SubtotalPrice != MoneyFromInt(200) {
		t.Errorf("order = %+v", o)
	}
	if o.Customer == nil || o.Customer.ID != 23300813 || o.Buyer == nil || o.Buyer.Mobile != "" {
		t.Errorf("customer/buyer = %+v %+v", o.Customer, o.Buyer)
	}
	if o.Statuses == nil || o.Statuses.FinancialStatus != FinancialStatusPaid || !o.Statuses.FinancialStatus.IsPaid() {
		t.Errorf("statuses = %+v", o.Statuses)
	}
	if o.Timings == nil || o.Timings.ClosedAt != nil || o.Timings.ConfirmedAt.String() != "2026-03-10 11:34:31" {
		t.Errorf("timings = %+v", o.Timings)
	}
	if o.Prices == nil || o.Prices.TotalLineItemsPrice != MoneyFromInt(1499) || o.Prices.TotalPrice != MoneyFromInt(200) {
		t.Errorf("prices = %+v", o.Prices)
	}
	if len(o.LineItems) != 2 {
		t.Fatalf("line items = %d", len(o.LineItems))
	}
	li := o.LineItems[0]
	if li.ID != 109940715 || li.Price != MoneyFromInt(999) || li.Cost != 0 || li.Weight != 0 {
		t.Errorf("line item = %+v", li)
	}
	if len(li.Discounts) != 1 || li.Discounts[0].Code != "bundle_discount" || li.Discounts[0].Discount != MoneyFromInt(866) {
		t.Errorf("discounts = %+v", li.Discounts)
	}
	if len(li.RelatedItems) != 1 || len(li.RelatedItems[0].Items) != 3 || li.RelatedItems[0].Items[0].ComboProductPriceDifference != MoneyFromInt(57) {
		t.Errorf("related = %+v", li.RelatedItems)
	}
	if li.RelatedItems[0].Items[0].ComboProductPriceDiffDetails[0] != MoneyFromInt(57) {
		t.Errorf("diff details = %v", li.RelatedItems[0].Items[0].ComboProductPriceDiffDetails)
	}
}

func TestCustomersGoldenRecentPurchases(t *testing.T) {
	var out []CustomerLineItem
	decodeGolden(t, "v1/GET_v1_customers_{id}_recent_purchases.json", &out)
	if len(out) != 3 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].ID != 109940716 || out[0].TotalDiscount != MoneyFromInt(433) || out[0].TotalPriceAfterDiscounts != MoneyFromInt(67) {
		t.Errorf("first = %+v", out[0])
	}
	if out[1].TotalDiscount != 0 || out[1].CreatedAt.String() != "2026-03-10 11:33:54" {
		t.Errorf("second = %+v", out[1])
	}
}

func TestCustomersGoldenSpendingOverview(t *testing.T) {
	var out CustomerSpendingOverview
	decodeGolden(t, "v1/GET_v1_customers_{id}_spending_overview.json", &out)
	if out.PaidAndValidTotalSpent != MoneyFromInt(1398) || out.PaidAndValidOrdersCount != 3 || out.PaidAndValidAverageSpent != MoneyFromInt(466) {
		t.Errorf("overview = %+v", out)
	}
}

func TestCustomersGoldenUIDProvider(t *testing.T) {
	var out CustomerUIDLookup
	decodeGolden(t, "v1/GET_v1_customers_{id}_uid_providers_line.json", &out)
	if out.CustomerID != 23300813 || out.UID != "REDACTED" || out.Message != "查詢成功" {
		t.Errorf("lookup = %+v", out)
	}
}

func TestCustomersGoldenVIPInfo(t *testing.T) {
	var out CustomerVIPInfo
	decodeGolden(t, "v1/GET_v1_customers_{id}_vip_info.json", &out)
	if out.CustomerID != 23300813 || out.CurrentGroup != nil || out.CurrentLevel != nil || out.NextLevel != nil {
		t.Errorf("vip info = %+v", out)
	}
	if out.ExtraInfo == nil || !out.ExtraInfo.StartAt.IsZero() || out.ExtraInfo.DifferenceOfTotalSpentForUpgrade != 0 {
		t.Errorf("extra info = %+v", out.ExtraInfo)
	}
}

func TestCustomersGoldenActivationURL(t *testing.T) {
	var out CustomerActivationURL
	decodeGolden(t, "v1/GET_v1_customers_{id}_account_activation_url.json", &out)
	if out.AccountActivationURL == "" {
		t.Error("empty url")
	}
}

func TestCustomersGoldenGroups(t *testing.T) {
	var out []CustomerGroup
	decodeGolden(t, "v1/GET_v1_customers_customer_groups.json", &out)
	if len(out) < 2 {
		t.Fatalf("len = %d", len(out))
	}
	g := out[0]
	if g.ID != 10693 || g.GroupType != CustomerGroupTypeDefault || g.Amount != 0 || g.Query == "" {
		t.Errorf("group = %+v", g)
	}
	if g.CreatedAt.String() != "2023-09-11 09:25:38" {
		t.Errorf("created_at = %v", g.CreatedAt)
	}
}

func TestCustomersGoldenGroupCheckStatus(t *testing.T) {
	c, _ := New("tok")
	status, err := customersDecodeGroupStatus(c, readGolden(t, "v1/GET_v1_customers_customer_groups_check_status_{id}.json"))
	if err != nil {
		t.Fatal(err)
	}
	if status.Done || status.Success != "QUEUED" || status.JobID != "1" || status.Message == "" {
		t.Errorf("status = %+v", status)
	}
	done, err := customersDecodeGroupStatus(c, []byte(`[{"id":1,"email":"a@example.com","name":"A","uid":"U1"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if !done.Done || len(done.Members) != 1 || done.Members[0].ID != 1 || done.Members[0].UID != "U1" {
		t.Errorf("members = %+v", done)
	}
}

func TestCustomersGoldenDefaultGenderOptions(t *testing.T) {
	var out []CustomerGenderOption
	decodeGolden(t, "v1/GET_v1_customers_default_gender_options.json", &out)
	if len(out) != 2 || out[0].DefaultGenderOption != "" {
		t.Errorf("options = %+v", out)
	}
}

func TestCustomersGoldenLookupByName(t *testing.T) {
	var out []CustomerNameMatch
	decodeGolden(t, "v1/GET_v1_customers_get_customer_id_by_name.json", &out)
	if len(out) != 2 || out[0].CustomerID != 40401725 || out[1].CustomerID != 38172973 {
		t.Errorf("matches = %+v", out)
	}
}

func TestCustomersGoldenTags(t *testing.T) {
	var out []CustomerTag
	decodeGolden(t, "v1/GET_v1_customers_tags.json", &out)
	if len(out) != 2 || out[0].Name != "REDACTED" {
		t.Errorf("tags = %+v", out)
	}
}

// Golden server round trips.

func TestCustomersGetRoundTrip(t *testing.T) {
	c := goldenServer(t)
	got, resp, err := c.Customers.Get(testCtx, 23300813)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 23300813 || got.Email != "redacted@example.com" || resp.StatusCode != 200 {
		t.Errorf("got %+v", got)
	}
}

func TestCustomersListRoundTrip(t *testing.T) {
	c := goldenServer(t)
	page, err := c.Customers.List(testCtx, &CustomerListOptions{ListOptions: ListOptions{PerPage: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) < 2 || page.Items[0].ID != 23300813 || page.Pagination.Total != 1 {
		t.Errorf("page = %+v", page.Pagination)
	}
	v2, err := c.Customers.ListV2(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(v2.Items) != 2 || v2.Items[1].ID != 24077804 {
		t.Errorf("v2 items = %+v", v2.Items)
	}
}

func TestCustomersSubResourceRoundTrips(t *testing.T) {
	c := goldenServer(t)
	const id = 23300813
	if bp, _, err := c.Customers.ListBonusPoints(testCtx, id); err != nil || len(bp) != 2 {
		t.Errorf("bonus points: %v %d", err, len(bp))
	}
	if sc, err := c.Customers.ListShopCoupons(testCtx, id, nil); err != nil || len(sc.Items) != 1 {
		t.Errorf("shop coupons: %v", err)
	}
	if one, _, err := c.Customers.GetShopCoupon(testCtx, id, 345196115); err != nil || one.ID != 345196115 {
		t.Errorf("shop coupon: %v %+v", err, one)
	}
	if uid, _, err := c.Customers.GetUIDProvider(testCtx, id, CustomerUIDProviderLine); err != nil || uid.UID != "REDACTED" {
		t.Errorf("uid provider: %v %+v", err, uid)
	}
	if vip, _, err := c.Customers.VIPInfo(testCtx, id); err != nil || vip.CustomerID != id {
		t.Errorf("vip info: %v %+v", err, vip)
	}
	if orders, err := c.Customers.ListOrders(testCtx, id, nil); err != nil || len(orders.Items) != 2 {
		t.Errorf("orders: %v", err)
	}
	if groups, err := c.Customers.ListGroups(testCtx, nil); err != nil || len(groups.Items) < 2 {
		t.Errorf("groups: %v", err)
	}
	if st, _, err := c.Customers.CheckGroupStatus(testCtx, "1", nil); err != nil || st.Success != "QUEUED" {
		t.Errorf("check status: %v %+v", err, st)
	}
	if ids, _, err := c.Customers.LookupIDs(testCtx, &CustomerLookupOptions{Emails: []string{"x@example.com"}}); err != nil || ids == nil || len(ids) != 0 {
		t.Errorf("lookup ids: %v %v", err, ids)
	}
}

func TestCustomersGetByUIDProviderNullIsNotFound(t *testing.T) {
	c := goldenServer(t)
	_, _, err := c.Customers.GetByUIDProvider(testCtx, CustomerUIDProviderLine, "U0")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestCustomersErrorShapes(t *testing.T) {
	cases := map[string]error{
		"errors/GET_v1_customers_{id}.json":                        ErrNotFound,
		"errors/GET_v1_customers_{id}_coupons_{id}.json":           ErrNotFound,
		"errors/GET_v1_customers_{id}_uid_providers_facebook.json": ErrValidation,
	}
	statuses := map[error]int{ErrNotFound: 404, ErrValidation: 422}
	for name, want := range cases {
		body := readGolden(t, name)
		c := customersRecorder(t, statuses[want], string(body))
		_, _, err := c.client.Customers.Get(testCtx, 1)
		if !errors.Is(err, want) {
			t.Errorf("%s: got %v", name, err)
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) || len(apiErr.Messages) != 1 {
			t.Errorf("%s: messages = %v", name, err)
		}
	}
}
