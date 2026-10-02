package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

const (
	testSecret = "test-app-secret"
	testShop   = "example.cyberbiz.co"
	testCustom = "www.example-shop.com"
)

// readSample loads a redacted sample body from testdata.
func readSample(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read sample %s: %v", name, err)
	}
	return b
}

// newRequest builds a signed Inbound the way CYBERBIZ sends it.
func newRequest(t *testing.T, event EventType, body []byte, secret string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/webhooks/cyberbiz", bytes.NewReader(body))
	r.Header.Set("User-Agent", UserAgent)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set(HeaderDomain, testShop)
	r.Header.Set(HeaderShopDomain, testCustom)
	r.Header.Set(HeaderEvent, string(event))
	r.Header.Set(HeaderSignature, Sign(body, secret))
	r.Header.Set(HeaderDomainHMAC, SignDomain(testShop, secret))
	return r
}

// sampleEvent parses a sample body as an authenticated Event.
func sampleEvent(t *testing.T, event EventType, file string) *Event {
	t.Helper()
	body := readSample(t, file)
	e, err := Parse(context.Background(), newRequest(t, event, body, testSecret), StaticSecret(testSecret))
	if err != nil {
		t.Fatalf("Parse(%s): %v", file, err)
	}
	return e
}

func TestSignAndVerify(t *testing.T) {
	body := []byte(`{"id":1}`)
	raw := hmac.New(sha256.New, []byte(testSecret))
	raw.Write(body)
	sum := raw.Sum(nil)

	hexSig := Sign(body, testSecret)
	if len(hexSig) != 64 || strings.ToLower(hexSig) != hexSig {
		t.Fatalf("Sign must return 64 lowercase hex chars, got %q", hexSig)
	}
	b64Sig := base64.StdEncoding.EncodeToString(sum)

	cases := []struct {
		name      string
		signature string
		secret    string
		want      bool
	}{
		{"hex lowercase", hexSig, testSecret, true},
		{"hex uppercase", strings.ToUpper(hexSig), testSecret, true},
		{"hex with whitespace", " " + hexSig + "\n", testSecret, true},
		{"base64", b64Sig, testSecret, true},
		{"wrong secret hex", hexSig, "other", false},
		{"wrong secret base64", b64Sig, "other", false},
		{"tampered hex", "0" + hexSig[1:], testSecret, hexSig[0] == '0'},
		{"empty signature", "", testSecret, false},
		{"empty secret", hexSig, "", false},
		{"garbage", "not-a-signature", testSecret, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Verify(body, tc.signature, tc.secret); got != tc.want {
				t.Fatalf("Verify = %v, want %v", got, tc.want)
			}
		})
	}
	if Verify([]byte(`{"id":2}`), hexSig, testSecret) {
		t.Fatal("Verify accepted a signature for a different body")
	}
}

func TestSignDomainAndVerifyDomain(t *testing.T) {
	sig := SignDomain(testShop, testSecret)
	if !VerifyDomain(testShop, sig, testSecret) {
		t.Fatal("VerifyDomain rejected its own signature")
	}
	if VerifyDomain("other.cyberbiz.co", sig, testSecret) {
		t.Fatal("VerifyDomain accepted a signature for another domain")
	}
	if VerifyDomain(testShop, sig, "other") {
		t.Fatal("VerifyDomain accepted a signature under another secret")
	}
}

func TestParseHappyPath(t *testing.T) {
	body := readSample(t, "orders_paid.json")
	r := newRequest(t, EventOrdersPaid, body, testSecret)

	var seenShop string
	resolver := SecretResolverFunc(func(_ context.Context, shop string) (string, error) {
		seenShop = shop
		return testSecret, nil
	})
	e, err := Parse(context.Background(), r, resolver)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if seenShop != testShop {
		t.Errorf("resolver saw shop %q, want %q", seenShop, testShop)
	}
	if e.Type != EventOrdersPaid {
		t.Errorf("Type = %q", e.Type)
	}
	if e.Resource() != "orders" || e.Action() != "paid" {
		t.Errorf("Resource/Action = %q/%q", e.Resource(), e.Action())
	}
	if e.ShopDomain != testShop || e.CustomDomain != testCustom {
		t.Errorf("domains = %q/%q", e.ShopDomain, e.CustomDomain)
	}
	if e.Signature != Sign(body, testSecret) || e.DomainSignature != SignDomain(testShop, testSecret) {
		t.Error("signatures not recorded on the event")
	}
	if e.ReceivedAt.IsZero() {
		t.Error("ReceivedAt is zero")
	}
	if e.Header.Get("User-Agent") != UserAgent {
		t.Error("Header not copied")
	}
	if !bytes.Equal(e.Raw, body) {
		t.Error("Raw differs from the request body")
	}

	// The body is readable again after Parse.
	again, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("re-read body: %v", err)
	}
	if !bytes.Equal(again, body) {
		t.Error("r.Body was not restored")
	}
}

