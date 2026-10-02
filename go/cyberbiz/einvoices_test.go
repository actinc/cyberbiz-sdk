package cyberbiz

import (
	"errors"
	"testing"
)

const einvoicesSampleBody = `{"title":"測試公司","order_id":4977,"company_no":"12345678","invoice_no":"AB12345678",
	"invoice_status":"issue","invoice_at":"2026-07-10 20:22:05","invalid_at":null,"random_num":"1234",
	"invoice_type":"company","love_code":"","phone_barcode":"","nature_person":""}`

func TestEinvoiceDecodesSample(t *testing.T) {
	c, _ := New("tok")
	var inv Einvoice
	if err := c.decode([]byte(einvoicesSampleBody), &inv); err != nil {
		t.Fatal(err)
	}
	if inv.OrderID != 4977 || inv.InvoiceNo != "AB12345678" || inv.InvoiceStatus != InvoiceStatusIssue || inv.InvoiceType != InvoiceTypeCompany {
		t.Errorf("invoice = %+v", inv)
	}
	if inv.InvoiceAt.String() != "2026-07-10 20:22:05" || inv.InvalidAt != nil {
		t.Errorf("times = %v %v", inv.InvoiceAt, inv.InvalidAt)
	}
}

func TestEinvoiceGoldenErrors(t *testing.T) {
	c := goldenServer(t)
	for name, call := range map[string]func() error{
		"get": func() error {
			_, _, err := c.Einvoices.Get(testCtx, 1)
			return err
		},
		"get_einvoice": func() error {
			_, _, err := c.Einvoices.GetByNumber(testCtx, "AB00000000")
			return err
		},
	} {
		err := call()
		var apiErr *APIError
		if !errors.As(err, &apiErr) || !errors.Is(err, ErrNotFound) {
			t.Errorf("%s: got %v", name, err)
			continue
		}
		if len(apiErr.Messages) != 1 || apiErr.Messages[0] != "無此資源" {
			t.Errorf("%s: messages = %q", name, apiErr.Messages)
		}
	}
}

func TestEinvoiceRequests(t *testing.T) {
	c, call := discountsSpyClient(t, 200, einvoicesSampleBody)

	inv, _, err := c.Einvoices.Get(testCtx, 42)
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/einvoices/42")
	if inv.OrderID != 4977 {
		t.Errorf("invoice = %+v", inv)
	}

	if _, _, err := c.Einvoices.GetByNumber(testCtx, "AB12345678"); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/einvoices/get_einvoice")
	if call.Query.Get("invoice_no") != "AB12345678" {
		t.Errorf("query = %v", call.Query)
	}

	at, _ := ParseTime("2026-07-10 20:22:05")
	title := ""
	if _, _, err := c.Einvoices.UpdateByOrder(testCtx, 4977, &EinvoiceUpdateRequest{
		InvoiceType: InvoiceTypeClassCompany, InvoiceStatus: InvoiceStatusNotIssued, InvoiceAt: at, Title: &title,
	}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/einvoices/4977")
	discountsAssertJSONBody(t, call, `{"invoice_type":"class_company","invoice_status":"nil","invoice_at":"2026-07-10 20:22:05","title":""}`)

	c2, call2 := discountsSpyClient(t, 201, ``)
	_, err = c2.Einvoices.CreateOffline(testCtx, &OfflineEinvoiceRequest{
		CustomerID: 7,
		OfflineEinvoice: OfflineEinvoice{InvNum: "AB12345678", InvoiceTime: "12:00:00", InvStatus: "已確認", SellerName: "店家",
			InvPeriod: "11508", InvDate: "20260710", SellerAddress: "台北市", SellerBan: "12345678",
			Details: []OfflineEinvoiceDetail{{UnitPrice: "100", Amount: "100", Quantity: "1", RowNum: "1", Description: "茶", SKU: "T1"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call2, "POST", "/v1/offline_einvoices")
	discountsAssertJSONBody(t, call2, `{"customer_id":7,"offline_einvoice":{"invNum":"AB12345678","invoiceTime":"12:00:00","invStatus":"已確認",
		"sellerName":"店家","invPeriod":"11508","invDate":"20260710","sellerAddress":"台北市","sellerBan":"12345678",
		"details":[{"unitPrice":"100","amount":"100","quantity":"1","rowNum":"1","description":"茶","sku":"T1"}]}}`)
}
