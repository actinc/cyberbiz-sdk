package gendocs

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// Synthetic values used in the webhook header samples.
const (
	webhookSecret     = "example-app-secret"
	webhookUserAgent  = "CyberbizAppWebhook/1.0"
	webhookShopDomain = "www.example-shop.com"
)

// webhookBuilder assembles the synthetic payload of every Event from the
// reference field tables, taking values from the Golden order / customer /
// product objects where the tables refer to them.
type webhookBuilder struct {
	src   *webhookSource
	bases map[string]*OMap // section -> synthetic Golden object
	loc   *locale
	log   *report
	syn   *synthesizer
	seen  map[string]int
}

func newWebhookBuilder(src *webhookSource, golden *goldenSet, loc *locale, log *report) *webhookBuilder {
	b := &webhookBuilder{src: src, bases: map[string]*OMap{}, loc: loc, log: log, seen: map[string]int{}}
	b.syn = newSynthesizer()
	if o := golden.success("GET", "/v1/orders/{order_id}"); usableGolden(o) {
		b.bases["Order"] = b.syn.Value("", o.Body, nil).(*OMap)
	}
	if c := golden.success("GET", "/v1/customers/{customer_id}"); usableGolden(c) {
		cust := b.syn.Value("", c.Body, nil).(*OMap)
		if !cust.Has("id") {
			cust.Set("id", json.Number("1"))
		}
		b.bases["Customer"] = cust
	}
	if p := golden.success("GET", "/v1/products/{product_id}"); usableGolden(p) {
		prod := b.syn.Value("", p.Body, nil).(*OMap)
		if !prod.Has("id") {
			prod.Set("id", json.Number("1"))
		}
		b.bases["Product"] = prod
		if vs, ok := prod.Get("product_variants"); ok {
			if arr, ok := vs.([]any); ok && len(arr) > 0 {
				b.bases["Product Variant"] = arr[0].(*OMap)
			}
		}
	}
	return b
}

// payload builds the sample body of one event.
func (b *webhookBuilder) payload(ev webhookEvent) *OMap {
	b.seen = map[string]int{}
	return b.object(ev.Payload, b.bases[ev.Payload], 0)
}

// object builds a section's object; base supplies recorded values.
func (b *webhookBuilder) object(section string, base *OMap, depth int) *OMap {
	out := NewOMap()
	fields := b.src.fields(section)
	if len(fields) == 0 && base != nil {
		return base
	}
	b.seen[section]++
	defer func() { b.seen[section]-- }()
	for _, f := range fields {
		var baseVal any
		if base != nil {
			baseVal, _ = base.Get(f.Name)
		}
		out.Set(f.Name, b.value(section, f, baseVal, depth+1))
	}
	if base != nil {
		for _, k := range base.Keys() {
			if !out.Has(k) {
				v, _ := base.Get(k)
				out.Set(k, v)
			}
		}
	}
	return out
}

// value builds one field from its declared type, the recorded value and
// the referenced section.
func (b *webhookBuilder) value(parent string, f whField, base any, depth int) any {
	sub := b.section(parent, f.Name)
	isList := strings.HasPrefix(f.Type, "[")
	if sub != "" && depth < 8 && b.seen[sub] == 0 {
		if isList {
			var first *OMap
			if arr, ok := base.([]any); ok && len(arr) > 0 {
				first, _ = arr[0].(*OMap)
			}
			if first == nil && base != nil {
				return base // recorded as an empty list
			}
			return []any{b.object(sub, first, depth)}
		}
		bm, _ := base.(*OMap)
		if base != nil && bm == nil {
			return base // recorded as null
		}
		return b.object(sub, bm, depth)
	}
	if base != nil {
		return base
	}
	return b.generic(f)
}

func (b *webhookBuilder) section(parent, key string) string {
	if s, ok := webhookSectionByKey[parent+"/"+key]; ok {
		return s
	}
	return webhookSectionByKey[key]
}