func TestParseMissingHeaders(t *testing.T) {
	for _, name := range []string{HeaderEvent, HeaderDomain, HeaderSignature} {
		t.Run(name, func(t *testing.T) {
			r := newRequest(t, EventOrdersPaid, []byte(`{}`), testSecret)
			r.Header.Del(name)
			_, err := Parse(context.Background(), r, StaticSecret(testSecret))
			if !errors.Is(err, ErrMissingHeader) {
				t.Fatalf("err = %v, want ErrMissingHeader", err)
			}
			if !strings.Contains(err.Error(), name) {
				t.Errorf("error %q does not name the header", err)
			}
		})
	}
	// The domain HMAC header is optional.
	r := newRequest(t, EventOrdersPaid, []byte(`{}`), testSecret)
	r.Header.Del(HeaderDomainHMAC)
	e, err := Parse(context.Background(), r, StaticSecret(testSecret))
	if err != nil {
		t.Fatalf("Parse without domain HMAC: %v", err)
	}
	if e.DomainSignature != "" {
		t.Errorf("DomainSignature = %q, want empty", e.DomainSignature)
	}
}

func TestParseUnknownShop(t *testing.T) {
	r := newRequest(t, EventOrdersPaid, []byte(`{}`), testSecret)
	boom := errors.New("no such shop")
	_, err := Parse(context.Background(), r, SecretResolverFunc(func(context.Context, string) (string, error) {
		return "", boom
	}))
	if !errors.Is(err, ErrUnknownShop) || !errors.Is(err, boom) {
		t.Fatalf("err = %v, want ErrUnknownShop wrapping the resolver error", err)
	}

	r = newRequest(t, EventOrdersPaid, []byte(`{}`), testSecret)
	_, err = Parse(context.Background(), r, StaticSecret(""))
	if !errors.Is(err, ErrUnknownShop) {
		t.Fatalf("empty secret: err = %v, want ErrUnknownShop", err)
	}
}

func TestParseInvalidSignature(t *testing.T) {
	r := newRequest(t, EventOrdersPaid, []byte(`{}`), "wrong-secret")
	_, err := Parse(context.Background(), r, StaticSecret(testSecret))
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("err = %v, want ErrInvalidSignature", err)
	}

	// Tampered body under a valid header.
	r = newRequest(t, EventOrdersPaid, []byte(`{"a":1}`), testSecret)
	r.Body = io.NopCloser(strings.NewReader(`{"a":2}`))
	_, err = Parse(context.Background(), r, StaticSecret(testSecret))
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("tampered: err = %v, want ErrInvalidSignature", err)
	}
}

