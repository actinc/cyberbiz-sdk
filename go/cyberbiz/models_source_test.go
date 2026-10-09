package cyberbiz

import (
	"errors"
	"net/http"
	"testing"
)

// These tests pin response shapes that the recorded Golden Files cannot
// show (the test shop returned empty lists for these resources) and that
// were confirmed against the platform source instead (CBSDK-44).

// sourceShapeClient serves body with status for every request.
func sourceShapeClient(t *testing.T, status int, body string) *Client {
	t.Helper()
	c, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	return c
}

func TestBranchStoreCoordinatesAcceptDecimalStrings(t *testing.T) {
	c, _ := New("tok")
	tests := []struct {
		body     string
		lat, lng Coordinate
	}{
		{`{"lat":"25.0330000","lng":"121.5654000"}`, 25.033, 121.5654},
		{`{"lat":25.033,"lng":121.5654}`, 25.033, 121.5654},
		{`{"lat":null,"lng":""}`, 0, 0},
	}
	for _, tt := range tests {
		var bs BranchStore
		if err := c.decode([]byte(tt.body), &bs); err != nil {
			t.Fatalf("%s: %v", tt.body, err)
		}
		if bs.Lat != tt.lat || bs.Lng != tt.lng {
			t.Errorf("%s: got %v,%v", tt.body, bs.Lat, bs.Lng)
		}
		var ob OrderBranchStore
		if err := c.decode([]byte(tt.body), &ob); err != nil || ob.Lat != tt.lat {
			t.Errorf("order store %s: %v %v", tt.body, ob.Lat, err)
		}
	}
	var bs BranchStore
	if err := c.decode([]byte(`{"lat":"north"}`), &bs); err == nil {
		t.Error("non-numeric coordinate accepted")
	}
}

func TestPosProductTagsAreObjects(t *testing.T) {
	c, _ := New("tok")
	var p PosProduct
	if err := c.decode([]byte(`{"id":1,"tags":[{"id":7,"name":"新品","category":0}]}`), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Tags) != 1 || p.Tags[0].ID != 7 || p.Tags[0].Name != "新品" {
		t.Errorf("tags = %+v", p.Tags)
	}
}

func TestPeriodicOrderNumberAndScheduleShapes(t *testing.T) {
	c, _ := New("tok")
	var po PeriodicOrder
	body := `{"id":1,"number":1024,"periodic":{"type":"monthly","month":"2","week":"1","day":3}}`
	if err := c.decode([]byte(body), &po); err != nil {
		t.Fatal(err)
	}
	if po.Number != "1024" {
		t.Errorf("number = %q", po.Number)
	}
	if po.Periodic == nil || po.Periodic.Month != 2 || po.Periodic.Week != 1 || po.Periodic.Day != 3 {
		t.Errorf("periodic = %+v", po.Periodic)
	}
	if err := c.decode([]byte(`{"periodic":{"day":"x"}}`), &po); err == nil {
		t.Error("non-numeric schedule value accepted")
	}
}

func TestAffiliatesListUnwrapsEnvelope(t *testing.T) {
	body := `{"affiliate_vendor_orders":[{"uid":"u5","order":{"order_number":1020,"line_items":[{"id":1,
		"related_items":[{"quantity":2,"items":[{"id":11,"product_id":3}]}]}]}}]}`
	page, err := sourceShapeClient(t, 200, body).Affiliates.List(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].UID != "u5" || page.Items[0].Order == nil {
		t.Fatalf("items = %+v", page.Items)
	}
	ri := page.Items[0].Order.LineItems[0].RelatedItems
	if len(ri) != 1 || ri[0].Quantity != 2 || len(ri[0].Items) != 1 || ri[0].Items[0].ID != 11 {
		t.Errorf("related items = %+v", ri)
	}

	// A page past the end is a bare empty array.
	page, err = sourceShapeClient(t, 200, `[]`).Affiliates.List(testCtx, nil)
	if err != nil || len(page.Items) != 0 {
		t.Errorf("empty page = %+v, %v", page, err)
	}
}

