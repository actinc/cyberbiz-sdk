package cyberbiz

import "testing"

func TestVIPGroupsGoldenOverview(t *testing.T) {
	var out VIPGroupsOverview
	decodeGolden(t, "v1/GET_v1_vip_groups.json", &out)
	if len(out.Current) != 0 || out.Current == nil {
		t.Errorf("current = %#v", out.Current)
	}
	if len(out.Draft) != 1 {
		t.Fatalf("draft = %+v", out.Draft)
	}
	g := out.Draft[0]
	if g.ID != 3924 || g.Position != -1 || g.DescriptionURL != "" {
		t.Errorf("group = %+v", g)
	}
	if g.CustomerTags == nil || g.VIPGroupLevels == nil {
		t.Errorf("empty arrays should decode to empty slices: %+v", g)
	}
}

func TestVIPGroupsGoldenGet(t *testing.T) {
	c := goldenServer(t)
	g, resp, err := c.VIPGroups.Get(testCtx, 3924)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || g.ID != 3924 || g.Name != "REDACTED" {
		t.Errorf("group = %+v", g)
	}
	ov, _, err := c.VIPGroups.List(testCtx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ov.Draft) != 1 || ov.Draft[0].ID != 3924 {
		t.Errorf("overview = %+v", ov)
	}
}

func TestVIPGroupLevelDecodesPostmanShape(t *testing.T) {
	// The swagger contradicts itself on several level fields; this pins the
	// interpretation the model takes.
	c, _ := New("tok")
	var lvl VIPGroupLevel
	body := `{"id":1,"position":2,"name":"金卡","validity_days":365,"upgrade_condition_total_spent":null,
		"upgrade_condition_total_spent_in_validity_days":10000.0,"bonus_point_enabled":true,"bonus_point_threshold":100,
		"bonus_point_value":2,"birth_gift_before_days":7,"birth_gift_setting":{"bonus":{"enabled":true,"value":50,"expiry_days":0},
		"coupon":{"enabled":false,"presets":[{"coupon_type_id":2,"value":10,"code":"BDAY","order_price_threshold":0,
		"usable_days":30,"usage_limit":1,"product_tags":[],"restrict_strategy":"unrestricted","restrict_campaigns":[]}]}},
		"order_discount_enabled":true,"order_discount_value":95.5,"order_discount_setting":{"restrict_strategy":"forbidden","restrict_campaigns":["vip_discount"]},
		"free_shipping_enabled":false,"free_shipping_setting":null}`
	if err := c.decode([]byte(body), &lvl); err != nil {
		t.Fatal(err)
	}
	if lvl.UpgradeConditionTotalSpent != 0 || lvl.UpgradeConditionTotalSpentInValidityDays != MoneyFromInt(10000) {
		t.Errorf("conditions = %+v", lvl)
	}
	if lvl.BirthGiftSetting == nil || !lvl.BirthGiftSetting.Bonus.Enabled || lvl.BirthGiftSetting.Coupon.Presets[0].CouponTypeID != VIPCouponTypePercent {
		t.Errorf("gift = %+v", lvl.BirthGiftSetting)
	}
	if lvl.OrderDiscountValue != 95.5 || lvl.OrderDiscountSetting.RestrictStrategy != RestrictStrategyForbidden || lvl.FreeShippingSetting != nil {
		t.Errorf("discount = %+v", lvl)
	}
}

func TestVIPGroupsRequests(t *testing.T) {
	c, call := discountsSpyClient(t, 201, `{"id":5,"name":"x","customer_tags":[],"vip_group_levels":[]}`)

	url := ""
	pos := 0
	g, _, err := c.VIPGroups.Create(testCtx, &VIPGroupCreateRequest{Name: "金卡", CustomerTags: "gold,vip", DescriptionURL: &url, Position: &pos})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/vip_groups")
	discountsAssertJSONBody(t, call, `{"name":"金卡","customer_tags":"gold,vip","description_url":"","position":0}`)
	if g.ID != 5 {
		t.Errorf("group = %+v", g)
	}

	if _, err := c.VIPGroups.Delete(testCtx, 5); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "DELETE", "/v1/vip_groups/5")

	if _, err := c.VIPGroups.ApplyDraft(testCtx, NewDate(2026, 10, 1)); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/vip_groups/apply_draft")
	discountsAssertJSONBody(t, call, `{"start_date":"2026-10-01"}`)

	if _, err := c.VIPGroups.CancelDraft(testCtx); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/vip_groups/cancel_draft")
	if len(call.Body) != 0 {
		t.Errorf("cancel_draft sent a body: %s", call.Body)
	}
}

func TestVIPGroupLevelRequests(t *testing.T) {
	c, call := discountsSpyClient(t, 200, `{"id":9,"name":"x"}`)

	name := "銀卡"
	days := 365
	spent := Money(0)
	on := true
	pct := 90.0
	lvl, _, err := c.VIPGroups.CreateLevel(testCtx, 5, &VIPGroupLevelRequest{
		Name: &name, ValidityDays: &days, UpgradeConditionTotalSpent: &spent,
		OrderDiscountEnabled: &on, OrderDiscountValue: &pct,
		OrderDiscountSetting: &VIPRestrictSetting{RestrictStrategy: RestrictStrategyUnrestricted, RestrictCampaigns: []RestrictCampaign{}},
		BirthGiftSetting:     &VIPGiftSetting{Bonus: &VIPBonusSetting{Enabled: true, Value: MoneyFromInt(50)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/vip_groups/5/vip_group_levels")
	discountsAssertJSONBody(t, call, `{"name":"銀卡","validity_days":365,"upgrade_condition_total_spent":0,
		"order_discount_enabled":true,"order_discount_value":90,
		"order_discount_setting":{"restrict_strategy":"unrestricted","restrict_campaigns":[]},
		"birth_gift_setting":{"bonus":{"enabled":true,"value":50,"expiry_days":0},"coupon":null}}`)
	if lvl.ID != 9 {
		t.Errorf("level = %+v", lvl)
	}

	if _, _, err := c.VIPGroups.UpdateLevel(testCtx, 5, 9, &VIPGroupLevelRequest{Name: &name}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/vip_groups/5/vip_group_levels/9")
	discountsAssertJSONBody(t, call, `{"name":"銀卡"}`)

	if _, err := c.VIPGroups.DeleteLevel(testCtx, 5, 9); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "DELETE", "/v1/vip_groups/5/vip_group_levels/9")
}