func TestParseBase64Signature(t *testing.T) {
	body := []byte(`{"id":7}`)
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(body)
	r := newRequest(t, EventOrdersPaid, body, testSecret)
	r.Header.Set(HeaderSignature, base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	if _, err := Parse(context.Background(), r, StaticSecret(testSecret)); err != nil {
		t.Fatalf("Parse with base64 signature: %v", err)
	}
}

func TestParseInvalidDomainSignature(t *testing.T) {
	r := newRequest(t, EventOrdersPaid, []byte(`{}`), testSecret)
	r.Header.Set(HeaderDomainHMAC, SignDomain("other.cyberbiz.co", testSecret))
	_, err := Parse(context.Background(), r, StaticSecret(testSecret))
	if !errors.Is(err, ErrInvalidDomainSignature) {
		t.Fatalf("err = %v, want ErrInvalidDomainSignature", err)
	}
	if errors.Is(err, ErrInvalidSignature) {
		t.Error("domain mismatch must not be reported as ErrInvalidSignature")
	}

	// WithoutDomainCheck ignores the mismatch.
	r = newRequest(t, EventOrdersPaid, []byte(`{}`), testSecret)
	r.Header.Set(HeaderDomainHMAC, SignDomain("other.cyberbiz.co", testSecret))
	if _, err := Parse(context.Background(), r, StaticSecret(testSecret), WithoutDomainCheck()); err != nil {
		t.Fatalf("WithoutDomainCheck: %v", err)
	}
}

func TestParseBodyTooLarge(t *testing.T) {
	body := bytes.Repeat([]byte("x"), 100)
	r := newRequest(t, EventOrdersPaid, body, testSecret)
	_, err := Parse(context.Background(), r, StaticSecret(testSecret), WithMaxBodyBytes(99))
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("err = %v, want ErrBodyTooLarge", err)
	}
	// Even then the full body can still be read downstream.
	again, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	if !bytes.Equal(again, body) {
		t.Errorf("restored body has %d bytes, want %d", len(again), len(body))
	}

	// Exactly at the limit is fine.
	r = newRequest(t, EventOrdersPaid, body, testSecret)
	if _, err := Parse(context.Background(), r, StaticSecret(testSecret), WithMaxBodyBytes(100)); err != nil {
		t.Fatalf("at limit: %v", err)
	}
}

func TestParseNilResolver(t *testing.T) {
	r := newRequest(t, EventOrdersPaid, []byte(`{}`), testSecret)
	if _, err := Parse(context.Background(), r, nil); err == nil {
		t.Fatal("Parse accepted a nil resolver")
	}
}

func TestParseEmptyBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set(HeaderDomain, testShop)
	r.Header.Set(HeaderEvent, string(EventAppsUninstall))
	r.Header.Set(HeaderSignature, Sign(nil, testSecret))
	e, err := Parse(context.Background(), r, StaticSecret(testSecret))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(e.Raw) != 0 {
		t.Errorf("Raw = %q, want empty", e.Raw)
	}
}

func TestEventTypeHelpers(t *testing.T) {
	// CYBERBIZ documents 35 events for App webhooks.
	if len(AllEvents) != 35 {
		t.Fatalf("AllEvents has %d entries, want 35", len(AllEvents))
	}
	seen := map[EventType]bool{}
	for _, e := range AllEvents {
		if seen[e] {
			t.Errorf("duplicate event %q", e)
		}
		seen[e] = true
		if !e.Known() {
			t.Errorf("%q not Known", e)
		}
		if e.Resource() == "" || e.Action() == "" {
			t.Errorf("%q has empty resource or action", e)
		}
	}
	if EventType("orders/teleported").Known() {
		t.Error("unknown event reported as Known")
	}
	if EventType("noslash").Resource() != "noslash" || EventType("noslash").Action() != "" {
		t.Error("slashless type not split as expected")
	}
}

func TestEventDecode(t *testing.T) {
	e := &Event{Type: EventAppsUninstall, Raw: []byte(`{"app_uuid":"u","app_uuid":"dup","extra":1}`)}
	var v struct {
		AppUUID string `json:"app_uuid"`
	}
	if err := e.Decode(&v); err != nil {
		t.Fatalf("Decode with duplicate names and unknown field: %v", err)
	}
	e.Raw = []byte("{\"app_uuid\":\"\xff\"}")
	if err := e.Decode(&v); err != nil {
		t.Fatalf("Decode with invalid UTF-8: %v", err)
	}
	e.Raw = []byte(`{not json`)
	if err := e.Decode(&v); err == nil {
		t.Fatal("Decode accepted malformed JSON")
	}
}

