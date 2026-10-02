package gendocs

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Fixed synthetic values every Sample uses. They are obviously fake but keep
// the shape of real data.
const (
	synthName       = "王小明"
	synthEmail      = "customer@example.com"
	synthStaffEmail = "staff@example.com"
	synthMobile     = "0912345678"
	synthShopDomain = "example.cyberbiz.co"
	synthImageURL   = "//example.cyberbiz.co/media/sample-product.jpg"
	synthTimestamp  = "2026-09-01 10:00:00"
	synthDate       = "2026-09-01"
	synthToken      = "synthetic-token-do-not-use"
)

var (
	reEmail       = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	reTWMobile    = regexp.MustCompile(`\b09\d{8}\b`)
	reISOTime     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z$`)
	reShopHost    = regexp.MustCompile(`(?i)\b[a-z0-9-]+\.cyberbiz\.(?:co|io)\b`)
	reMediaPath   = regexp.MustCompile(`/media/[A-Za-z0-9_\-]+\.(jpe?g|png|gif|webp)(\?sha=[0-9a-f]+)?`)
	reConfirmTok  = regexp.MustCompile(`confirmation_token=[A-Za-z0-9]+`)
	reIDKey       = regexp.MustCompile(`(^|_)id$`)
	reIDsKey      = regexp.MustCompile(`(^|_)ids$`)
	reTimeKey     = regexp.MustCompile(`(^|_)(at|time|date|deadline|from|to)$`)
	reRedactedVal = regexp.MustCompile(`(?i)^<?redacted>?$`)
)

// synthesizer rewrites a decoded JSON value into a Sample: identifiers are
// renumbered to small integers (consistently within one document), personal
// data is replaced by fixed fake values, and hostnames point at example.com.
type synthesizer struct {
	ids      map[string]json.Number
	nextID   int64
	orderNos map[string]json.Number
	nextNo   int64
	skuSeq   int
}

func newSynthesizer() *synthesizer {
	return &synthesizer{ids: map[string]json.Number{}, nextID: 1, orderNos: map[string]json.Number{}, nextNo: 1001}
}

// Value returns a synthetic copy of v. key is the JSON key the value sits
// under ("" at the root) and parent is the enclosing object, if any.
func (s *synthesizer) Value(key string, v any, parent *OMap) any {
	switch t := v.(type) {
	case *OMap:
		out := NewOMap()
		for _, k := range t.Keys() {
			cv, _ := t.Get(k)
			out.Set(k, s.Value(k, cv, t))
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, e := range t {
			out = append(out, s.Value(key, e, parent))
		}
		return out
	case json.Number:
		return s.number(key, t)
	case string:
		return s.str(key, t, parent)
	default:
		return v
	}
}

func (s *synthesizer) number(key string, n json.Number) json.Number {
	switch {
	case key == "order_number":
		return s.orderNumber(n)
	case reIDKey.MatchString(key) || reIDsKey.MatchString(key) || key == "customer_id":
		return s.id(n)
	}
	// Random-looking floats from the Postman placeholders are rounded.
	if strings.Contains(n.String(), ".") {
		if f, err := n.Float64(); err == nil {
			if len(n.String()) > 8 {
				return json.Number(strconv.FormatFloat(math.Round(f*100)/100, 'f', -1, 64))
			}
		}
	}
	return n
}

func (s *synthesizer) id(n json.Number) json.Number {
	if v, ok := s.ids[n.String()]; ok {
		return v
	}
	if strings.Contains(n.String(), ".") {
		return n
	}
	v := json.Number(strconv.FormatInt(s.nextID, 10))
	s.nextID++
	s.ids[n.String()] = v
	return v
}

func (s *synthesizer) orderNumber(n json.Number) json.Number {
	if v, ok := s.orderNos[n.String()]; ok {
		return v
	}
	v := json.Number(strconv.FormatInt(s.nextNo, 10))
	s.nextNo++
	s.orderNos[n.String()] = v
	return v
}

// str rewrites one string value. Empty strings keep their shape.
func (s *synthesizer) str(key, v string, parent *OMap) any {
	if v == "" {
		return v
	}
	if gen, ok := s.keyed(key, v, parent); ok {
		return gen
	}
	redacted := reRedactedVal.MatchString(v) || strings.Contains(v, "REDACTED")
	if redacted {
		return "範例"
	}
	if key == "order_name" || strings.HasPrefix(v, "#") && len(v) < 8 {
		if no := s.orderNameNumber(v); no != "" {
			return "#" + no
		}
	}
	return s.scrub(key, v)
}

// orderNameNumber maps "#1101" onto the synthetic order number.
func (s *synthesizer) orderNameNumber(v string) string {
	raw := strings.TrimPrefix(v, "#")
	if _, err := strconv.Atoi(raw); err != nil {
		return ""
	}
	return s.orderNumber(json.Number(raw)).String()
}

// keyed returns a generated value for keys that always carry personal or
// shop-specific data, regardless of the recorded value.
func (s *synthesizer) keyed(key, v string, parent *OMap) (any, bool) {
	switch key {
	case "name":
		return s.nameFor(parent), true
	case "email", "pos_user_email", "shop_email":
		if key == "shop_email" {
			return "shop@example.com", true
		}
		return synthEmail, true
	case "delegate":
		if strings.Contains(v, "@") {
			return synthStaffEmail, true
		}
	case "mobile", "phone":
		return synthMobile, true
	case "address":
		return "台北市信義區市府路1號", true
	case "address1":
		return "市府路1號", true
	case "zip":
		return "110", true
	case "city", "county":
		return "台北市", true
	case "district":
		return "信義區", true
	case "company":
		return "範例有限公司", true
	case "company_no":
		return "12345678", true
	case "title", "english_title":
		return s.titleFor(key, parent), true
	case "variant_title":
		return "範例商品 - 紅色", true
	case "sku":
		s.skuSeq++
		return fmt.Sprintf("SKU-%03d", s.skuSeq), true
	case "qc":
		return "QC-001", true
	case "merchant_trade_no":
		return "S1#1001", true
	case "transaction_number", "logistics_id", "allpay_logistics_id":
		return "TXN0000001", true
	case "tracking_number":
		return "TRK000000001", true
	case "payment_url":
		return "https://example.com/pay/1001", true
	case "account_activation_url":
		return "https://example.cyberbiz.co/account/customer/activate?confirmation_token=SYNTHETIC", true
	case "token", "access_token":
		return synthToken, true
	case "uid":
		return "U0000000000000000000000000000001", true
	case "invoice_no":
		return "AB12345678", true
	case "paper_invoice_no":
		return "PA12345678", true
	case "random_num":
		return "1234", true
	case "card4no":
		return "4242", true
	case "referral_code", "checkout_referral_code", "register_referral_code", "code":
		if key == "code" && !looksLikeCoupon(v) {
			break
		}
		return "SAMPLE100", true
	case "webhook_url":
		return "https://example.com/webhooks/cyberbiz", true
	case "vendor_type":
		return "example-app", true
	case "primary_domain":
		return synthShopDomain, true
	case "store_no", "store_number":
		return "S001", true
	case "job_id":
		return "5f1a2b3c4d5e6f7a8b9c0d1e", true
	case "handle":
		if parent != nil && (parent.Has("product_variants") || parent.Has("product_url")) {
			return "sample-product", true
		}
	case "photo", "url", "url_thumb", "url_content", "og_image_url", "icon_image_url", "photo_urls":
		return synthImageURL, true
	case "product_url":
		return "//example.cyberbiz.co/products/sample-product", true
	case "note", "warehouse_note", "customer_note":
		return "範例備註", true
	case "description_url":
		return "https://example.cyberbiz.co/pages/vip", true
	}
	return nil, false
}

func looksLikeCoupon(v string) bool {
	return len(v) >= 4 && strings.ToUpper(v) == v
}

// nameFor picks a name that fits the object: a person for customer-like
// objects, otherwise a labelled placeholder.
func (s *synthesizer) nameFor(parent *OMap) string {
	switch {
	case parent == nil:
		return "範例名稱"
	case parent.Has("email") || parent.Has("mobile") || parent.Has("birthday") || parent.Has("phone") && parent.Has("address"):
		return synthName
	case parent.Has("store_no") || parent.Has("opening_hours"):
		return "範例門市"
	case parent.Has("sku") || parent.Has("option1"):
		return "範例款式"
	case parent.Has("system_default"):
		return "主選單"
	case parent.Has("vip_group_levels"):
		return "VIP 群組"
	case parent.Has("validity_days"):
		return "金卡會員"
	case parent.Has("primary_domain"):
		return "範例商店"
	case parent.Has("manifest_version") || parent.Has("webhook_events"):
		return "Example App"
	}
	return "範例名稱"
}

func (s *synthesizer) titleFor(key string, parent *OMap) string {
	if key == "english_title" {
		return "Sample Product"
	}
	if parent != nil && (parent.Has("product_variants") || parent.Has("product_id") || parent.Has("sku") || parent.Has("price")) {
		return "範例商品"
	}
	if parent != nil && (parent.Has("handle") || parent.Has("published")) {
		return "範例群組"
	}
	return "範例標題"
}

// scrub removes whatever personal or shop-specific data slipped through the
// keyed rules: e-mail addresses, Taiwanese mobile numbers, real hostnames,
// media paths, and Postman placeholder timestamps.
func (s *synthesizer) scrub(key, v string) any {
	if v == "string" {
		return s.placeholder(key)
	}
	if reISOTime.MatchString(v) {
		if strings.HasSuffix(key, "_at") {
			return synthTimestamp
		}
		return synthDate
	}
	v = reEmail.ReplaceAllString(v, synthEmail)
	v = reTWMobile.ReplaceAllString(v, synthMobile)
	v = synthHosts(v)
	v = reMediaPath.ReplaceAllString(v, "/media/sample-product.jpg")
	v = reConfirmTok.ReplaceAllString(v, "confirmation_token=SYNTHETIC")
	return v
}

// placeholder replaces the literal "string" the Postman export uses.
func (s *synthesizer) placeholder(key string) string {
	switch {
	case reTimeKey.MatchString(key):
		return synthTimestamp
	case strings.Contains(key, "html"):
		return "<p>範例內容</p>"
	case strings.Contains(key, "url"):
		return "https://example.cyberbiz.co/sample"
	case key == "password":
		return "S3cret-example"
	case key == "gender":
		return "女"
	case key == "birthday":
		return "1990-01-01"
	case key == "country_calling_code":
		return "+886"
	case key == "tags_text":
		return "VIP,新客"
	}
	return "範例"
}
