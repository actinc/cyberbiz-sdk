package cyberbiz

import (
	"errors"
	"testing"
)

const posShopCouponJSON = `{"title":"門市九折","code":"POS10","coupon_type_name":"百分比","coupon_value":"10%","order_price_threshold":500.0,"start_date":"2026-01-01","end_date":null,"concurrently_apply":true,"usage_limit":100,"can_accumulate_bonus":false,"usage_unlimited":false,"used_times":3,"gift_order_id":null,"gift_days":null,"account_usage_limit_enabled":false,"account_usage_limit":0,"restrict_strategy":"unrestricted","restrict_campaigns":[],"tags":[],"pos_shop_ids":[1446,1554],"coupon_status":"has_expire_date","gift_order_status":"","valid":true,"customer_used_times":0,"customer_usable":true}`

const posWalletTransactionsJSON = `[
 {"type":"refund","amount":-157371.0,"balance_after":0.0,"created_at":"2026-02-26 12:04+0800","payment_name":"現金",
  "invalid_einvoices":[{"title":"","company_no":"","invoice_no":"KW00000001","invoice_status":"issue_invalid","invoice_at":"2026-02-26 12:01+0800","invalid_at":"2026-02-26 12:04+0800","random_num":"8988","invoice_type":"default","love_code":"","phone_barcode":"","nature_person":""}],
  "allowance_einvoices":[]},
 {"type":"topup","amount":996.0,"balance_after":57314.0,"created_at":"2026-02-26 11:56+0800","payment_name":"現金","financial_status":"paid",
  "einvoice":{"title":"","company_no":"","invoice_no":null,"invoice_status":null,"invoice_at":null,"invalid_at":null,"random_num":null,"invoice_type":"paper_invoice","love_code":"","phone_barcode":"","nature_person":""},
  "paper_invoice_no":"JJ00000001"},
 {"type":"consumption","amount":-1841.0,"balance_after":39985.0,"created_at":"2026-01-12 15:42+0800","payment_name":null}
]`

func TestPosShopsGolden(t *testing.T) {
	var shops []PosShop
	decodeGolden(t, "v1/GET_v1_pos_shops.json", &shops)
	if len(shops) != 2 || shops[0].ID != 1446 || shops[1].ID != 1554 {
		t.Fatalf("shops = %+v", shops)
	}
	s := shops[0]
	if s.Phone != "0912345678" || s.County != "台北市" || s.VATNumber != "REDACTED" || s.Address != "" {
		t.Errorf("shop = %+v", s)
	}
	if s.Deadline.String() != "2026-10-01" || !s.CanFindOthersOrder || !s.CanAccessCustomers {
		t.Errorf("deadline = %v, flags = %v %v", s.Deadline, s.CanFindOthersOrder, s.CanAccessCustomers)
	}
	var one PosShop
	decodeGolden(t, "v1/GET_v1_pos_shops_{id}.json", &one)
	if one.ID != 1446 || one.District != "松山區" {
		t.Errorf("detail = %+v", one)
	}
}

func TestPosGolden(t *testing.T) {
	var poses []Pos
	decodeGolden(t, "v1/GET_v1_pos_shops_{id}_poses.json", &poses)
	if len(poses) != 1 || poses[0].ID != 1627 || poses[0].PettyCash != 0 || poses[0].AdminPasswordEnabled {
		t.Errorf("poses = %+v", poses)
	}
	var one Pos
	decodeGolden(t, "v1/GET_v1_pos_shops_{id}_poses_{id}.json", &one)
	if one.ID != 1627 || one.Name != "REDACTED" {
		t.Errorf("pos = %+v", one)
	}
}