func TestWrongEvent(t *testing.T) {
	e := sampleEvent(t, EventOrdersPaid, "orders_paid.json")
	if _, err := e.Customer(); !errors.Is(err, ErrWrongEvent) {
		t.Errorf("Customer() on an orders event: %v", err)
	}
	if _, err := e.App(); !errors.Is(err, ErrWrongEvent) {
		t.Errorf("App() on an orders event: %v", err)
	}
	bad := &Event{Type: EventOrdersPaid, Raw: []byte(`{"id":"not a number"}`)}
	if _, err := bad.Order(); err == nil || errors.Is(err, ErrWrongEvent) {
		t.Errorf("Order() on a malformed body: %v", err)
	}
}

func TestOrderPayloadPaid(t *testing.T) {
	e := sampleEvent(t, EventOrdersPaid, "orders_paid.json")
	o, err := e.Order()
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	if o.ID != 500001 || o.OrderNumber != 10001 || o.OrderName != "#10001" {
		t.Errorf("identity = %d/%d/%q", o.ID, o.OrderNumber, o.OrderName)
	}
	if o.SubtotalPrice != cyberbiz.MoneyFromInt(480) {
		t.Errorf("SubtotalPrice = %v", o.SubtotalPrice)
	}
	if o.CreatedAt.String() != "2026-01-09 13:10:05" {
		t.Errorf("CreatedAt = %v", o.CreatedAt)
	}
	if o.Customer == nil || o.Customer.ID != 900001 || o.Customer.Email != "customer@example.com" {
		t.Errorf("Customer = %+v", o.Customer)
	}
	if o.Customer.Birthday.String() != "1990-01-01" {
		t.Errorf("Customer.Birthday = %v", o.Customer.Birthday)
	}
	if o.Buyer == nil || o.Buyer.Mobile != "0900000000" {
		t.Errorf("Buyer = %+v", o.Buyer)
	}
	if o.Receiver == nil || o.Receiver.CVSStoreID != "000001" || o.Receiver.DetailAddress == nil {
		t.Errorf("Receiver = %+v", o.Receiver)
	}
	if len(o.LineItems) != 2 {
		t.Fatalf("LineItems = %d", len(o.LineItems))
	}
	li := o.LineItems[0]
	if li.Price != cyberbiz.MoneyFromInt(960) || li.TotalDiscount != cyberbiz.MoneyFromInt(480) || li.Quantity != 1 {
		t.Errorf("line item 0 = %+v", li)
	}
	if len(li.Discounts) != 1 || li.Discounts[0].Code != "coupon_discount" || li.Discounts[0].Discount != cyberbiz.MoneyFromInt(480) {
		t.Errorf("line item discounts = %+v", li.Discounts)
	}
	if o.LineItems[1].ItemType != "gift" || o.LineItems[1].TotalDiscount != 0 {
		t.Errorf("gift line = %+v", o.LineItems[1])
	}
	if o.ShippingVendor == nil || o.ShippingVendor.Type != "seven" {
		t.Errorf("ShippingVendor = %+v", o.ShippingVendor)
	}
	if len(o.Fulfillments) != 1 || o.Fulfillments[0].Status != cyberbiz.FulfillmentStatusUnshipped || len(o.Fulfillments[0].LineItems) != 2 {
		t.Errorf("Fulfillments = %+v", o.Fulfillments)
	}
	if o.Prices == nil || o.Prices.TotalPrice != cyberbiz.MoneyFromInt(560) || o.Prices.ShippingRatePrice != cyberbiz.MoneyFromInt(80) {
		t.Errorf("Prices = %+v", o.Prices)
	}
	d := o.Prices.Discounts
	if d == nil || d.ShopDiscount != nil || d.CouponDiscount == nil || d.CouponDiscount.Amount != cyberbiz.MoneyFromInt(480) {
		t.Errorf("Discounts = %+v", d)
	}
	if len(d.CouponDiscounts) != 1 || d.CouponDiscounts[0].CouponID != 300001 {
		t.Errorf("CouponDiscounts = %+v", d.CouponDiscounts)
	}
	if o.EInvoice == nil || o.EInvoice.InvoiceStatus != cyberbiz.InvoiceStatusIssue || o.EInvoice.InvoiceType != cyberbiz.InvoiceTypePhoneBarcode || o.EInvoice.InvalidAt != nil {
		t.Errorf("EInvoice = %+v", o.EInvoice)
	}
	if o.Statuses == nil || o.Statuses.OrderStatus != cyberbiz.OrderStatusOpen || o.Statuses.FinancialStatus != cyberbiz.FinancialStatusPaid {
		t.Errorf("Statuses = %+v", o.Statuses)
	}
	if o.Timings == nil || o.Timings.ConfirmedAt == nil || o.Timings.RefundAt != nil || o.Timings.CancelledAt != nil {
		t.Errorf("Timings = %+v", o.Timings)
	}
	if o.BranchStore != nil || o.CustomerCancelReasonDetail != nil || o.POSInfo == nil || o.POSInfo.POSID != 0 {
		t.Errorf("nullable objects: branch=%v cancel=%v pos=%+v", o.BranchStore, o.CustomerCancelReasonDetail, o.POSInfo)
	}
	if o.LinkedOrderInfo == nil || o.LinkedOrderInfo.Source != "none" {
		t.Errorf("LinkedOrderInfo = %+v", o.LinkedOrderInfo)
	}
	if len(o.ExchangeHistories) != 0 || len(o.Tags) != 0 || len(o.SerialNumbers) != 0 {
		t.Error("expected empty arrays")
	}
}