// generic produces a value from the declared type alone.
func (b *webhookBuilder) generic(f whField) any {
	switch strings.Trim(f.Type, "[]") {
	case "Integer":
		if strings.HasPrefix(f.Type, "[") {
			return []any{json.Number("1")}
		}
		return json.Number("1")
	case "Float":
		return json.Number("100.0")
	case "Boolean":
		return true
	case "Date":
		return synthTimestamp
	case "Object", "Hash":
		if strings.HasPrefix(f.Type, "[") {
			return []any{}
		}
		return NewOMap()
	}
	if ed, ok := enumDocs[f.Name]; ok {
		return ed.Values[0]
	}
	if v, ok := webhookFieldSamples[f.Name]; ok {
		return v
	}
	if strings.HasPrefix(f.Type, "[") {
		return []any{"範例"}
	}
	if v, ok := b.syn.keyed(f.Name, "sample", nil); ok {
		return v
	}
	return b.syn.placeholder(f.Name)
}

// webhookFieldSamples are values for documented fields the Golden Files do
// not carry and whose meaning a generic placeholder would obscure.
var webhookFieldSamples = map[string]any{
	"code":                   "SAMPLE100",
	"coupon_type_name":       "金額",
	"coupon_value":           "100",
	"concurrently_apply":     "true",
	"restrict_strategy":      "unrestricted",
	"restrict_campaigns":     []any{"shop_discount"},
	"tags":                   []any{"VIP"},
	"source":                 "shop",
	"status":                 "enabled",
	"provider_type":          "line",
	"item_type":              "normal",
	"source_type":            "BranchStore",
	"app_uuid":               "00000000-0000-4000-8000-000000000001",
	"app_version_uuid":       "00000000-0000-4000-8000-000000000002",
	"app_client_id":          "example-app-client-id",
	"opening_hours":          "09:00-21:00",
	"lat":                    json.Number("25.0330"),
	"lng":                    json.Number("121.5654"),
	"birth_gift_before_days": "7",
	"description_url":        "https://example.cyberbiz.co/pages/vip",
	"customer_tags":          []any{"VIP"},
	"product_tags":           []any{"VIP"},
	"temperature_types":      []any{"常溫"},
	"required_customer_tags": []any{},
	"serial_numbers":         []any{},
	"types":                  "大,中,小",
	"county":                 "台北市",
	"pos_info":               "",
	"channel_shop_name":      "範例通路商店",
	"body":                   "退款說明",
	"reason_detail":          "顧客改變心意",
	"tracking_url":           "https://example.com/track/TRK000000001",
	"pos_user_email":         synthStaffEmail,
	"pos_name":               "POS 1",
	"filter":                 "商品價格 小於 1000",
}

// signatures computes the synthetic signature headers for a body.
func signatures(body []byte) (hexSig, b64Sig, domainSig string) {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	sum := mac.Sum(nil)
	dm := hmac.New(sha256.New, []byte(webhookSecret))
	dm.Write([]byte(synthShopDomain))
	return hex.EncodeToString(sum), base64.StdEncoding.EncodeToString(sum), hex.EncodeToString(dm.Sum(nil))
}

