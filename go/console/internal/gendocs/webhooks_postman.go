package gendocs

import (
	"fmt"
	"strings"
)

// FileWebhookPostman is the Postman collection an integrator runs against
// their own webhook receiver.
const FileWebhookPostman = "cyberbiz-webhooks.postman_collection.json"

// webhookResourceOrder fixes the folder order of the collection.
var webhookResourceOrder = []string{
	"orders", "customers", "uid_providers", "bonus_points", "comment_bonus",
	"express_delivery_orders", "products", "variants", "coupons", "customer_vip_level", "apps",
}

// webhookVariables are the collection variables with synthetic defaults.
var webhookVariables = []struct{ Key, Value, Description string }{
	{"webhookUrl", "https://example.com/webhooks/cyberbiz", pmVarWebhookURL},
	{"shopDomain", synthShopDomain, pmVarShopDomain},
	{"customDomain", "www.example.com", pmVarCustomDomain},
	{"appSecret", webhookSecret, pmVarAppSecret},
	{"signatureEncoding", "hex", pmVarSignatureEncoding},
}

// webhookPreRequestScript computes both signatures over the exact body bytes
// with Postman's built-in CryptoJS. Environment values override collection
// variables because pm.variables resolves the full scope chain.
var webhookPreRequestScript = []string{
	"const secret = pm.variables.replaceIn(pm.variables.get('appSecret'));",
	"const raw = pm.request.body && pm.request.body.mode === 'raw' ? pm.request.body.raw : '';",
	"const enc = pm.variables.get('signatureEncoding') === 'base64' ? CryptoJS.enc.Base64 : CryptoJS.enc.Hex;",
	"pm.variables.set('cyberbizHmac', CryptoJS.HmacSHA256(raw, secret).toString(enc));",
	"pm.variables.set('cyberbizDomainHmac', CryptoJS.HmacSHA256(pm.variables.get('shopDomain'), secret).toString(enc));",
}

// webhookTestScript checks the receiver answers like CYBERBIZ expects.
var webhookTestScript = []string{
	"pm.test('receiver answers 200', function () {",
	"  pm.expect(pm.response.code, 'your receiver should answer 200 quickly and process asynchronously; 401 usually means its App Secret differs from appSecret').to.eql(200);",
	"});",
}

// eventPayload is one Event with its synthetic body, pretty-printed with a
// trailing newline exactly as webhooks.md shows it.
type eventPayload struct {
	Event webhookEvent
	Body  []byte
}

// webhookPayloads builds every Event's payload in reference order. Both
// webhooks.md and the Postman collection call it, so they cannot drift.
func webhookPayloads(src *webhookSource, golden *goldenSet, loc *locale, log *report) ([]eventPayload, error) {
	b := newWebhookBuilder(src, golden, loc, log)
	out := make([]eventPayload, 0, len(webhookEvents))
	for _, ev := range webhookEvents {
		body, err := EncodeJSONIndent(b.payload(ev), "  ")
		if err != nil {
			return nil, err
		}
		out = append(out, eventPayload{Event: ev, Body: body})
	}
	return out, nil
}

// buildWebhookPostman renders the receiver-testing collection for a locale.
func buildWebhookPostman(src *webhookSource, golden *goldenSet, loc *locale, log *report) (*OMap, error) {
	payloads, err := webhookPayloads(src, golden, loc, log)
	if err != nil {
		return nil, err
	}
	col := NewOMap()
	col.Set("info", NewOMap().
		Set("_postman_id", stableUUID("CYBERBIZ Webhooks")).
		Set("name", "CYBERBIZ Webhooks").
		Set("description", webhookPostmanDescription(loc)).
		Set("version", DocVersion).
		Set("schema", "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"))
	col.Set("event", []any{
		NewOMap().Set("listen", "prerequest").Set("script", NewOMap().Set("type", "text/javascript").Set("exec", stringsToAny(webhookPreRequestScript))),
		NewOMap().Set("listen", "test").Set("script", NewOMap().Set("type", "text/javascript").Set("exec", stringsToAny(webhookTestScript))),
	})
	var vars []any
	for _, v := range webhookVariables {
		vars = append(vars, NewOMap().Set("key", v.Key).Set("value", v.Value).Set("description", loc.T(v.Description)))
	}
	col.Set("variable", vars)
	folders := NewOMap()
	for _, r := range webhookResourceOrder {
		folders.Set(r, NewOMap().Set("name", r).Set("description", fmt.Sprintf(loc.T(pmFolderTemplate), r)).Set("item", []any{}))
	}
	for _, p := range payloads {
		resource, _, _ := strings.Cut(p.Event.Code, "/")
		f, ok := folders.Get(resource)
		if !ok {
			return nil, fmt.Errorf("webhook postman: event %s has no folder", p.Event.Code)
		}
		items, _ := f.(*OMap).Get("item")
		f.(*OMap).Set("item", append(items.([]any), webhookPostmanItem(p, loc)))
	}
	var items []any
	for _, r := range webhookResourceOrder {
		f, _ := folders.Get(r)
		if list, _ := f.(*OMap).Get("item"); len(list.([]any)) > 0 {
			items = append(items, f)
		}
	}
	col.Set("item", items)
	return col, nil
}