func TestOrderPayloadRefunded(t *testing.T) {
	e := sampleEvent(t, EventOrdersRefunded, "orders_refunded.json")
	o, err := e.Order()
	if err != nil {
		t.Fatalf("Order: %v", err)
	}
	if o.Statuses.OrderStatus != cyberbiz.OrderStatusCancelled || o.Statuses.FinancialStatus != cyberbiz.FinancialStatusRefunded {
		t.Errorf("Statuses = %+v", o.Statuses)
	}
	if o.Timings.RefundAt == nil || o.Timings.RefundAt.String() != "2026-01-09 23:45:41" {
		t.Errorf("RefundAt = %v", o.Timings.RefundAt)
	}
	if o.Timings.CancelledAt == nil || o.Timings.CancelledAt.String() != "2026-01-09 23:45:39" {
		t.Errorf("CancelledAt = %v", o.Timings.CancelledAt)
	}
	if o.Timings.ClosedAt != nil {
		t.Errorf("ClosedAt = %v, want nil", o.Timings.ClosedAt)
	}
	if len(o.LineItems) != 9 {
		t.Errorf("LineItems = %d, want 9", len(o.LineItems))
	}
	if o.ReferralCode != "REF000" || o.EInvoice.InvoiceType != cyberbiz.InvoiceTypeDonate || o.EInvoice.LoveCode != "00000" {
		t.Errorf("referral/invoice = %q/%+v", o.ReferralCode, o.EInvoice)
	}
	if o.Receiver.DetailAddress.City != "Taipei City" || o.Receiver.CVSStoreID != "" {
		t.Errorf("Receiver = %+v", o.Receiver)
	}
	if o.Prices.Discounts.CouponDiscount != nil || len(o.Prices.Discounts.CouponDiscounts) != 0 {
		t.Errorf("coupon discounts should be absent: %+v", o.Prices.Discounts)
	}
}

