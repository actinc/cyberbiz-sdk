package cyberbiz

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// customersRecorded is what the test server saw on the last request.
type customersRecorded struct {
	client *Client
	method string
	path   string
	query  string
	body   string
}

// customersRecorder serves status/body for every request and records the
// request it received.
func customersRecorder(t *testing.T, status int, body string) *customersRecorded {
	t.Helper()
	rec := &customersRecorded{}
	rec.client, _ = newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path
		rec.query = r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		rec.body = string(b)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}, WithMaxRetries(0))
	return rec
}

func (r *customersRecorded) expect(t *testing.T, method, path, query string) {
	t.Helper()
	if r.method != method || r.path != path {
		t.Errorf("request = %s %s, want %s %s", r.method, r.path, method, path)
	}
	if r.query != query {
		t.Errorf("query = %q, want %q", r.query, query)
	}
}

func (r *customersRecorded) expectBody(t *testing.T, want string) {
	t.Helper()
	if r.body != want {
		t.Errorf("body = %s, want %s", r.body, want)
	}
}

func customersPtr[T any](v T) *T { return &v }

func TestCustomersListQuery(t *testing.T) {
	rec := customersRecorder(t, 200, `[]`)
	start, _ := ParseTime("2026-01-01 00:00:00")
	_, err := rec.client.Customers.List(testCtx, &CustomerListOptions{
		ListOptions:        ListOptions{Page: 2, PerPage: 10},
		UpdatedAtStartTime: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "GET", "/v1/customers", "page=2&per_page=10&updated_at_start_time=2026-01-01+00%3A00%3A00")
}

func TestCustomersAllWalksPages(t *testing.T) {
	var pages []string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		pages = append(pages, r.URL.Query().Get("page"))
		if r.URL.Query().Get("page") == "1" {
			w.Header().Set("X-Next-Page", "2")
			_, _ = w.Write([]byte(`[{"id":1}]`))
			return
		}
		_, _ = w.Write([]byte(`[{"id":2}]`))
	})
	var ids []int64
	for cu, err := range c.Customers.All(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, cu.ID)
	}
	if len(ids) != 2 || len(pages) != 2 {
		t.Errorf("ids = %v pages = %v", ids, pages)
	}
}

