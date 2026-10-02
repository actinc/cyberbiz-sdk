package cyberbiz

import "testing"

func TestBonusRuleGolden(t *testing.T) {
	var out BonusRule
	decodeGolden(t, "v1/GET_v1_bonus_rule.json", &out)
	if !out.Enabled || out.SignupBonus != 0 || out.ConsumptionPrice != MoneyFromInt(50) || out.ConsumptionBonus != MoneyFromInt(1) {
		t.Errorf("rule = %+v", out)
	}
	if out.ExpireDay != 0 || out.ThresholdOfTotalPrice != 0 || out.UsageLimitType != BonusUsageLimitAmount {
		t.Errorf("rule = %+v", out)
	}
	if out.UsageLimitValue != MoneyFromInt(10) || out.ConversionBonus != 1 {
		t.Errorf("rule = %+v", out)
	}
}

func TestBonusRuleGet(t *testing.T) {
	c := goldenServer(t)
	rule, resp, err := c.BonusRule.Get(testCtx)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Request.URL.Path != "/v1/bonus_rule" || resp.Request.Method != "GET" {
		t.Errorf("request = %s %s", resp.Request.Method, resp.Request.URL.Path)
	}
	if rule.ConsumptionPrice != MoneyFromInt(50) {
		t.Errorf("rule = %+v", rule)
	}
}
