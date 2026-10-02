package gendocs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const miniWebhookMD = `# doc

## Webhook 事件列表

| 事件代碼 | 說明 |
|---------|------|
| ` + "`orders/paid`" + ` | 訂單付款 |

## HTTP Headers

| Header 名稱 | 範例值 | 說明 |
|------------|--------|------|
| ` + "`User-Agent`" + ` | ` + "`CyberbizWebhook/1.0`" + ` | 固定值 |

## Payload 資料結構

### 3. 訂單 (Order)

| 屬性 | 類型 | 說明 |
|-----|------|------|
| ` + "`id`" + ` | ` + "`Integer`" + ` | 訂單 ID |
| ` + "`order_number`" + ` | ` + "`String`" + ` | 訂單編號 |
| ` + "`buyer`" + ` | ` + "`Object`" + ` | 購買會員資訊 |
| ` + "`line_items`" + ` | ` + "`[Object]`" + ` | 訂單商品資訊陣列 |
| ` + "`statuses`" + ` | ` + "`Object`" + ` | 狀態資訊 |

#### 購買會員資訊 (Buyer)

| 屬性 | 類型 | 說明 |
|-----|------|------|
| ` + "`email`" + ` | ` + "`String`" + ` | 購買會員 Email |
| ` + "`mobile`" + ` | ` + "`String`" + ` | 購買會員手機 |

#### 訂單商品資訊 (Line Item)

| 屬性 | 類型 | 說明 |
|-----|------|------|
| ` + "`id`" + ` | ` + "`Integer`" + ` | ID |
| ` + "`price`" + ` | ` + "`Float`" + ` | 金額 |

#### 狀態資訊 (Statuses)

| 屬性 | 類型 | 說明 |
|-----|------|------|
| ` + "`financial_status`" + ` | ` + "`String`" + ` | 付款狀態 |
`

func TestLoadWebhookSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wh.md")
	if err := os.WriteFile(path, []byte(miniWebhookMD), 0o644); err != nil {
		t.Fatal(err)
	}
	src, err := loadWebhookSource(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(src.Events) != 1 || src.Events[0].Code != "orders/paid" {
		t.Errorf("events: %+v", src.Events)
	}
	if got := src.Order; strings.Join(got, ",") != "Order,Buyer,Line Item,Statuses" {
		t.Errorf("sections: %v", got)
	}
	if f := src.fields("Order"); len(f) != 5 || f[2].Type != "Object" || f[3].Type != "[Object]" || f[0].Desc != "訂單 ID" {
		t.Errorf("order fields: %+v", f)
	}
	if _, ok := src.Sections["HTTP Headers"]; ok {
		t.Error("the header table is not a payload section")
	}
}

func TestWebhookPayloadAndSignature(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wh.md")
	if err := os.WriteFile(path, []byte(miniWebhookMD), 0o644); err != nil {
		t.Fatal(err)
	}
	src, _ := loadWebhookSource(path)
	golden := &goldenSet{}
	if body, err := DecodeJSON([]byte(`{"id": 99, "order_number": 1101, "buyer": {"email": "real@gmail.com", "mobile": "0955123123"}, "line_items": [{"id": 5, "price": 12.0, "extra": "kept"}], "statuses": {"financial_status": "paid"}}`)); err == nil {
		golden.files = append(golden.files, &goldenFile{Group: "v1", Method: "GET", Name: "v1_orders_{id}", Body: body, Status: 200})
	}
	b := newWebhookBuilder(src, golden, localeEN, newReport())
	payload := b.payload(webhookEvent{Code: "orders/paid", Payload: "Order"})
	if payload.Keys()[0] != "id" || !payload.Has("statuses") {
		t.Errorf("payload must follow the table order: %v", payload.Keys())
	}
	buyer, _ := payload.Get("buyer")
	if email, _ := buyer.(*OMap).Get("email"); email != synthEmail {
		t.Errorf("buyer email must be synthetic, got %v", email)
	}
	items, _ := payload.Get("line_items")
	if extra, _ := items.([]any)[0].(*OMap).Get("extra"); extra != "kept" {
		t.Error("recorded fields beyond the table must be kept for shape fidelity")
	}
	if on, _ := payload.Get("order_number"); on.(interface{ String() string }).String() != "1001" {
		t.Errorf("order_number: %v", on)
	}

	body := []byte(`{"id": 1}`)
	hexSig, b64Sig, domainSig := signatures(body)
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	if hexSig != hex.EncodeToString(mac.Sum(nil)) || len(hexSig) != 64 {
		t.Error("hex signature mismatch")
	}
	if !strings.HasSuffix(b64Sig, "=") && len(b64Sig) != 44 {
		t.Error("base64 signature must be 44 chars")
	}
	if len(domainSig) != 64 {
		t.Error("domain signature must be hex")
	}

	md, err := buildWebhooksMarkdown(src, golden, localeEN, newReport())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(md, "\n")
	if lines[0] != "# CYBERBIZ Webhooks" || lines[2] != "Version: "+DocVersion {
		t.Errorf("header/version lines wrong: %q %q", lines[0], lines[2])
	}
	if idx := strings.LastIndex(md, "\n## "); !strings.HasPrefix(md[idx:], "\n## Release Notes") {
		t.Error("release notes must be the last section")
	}
	for _, ev := range webhookEvents {
		if !strings.Contains(md, "X-Cyberbiz-Event: "+ev.Code+"\n") {
			t.Errorf("no header sample for %s", ev.Code)
		}
	}
	if err := scanOutput("webhooks", []byte(md)); err != nil {
		t.Fatal(err)
	}
	col, err := buildWebhookPostman(src, golden, localeEN, newReport())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := EncodeJSONIndent(col, "  ")
	if err != nil {
		t.Fatal(err)
	}
	checkWebhookCollection(t, []byte(md), raw)
	if !strings.Contains(string(raw), "CryptoJS.HmacSHA256(raw, secret)") || !strings.Contains(string(raw), "\"listen\": \"test\"") {
		t.Error("collection must carry the pre-request and test scripts")
	}
}
