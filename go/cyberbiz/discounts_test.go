package cyberbiz

import (
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// discountsSpyCall is what the last request to a discountsSpyClient looked like.
type discountsSpyCall struct {
	Method      string
	Path        string
	Query       url.Values
	Body        []byte
	ContentType string
}

// discountsSpyClient returns a client whose server records every request into the
// returned discountsSpyCall and answers with status and body.
func discountsSpyClient(t *testing.T, status int, body string) (*Client, *discountsSpyCall) {
	t.Helper()
	call := &discountsSpyCall{}
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		*call = discountsSpyCall{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(),
			Body: data, ContentType: r.Header.Get("Content-Type")}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}, WithMaxRetries(0))
	return c, call
}

// discountsAssertJSONBody fails unless the recorded body is exactly want, compared as
// JSON so key order does not matter.
func discountsAssertJSONBody(t *testing.T, call *discountsSpyCall, want string) {
	t.Helper()
	var got, exp any
	if err := json.Unmarshal(call.Body, &got); err != nil {
		t.Fatalf("body is not JSON: %v: %s", err, call.Body)
	}
	if err := json.Unmarshal([]byte(want), &exp); err != nil {
		t.Fatalf("want is not JSON: %v", err)
	}
	g, _ := json.Marshal(got, json.Deterministic(true))
	e, _ := json.Marshal(exp, json.Deterministic(true))
	if string(g) != string(e) {
		t.Errorf("body = %s\nwant   %s", g, e)
	}
}

// discountsAssertCall checks the recorded method and path.
func discountsAssertCall(t *testing.T, call *discountsSpyCall, method, path string) {
	t.Helper()
	if call.Method != method || call.Path != path {
		t.Errorf("request = %s %s, want %s %s", call.Method, call.Path, method, path)
	}
}

func TestDiscountGoldenList(t *testing.T) {
	var out []Discount
	decodeGolden(t, "v1/GET_v1_discounts.json", &out)
	if len(out) != 1 {
		t.Fatalf("len = %d", len(out))
	}
	d := out[0]
	if d.ID != 72629 || d.DiscountTypeName != "金額" || d.DiscountValue != "5.0元" {
		t.Errorf("discount = %+v", d)
	}
	if d.Threshold != MoneyFromInt(10) {
		t.Errorf("threshold = %v", d.Threshold)
	}
	if !d.StartAt.IsZero() || !d.EndAt.IsZero() {
		t.Errorf("null times should decode to zero: %v %v", d.StartAt, d.EndAt)
	}
	if d.SalesChannelID != SalesChannelAll {
		t.Errorf("sales channel = %v", d.SalesChannelID)
	}
}

func TestDiscountGoldenGet(t *testing.T) {
	c := goldenServer(t)
	d, resp, err := c.Discounts.Get(testCtx, 72629)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || d.ID != 72629 || d.Threshold != MoneyFromInt(10) {
		t.Errorf("got %+v", d)
	}
}