func TestCustomerPayload(t *testing.T) {
	e := sampleEvent(t, EventCustomersCreate, "customers_create.json")
	c, err := e.Customer()
	if err != nil {
		t.Fatalf("Customer: %v", err)
	}
	if c.ID != 900001 || c.Status != "enabled" || c.Name != "Test Customer" {
		t.Errorf("identity = %+v", c)
	}
	if !c.EnableCVSPickup || !c.AcceptsMarketing || c.Address != nil {
		t.Errorf("flags/address = %+v", c)
	}
	if c.Birthday.String() != "1990-01-01" || c.CreatedAt.String() != "2026-01-11 11:51:21" {
		t.Errorf("dates = %v/%v", c.Birthday, c.CreatedAt)
	}
	if c.ConfirmedAt != nil || c.MobileSMSConfirmedAt != nil {
		t.Errorf("unconfirmed customer has confirmation times: %v/%v", c.ConfirmedAt, c.MobileSMSConfirmedAt)
	}
	if len(c.UIDProviders) != 1 || c.UIDProviders[0].ProviderType != "line" {
		t.Errorf("UIDProviders = %+v", c.UIDProviders)
	}
	if c.OtherAccumulatedConsumption != 0 || !c.OtherAccumulatedConsumptionExpiredAt.IsZero() {
		t.Errorf("other consumption = %v/%v", c.OtherAccumulatedConsumption, c.OtherAccumulatedConsumptionExpiredAt)
	}

	e = sampleEvent(t, EventCustomersUpdate, "customers_update.json")
	c, err = e.Customer()
	if err != nil {
		t.Fatalf("Customer(update): %v", err)
	}
	if c.ConfirmedAt == nil || c.ConfirmedAt.String() != "2024-06-04 13:06:42" {
		t.Errorf("ConfirmedAt = %v", c.ConfirmedAt)
	}
	if c.MobileSMSConfirmedAt == nil || c.MobileSMSConfirmedAt.String() != "2024-06-04 13:10:28" {
		t.Errorf("MobileSMSConfirmedAt = %v", c.MobileSMSConfirmedAt)
	}
	if c.Address == nil || c.Address.DetailAddress == nil || c.Address.Address != "  " {
		t.Errorf("Address = %+v", c.Address)
	}
	if c.BonusRemain != 5 || c.EnableHomeDeliveryCOD {
		t.Errorf("BonusRemain/EnableHomeDeliveryCOD = %v/%v", c.BonusRemain, c.EnableHomeDeliveryCOD)
	}
}

func TestBonusPointPayload(t *testing.T) {
	e := sampleEvent(t, EventBonusPointsCreate, "bonus_points_create.json")
	b, err := e.BonusPoint()
	if err != nil {
		t.Fatalf("BonusPoint: %v", err)
	}
	if b.ID != 100001 || b.CustomerID != 900001 || b.Points != 100 || b.UnusedPoints != 100 {
		t.Errorf("payload = %+v", b)
	}
	if b.Deadline.String() != "2027-01-09 22:00:37" || b.OrderID != 0 || b.ConsumptionPrice != 0 {
		t.Errorf("deadline/order/consumption = %v/%d/%v", b.Deadline, b.OrderID, b.ConsumptionPrice)
	}

	e = sampleEvent(t, EventBonusPointsUpdate, "bonus_points_update.json")
	b, err = e.BonusPoint()
	if err != nil {
		t.Fatalf("BonusPoint(update): %v", err)
	}
	if b.Points != 29 || b.UnusedPoints != 0 {
		t.Errorf("update = %+v", b)
	}

	// comment_bonus/* carries the same payload.
	e = sampleEvent(t, EventCommentBonusCreate, "bonus_points_create.json")
	if _, err := e.BonusPoint(); err != nil {
		t.Errorf("BonusPoint(comment_bonus): %v", err)
	}
}

func TestUIDProviderPayload(t *testing.T) {
	e := sampleEvent(t, EventUIDProvidersCreate, "uid_providers_create.json")
	u, err := e.UIDProvider()
	if err != nil {
		t.Fatalf("UIDProvider: %v", err)
	}
	if u.ProviderType != "line" || !strings.HasPrefix(u.UID, "U0") {
		t.Errorf("payload = %+v", u)
	}
}

func TestProductPayload(t *testing.T) {
	e := sampleEvent(t, EventProductsUpdate, "products_update.json")
	p, err := e.Product()
	if err != nil {
		t.Fatalf("Product: %v", err)
	}
	if p.ID != 700001 || !p.Published || p.Price != cyberbiz.MoneyFromInt(960) || p.SellWeight != 12 {
		t.Errorf("payload = %+v", p)
	}
	if p.SellFrom.String() != "2026-01-01 00:00:00" || !p.SellTo.IsZero() {
		t.Errorf("sell window = %v/%v", p.SellFrom, p.SellTo)
	}
	if len(p.ProductVariants) != 1 || p.ProductVariants[0].SKU != "SKU-1" || p.ProductVariants[0].InventoryQuantity != 25 {
		t.Errorf("variants = %+v", p.ProductVariants)
	}
	if len(p.Tags) != 1 || p.Tags[0].Name != "sample-tag" || len(p.Photos) != 1 || p.Photos[0].Position != 1 {
		t.Errorf("tags/photos = %+v/%+v", p.Tags, p.Photos)
	}
	if p.Channel == nil || p.Channel.Name != "Web" || p.SpecialCollection != nil || p.BranchStore != nil {
		t.Errorf("nested = %+v/%+v/%+v", p.Channel, p.SpecialCollection, p.BranchStore)
	}
	if len(p.CustomCollections) != 1 || p.CustomCollections[0].Handle != "sample-collection" {
		t.Errorf("collections = %+v", p.CustomCollections)
	}
	if !p.SEOMetaTags.IsValid() || p.GoogleProductCategoryID != 0 {
		t.Errorf("seo/google = %s/%d", p.SEOMetaTags, p.GoogleProductCategoryID)
	}
}