func webhookPostmanItem(p eventPayload, loc *locale) *OMap {
	headers := []any{
		pmHeader("Content-Type", "application/json"),
		pmHeader("User-Agent", webhookUserAgent),
		pmHeader("X-Cyberbiz-Domain", "{{shopDomain}}"),
		pmHeader("X-Cyberbiz-Shop-Domain", "{{customDomain}}"),
		pmHeader("X-Cyberbiz-Event", p.Event.Code),
		pmHeader("X-Cyberbiz-Hmac-Sha256", "{{cyberbizHmac}}"),
		pmHeader("X-Cyberbiz-Domain-Hmac-Sha256", "{{cyberbizDomainHmac}}"),
	}
	req := NewOMap().
		Set("method", "POST").
		Set("header", headers).
		Set("body", NewOMap().
			Set("mode", "raw").
			Set("raw", strings.TrimRight(string(p.Body), "\n")).
			Set("options", NewOMap().Set("raw", NewOMap().Set("language", "json")))).
		Set("url", NewOMap().Set("raw", "{{webhookUrl}}").Set("host", []any{"{{webhookUrl}}"})).
		Set("description", fmt.Sprintf(loc.T(pmRequestTemplate), p.Event.Code, loc.T(p.Event.Description), p.Event.Payload)+"\n\n"+loc.T(pmUnauthorizedHint))
	return NewOMap().Set("name", p.Event.Code).Set("request", req).Set("response", []any{})
}

func pmHeader(k, v string) *OMap {
	return NewOMap().Set("key", k).Set("value", v)
}

func stringsToAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func webhookPostmanDescription(loc *locale) string {
	return "CYBERBIZ Webhooks\n\n" + loc.T(pmWebhookUsage) + "\n\n" + releaseNotesMarkdown(loc)
}

// Authored text of the receiver-testing collection.
const (
	pmVarWebhookURL        = "Your receiver's HTTPS URL. Override it in your own Postman environment (*.postman_environment.json files are never committed)."
	pmVarShopDomain        = "The Shop Domain sent as X-Cyberbiz-Domain and signed into X-Cyberbiz-Domain-Hmac-Sha256. Override it in your environment."
	pmVarCustomDomain      = "The merchant's custom storefront hostname sent as X-Cyberbiz-Shop-Domain. Override it in your environment."
	pmVarAppSecret         = "The App Secret used to sign the requests. Set the real one in your own Postman environment, never in this collection."
	pmVarSignatureEncoding = "`hex` or `base64`. Production App webhooks send hex (64 characters) although CYBERBIZ's documentation says base64; see ADR-0001. Switch to `base64` to test the documented form."
	pmFolderTemplate       = "Events of the `%s` resource."
	pmRequestTemplate      = "CYBERBIZ sends `%s` when: %s The body is the `%s` payload object, exactly as documented in webhooks.md; the pre-request script signs it with `appSecret`."
	pmUnauthorizedHint     = "A 401 from your receiver usually means its App Secret differs from `appSecret`."
	pmWebhookUsage         = `## Usage

This collection replays every documented CYBERBIZ Event against your own webhook receiver, with the same headers, payloads and signatures CYBERBIZ sends.

1. Create a Postman environment with your receiver's ` + "`webhookUrl`" + `, your shop's ` + "`shopDomain`" + ` / ` + "`customDomain`" + ` and the real ` + "`appSecret`" + `; never commit it.
2. Send a request. The collection's pre-request script computes ` + "`X-Cyberbiz-Hmac-Sha256`" + ` (HMAC-SHA256 of the exact body bytes) and ` + "`X-Cyberbiz-Domain-Hmac-Sha256`" + ` (HMAC-SHA256 of ` + "`shopDomain`" + `) with ` + "`appSecret`" + `, in the encoding selected by ` + "`signatureEncoding`" + `.
3. The collection's test expects a 200: your receiver should answer quickly and process the Event asynchronously.

Payloads are synthetic Samples generated from the same source as webhooks.md; they contain no variables so the signature always matches the bytes sent.`
	whPostmanHeading = "Testing your receiver with Postman"
	whPostmanText    = "The collection `cyberbiz-webhooks.postman_collection.json` next to this file replays every Event below against your own receiver: set `webhookUrl`, `shopDomain`, `customDomain` and `appSecret` in a Postman environment (never commit it), and its pre-request script signs each request exactly as CYBERBIZ does (`signatureEncoding` selects hex, the production form, or base64, the documented form)."
)

var webhookPostmanStrings = []string{
	pmVarWebhookURL, pmVarShopDomain, pmVarCustomDomain, pmVarAppSecret, pmVarSignatureEncoding,
	pmFolderTemplate, pmRequestTemplate, pmUnauthorizedHint, pmWebhookUsage, whPostmanHeading, whPostmanText,
}