func TestCustomersCreateBodyOmitsUnset(t *testing.T) {
	rec := customersRecorder(t, 201, `{"id":5,"name":"Amy"}`)
	got, _, err := rec.client.Customers.Create(testCtx, &CustomerCreateRequest{
		Name:             "Amy",
		Email:            "amy@example.com",
		AcceptsMarketing: customersPtr(false),
		Birthday:         NewDate(1990, 5, 1),
		Address:          &CustomerAddressRequest{City: "台北市", Phone: "0912345678"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "POST", "/v1/customers", "")
	rec.expectBody(t, `{"name":"Amy","email":"amy@example.com","accepts_marketing":false,"birthday":"1990-05-01","address":{"city":"台北市","phone":"0912345678"}}`)
	if got.ID != 5 {
		t.Errorf("got %+v", got)
	}
}

func TestCustomersUpdateBodyAndNullClears(t *testing.T) {
	rec := customersRecorder(t, 200, `{"name":"Amy"}`)
	got, _, err := rec.client.Customers.Update(testCtx, 7, &CustomerUpdateRequest{
		Status:                      CustomerStatusDisabled,
		OtherAccumulatedConsumption: customersPtr(Money(0)),
		ConfirmedAt:                 &Time{},
		EnableSendBirthGift:         customersPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "PUT", "/v1/customers/7", "")
	rec.expectBody(t, `{"status":"disabled","enable_send_birth_gift":true,"other_accumulated_consumption":0,"confirmed_at":null}`)
	if got.ID != 7 {
		t.Errorf("id not filled: %+v", got)
	}
}

func TestCustomersLookups(t *testing.T) {
	rec := customersRecorder(t, 206, `[]`)
	ids, resp, err := rec.client.Customers.LookupIDs(testCtx, &CustomerLookupOptions{
		Emails:  []string{"a@example.com", "b@example.com"},
		Mobiles: []string{"0911111111"},
	})
	if err != nil || ids == nil || len(ids) != 0 || resp.StatusCode != 206 {
		t.Fatalf("ids=%v resp=%v err=%v", ids, resp, err)
	}
	rec.expect(t, "GET", "/v1/customers/get_customer_id", "customer_emails=a%40example.com%2Cb%40example.com&customer_mobiles=0911111111")

	rec = customersRecorder(t, 200, `[{"customer_id":1,"customer_name":"Amy"}]`)
	names, _, err := rec.client.Customers.LookupIDsByName(testCtx, "A", 10)
	if err != nil || len(names) != 1 || names[0].CustomerID != 1 {
		t.Fatalf("names=%v err=%v", names, err)
	}
	rec.expect(t, "GET", "/v1/customers/get_customer_id_by_name", "customer_name=A&limit=10")

	rec = customersRecorder(t, 200, `[{"default_gender_option":"other"}]`)
	opts, _, err := rec.client.Customers.DefaultGenderOptions(testCtx)
	if err != nil || len(opts) != 1 || opts[0].DefaultGenderOption != "other" {
		t.Fatalf("opts=%v err=%v", opts, err)
	}
	rec.expect(t, "GET", "/v1/customers/default_gender_options", "")

	rec = customersRecorder(t, 200, `[{"name":"vip"}]`)
	tags, err := rec.client.Customers.ListTags(testCtx, &ListOptions{Page: 3})
	if err != nil || len(tags.Items) != 1 {
		t.Fatalf("tags=%v err=%v", tags, err)
	}
	rec.expect(t, "GET", "/v1/customers/tags", "page=3")

	rec = customersRecorder(t, 200, `{"account_activation_url":"https://x/activate"}`)
	u, _, err := rec.client.Customers.AccountActivationURL(testCtx, 9)
	if err != nil || u.AccountActivationURL != "https://x/activate" {
		t.Fatalf("url=%v err=%v", u, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/account_activation_url", "")
}

func TestCustomersActions(t *testing.T) {
	rec := customersRecorder(t, 201, ``)
	if _, err := rec.client.Customers.ConsumeBonusPoints(testCtx, 9, &CustomerConsumeBonusPointsRequest{ConsumeAll: customersPtr(true)}); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "POST", "/v1/customers/9/consume_bonus_points", "")
	rec.expectBody(t, `{"consume_all":true}`)

	rec = customersRecorder(t, 201, ``)
	if _, err := rec.client.Customers.ConsumeBonusPoints(testCtx, 9, &CustomerConsumeBonusPointsRequest{Amount: MoneyFromInt(20)}); err != nil {
		t.Fatal(err)
	}
	rec.expectBody(t, `{"amount":20}`)

	rec = customersRecorder(t, 201, ``)
	_, err := rec.client.Customers.UpdateRegisterCode(testCtx, 9, &CustomerRegisterCodeRequest{
		SecretKey: "k", RegisterCode: "REF1", AcceptsMarketing: customersPtr(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "POST", "/v1/customers/9/update_register_code_and_accepts_marketing", "")
	rec.expectBody(t, `{"secret_key":"k","register_code":"REF1","accepts_marketing":false}`)

	rec = customersRecorder(t, 200, ``)
	if _, err := rec.client.Customers.SetUIDProvider(testCtx, 9, CustomerUIDProviderLine, "U123"); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "PUT", "/v1/customers/9/uid_providers/line", "")
	rec.expectBody(t, `{"uid":"U123"}`)

	rec = customersRecorder(t, 200, `{"customer_id":9,"uid":"U123","message":"ok"}`)
	got, _, err := rec.client.Customers.GetUIDProvider(testCtx, 9, CustomerUIDProviderLineAt)
	if err != nil || got.UID != "U123" {
		t.Fatalf("got=%v err=%v", got, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/uid_providers/line_at", "")
}

func TestCustomersReports(t *testing.T) {
	rec := customersRecorder(t, 200, `{"paid_and_valid_total_spent":100,"paid_and_valid_orders_count":1,"paid_and_valid_average_spent":100}`)
	ov, _, err := rec.client.Customers.SpendingOverview(testCtx, 9, &CustomerSpendingOptions{
		StartDate: NewDate(2026, 1, 1), EndDate: NewDate(2026, 1, 31),
	})
	if err != nil || ov.PaidAndValidTotalSpent != MoneyFromInt(100) {
		t.Fatalf("ov=%v err=%v", ov, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/spending_overview", "end_date=2026-01-31&start_date=2026-01-01")

	rec = customersRecorder(t, 200, `[{"id":1,"price":50.0}]`)
	items, _, err := rec.client.Customers.RecentPurchases(testCtx, 9, &CustomerRecentPurchasesOptions{
		StartDate: NewDate(2026, 1, 1), EndDate: NewDate(2026, 1, 31), MaxProducts: 5,
	})
	if err != nil || len(items) != 1 || items[0].Price != MoneyFromInt(50) {
		t.Fatalf("items=%v err=%v", items, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/recent_purchases", "end_date=2026-01-31&max_products=5&start_date=2026-01-01")

	rec = customersRecorder(t, 200, `[{"id":3,"product_id":2,"price":10.5,"photo_urls":["a"]}]`)
	cart, _, err := rec.client.Customers.CartItems(testCtx, 9)
	if err != nil || len(cart) != 1 || cart[0].Price != Money(1050) || cart[0].PhotoURLs[0] != "a" {
		t.Fatalf("cart=%v err=%v", cart, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/customer_cart_items", "")

	rec = customersRecorder(t, 200, `[{"id":4,"status":"replied","comments":[{"id":1,"role":"admin","content":"hi","created_at":"2026-01-02 03:04:05"}],"created_at":"2026-01-01 00:00:00"}]`)
	posts, err := rec.client.Customers.ListMessagePosts(testCtx, 9, &ListOptions{PerPage: 5})
	if err != nil || len(posts.Items) != 1 || posts.Items[0].Status != CustomerMessageStatusReplied || posts.Items[0].Comments[0].CreatedAt.String() != "2026-01-02 03:04:05" {
		t.Fatalf("posts=%+v err=%v", posts, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/message_posts", "per_page=5")

	rec = customersRecorder(t, 200, `[{"id":1,"order_number":10}]`)
	orders, err := rec.client.Customers.ListOrders(testCtx, 9, &ListOptions{Page: 2})
	if err != nil || len(orders.Items) != 1 || orders.Items[0].OrderNumber != 10 {
		t.Fatalf("orders=%+v err=%v", orders, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/orders", "page=2")
	for _, err := range rec.client.Customers.AllOrders(testCtx, 9, nil) {
		if err != nil {
			t.Fatal(err)
		}
	}
	rec.expect(t, "GET", "/v1/customers/9/orders", "page=1&per_page=50")

	rec = customersRecorder(t, 200, `{"customer_id":9}`)
	if _, _, err := rec.client.Customers.VIPInfo(testCtx, 9); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "GET", "/v1/customers/9/vip_info", "")
}

func TestCustomersV2Requests(t *testing.T) {
	rec := customersRecorder(t, 200, `[{"id":1,"vip_info":{"customer_id":1}}]`)
	page, err := rec.client.Customers.ListV2(testCtx, &CustomerListV2Options{
		IDs:     []int64{1, 2},
		Include: []CustomerInclude{CustomerIncludeUIDProviders, CustomerIncludeVIPInfo},
	})
	if err != nil || len(page.Items) != 1 || page.Items[0].VIPInfo == nil {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	rec.expect(t, "GET", "/v2/customers", "ids=1%2C2&include_params=uid_providers%2Cvip_info")

	rec = customersRecorder(t, 200, `[{"id":1}]`)
	for _, err := range rec.client.Customers.AllV2(testCtx, &CustomerListV2Options{ListOptions: ListOptions{PerPage: 5}}) {
		if err != nil {
			t.Fatal(err)
		}
	}
	rec.expect(t, "GET", "/v2/customers", "page=1&per_page=5")

	rec = customersRecorder(t, 200, `{"id":1,"uid_providers":[{"provider_type":"line","uid":"U1"}]}`)
	cu, _, err := rec.client.Customers.GetByUIDProvider(testCtx, CustomerUIDProviderLine, "U1")
	if err != nil || cu.ID != 1 || cu.UIDProviders[0].UID != "U1" {
		t.Fatalf("cu=%+v err=%v", cu, err)
	}
	rec.expect(t, "GET", "/v2/customers/by_uid_provider", "provider=line&uid=U1")

	rec = customersRecorder(t, 200, `{"id":1}`)
	cu, _, err = rec.client.Customers.OAuth(testCtx, &CustomerOAuthRequest{UID: "U1", Provider: CustomerUIDProviderLine})
	if err != nil || cu.ID != 1 {
		t.Fatalf("cu=%+v err=%v", cu, err)
	}
	rec.expect(t, "POST", "/v2/customer_oauth", "")
	rec.expectBody(t, `{"uid":"U1","provider":"line"}`)
}

func TestCustomersBonusPointRequests(t *testing.T) {
	rec := customersRecorder(t, 201, `{"id":1,"points":10.0}`)
	bp, _, err := rec.client.Customers.CreateBonusPoints(testCtx, 9, &CustomerBonusPointCreateRequest{
		Title: "gift", Points: MoneyFromInt(10),
	})
	if err != nil || bp.Points != MoneyFromInt(10) {
		t.Fatalf("bp=%+v err=%v", bp, err)
	}
	rec.expect(t, "POST", "/v1/customers/9/bonus_points", "")
	rec.expectBody(t, `{"title":"gift","points":10,"deadline":"0"}`)

	rec = customersRecorder(t, 201, `{"id":1}`)
	deadline, _ := ParseTime("2027-01-01 00:00:00")
	_, _, err = rec.client.Customers.CreateBonusPoints(testCtx, 9, &CustomerBonusPointCreateRequest{
		Title: "gift", Points: MoneyFromInt(10), ConsumptionPrice: customersPtr(Money(0)), Deadline: deadline,
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expectBody(t, `{"title":"gift","points":10,"consumption_price":0,"deadline":"2027-01-01 00:00:00"}`)

	rec = customersRecorder(t, 200, `{"id":2}`)
	_, _, err = rec.client.Customers.UpdateBonusPoints(testCtx, 9, 2, &CustomerBonusPointUpdateRequest{
		BonusPointsDiff: customersPtr(MoneyFromInt(-5)), UnusedPoints: customersPtr(Money(0)),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "PUT", "/v1/customers/9/bonus_points/2", "")
	rec.expectBody(t, `{"bonus_points_diff":-5,"unused_points":0}`)

	rec = customersRecorder(t, 200, `{"id":2}`)
	if _, err := rec.client.Customers.DeleteBonusPoints(testCtx, 9, 2); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "DELETE", "/v1/customers/9/bonus_points/2", "")

	rec = customersRecorder(t, 200, `[]`)
	if _, _, err := rec.client.Customers.ListBonusPoints(testCtx, 9); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "GET", "/v1/customers/9/bonus_points", "")
}

func TestCustomersCouponRequests(t *testing.T) {
	rec := customersRecorder(t, 200, `[]`)
	_, err := rec.client.Customers.ListCoupons(testCtx, 9, &CustomerCouponListOptions{
		ListOptions:             ListOptions{PerPage: 20},
		OrderPriceThresholdLTEQ: MoneyFromInt(500),
		CouponTypes:             []CustomerCouponType{CustomerCouponTypeAmount, CustomerCouponTypePercent},
		Tags:                    []string{"a", "b"},
		OnlyValid:               customersPtr(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "GET", "/v1/customers/9/coupons", "coupon_types=amount%2Cpercent&only_valid=false&order_price_threshold_lteq=500&per_page=20&tags=a%2Cb")

	rec = customersRecorder(t, 200, `{"title":"x","coupon_status":"expired"}`)
	cp, _, err := rec.client.Customers.GetCoupon(testCtx, 9, 3)
	if err != nil || cp.ID != 3 || cp.Usable() {
		t.Fatalf("cp=%+v err=%v", cp, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/coupons/3", "")

	rec = customersRecorder(t, 201, `{"id":4}`)
	_, _, err = rec.client.Customers.CreateCoupon(testCtx, 9, &CustomerCouponCreateRequest{
		Title: "t", Code: "C1", CouponType: CustomerCouponTypeAmount, Value: MoneyFromInt(50),
		OrderPriceThreshold: 0, UsageLimit: 1, EndDate: NewDate(2026, 12, 31),
		RestrictStrategy: CustomerCouponRestrictForbidden, RestrictCampaigns: []string{"vip_discount"},
		ProductIDs: []int64{7},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "POST", "/v1/customers/9/coupons", "")
	rec.expectBody(t, `{"title":"t","code":"C1","coupon_type":"amount","value":50,"order_price_threshold":0,"usage_limit":1,"end_date":"2026-12-31","restrict_strategy":"forbidden","restrict_campaigns":["vip_discount"],"product_ids":[7]}`)

	rec = customersRecorder(t, 200, `{"id":4}`)
	_, _, err = rec.client.Customers.UpdateCoupon(testCtx, 9, 4, &CustomerCouponUpdateRequest{
		UsedTimes: customersPtr(0), ConcurrentlyApply: customersPtr(false), Tags: []string{"x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "PUT", "/v1/customers/9/coupons/4", "")
	rec.expectBody(t, `{"used_times":0,"concurrently_apply":false,"tags":["x"]}`)

	rec = customersRecorder(t, 200, `{"id":4}`)
	if _, err := rec.client.Customers.DeleteCoupon(testCtx, 9, 4); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "DELETE", "/v1/customers/9/coupons/4", "")

	rec = customersRecorder(t, 200, `[]`)
	if _, err := rec.client.Customers.ListShopCoupons(testCtx, 9, &CustomerShopCouponListOptions{OnlyValid: customersPtr(true)}); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "GET", "/v1/customers/9/shop_coupons", "only_valid=true")

	rec = customersRecorder(t, 201, `{"used_times":2}`)
	sc, _, err := rec.client.Customers.DecrementShopCoupon(testCtx, 9, 5)
	if err != nil || sc.ID != 5 || sc.UsedTimes != 2 {
		t.Fatalf("sc=%+v err=%v", sc, err)
	}
	rec.expect(t, "POST", "/v1/customers/9/shop_coupons/5/decrement", "")
	if rec.body != "" {
		t.Errorf("decrement sent a body: %s", rec.body)
	}
}

func TestCustomersCustomFieldRequests(t *testing.T) {
	rec := customersRecorder(t, 200, `[{"id":1,"name":"n","label":"L","value":"v"}]`)
	page, err := rec.client.Customers.ListCustomFields(testCtx, 9, nil)
	if err != nil || len(page.Items) != 1 || page.Items[0].Label != "L" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/custom_fields", "")

	rec = customersRecorder(t, 200, `{"name":"n","value":"v"}`)
	f, _, err := rec.client.Customers.GetCustomField(testCtx, 9, 1)
	if err != nil || f.ID != 1 {
		t.Fatalf("f=%+v err=%v", f, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/custom_fields/1", "")

	rec = customersRecorder(t, 201, `{"id":2}`)
	if _, _, err := rec.client.Customers.CreateCustomField(testCtx, 9, &CustomerCustomFieldCreateRequest{Name: "n", Value: ""}); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "POST", "/v1/customers/9/custom_fields", "")
	rec.expectBody(t, `{"name":"n","value":""}`)

	rec = customersRecorder(t, 200, `[]`)
	_, _, err = rec.client.Customers.UpsertCustomFields(testCtx, 9, []CustomerCustomFieldInput{
		{ID: 2, Name: "n", Value: "v2"}, {Name: "m", Value: "w"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "PUT", "/v1/customers/9/custom_fields", "")
	rec.expectBody(t, `{"custom_fields":[{"id":2,"name":"n","value":"v2"},{"name":"m","value":"w"}]}`)

	rec = customersRecorder(t, 200, `{"id":2}`)
	if _, _, err := rec.client.Customers.UpdateCustomField(testCtx, 9, 2, &CustomerCustomFieldUpdateRequest{Value: customersPtr("")}); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "PUT", "/v1/customers/9/custom_fields/2", "")
	rec.expectBody(t, `{"value":""}`)

	rec = customersRecorder(t, 200, `{"id":2}`)
	if _, err := rec.client.Customers.DeleteCustomField(testCtx, 9, 2); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "DELETE", "/v1/customers/9/custom_fields/2", "")

	rec = customersRecorder(t, 200, `[]`)
	if _, err := rec.client.Customers.DeleteCustomFields(testCtx, 9, []int64{2, 3}); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "DELETE", "/v1/customers/9/custom_fields", "custom_field_ids%5B%5D=2&custom_field_ids%5B%5D=3")
}

func TestCustomersGroupRequests(t *testing.T) {
	rec := customersRecorder(t, 200, `[]`)
	_, err := rec.client.Customers.ListGroups(testCtx, &CustomerGroupListOptions{
		IDs: []int64{1, 2}, Name: "vip", GroupType: CustomerGroupTypeCustom,
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "GET", "/v1/customers/customer_groups", "customer_group_ids=1%2C2&group_type=custom&name=vip")

	rec = customersRecorder(t, 201, `{"job_id":"42"}`)
	job, _, err := rec.client.Customers.ScheduleGroupFilter(testCtx, 8)
	if err != nil || job.JobID != "42" {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	rec.expect(t, "POST", "/v1/customers/customer_groups/schedule_customer_filter_worker", "")
	rec.expectBody(t, `{"customer_group_id":8}`)

	rec = customersRecorder(t, 200, `{"job_id":"43"}`)
	job, _, err = rec.client.Customers.UpdateGroupAmounts(testCtx)
	if err != nil || job.JobID != "43" {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	rec.expect(t, "PUT", "/v1/customers/customer_groups/update_customer_group_amounts", "")
	if rec.body != "" {
		t.Errorf("unexpected body %q", rec.body)
	}

	rec = customersRecorder(t, 200, `[{"id":1,"email":"a@example.com"}]`)
	st, _, err := rec.client.Customers.CheckGroupStatus(testCtx, "42", &ListOptions{Page: 2})
	if err != nil || !st.Done || len(st.Members) != 1 {
		t.Fatalf("st=%+v err=%v", st, err)
	}
	rec.expect(t, "GET", "/v1/customers/customer_groups/check_status/42", "page=2")
}

func TestCustomersOtherValidOrderRequests(t *testing.T) {
	rec := customersRecorder(t, 200, `[{"id":1,"price":100.0,"valid_at":"2026-01-02","invalid_at":null}]`)
	page, err := rec.client.Customers.ListOtherValidOrders(testCtx, 9, &ListOptions{Page: 1})
	if err != nil || len(page.Items) != 1 || page.Items[0].InvalidAt != nil || page.Items[0].ValidAt.String() != "2026-01-02 00:00:00" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/other_valid_orders", "page=1")

	rec = customersRecorder(t, 200, `{"id":1,"extra_infos":{"channel_name":"shopee","products":[{"name":"p","quantity":2,"price":10.0}]}}`)
	o, _, err := rec.client.Customers.GetOtherValidOrder(testCtx, 9, 1)
	if err != nil || o.ExtraInfos == nil || o.ExtraInfos.Products[0].Price != MoneyFromInt(10) {
		t.Fatalf("o=%+v err=%v", o, err)
	}
	rec.expect(t, "GET", "/v1/customers/9/other_valid_orders/1", "")

	rec = customersRecorder(t, 201, `{"id":2}`)
	_, _, err = rec.client.Customers.CreateOtherValidOrder(testCtx, 9, &CustomerOtherValidOrderCreateRequest{
		Price: MoneyFromInt(300), ValidAt: NewDate(2026, 2, 3),
		ExtraInfos: &CustomerOtherValidOrderExtraInfo{ChannelName: "shopee", Products: []CustomerOtherValidOrderProduct{{Name: "p", Quantity: 1, Price: MoneyFromInt(300)}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "POST", "/v1/customers/9/other_valid_orders", "")
	rec.expectBody(t, `{"price":300,"valid_at":"2026-02-03","extra_infos":{"channel_name":"shopee","products":[{"name":"p","quantity":1,"price":300}]}}`)

	rec = customersRecorder(t, 200, `{"id":2}`)
	_, _, err = rec.client.Customers.UpdateOtherValidOrder(testCtx, 9, 2, &CustomerOtherValidOrderUpdateRequest{Price: customersPtr(Money(0))})
	if err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "PUT", "/v1/customers/9/other_valid_orders/2", "")
	rec.expectBody(t, `{"price":0}`)

	rec = customersRecorder(t, 200, `{"id":2}`)
	if _, err := rec.client.Customers.DeleteOtherValidOrder(testCtx, 9, 2); err != nil {
		t.Fatal(err)
	}
	rec.expect(t, "DELETE", "/v1/customers/9/other_valid_orders/2", "")

	rec = customersRecorder(t, 201, `{"id":2,"invalid_at":"2026-03-01 10:00:00"}`)
	inv, _, err := rec.client.Customers.InvalidateOtherValidOrder(testCtx, 9, 2)
	if err != nil || inv.InvalidAt == nil || !strings.HasPrefix(inv.InvalidAt.String(), "2026-03-01") {
		t.Fatalf("inv=%+v err=%v", inv, err)
	}
	rec.expect(t, "POST", "/v1/customers/9/other_valid_orders/2/invalid", "")
}
