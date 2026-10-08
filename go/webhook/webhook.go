package webhook

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Header names CYBERBIZ sends on every Inbound.
const (
	HeaderDomain     = "X-Cyberbiz-Domain"
	HeaderShopDomain = "X-Cyberbiz-Shop-Domain"
	HeaderEvent      = "X-Cyberbiz-Event"
	HeaderSignature  = "X-Cyberbiz-Hmac-Sha256"
	HeaderDomainHMAC = "X-Cyberbiz-Domain-Hmac-Sha256"
	UserAgent        = "CyberbizAppWebhook/1.0"
)

// MaxBodyBytes is the default limit on an Inbound body. Real order payloads
// are tens of kilobytes; 2 MiB leaves ample room while bounding memory.
const MaxBodyBytes = 2 << 20

// Sentinel errors. Every error [Parse] returns wraps one of these, so
// callers can classify with errors.Is.
var (
	// ErrMissingHeader means a required X-Cyberbiz-* header is absent. The
	// wrapping error names the header.
	ErrMissingHeader = errors.New("webhook: missing header")
	// ErrUnknownShop means the SecretResolver has no App Secret for the
	// Shop Domain in X-Cyberbiz-Domain.
	ErrUnknownShop = errors.New("webhook: unknown shop")
	// ErrInvalidSignature means X-Cyberbiz-Hmac-Sha256 does not match the
	// body under the resolved App Secret.
	ErrInvalidSignature = errors.New("webhook: invalid signature")
	// ErrInvalidDomainSignature means X-Cyberbiz-Domain-Hmac-Sha256 was
	// present but does not match the Shop Domain under the resolved App
	// Secret. The body signature had already verified.
	ErrInvalidDomainSignature = errors.New("webhook: invalid domain signature")
	// ErrBodyTooLarge means the body exceeded the configured limit.
	ErrBodyTooLarge = errors.New("webhook: body too large")
)

// SecretResolver returns the App Secret for a Shop Domain. Return an error,
// or an empty secret, for a Shop the integration does not know; Parse maps
// both to [ErrUnknownShop].
type SecretResolver interface {
	WebhookSecret(ctx context.Context, shopDomain string) (string, error)
}

// SecretResolverFunc adapts a function to the [SecretResolver] interface.
type SecretResolverFunc func(ctx context.Context, shopDomain string) (string, error)

// WebhookSecret calls f.
func (f SecretResolverFunc) WebhookSecret(ctx context.Context, shopDomain string) (string, error) {
	return f(ctx, shopDomain)
}

// StaticSecret returns a resolver that answers every Shop Domain with the
// same App Secret, for single-shop integrations.
func StaticSecret(secret string) SecretResolver {
	return SecretResolverFunc(func(context.Context, string) (string, error) {
		return secret, nil
	})
}

// Event is one authenticated Inbound.
type Event struct {
	// Type is the Event from X-Cyberbiz-Event, e.g. "orders/paid".
	Type EventType
	// AppID identifies the App whose secret verified the body Signature, as
	// returned by a [CredentialResolver]. It is empty when the resolver is a
	// plain [SecretResolver].
	AppID string
	// ShopDomain is the Shop Domain from X-Cyberbiz-Domain, e.g.
	// "example.cyberbiz.co". It identifies the Shop.
	ShopDomain string
	// CustomDomain is the merchant's storefront hostname from
	// X-Cyberbiz-Shop-Domain. Informational; it is not an identifier.
	CustomDomain string
	// Signature is the verified value of X-Cyberbiz-Hmac-Sha256.
	Signature string
	// DomainSignature is the value of X-Cyberbiz-Domain-Hmac-Sha256, empty
	// when the header was absent.
	DomainSignature string
	// ReceivedAt is when Parse ran.
	ReceivedAt time.Time
	// Header is a copy of the request headers.
	Header http.Header
	// Raw is the request body, byte for byte as CYBERBIZ sent it.
	Raw []byte
}

// Decode unmarshals the raw body into v with encoding/json/v2, tolerating
// invalid UTF-8 and duplicate object names as the rest of the SDK does
// (ADR-0005).
func (e *Event) Decode(v any) error {
	return json.Unmarshal(e.Raw, v,
		jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true))
}

type options struct {
	maxBodyBytes int64
	domainCheck  bool
}

// Option configures [Parse] and [Handler].
type Option func(*options)

// WithMaxBodyBytes replaces [MaxBodyBytes] as the body limit.
func WithMaxBodyBytes(n int64) Option {
	return func(o *options) { o.maxBodyBytes = n }
}

// WithoutDomainCheck skips verification of X-Cyberbiz-Domain-Hmac-Sha256.
// The body signature is still required and verified.
func WithoutDomainCheck() Option {
	return func(o *options) { o.domainCheck = false }
}