func TestVariantPayload(t *testing.T) {
	e := sampleEvent(t, EventVariantsUpdate, "variants_update.json")
	v, err := e.Variant()
	if err != nil {
		t.Fatalf("Variant: %v", err)
	}
	if v.ID != 750001 || v.ProductID != 700001 || v.CompareAtPrice != cyberbiz.MoneyFromInt(1200) || v.Cost != cyberbiz.MoneyFromInt(400) {
		t.Errorf("payload = %+v", v)
	}
	if v.Weight != 0.5 || v.MaxUsableBonus != 50 || v.Option2 != "" || !v.InventoryManagement {
		t.Errorf("scalars = %+v", v)
	}
	if len(v.PIMInfos) != 1 || v.PIMInfos[0].ChannelShopName != "Sample Channel" || !v.PIMInfos[0].IsConnected {
		t.Errorf("PIMInfos = %+v", v.PIMInfos)
	}
}

func TestCouponPayload(t *testing.T) {
	e := sampleEvent(t, EventCouponsCreate, "coupons_create.json")
	c, err := e.Coupon()
	if err != nil {
		t.Fatalf("Coupon: %v", err)
	}
	if c.CustomerID != 900001 || c.Code != "SAMPLECODE0000" || c.CouponValue != cyberbiz.MoneyFromInt(100) {
		t.Errorf("payload = %+v", c)
	}
	if c.OrderPriceThreshold != cyberbiz.MoneyFromInt(500) || c.EndDate.String() != "2026-03-31 23:59:59" {
		t.Errorf("threshold/end = %v/%v", c.OrderPriceThreshold, c.EndDate)
	}
	if c.CouponStatus != cyberbiz.CouponStatusHasExpireDate || c.GiftOrderStatus != cyberbiz.GiftOrderStatusClosed || !c.Usable() {
		t.Errorf("status = %v/%v usable=%v", c.CouponStatus, c.GiftOrderStatus, c.Usable())
	}
	if c.AccountUsageLimit != 0 || len(c.Tags) != 1 || len(c.POSShopIDs) != 0 {
		t.Errorf("limits/tags = %+v", c)
	}
}

func TestVIPLevelPayload(t *testing.T) {
	e := sampleEvent(t, EventCustomerVIPLevelUpdate, "customer_vip_level_update.json")
	v, err := e.VIPLevel()
	if err != nil {
		t.Fatalf("VIPLevel: %v", err)
	}
	if v.CustomerID != 900001 || v.CurrentGroup == nil || len(v.CurrentGroup.VIPGroupLevels) != 2 {
		t.Errorf("group = %+v", v.CurrentGroup)
	}
	if v.CurrentLevel == nil || v.CurrentLevel.Name != "Silver" || v.CurrentLevel.UpgradeConditionTotalSpent != cyberbiz.MoneyFromInt(1000) {
		t.Errorf("current = %+v", v.CurrentLevel)
	}
	if v.NextLevel == nil || v.NextLevel.Name != "Gold" || v.NextLevel.BirthGiftSetting != nil {
		t.Errorf("next = %+v", v.NextLevel)
	}
	gold := v.CurrentGroup.VIPGroupLevels[1]
	if gold.BirthGiftSetting == nil || gold.BirthGiftSetting.Coupon == nil || gold.BirthGiftSetting.Coupon.Presets == nil {
		t.Fatalf("gold gift setting = %+v", gold.BirthGiftSetting)
	}
	if p := gold.BirthGiftSetting.Coupon.Presets; p.CouponTypeID != 1 || p.Value != cyberbiz.MoneyFromInt(200) || p.Code != "GOLDBDAY" {
		t.Errorf("presets = %+v", p)
	}
	if v.ExtraInfo == nil || v.ExtraInfo.EndAt.String() != "2027-01-09 13:10:05" || v.ExtraInfo.DifferenceOfTotalSpentForUpgrade != cyberbiz.MoneyFromInt(4040) {
		t.Errorf("extra = %+v", v.ExtraInfo)
	}
}

