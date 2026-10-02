package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// Sign returns the Signature CYBERBIZ sends for body: the lowercase hex
// HMAC-SHA256 of body keyed by the App Secret. Use it to build requests in
// tests and in the Console; production webhooks are verified with [Verify].
func Sign(body []byte, secret string) string {
	return hex.EncodeToString(mac(body, secret))
}

// SignDomain returns the Domain Signature for shopDomain: the lowercase hex
// HMAC-SHA256 of the Shop Domain keyed by the App Secret.
func SignDomain(shopDomain, secret string) string {
	return hex.EncodeToString(mac([]byte(shopDomain), secret))
}

// Verify reports whether signature is the HMAC-SHA256 of body keyed by
// secret. The digest is accepted as lowercase or uppercase hex (what CYBERBIZ
// sends today) or as standard base64 (what its documentation says); see
// ADR-0001. Both comparisons run in constant time. An empty secret or an
// empty signature never verifies.
func Verify(body []byte, signature, secret string) bool {
	return verifyMAC(mac(body, secret), signature, secret)
}

// VerifyDomain reports whether signature is the HMAC-SHA256 of shopDomain
// keyed by secret, with the same encodings as [Verify].
func VerifyDomain(shopDomain, signature, secret string) bool {
	return verifyMAC(mac([]byte(shopDomain), secret), signature, secret)
}

func mac(msg []byte, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(msg)
	return h.Sum(nil)
}

// verifyMAC compares the expected digest against signature in both
// encodings. Both comparisons are always evaluated so the time taken does not
// reveal which encoding matched.
func verifyMAC(expected []byte, signature, secret string) bool {
	if secret == "" || signature == "" {
		return false
	}
	signature = strings.TrimSpace(signature)
	hexOK := subtle.ConstantTimeCompare(
		[]byte(hex.EncodeToString(expected)), []byte(strings.ToLower(signature))) == 1
	b64OK := subtle.ConstantTimeCompare(
		[]byte(base64.StdEncoding.EncodeToString(expected)), []byte(signature)) == 1
	return hexOK || b64OK
}
