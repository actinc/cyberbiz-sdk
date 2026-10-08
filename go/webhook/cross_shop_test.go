package webhook

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Cross-shop replay (CBSDK-39): shop A's signed delivery resent with shop
// B's X-Cyberbiz-Domain. CYBERBIZ gives every shop its own App Secret, so
// the resolver returns B's secret and the body Signature, which A signed,
// fails. The Domain Signature stays optional (checked only when present).

const (
	crossShopA       = "shop-a.cyberbiz.co"
	crossShopB       = "shop-b.cyberbiz.co"
	crossShopSecretA = "shop-a-test-secret"
	crossShopSecretB = "shop-b-test-secret"
)

// crossShopSecrets resolves a different App Secret for each shop.
func crossShopSecrets() SecretResolver {
	return SecretResolverFunc(func(_ context.Context, shop string) (string, error) {
		switch shop {
		case crossShopA:
			return crossShopSecretA, nil
		case crossShopB:
			return crossShopSecretB, nil
		}
		return "", errors.New("unknown shop")
	})
}

// crossShopRequest builds shop A's delivery (body and body Signature under
// A's secret) labelled with domain, carrying domainSig as the Domain
// Signature when it is not empty.
func crossShopRequest(t *testing.T, domain, domainSig string) *http.Request {
	t.Helper()
	body := []byte(`{"id":1001,"name":"#1001"}`)
	r := httptest.NewRequest(http.MethodPost, "/webhooks/cyberbiz", bytes.NewReader(body))
	r.Header.Set(HeaderEvent, string(EventOrdersPaid))
	r.Header.Set(HeaderDomain, domain)
	r.Header.Set(HeaderSignature, Sign(body, crossShopSecretA))
	if domainSig != "" {
		r.Header.Set(HeaderDomainHMAC, domainSig)
	}
	return r
}

func TestParseCrossShopReplay(t *testing.T) {
	cases := []struct {
		name      string
		domainSig string
	}{
		{"no domain signature", ""},
		{"B's domain signature", SignDomain(crossShopB, crossShopSecretB)},
		{"A's domain signature", SignDomain(crossShopA, crossShopSecretA)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, err := Parse(context.Background(), crossShopRequest(t, crossShopB, tc.domainSig), crossShopSecrets())
			if !errors.Is(err, ErrInvalidSignature) {
				t.Fatalf("err = %v, want %v", err, ErrInvalidSignature)
			}
			if e != nil {
				t.Error("Parse returned an Event")
			}
			if got := statusFor(err); got != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", got)
			}
		})
	}
}

func TestParseCrossShopControl(t *testing.T) {
	r := crossShopRequest(t, crossShopA, SignDomain(crossShopA, crossShopSecretA))
	e, err := Parse(context.Background(), r, crossShopSecrets())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.ShopDomain != crossShopA {
		t.Errorf("ShopDomain = %q, want %q", e.ShopDomain, crossShopA)
	}
}

func TestHandlerCrossShopReplay(t *testing.T) {
	h := Handler(crossShopSecrets(), func(context.Context, *Event) error {
		t.Error("handler ran for a cross-shop replay")
		return nil
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, crossShopRequest(t, crossShopB, SignDomain(crossShopB, crossShopSecretB)))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}