// buildWebhooksMarkdown renders webhooks.md for one locale.
func buildWebhooksMarkdown(src *webhookSource, golden *goldenSet, loc *locale, log *report) (string, error) {
	payloads, err := webhookPayloads(src, golden, loc, log)
	if err != nil {
		return "", err
	}
	t := loc.T
	var md strings.Builder
	md.WriteString("# CYBERBIZ Webhooks\n\n")
	fmt.Fprintf(&md, "Version: %s\n\n", DocVersion)
	md.WriteString(t(webhookIntro))
	md.WriteString("\n\n")
	md.WriteString(fmt.Sprintf(t(webhookHeadersTemplate), webhookUserAgent, synthShopDomain, webhookShopDomain))
	md.WriteString("\n\n")
	md.WriteString(t(webhookVerifyText))
	md.WriteString("\n\n")
	md.WriteString(t(webhookRetryText))
	fmt.Fprintf(&md, "\n\n## %s\n\n%s\n", t(whPostmanHeading), t(whPostmanText))
	fmt.Fprintf(&md, "\n## %s\n\n| %s | %s | %s |\n| --- | --- | --- |\n", t(whEventsHeading), t(whColEvent), t(whColDescription), t(whColPayload))
	for _, ev := range webhookEvents {
		fmt.Fprintf(&md, "| `%s` | %s | [%s](#%s) |\n", ev.Code, t(ev.Description), ev.Payload, anchor(ev.Payload))
	}
	fmt.Fprintf(&md, "\n### %s\n\n%s\n\n", t(whUndocumentedHeading), t(whUndocumentedText))
	for _, e := range undocumentedEvents {
		fmt.Fprintf(&md, "- `%s`\n", e)
	}
	fmt.Fprintf(&md, "\n## %s\n\n%s\n", t(whPayloadsHeading), t(whPayloadsText))
	for _, sec := range src.Order {
		fmt.Fprintf(&md, "\n### %s\n\n", sec)
		if sec == "Cart" {
			fmt.Fprintf(&md, "%s\n\n", t(whCartNote))
		}
		fmt.Fprintf(&md, "| %s | %s | %s |\n| --- | --- | --- |\n", t(whColField), t(whColType), t(whColDescription))
		for _, f := range src.fields(sec) {
			desc := loc.Source(f.Desc)
			if loc.glossary && containsCJK(desc) {
				log.untranslated[desc]++
			}
			fmt.Fprintf(&md, "| `%s` | `%s` | %s |\n", f.Name, f.Type, escapePipes(desc))
		}
	}
	fmt.Fprintf(&md, "\n## %s\n\n%s\n", t(whSamplesHeading), fmt.Sprintf(t(whSamplesText), webhookSecret))
	for _, p := range payloads {
		ev, body := p.Event, p.Body
		hexSig, b64Sig, domainSig := signatures(body)
		fmt.Fprintf(&md, "\n### `%s`\n\n%s\n\n```http\n", ev.Code, t(ev.Description))
		fmt.Fprintf(&md, "POST /webhooks/cyberbiz HTTP/1.1\nHost: example.com\nContent-Type: application/json\nUser-Agent: %s\nX-Cyberbiz-Domain: %s\nX-Cyberbiz-Shop-Domain: %s\nX-Cyberbiz-Event: %s\nX-Cyberbiz-Hmac-Sha256: %s\nX-Cyberbiz-Domain-Hmac-Sha256: %s\n```\n\n",
			webhookUserAgent, synthShopDomain, webhookShopDomain, ev.Code, hexSig, domainSig)
		fmt.Fprintf(&md, "%s `%s`\n\n", t(whBase64Text), b64Sig)
		fmt.Fprintf(&md, "```json\n%s```\n", body)
	}
	md.WriteString("\n" + releaseNotesMarkdown(loc) + "\n")
	return md.String(), nil
}