func TestMessagePostOrderAndAdminAreObjects(t *testing.T) {
	c, _ := New("tok")
	var mp CustomerMessagePost
	body := `{"id":1,"order":{"id":9,"type":"Order","name":"#1001"},
		"comments":[{"id":2,"role":"admin","admin":{"id":3,"name":"店長","email":"admin@example.com"}},
		{"id":4,"role":"customer","admin":null}]}`
	if err := c.decode([]byte(body), &mp); err != nil {
		t.Fatal(err)
	}
	if mp.Order == nil || mp.Order.ID != 9 || mp.Order.Name != "#1001" {
		t.Errorf("order = %+v", mp.Order)
	}
	if mp.Comments[0].Admin == nil || mp.Comments[0].Admin.Name != "店長" || mp.Comments[1].Admin != nil {
		t.Errorf("comments = %+v", mp.Comments)
	}
	var none CustomerMessagePost
	if err := c.decode([]byte(`{"id":1,"order":null}`), &none); err != nil || none.Order != nil {
		t.Errorf("null order: %+v %v", none.Order, err)
	}
}

func TestUpdateGroupAmountsReportsBusy(t *testing.T) {
	c := sourceShapeClient(t, 200, `{"success":false,"errors":["請稍後再試"]}`)
	job, _, err := c.Customers.UpdateGroupAmounts(testCtx)
	var apiErr *APIError
	if job != nil || !errors.As(err, &apiErr) || len(apiErr.Messages) != 1 || apiErr.Messages[0] != "請稍後再試" {
		t.Fatalf("job = %+v err = %v", job, err)
	}

	job, _, err = sourceShapeClient(t, 200, `{"job_id":"abc"}`).Customers.UpdateGroupAmounts(testCtx)
	if err != nil || job.JobID != "abc" {
		t.Errorf("job = %+v err = %v", job, err)
	}
}

func TestSupportShippingTrackingNumbersIsAList(t *testing.T) {
	c, _ := New("tok")
	var res SupportShippingBatchResult
	body := `{"request_id":"r","results":[{"order_id":1,"tracking_numbers":[
		{"tracking_company":"tcat","tracking_number":"T1"},{"tracking_company":"tcat","tracking_number":"T2"}]},
		{"order_id":2,"tracking_numbers":[]}]}`
	if err := c.decode([]byte(body), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 2 || len(res.Results[0].TrackingNumbers) != 2 || res.Results[0].TrackingNumbers[1].TrackingNumber != "T2" {
		t.Errorf("results = %+v", res.Results)
	}
}

func TestVIPLabelsDecodeToCodes(t *testing.T) {
	c, _ := New("tok")
	tests := []struct {
		rule, promotion string
		wantRule        VIPRuleType
		wantPromotion   VIPPromotionType
	}{
		{"訂單數量", "享優惠", VIPRuleOrderCount, VIPPromotionDiscount},
		{"金額累積", "訂單免運費", VIPRuleTotalPrice, VIPPromotionFreeShipping},
		{"顧客標籤", "free_shipping", VIPRuleCustomersTag, VIPPromotionFreeShipping},
		{"新規則", "新優惠", "新規則", "新優惠"},
	}
	for _, tt := range tests {
		var v VIPCollection
		body := `{"id":1,"rule":{"rule_type":"` + tt.rule + `"},"promotion":{"promotion_type":"` + tt.promotion + `"}}`
		if err := c.decode([]byte(body), &v); err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		if v.Rule.RuleType != tt.wantRule || v.Promotion.PromotionType != tt.wantPromotion {
			t.Errorf("%s/%s: got %q/%q", tt.rule, tt.promotion, v.Rule.RuleType, v.Promotion.PromotionType)
		}
	}
}

func TestRegisterCouponRuleExpireDayIsANumber(t *testing.T) {
	c, _ := New("tok")
	var r RegisterCouponRule
	if err := c.decode([]byte(`{"enabled":true,"expire_day":30}`), &r); err != nil {
		t.Fatal(err)
	}
	if r.ExpireDay != 30 {
		t.Errorf("expire_day = %d", r.ExpireDay)
	}
}