func TestDiscountGoldenListRoundTrip(t *testing.T) {
	c := goldenServer(t)
	page, err := c.Discounts.List(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "REDACTED" || page.Pagination.Total != 1 {
		t.Errorf("page = %+v", page)
	}
}

func TestDiscountRequests(t *testing.T) {
	lc, lcall := discountsSpyClient(t, 200, `[]`)
	if _, err := lc.Discounts.List(testCtx, &ListOptions{Page: 2, PerPage: 10}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "GET", "/v1/discounts")
	if lcall.Query.Get("page") != "2" || lcall.Query.Get("per_page") != "10" {
		t.Errorf("query = %v", lcall.Query)
	}

	c, call := discountsSpyClient(t, 200, `{"name":"x","threshold":10.0}`)

	days := 3
	start, _ := ParseTime("2026-01-01 00:00:00")
	_, _, err := c.Discounts.Create(testCtx, &DiscountCreateRequest{
		Name: "新春", DiscountType: DiscountTypePercent, Value: MoneyFromInt(10),
		Threshold: MoneyFromInt(500), StartAt: start, Days: &days, SalesChannelID: SalesChannelEC,
	})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/discounts")
	discountsAssertJSONBody(t, call, `{"name":"新春","discount_type":"percent","value":10,"threshold":500,
		"start_at":"2026-01-01 00:00:00","days":3,"sales_channel_id":2}`)

	off := false
	zero := Money(0)
	if _, _, err := c.Discounts.Update(testCtx, 7, &DiscountUpdateRequest{ConcurrentlyApply: &off, Threshold: &zero}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/discounts/7")
	discountsAssertJSONBody(t, call, `{"concurrently_apply":false,"threshold":0}`)

	if _, err := c.Discounts.Delete(testCtx, 7); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "DELETE", "/v1/discounts/7")

	if _, _, err := c.Discounts.UpdateRegisterCouponRule(testCtx, &RegisterCouponRuleRequest{
		Enabled: true, Value: MoneyFromInt(100), OrderPriceThreshold: MoneyFromInt(1000), UsageLimit: 1, ExpireDay: 30,
	}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/discounts/register_coupon_rule")
	discountsAssertJSONBody(t, call, `{"enabled":true,"value":100,"order_price_threshold":1000,"usage_limit":1,"expire_day":30,"concurrently_apply":false}`)
}

func TestDiscountAllWalksPages(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "1" {
			w.Header().Set("X-Next-Page", "2")
			_, _ = w.Write([]byte(`[{"id":1},{"id":2}]`))
			return
		}
		_, _ = w.Write([]byte(`[{"id":3}]`))
	})
	var n int
	for _, err := range c.Discounts.All(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 3 {
		t.Errorf("walked %d items", n)
	}
}

func TestShopCouponGolden(t *testing.T) {
	var list []ShopCoupon
	decodeGolden(t, "v1/GET_v1_shop_coupons.json", &list)
	if len(list) != 1 {
		t.Fatalf("len = %d", len(list))
	}
	cp := list[0]
	if cp.ID != 345196115 || cp.Code != "SHMHHDNL" || cp.Title != "測試優惠碼" {
		t.Errorf("coupon = %+v", cp)
	}
	if cp.OrderPriceThreshold != 0 || !cp.ConcurrentlyApply || cp.UsageLimit != 1000000000 || cp.UsedTimes != 1 {
		t.Errorf("coupon = %+v", cp)
	}
	if cp.GiftOrderID != 0 || cp.GiftDays != 0 || !cp.StartDate.IsZero() {
		t.Errorf("null fields should be zero: %+v", cp)
	}
	if cp.RestrictStrategy != RestrictStrategyUnrestricted || cp.CouponStatus != CouponStatusNoExpireDate || cp.GiftOrderStatus != "" {
		t.Errorf("enums = %q %q %q", cp.RestrictStrategy, cp.CouponStatus, cp.GiftOrderStatus)
	}
	if !cp.Usable() || !cp.Valid || !cp.CustomerUsable {
		t.Errorf("usable flags: %+v", cp)
	}
	if cp.ProductIDs == nil || cp.Tags == nil || cp.PosShopIDs == nil {
		t.Errorf("empty arrays should decode to empty slices")
	}

	var one ShopCoupon
	decodeGolden(t, "v1/GET_v1_shop_coupons_{id}.json", &one)
	if one.ID != 0 || one.Code != "SHMHHDNL" {
		t.Errorf("detail = %+v", one)
	}
}

func TestShopCouponGoldenServer(t *testing.T) {
	c := goldenServer(t)
	cp, _, err := c.Discounts.GetCoupon(testCtx, 345196115)
	if err != nil {
		t.Fatal(err)
	}
	if cp.ID != 345196115 || cp.Code != "SHMHHDNL" {
		t.Errorf("coupon = %+v", cp)
	}
	page, err := c.Discounts.ListCoupons(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != 345196115 {
		t.Errorf("page = %+v", page.Items)
	}
}

func TestShopCouponRequests(t *testing.T) {
	lc, lcall := discountsSpyClient(t, 200, `[]`)
	valid := true
	_, err := lc.Discounts.ListCoupons(testCtx, &ShopCouponListOptions{
		ListOptions:            ListOptions{PerPage: 5},
		OrderPriceThresholdLTE: MoneyFromInt(100),
		CouponTypes:            []CouponType{CouponTypeAmount, CouponTypePercent},
		Tags:                   []string{"a", "b"},
		OnlyValid:              &valid,
	})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "GET", "/v1/shop_coupons")
	want := url.Values{"per_page": {"5"}, "order_price_threshold_lteq": {"100"},
		"coupon_types": {"amount,percent"}, "tags": {"a,b"}, "only_valid": {"true"}}
	if lcall.Query.Encode() != want.Encode() {
		t.Errorf("query = %v", lcall.Query)
	}

	c, call := discountsSpyClient(t, 201, `{"title":"x"}`)

	enabled := true
	limit := 2
	_, _, err = c.Discounts.CreateCoupon(testCtx, &ShopCouponCreateRequest{
		Title: "VIP", Code: "VIP100", CouponType: CouponTypeAmount, Value: MoneyFromInt(100),
		OrderPriceThreshold: MoneyFromInt(1000), StartDate: NewDate(2026, 1, 1), UsageLimit: 10,
		AccountUsageLimitEnabled: &enabled, AccountUsageLimit: &limit,
		RestrictStrategy: RestrictStrategyForbidden, RestrictCampaigns: []RestrictCampaign{RestrictCampaignVIPDiscount},
		Tags: []string{"hot"}, ProductIDs: []int64{1, 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/shop_coupons")
	discountsAssertJSONBody(t, call, `{"title":"VIP","code":"VIP100","coupon_type":"amount","value":100,
		"order_price_threshold":1000,"start_date":"2026-01-01","concurrently_apply":false,"usage_limit":10,
		"account_usage_limit_enabled":true,"account_usage_limit":2,"restrict_strategy":"forbidden",
		"restrict_campaigns":["vip_discount"],"tags":["hot"],"product_ids":[1,2]}`)

	empty := ""
	used := 0
	if _, _, err := c.Discounts.UpdateCoupon(testCtx, 9, &ShopCouponUpdateRequest{Title: &empty, UsedTimes: &used}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/shop_coupons/9")
	discountsAssertJSONBody(t, call, `{"title":"","used_times":0}`)

	if _, err := c.Discounts.DeleteCoupon(testCtx, 9); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "DELETE", "/v1/shop_coupons/9")
	if len(call.Body) != 0 {
		t.Errorf("delete sent a body: %s", call.Body)
	}
}

func TestDiscountGetNullIsNotFound(t *testing.T) {
	c, _ := discountsSpyClient(t, 200, "null")
	_, _, err := c.Discounts.Get(testCtx, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(err.Error(), "v1/discounts/1") {
		t.Errorf("error = %v", err)
	}
}
