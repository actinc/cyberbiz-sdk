// Package redact strips personal and secret data from CYBERBIZ API payloads
// before they are stored as Golden Files or shown in the Console. It rewrites
// JSON as a token stream, so key order, nesting, and value types are kept and
// a redacted document still validates against the same schema as the
// original.
//
// Three rules decide what happens to a string value:
//
//   - the object key matches a name in [SensitiveKeys] or a suffix/prefix in
//     [SensitivePatterns]: the value is replaced by a placeholder of the same
//     shape (an e-mail stays an e-mail, a date stays a date);
//   - the string value looks like an e-mail address, a Taiwanese mobile
//     number, or a JWT: same replacement, whatever the key;
//   - the value names a host: any Shop's CYBERBIZ subdomain becomes
//     example.cyberbiz.co and every other non-CYBERBIZ host becomes
//     example.com, so a Golden File never identifies the merchant or an
//     integrator's servers. Keys in [HashedKeys] (merchant identifiers such as
//     SKUs) are replaced by a stable hash-derived token instead.
package redact

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/jsontext"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// SensitiveKeys are JSON object keys whose values are always replaced,
// regardless of nesting. Matching is case-insensitive.
var SensitiveKeys = map[string]bool{
	"name": true, "first_name": true, "last_name": true, "full_name": true,
	"nickname": true, "email": true, "mobile": true, "phone": true, "tel": true,
	"birthday": true, "note": true, "card4no": true, "transaction_number": true,
	"merchant_trade_no": true, "paper_company_no": true, "uid": true,
	"payment_url": true, "activation_url": true, "token": true, "access_token": true,
	"line_uid": true, "ip": true, "id_number": true, "vat_number": true,
	"company_no": true, "love_code": true, "password": true, "secret": true,
	"remote_addr": true, "customer_email": true, "customer_mobile": true,
	"customer_name": true, "receiver_name": true, "buyer_name": true,
}

// SensitivePatterns are matched against the lower-cased key with
// strings.HasPrefix (entries ending in "*") or strings.HasSuffix (entries
// starting with "*").
var SensitivePatterns = []string{
	"address*", "*_address", "*invoice_no", "referral_code*", "*_referral_code",
	"*_email", "*_mobile", "*_phone", "carrier*", "*_token", "*_secret",
}

// HashedKeys are merchant identifiers that must not survive verbatim but
// whose values must stay distinct and stable across runs (a SKU appears in
// several Golden Files and tests compare them). They are replaced by
// "SKU-" plus a short hash of the original value.
var HashedKeys = map[string]bool{
	"sku": true, "vendor_type": true, "store_no": true, "store_number": true,
}

// SensitiveHeaders are HTTP headers whose values are replaced.
var SensitiveHeaders = map[string]bool{
	"Authorization": true, "Cookie": true, "Set-Cookie": true,
	"X-Forwarded-For": true, "X-Real-Ip": true, "Proxy-Authorization": true,
}

var (
	// shopHostRE matches a Shop's CYBERBIZ-issued hostname. The platform's own
	// hosts (api, app-store-api, app, www) are excluded.
	shopHostRE = regexp.MustCompile(`\b(?:[a-z0-9-]+\.)*?([a-z0-9-]+)\.cyberbiz\.(?:co|io)\b`)
	// urlHostRE matches the host of an absolute or protocol-relative URL.
	urlHostRE = regexp.MustCompile(`((?:https?:)?//)([A-Za-z0-9.-]+\.[A-Za-z]{2,})`)
	// bareHostRE matches a value that is nothing but a hostname.
	bareHostRE = regexp.MustCompile(`^(?:[a-z0-9-]+\.)+[a-z]{2,}$`)
	emailRE    = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	mobileRE   = regexp.MustCompile(`^(\+?886-?|0)9\d{2}-?\d{3}-?\d{3}$`)
	jwtRE      = regexp.MustCompile(`^ey[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)
	dateRE     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)
	urlRE      = regexp.MustCompile(`^https?://`)
)

// queryTokenRE matches a token passed in a URL query, such as an account
// activation link's confirmation_token, up to the end of its value.
var queryTokenRE = regexp.MustCompile(`(?i)([?&][a-z_]*token=)[^&#"\s]+`)

// QueryTokens replaces the value of every *token= URL query parameter in s,
// so a link copied from a real response cannot be used.
func QueryTokens(s string) string {
	return queryTokenRE.ReplaceAllString(s, "${1}"+Placeholder)
}

// Placeholder is the value written for a redacted string that has no more
// specific shape.
const Placeholder = "REDACTED"

// ExampleShopHost replaces every Shop's CYBERBIZ subdomain.
const ExampleShopHost = "example.cyberbiz.co"

// ExampleHost replaces every other host that is not CYBERBIZ's own.
const ExampleHost = "example.com"

// platformHosts are CYBERBIZ-owned hostnames that identify no merchant.
var platformHosts = map[string]bool{
	"api": true, "app-store-api": true, "app": true, "www": true, "api-doc": true, "example": true,
}

// keepHost reports whether a URL host is safe to leave as it is: CYBERBIZ's
// own infrastructure, its CDN, the documentation placeholders, or localhost.
func keepHost(host string) bool {
	h := strings.ToLower(host)
	switch {
	case h == "localhost", h == ExampleHost, strings.HasSuffix(h, "."+ExampleHost):
		return true
	case strings.HasSuffix(h, ".cybassets.com"), h == "cyberbiz.io", h == "cyberbiz.co":
		return true
	case strings.HasSuffix(h, ".cyberbiz.co") || strings.HasSuffix(h, ".cyberbiz.io"):
		return platformHosts[strings.Split(h, ".")[0]]
	}
	return false
}

// Hosts rewrites every hostname in s that could identify a merchant or an
// integrator. It is applied to every string value and header value.
func Hosts(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	out := shopHostRE.ReplaceAllStringFunc(s, func(m string) string {
		sub := shopHostRE.FindStringSubmatch(m)
		if len(sub) > 1 && platformHosts[strings.ToLower(sub[1])] {
			return m
		}
		return ExampleShopHost
	})
	out = urlHostRE.ReplaceAllStringFunc(out, func(m string) string {
		sub := urlHostRE.FindStringSubmatch(m)
		if keepHost(sub[2]) {
			return m
		}
		return sub[1] + ExampleHost
	})
	if bareHostRE.MatchString(out) && !keepHost(out) {
		return ExampleHost
	}
	return out
}

// Hashed returns the stable replacement for a merchant identifier such as a
// SKU: "SKU-" followed by six hex characters derived from the value.
func Hashed(key, value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.ToLower(key) + ":" + value))
	return "SKU-" + hex.EncodeToString(sum[:3])
}