func anchor(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

func escapePipes(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// Authored prose of webhooks.md. Every constant here is a key of the locale
// tables; the English text is what the en locale writes.
const (
	whEventsHeading       = "Events"
	whColEvent            = "Event"
	whColDescription      = "Description"
	whColPayload          = "Payload"
	whColField            = "Field"
	whColType             = "Type"
	whUndocumentedHeading = "Events without documented payloads"
	whUndocumentedText    = "These events appear in app manifests (`webhook_events` of `GET /settings`) but the reference documents no payload for them:"
	whPayloadsHeading     = "Payload objects"
	whPayloadsText        = "Field types are as documented by CYBERBIZ. Timestamps are `YYYY-MM-DD HH:MM:SS` in Asia/Taipei; money fields are floats; any field may be `null`."
	whCartNote            = "CYBERBIZ documents this payload without naming an event for it; no delivery has been observed and the SDK defines no cart Event. It is listed for completeness only."
	whSamplesHeading      = "Samples"
	whSamplesText         = "Every sample below is synthetic. The signature headers are computed with the App Secret `%s` over the exact JSON bytes shown (pretty-printed, two-space indent, trailing newline), so the samples can be used to test a verifier."
	whBase64Text          = "Base64 form of the same signature (as the CYBERBIZ documentation describes it):"
)

const webhookIntro = `CYBERBIZ posts an HTTP request to the HTTPS URL an app registers (` + "`webhook_url`" + ` in the app
manifest) whenever a subscribed Event happens in a Shop. This document lists every Event, the
HTTP headers CYBERBIZ sends, how to verify the Signature, and a synthetic Sample of every payload.

## Delivery

- Method ` + "`POST`" + `, body ` + "`application/json`" + `, UTF-8. The body is the payload object of the Event, not an envelope.
- One request per Event per resource. Several Events for the same resource (e.g. ` + "`orders/create`" + ` and ` + "`orders/paid`" + `, or two ` + "`customers/update`" + `) can arrive within the same second and in any order; order by ` + "`updated_at`" + ` and the resource id, not by arrival time.
- Answer ` + "`2xx`" + ` quickly (a few seconds). Do the work asynchronously.`

// webhookHeadersTemplate is formatted with the User-Agent, the Shop Domain
// and the custom domain sample values.
const webhookHeadersTemplate = `### Headers

| Header | Example | Description |
| --- | --- | --- |
| ` + "`User-Agent`" + ` | ` + "`%s`" + ` | Fixed for App webhooks (observed). The documentation says ` + "`CyberbizWebhook/1.0`" + `, which is the older shop-webhook flavour. |
| ` + "`X-Cyberbiz-Domain`" + ` | ` + "`%s`" + ` | The Shop Domain, i.e. the shop's CYBERBIZ-issued hostname. Use it to find the shop's App Secret. |
| ` + "`X-Cyberbiz-Shop-Domain`" + ` | ` + "`%s`" + ` | The merchant's custom storefront hostname. Informational; not an identifier. |
| ` + "`X-Cyberbiz-Event`" + ` | ` + "`orders/paid`" + ` | The Event, named ` + "`resource/action`" + `. |
| ` + "`X-Cyberbiz-Hmac-Sha256`" + ` | ` + "`3f2a…`" + ` (64 hex chars) | The Signature: HMAC-SHA256 of the raw request body keyed by the App Secret. |
| ` + "`X-Cyberbiz-Domain-Hmac-Sha256`" + ` | ` + "`9b1c…`" + ` (64 hex chars) | The Domain Signature: HMAC-SHA256 of the ` + "`X-Cyberbiz-Domain`" + ` value keyed by the App Secret. |`

const webhookVerifyText = `## Verifying the signature

1. Read the raw body bytes before parsing them.
2. Compute ` + "`HMAC-SHA256(app_secret, body)`" + `.
3. Compare, in constant time, the **hex** encoding of the digest with ` + "`X-Cyberbiz-Hmac-Sha256`" + `. Production App webhooks send hex (64 characters), although the CYBERBIZ documentation describes base64 — accept base64 as a fallback so a future change on the platform does not break verification.
4. Optionally check ` + "`X-Cyberbiz-Domain-Hmac-Sha256`" + ` the same way against ` + "`X-Cyberbiz-Domain`" + `; it binds the request to one shop but does not authenticate the payload.
5. Reject anything that fails with ` + "`401`" + ` and do not process the body.

` + "```go" + `
func verify(body []byte, header, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sum := mac.Sum(nil)
	hexOK := hmac.Equal([]byte(hex.EncodeToString(sum)), []byte(header))
	b64OK := hmac.Equal([]byte(base64.StdEncoding.EncodeToString(sum)), []byte(header))
	return hexOK || b64OK
}
` + "```"

const webhookRetryText = `## Retries and idempotency

- The retry policy is **not documented**: assume a failed delivery may or may not be retried, and that a retry may arrive minutes later or never. Reconcile through the API (` + "`GET /v1/orders/{order_id}`" + `, ...) when a gap is suspected.
- Deliveries carry no delivery id. Deduplicate on ` + "`(X-Cyberbiz-Event, payload id, payload updated_at)`" + `.
- Treat every handler as idempotent: the same Event can be delivered more than once, and Events for one resource can arrive out of order within the same second.`

// authoredConstants lists the prose constants of this package for the
// locale coverage test.
var authoredConstants = []string{
	webhookIntro, webhookHeadersTemplate, webhookVerifyText, webhookRetryText,
	whEventsHeading, whColEvent, whColDescription, whColPayload, whColField, whColType,
	whUndocumentedHeading, whUndocumentedText, whPayloadsHeading, whPayloadsText, whCartNote,
	whSamplesHeading, whSamplesText, whBase64Text,
	v1Intro, v2Intro, conventionsTemplate,
	recordedNote, nullNote, includeNote, scopeTemplate, serverNote, bearerNote,
	pageNote, perPageNote, offsetNote, successNote, unresolvedNote,
	observedNote, nullOnlyNote,
	postmanTokenNote, postmanGeneratedNote, postmanEnvNote,
}
