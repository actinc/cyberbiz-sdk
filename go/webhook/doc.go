// Package webhook receives and authenticates Inbound webhooks from the
// CYBERBIZ e-commerce platform.
//
// CYBERBIZ announces Shop changes by POSTing a JSON body to an HTTPS URL the
// App registered. Every request carries these headers:
//
//	User-Agent:                    CyberbizAppWebhook/1.0
//	X-Cyberbiz-Domain:             example.cyberbiz.co   (the Shop Domain)
//	X-Cyberbiz-Shop-Domain:        www.example-shop.com  (the merchant's custom storefront hostname)
//	X-Cyberbiz-Event:              orders/paid           (the Event, resource/action)
//	X-Cyberbiz-Hmac-Sha256:        <HMAC-SHA256(App Secret, raw body)>
//	X-Cyberbiz-Domain-Hmac-Sha256: <HMAC-SHA256(App Secret, X-Cyberbiz-Domain)>
//
// The Shop Domain is the key for looking up the Shop's App Secret; the
// custom domain is informational only. The body Signature is what
// authenticates the payload. CYBERBIZ's documentation says the digest is
// base64, but every production delivery carries 64 lowercase hex characters,
// so [Verify] accepts hex first and base64 second, both in constant time.
// The Domain Signature is checked when present and a mismatch is reported as
// [ErrInvalidDomainSignature], distinct from a bad body signature.
//
// Only the App webhook flavour (User-Agent CyberbizAppWebhook/1.0) is
// supported.
//
// # Usage
//
// [Parse] turns an *http.Request into an authenticated [Event]:
//
//	secrets := webhook.StaticSecret(os.Getenv("CYBERBIZ_APP_SECRET"))
//
//	func handle(w http.ResponseWriter, r *http.Request) {
//		e, err := webhook.Parse(r.Context(), r, secrets)
//		if err != nil {
//			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
//			return
//		}
//		switch e.Type {
//		case webhook.EventOrdersPaid:
//			order, err := e.Order()
//			...
//		}
//		w.WriteHeader(http.StatusOK)
//	}
//
// [Handler] wraps the same steps in an http.Handler with the conventional
// status codes. Multi-shop integrations implement [SecretResolver] to map a
// Shop Domain to its App Secret.
//
// The typed accessors ([Event.Order], [Event.Customer], ...) decode the body
// into the payload types in this package; [Event.Decode] decodes into any
// value, and [Event.Raw] is the untouched body for storage or forwarding.
//
// # Delivery semantics
//
// CYBERBIZ retries a delivery that does not receive a 2xx response; the
// retry schedule is not documented. It also emits several events for one
// customer within the same second (for example customers/update followed by
// bonus_points/create when an order is paid), and deliveries can arrive out
// of order. Handlers must therefore be idempotent, and a handler that writes
// state derived from more than one event should serialise its work per
// customer (or per order) rather than assume one delivery at a time.
//
// Respond quickly: do the minimum needed to persist the event, then return
// 200. Anything slow belongs in a queue.
//
// Nothing in this package logs, prints, or retains the App Secret or the
// request body beyond the returned [Event].
package webhook