func TestAppPayload(t *testing.T) {
	e := sampleEvent(t, EventAppsUninstall, "apps_uninstall.json")
	a, err := e.App()
	if err != nil {
		t.Fatalf("App: %v", err)
	}
	if a.AppClientID != "sample-client-id" || a.AppUUID == "" || a.AppVersionUUID == "" {
		t.Errorf("payload = %+v", a)
	}
}

func TestHandler(t *testing.T) {
	body := readSample(t, "bonus_points_create.json")
	okFn := func(context.Context, *Event) error { return nil }

	cases := []struct {
		name  string
		build func(t *testing.T) *http.Request
		fn    func(context.Context, *Event) error
		opts  []Option
		want  int
	}{
		{"ok", func(t *testing.T) *http.Request {
			return newRequest(t, EventBonusPointsCreate, body, testSecret)
		}, okFn, nil, http.StatusOK},
		{"handler error", func(t *testing.T) *http.Request {
			return newRequest(t, EventBonusPointsCreate, body, testSecret)
		}, func(context.Context, *Event) error { return errors.New("db down") }, nil, http.StatusInternalServerError},
		{"bad signature", func(t *testing.T) *http.Request {
			return newRequest(t, EventBonusPointsCreate, body, "wrong")
		}, okFn, nil, http.StatusUnauthorized},
		{"bad domain signature", func(t *testing.T) *http.Request {
			r := newRequest(t, EventBonusPointsCreate, body, testSecret)
			r.Header.Set(HeaderDomainHMAC, "deadbeef")
			return r
		}, okFn, nil, http.StatusUnauthorized},
		{"unknown shop", func(t *testing.T) *http.Request {
			r := newRequest(t, EventBonusPointsCreate, body, testSecret)
			r.Header.Set(HeaderDomain, "nobody.cyberbiz.co")
			return r
		}, okFn, nil, http.StatusUnauthorized},
		{"too large", func(t *testing.T) *http.Request {
			return newRequest(t, EventBonusPointsCreate, body, testSecret)
		}, okFn, []Option{WithMaxBodyBytes(10)}, http.StatusRequestEntityTooLarge},
		{"missing header", func(t *testing.T) *http.Request {
			r := newRequest(t, EventBonusPointsCreate, body, testSecret)
			r.Header.Del(HeaderEvent)
			return r
		}, okFn, nil, http.StatusBadRequest},
		{"wrong method", func(t *testing.T) *http.Request {
			r := newRequest(t, EventBonusPointsCreate, body, testSecret)
			r.Method = http.MethodGet
			return r
		}, okFn, nil, http.StatusMethodNotAllowed},
	}

	secrets := SecretResolverFunc(func(_ context.Context, shop string) (string, error) {
		if shop != testShop {
			return "", errors.New("unknown")
		}
		return testSecret, nil
	})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			h := Handler(secrets, func(ctx context.Context, e *Event) error {
				called = true
				if e.Type != EventBonusPointsCreate {
					t.Errorf("handler got event %q", e.Type)
				}
				return tc.fn(ctx, e)
			}, tc.opts...)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, tc.build(t))
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tc.want, rec.Body.String())
			}
			wantCalled := tc.want == http.StatusOK || tc.want == http.StatusInternalServerError
			if called != wantCalled {
				t.Errorf("handler called = %v, want %v", called, wantCalled)
			}
			if bytes.Contains(rec.Body.Bytes(), []byte(`"id"`)) {
				t.Error("response echoed the request body")
			}
		})
	}
}