func TestPosProductsGolden(t *testing.T) {
	var products []PosProduct
	decodeGolden(t, "v1/GET_v1_pos_shops_{id}_products.json", &products)
	if len(products) != 2 {
		t.Fatalf("len = %d", len(products))
	}
	p := products[0]
	if p.ID != 60640108 || p.Title != "女休閒短T" || p.Published || p.Price != MoneyFromInt(100) || p.TaxTypeID != "inclusive_tax" {
		t.Errorf("product = %+v", p)
	}
	if p.PosShop == nil || p.PosShop.ID != 1446 || len(p.TemperatureTypes) != 1 || p.TemperatureTypes[0] != "常溫" {
		t.Errorf("pos_shop = %+v, temperature_types = %v", p.PosShop, p.TemperatureTypes)
	}
	if !p.SellFrom.IsZero() || p.CreatedAt.String() != "2025-10-15 09:49:41" || p.GoogleProductCategoryID != 0 {
		t.Errorf("nullable fields = %+v", p)
	}
	if len(p.ProductVariants) != 1 || p.ProductVariants[0].ID != 73485627 || p.ProductVariants[0].Cost != MoneyFromInt(50) {
		t.Errorf("variants = %+v", p.ProductVariants)
	}

	var variants []PosProductVariant
	decodeGolden(t, "v1/GET_v1_pos_shops_{id}_product_variants.json", &variants)
	if len(variants) != 2 {
		t.Fatalf("len = %d", len(variants))
	}
	v := variants[1]
	if v.ID != 77163605 || v.ProductID != 63562770 || v.Price != MoneyFromInt(600) || v.Cost != 0 || v.CompareAtPrice != 0 {
		t.Errorf("variant = %+v", v)
	}
	if v.InventoryQuantity != 36 || v.Sold != 1 || v.SafetyInventoryQuantity != 0 || v.InventoryPolicy != "continue" || !v.RequiresShipping {
		t.Errorf("inventory = %+v", v)
	}
	if v.SKU != "SKU-b17d99" || v.UpdatedAt.String() != "2026-01-07 12:29:20" || !v.HoneycombSync {
		t.Errorf("fields = %+v", v)
	}
	if variants[0].SafetyInventoryQuantity != 10 || variants[0].CompareAtPrice != MoneyFromInt(120) {
		t.Errorf("first variant = %+v", variants[0])
	}
}

func TestPosShopCouponPluginErrorGolden(t *testing.T) {
	for _, tc := range []struct {
		golden string
		call   func(c *Client) error
	}{
		{"errors/GET_v1_pos_shop_coupons.json", func(c *Client) error { _, err := c.PosShops.ListCoupons(testCtx, nil); return err }},
		{"errors/GET_v1_pos_shop_coupons_{id}.json", func(c *Client) error { _, _, err := c.PosShops.GetCoupon(testCtx, 1); return err }},
	} {
		c, _ := stockRecordingClient(t, 403, string(readGolden(t, tc.golden)))
		err := tc.call(c)
		var apiErr *APIError
		if !errors.Is(err, ErrForbidden) || !errors.As(err, &apiErr) || apiErr.Messages[0] != "Must have pos_shop_coupon plugin" {
			t.Errorf("%s: got %v", tc.golden, err)
		}
	}
}