func applyOptions(opts []Option) options {
	o := options{maxBodyBytes: MaxBodyBytes, domainCheck: true}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// Parse reads and authenticates an Inbound. It requires the
// X-Cyberbiz-Event, X-Cyberbiz-Domain and X-Cyberbiz-Hmac-Sha256 headers,
// reads at most the configured body limit, resolves the App Secret for the
// Shop Domain, verifies the body Signature, and verifies the Domain
// Signature when its header is present.
//
// When secrets also implements [CredentialResolver], every candidate secret
// is tried: exactly one must verify the body, and its AppID is reported on
// the Event. No match is [ErrInvalidSignature]; several matches are
// [ErrAmbiguousSecret]; more than [MaxCredentials] candidates are
// [ErrTooManyCredentials]. The Domain Signature is verified with the
// matched secret.
//
// r.Body is replaced with a reader that yields the same bytes again, so
// middleware or handlers running after Parse can still read it. Parse never
// closes the original body.
func Parse(ctx context.Context, r *http.Request, secrets SecretResolver, opts ...Option) (*Event, error) {
	if secrets == nil {
		return nil, errors.New("webhook: nil SecretResolver")
	}
	o := applyOptions(opts)

	eventType, err := requireHeader(r, HeaderEvent)
	if err != nil {
		return nil, err
	}
	shopDomain, err := requireHeader(r, HeaderDomain)
	if err != nil {
		return nil, err
	}
	signature, err := requireHeader(r, HeaderSignature)
	if err != nil {
		return nil, err
	}

	body, err := readBody(r, o.maxBodyBytes)
	if err != nil {
		return nil, err
	}

	creds, err := resolveCredentials(ctx, secrets, shopDomain)
	if err != nil {
		return nil, err
	}
	cred, err := matchCredential(body, signature, creds)
	if err != nil {
		return nil, fmt.Errorf("%w for %s from %q", err, eventType, shopDomain)
	}
	domainSignature := r.Header.Get(HeaderDomainHMAC)
	if o.domainCheck && domainSignature != "" && !VerifyDomain(shopDomain, domainSignature, cred.Secret) {
		return nil, fmt.Errorf("%w for %q", ErrInvalidDomainSignature, shopDomain)
	}

	return &Event{
		Type:            EventType(eventType),
		AppID:           cred.AppID,
		ShopDomain:      shopDomain,
		CustomDomain:    r.Header.Get(HeaderShopDomain),
		Signature:       signature,
		DomainSignature: domainSignature,
		ReceivedAt:      time.Now(),
		Header:          r.Header.Clone(),
		Raw:             body,
	}, nil
}

func requireHeader(r *http.Request, name string) (string, error) {
	v := r.Header.Get(name)
	if v == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingHeader, name)
	}
	return v, nil
}

// readBody reads up to limit bytes and restores r.Body so the bytes can be
// read again. When the body is longer than limit, the bytes already read
// are still restored ahead of the unread remainder.
func readBody(r *http.Request, limit int64) ([]byte, error) {
	if r.Body == nil || r.Body == http.NoBody {
		return []byte{}, nil
	}
	original := r.Body
	body, err := io.ReadAll(io.LimitReader(original, limit+1))
	r.Body = restoredBody{
		Reader: io.MultiReader(bytes.NewReader(body), original),
		Closer: original,
	}
	if err != nil {
		return nil, fmt.Errorf("webhook: reading body: %w", err)
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%w: more than %d bytes", ErrBodyTooLarge, limit)
	}
	return body, nil
}

// restoredBody replays the bytes Parse consumed before the rest of the
// original body, and closes the original.
type restoredBody struct {
	io.Reader
	io.Closer
}

// Handler returns an http.Handler that parses each Inbound with [Parse] and
// passes the Event to fn. It responds 200 when fn returns nil, 401 for a
// bad signature or unknown Shop, 413 for an oversize body, 405 for a
// non-POST method, 500 for a configuration error ([ErrAmbiguousSecret],
// [ErrTooManyCredentials]), 400 for any other parse failure, and 500 when
// fn returns an error (which makes CYBERBIZ retry). The response body is only
// the status text; the request body is never echoed.
func Handler(secrets SecretResolver, fn func(ctx context.Context, e *Event) error, opts ...Option) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respond(w, http.StatusMethodNotAllowed)
			return
		}
		e, err := Parse(r.Context(), r, secrets, opts...)
		if err != nil {
			respond(w, statusFor(err))
			return
		}
		if err := fn(r.Context(), e); err != nil {
			respond(w, http.StatusInternalServerError)
			return
		}
		respond(w, http.StatusOK)
	})
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, ErrInvalidSignature),
		errors.Is(err, ErrInvalidDomainSignature),
		errors.Is(err, ErrUnknownShop):
		return http.StatusUnauthorized
	case errors.Is(err, ErrBodyTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, ErrAmbiguousSecret),
		errors.Is(err, ErrTooManyCredentials):
		return http.StatusInternalServerError
	}
	return http.StatusBadRequest
}

func respond(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, _ = io.WriteString(w, http.StatusText(code))
}