// Key reports whether an object key is sensitive.
func Key(key string) bool {
	k := strings.ToLower(key)
	if SensitiveKeys[k] {
		return true
	}
	for _, p := range SensitivePatterns {
		switch {
		case strings.HasSuffix(p, "*") && strings.HasPrefix(k, strings.TrimSuffix(p, "*")):
			return true
		case strings.HasPrefix(p, "*") && strings.HasSuffix(k, strings.TrimPrefix(p, "*")):
			return true
		}
	}
	return false
}

// Value reports whether a string value is sensitive on its own.
func Value(s string) bool {
	return emailRE.MatchString(s) || mobileRE.MatchString(s) || jwtRE.MatchString(s)
}

// String returns the replacement for a sensitive string, preserving shape.
func String(s string) string {
	switch {
	case s == "":
		return ""
	case emailRE.MatchString(s):
		return "redacted@example.com"
	case mobileRE.MatchString(s):
		return "0912345678"
	case jwtRE.MatchString(s):
		return "eyJhbGciOiJIUzI1NiJ9.REDACTED.REDACTED"
	case urlRE.MatchString(s):
		return "https://example.com/redacted"
	case dateRE.MatchString(s):
		if len(s) > 10 {
			return "1990-01-01 00:00:00"
		}
		return "1990-01-01"
	}
	return Placeholder
}

// JSON returns a copy of data with sensitive values replaced. Non-JSON input
// is returned unchanged with a nil error, since bodies are not always JSON.
func JSON(data []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return data, nil
	}
	dec := jsontext.NewDecoder(bytes.NewReader(data),
		jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true))
	var out bytes.Buffer
	enc := jsontext.NewEncoder(&out, jsontext.AllowInvalidUTF8(true),
		jsontext.AllowDuplicateNames(true), jsontext.Multiline(true), jsontext.WithIndent("  "))

	// keys[i] is the member name the parser is currently under at nesting
	// level i+1. Inside an array the level inherits the array's own key, so
	// strings in "tags": [...] are judged by the key "tags".
	var keys []string
	// tainted[i] is true when level i+1 sits inside a container whose own
	// key is sensitive (e.g. everything under "address": {...}).
	var tainted []bool
	for {
		tok, err := dec.ReadToken()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		switch tok.Kind() {
		case '{', '[':
			parent, inherited := "", false
			if len(keys) > 0 {
				parent = keys[len(keys)-1]
				inherited = tainted[len(keys)-1]
			}
			keys = append(keys, parent)
			tainted = append(tainted, inherited || Key(parent))
		case '}', ']':
			keys = keys[:len(keys)-1]
			tainted = tainted[:len(tainted)-1]
		case '"':
			depth := dec.StackDepth()
			kind, n := dec.StackIndex(depth)
			if kind == '{' && n%2 == 1 {
				// An object member name: remember it, never redact it.
				keys[depth-1] = tok.String()
				break
			}
			parent, inherited := "", false
			if depth > 0 {
				parent, inherited = keys[depth-1], tainted[depth-1]
			}
			s := tok.String()
			switch {
			case HashedKeys[strings.ToLower(parent)]:
				tok = jsontext.String(Hashed(parent, s))
			case inherited || Key(parent) || Value(s):
				tok = jsontext.String(String(s))
			default:
				tok = jsontext.String(Hosts(QueryTokens(s)))
			}
		}
		if err := enc.WriteToken(tok); err != nil {
			return nil, err
		}
	}
	return bytes.TrimRight(out.Bytes(), "\n"), nil
}

// Headers returns a copy of h with sensitive header values replaced.
func Headers(h http.Header) http.Header {
	out := make(http.Header, len(h))
	for k, vs := range h {
		if SensitiveHeaders[http.CanonicalHeaderKey(k)] {
			out[k] = []string{Placeholder}
			continue
		}
		clean := make([]string, len(vs))
		for i, v := range vs {
			clean[i] = Hosts(QueryTokens(v))
		}
		out[k] = clean
	}
	return out
}