func TestPosShopsGoldenRoundTrip(t *testing.T) {
	c := goldenServer(t)
	if p, err := c.PosShops.List(testCtx, nil); err != nil || len(p.Items) != 2 {
		t.Errorf("List: %v", err)
	}
	if s, _, err := c.PosShops.Get(testCtx, 1446); err != nil || s.ID != 1446 {
		t.Errorf("Get: %v", err)
	}
	if p, err := c.PosShops.ListPoses(testCtx, 1446, nil); err != nil || p.Items[0].ID != 1627 {
		t.Errorf("ListPoses: %v", err)
	}
	if pos, _, err := c.PosShops.GetPos(testCtx, 1446, 1627); err != nil || pos.ID != 1627 {
		t.Errorf("GetPos: %v", err)
	}
	if p, err := c.PosShops.ListProducts(testCtx, 1446, nil); err != nil || len(p.Items) != 2 {
		t.Errorf("ListProducts: %v", err)
	}
	if p, err := c.PosShops.ListProductVariants(testCtx, 1446, nil); err != nil || len(p.Items) != 2 {
		t.Errorf("ListProductVariants: %v", err)
	}
	n := 0
	for _, err := range c.PosShops.AllProductVariants(testCtx, 1446, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 2 {
		t.Errorf("AllProductVariants yielded %d", n)
	}
}

func TestPosShopCouponDecodes(t *testing.T) {
	c, _ := stockRecordingClient(t, 200, posShopCouponJSON)
	coupon, _, err := c.PosShops.GetCoupon(testCtx, 88)
	if err != nil {
		t.Fatal(err)
	}
	if coupon.ID != 88 || coupon.Code != "POS10" || coupon.CouponValue != "10%" || coupon.OrderPriceThreshold != MoneyFromInt(500) {
		t.Errorf("coupon = %+v", coupon)
	}
	if coupon.StartDate.String() != "2026-01-01" || !coupon.EndDate.IsZero() || !coupon.ConcurrentlyApply || coupon.GiftOrderID != 0 {
		t.Errorf("dates/flags = %+v", coupon)
	}
	if len(coupon.PosShopIDs) != 2 || coupon.PosShopIDs[1] != 1554 || coupon.CouponStatus != CouponStatusHasExpireDate {
		t.Errorf("shops = %v, status = %q", coupon.PosShopIDs, coupon.CouponStatus)
	}
	if !CouponUsable(coupon.CouponStatus, coupon.GiftOrderStatus) {
		t.Error("coupon should be usable")
	}
}

func TestPosWalletTransactionsDecode(t *testing.T) {
	c, _ := stockRecordingClient(t, 200, posWalletTransactionsJSON)
	txs, _, err := c.PosShops.WalletTransactions(testCtx, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 3 {
		t.Fatalf("len = %d", len(txs))
	}
	refund := txs[0]
	if refund.Type != PosWalletTransactionRefund || refund.Amount != MoneyFromInt(-157371) || refund.BalanceAfter != 0 {
		t.Errorf("refund = %+v", refund)
	}
	if refund.CreatedAt.String() != "2026-02-26 12:04:00" || refund.PaymentName != "現金" || refund.Einvoice != nil {
		t.Errorf("refund fields = %+v", refund)
	}
	if len(refund.InvalidEinvoices) != 1 || refund.InvalidEinvoices[0].InvoiceStatus != InvoiceStatusIssueInvalid ||
		refund.InvalidEinvoices[0].InvalidAt == nil || refund.InvalidEinvoices[0].InvalidAt.String() != "2026-02-26 12:04:00" {
		t.Errorf("invalid_einvoices = %+v", refund.InvalidEinvoices)
	}
	topup := txs[1]
	if topup.Type != PosWalletTransactionTopup || topup.FinancialStatus != PosWalletFinancialStatusPaid || topup.PaperInvoiceNo != "JJ00000001" {
		t.Errorf("topup = %+v", topup)
	}
	if topup.Einvoice == nil || topup.Einvoice.InvoiceType != InvoiceTypePaperInvoice || topup.Einvoice.InvoiceStatus != "" ||
		topup.Einvoice.InvalidAt != nil || !topup.Einvoice.InvoiceAt.IsZero() {
		t.Errorf("topup einvoice = %+v", topup.Einvoice)
	}
	if txs[2].Type != PosWalletTransactionConsumption || txs[2].PaymentName != "" || txs[2].Amount != MoneyFromInt(-1841) {
		t.Errorf("consumption = %+v", txs[2])
	}
}

func TestPosShopsRequestShapes(t *testing.T) {
	stockRunShapes(t, []stockShapeCase{
		{name: "List", method: "GET", path: "/v1/pos_shops", reply: "[]", query: "offset=5&page=2",
			call: func(c *Client) error {
				_, err := c.PosShops.List(testCtx, &ListOptions{Page: 2, Offset: 5})
				return err
			}},
		{name: "Get", method: "GET", path: "/v1/pos_shops/1446",
			call: func(c *Client) error { _, _, err := c.PosShops.Get(testCtx, 1446); return err }},
		{name: "Update", method: "PUT", path: "/v1/pos_shops/1446",
			body:   map[string]any{"name": "新店名", "VAT_number": "", "can_access_customers": false},
			absent: []string{"phone", "aes_key", "can_find_others_order"},
			call: func(c *Client) error {
				_, _, err := c.PosShops.Update(testCtx, 1446, &PosShopUpdateRequest{
					Name: stockPtr("新店名"), VATNumber: stockPtr(""), CanAccessCustomers: stockPtr(false)})
				return err
			}},
		{name: "ListPoses", method: "GET", path: "/v1/pos_shops/1446/poses", reply: "[]",
			call: func(c *Client) error { _, err := c.PosShops.ListPoses(testCtx, 1446, nil); return err }},
		{name: "GetPos", method: "GET", path: "/v1/pos_shops/1446/poses/1627",
			call: func(c *Client) error { _, _, err := c.PosShops.GetPos(testCtx, 1446, 1627); return err }},
		{name: "UpdatePos", method: "PUT", path: "/v1/pos_shops/1446/poses/1627",
			body: map[string]any{"petty_cash": 0.0, "admin_password_enabled": true}, absent: []string{"name"},
			call: func(c *Client) error {
				_, _, err := c.PosShops.UpdatePos(testCtx, 1446, 1627, &PosUpdateRequest{
					PettyCash: stockPtr(Money(0)), AdminPasswordEnabled: stockPtr(true)})
				return err
			}},
		{name: "ListProducts", method: "GET", path: "/v1/pos_shops/1446/products", reply: "[]", query: "per_page=20",
			call: func(c *Client) error {
				_, err := c.PosShops.ListProducts(testCtx, 1446, &ListOptions{PerPage: 20})
				return err
			}},
		{name: "ListProductVariants", method: "GET", path: "/v1/pos_shops/1446/product_variants", reply: "[]",
			call: func(c *Client) error { _, err := c.PosShops.ListProductVariants(testCtx, 1446, nil); return err }},
		{name: "ListCoupons", method: "GET", path: "/v1/pos_shop_coupons", reply: "[]",
			call: func(c *Client) error { _, err := c.PosShops.ListCoupons(testCtx, nil); return err }},
		{name: "GetCoupon", method: "GET", path: "/v1/pos_shop_coupons/88", reply: posShopCouponJSON,
			call: func(c *Client) error { _, _, err := c.PosShops.GetCoupon(testCtx, 88); return err }},
		{name: "CreateCoupon", method: "POST", path: "/v1/pos_shop_coupons", reply: posShopCouponJSON,
			body: map[string]any{"title": "門市九折", "code": "POS10", "coupon_type": "percent", "value": 10.0,
				"order_price_threshold": 500.0, "start_date": "2026-01-01", "concurrently_apply": false,
				"usage_limit": 100.0, "usage_unlimited": false, "pos_shop_ids": []any{1446.0, 1554.0}},
			absent: []string{"end_date", "can_accumulate_bonus"},
			call: func(c *Client) error {
				_, _, err := c.PosShops.CreateCoupon(testCtx, &PosShopCouponCreateRequest{
					Title: "門市九折", Code: "POS10", CouponType: PosShopCouponTypePercent, Value: MoneyFromInt(10),
					OrderPriceThreshold: MoneyFromInt(500), StartDate: NewDate(2026, 1, 1), UsageLimit: 100,
					UsageUnlimited: stockPtr(false), PosShopIDs: []int64{1446, 1554}})
				return err
			}},
		{name: "UpdateCoupon", method: "PUT", path: "/v1/pos_shop_coupons/88", reply: posShopCouponJSON,
			body:   map[string]any{"value": 0.0, "end_date": "2026-12-31"},
			absent: []string{"title", "code", "coupon_type", "start_date", "pos_shop_ids"},
			call: func(c *Client) error {
				cp, _, err := c.PosShops.UpdateCoupon(testCtx, 88, &PosShopCouponUpdateRequest{
					Value: stockPtr(Money(0)), EndDate: NewDate(2026, 12, 31)})
				if err == nil && cp.ID != 88 {
					t.Errorf("id = %d", cp.ID)
				}
				return err
			}},
		{name: "DeleteCoupon", method: "DELETE", path: "/v1/pos_shop_coupons/88", reply: posShopCouponJSON,
			call: func(c *Client) error { _, err := c.PosShops.DeleteCoupon(testCtx, 88); return err }},
		{name: "WalletBalance", method: "GET", path: "/v2/pos_wallets/5/balance", reply: `{"balance":157371.0,"currency":"TWD"}`,
			call: func(c *Client) error {
				b, _, err := c.PosShops.WalletBalance(testCtx, 5)
				if err == nil && (b.Balance != MoneyFromInt(157371) || b.Currency != "TWD") {
					t.Errorf("balance = %+v", b)
				}
				return err
			}},
		{name: "WalletTransactions", method: "GET", path: "/v2/pos_wallets/5/transactions", reply: "[]",
			call: func(c *Client) error { _, _, err := c.PosShops.WalletTransactions(testCtx, 5); return err }},
	})
}
