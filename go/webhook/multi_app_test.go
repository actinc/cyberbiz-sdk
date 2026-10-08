package webhook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Several Apps on one Shop (CBSDK-41): CYBERBIZ sends no App identifier, so
// the App is the one whose secret verifies the body Signature.

const (
	multiAppShop    = "shop-a.cyberbiz.co"
	multiAppSecretA = "app-a-test-secret"
	multiAppSecretB = "app-b-test-secret"
)

// twoApps resolves two Apps with different secrets for multiAppShop.
func twoApps() SecretResolver {
	return AppSecrets(map[string]map[string]string{
		multiAppShop: {"app-a": multiAppSecretA, "app-b": multiAppSecretB},
	})
}

// multiAppRequest builds a delivery for multiAppShop signed with secret,
// carrying domainSig as the Domain Signature when it is not empty.
func multiAppRequest(t *testing.T, secret, domainSig string) *http.Request {
	t.Helper()
	body := []byte(`{"id":1001,"name":"#1001"}`)
	r := httptest.NewRequest(http.MethodPost, "/webhooks/cyberbiz", bytes.NewReader(body))
	r.Header.Set(HeaderEvent, string(EventOrdersPaid))
	r.Header.Set(HeaderDomain, multiAppShop)
	r.Header.Set(HeaderSignature, Sign(body, secret))
	if domainSig != "" {
		r.Header.Set(HeaderDomainHMAC, domainSig)
	}
	return r
}

// AC1
func TestParseMultiAppIdentifiesTheApp(t *testing.T) {
	for _, tc := range []struct{ secret, want string }{
		{multiAppSecretA, "app-a"},
		{multiAppSecretB, "app-b"},
	} {
		e, err := Parse(context.Background(), multiAppRequest(t, tc.secret, ""), twoApps())
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if e.AppID != tc.want {
			t.Errorf("AppID = %q, want %q", e.AppID, tc.want)
		}
	}
}

// AC2
func TestParseMultiAppNoCandidateVerifies(t *testing.T) {
	e, err := Parse(context.Background(), multiAppRequest(t, "app-c-test-secret", ""), twoApps())
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("err = %v, want %v", err, ErrInvalidSignature)
	}
	if e != nil {
		t.Error("Parse returned an Event")
	}
	if got := statusFor(err); got != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", got)
	}
}

// AC3
func TestParseMultiAppSharedSecretIsAmbiguous(t *testing.T) {
	same := AppSecrets(map[string]map[string]string{
		multiAppShop: {"app-a": multiAppSecretA, "app-b": multiAppSecretA},
	})
	e, err := Parse(context.Background(), multiAppRequest(t, multiAppSecretA, ""), same)
	if !errors.Is(err, ErrAmbiguousSecret) {
		t.Fatalf("err = %v, want %v", err, ErrAmbiguousSecret)
	}
	if e != nil {
		t.Error("Parse returned an Event")
	}
	if got := statusFor(err); got != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", got)
	}
}

// AC5
func TestParseMultiAppTooManyCandidates(t *testing.T) {
	many := CredentialResolverFunc(func(context.Context, string) ([]Credential, error) {
		creds := make([]Credential, MaxCredentials+1)
		for i := range creds {
			creds[i] = Credential{AppID: fmt.Sprintf("app-%d", i), Secret: fmt.Sprintf("secret-%d", i)}
		}
		creds[0].Secret = multiAppSecretA
		return creds, nil
	})
	e, err := Parse(context.Background(), multiAppRequest(t, multiAppSecretA, ""), many)
	if !errors.Is(err, ErrTooManyCredentials) {
		t.Fatalf("err = %v, want %v", err, ErrTooManyCredentials)
	}
	if e != nil {
		t.Error("Parse returned an Event")
	}
	if got := statusFor(err); got != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", got)
	}
}

func TestParseMultiAppAcceptsTheCap(t *testing.T) {
	atCap := CredentialResolverFunc(func(context.Context, string) ([]Credential, error) {
		creds := make([]Credential, MaxCredentials)
		for i := range creds {
			creds[i] = Credential{AppID: fmt.Sprintf("app-%d", i), Secret: fmt.Sprintf("secret-%d", i)}
		}
		return creds, nil
	})
	e, err := Parse(context.Background(), multiAppRequest(t, "secret-15", ""), atCap)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.AppID != "app-15" {
		t.Errorf("AppID = %q, want app-15", e.AppID)
	}
}

func TestParseMultiAppDomainSignatureUsesTheMatchedSecret(t *testing.T) {
	ok := multiAppRequest(t, multiAppSecretB, SignDomain(multiAppShop, multiAppSecretB))
	e, err := Parse(context.Background(), ok, twoApps())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.AppID != "app-b" {
		t.Errorf("AppID = %q, want app-b", e.AppID)
	}

	other := multiAppRequest(t, multiAppSecretB, SignDomain(multiAppShop, multiAppSecretA))
	_, err = Parse(context.Background(), other, twoApps())
	if !errors.Is(err, ErrInvalidDomainSignature) {
		t.Fatalf("err = %v, want %v", err, ErrInvalidDomainSignature)
	}
}

func TestParseMultiAppFailsClosedWithoutSecrets(t *testing.T) {
	resolvers := map[string]SecretResolver{
		"no credentials": CredentialResolverFunc(func(context.Context, string) ([]Credential, error) {
			return nil, nil
		}),
		"only empty secrets": CredentialResolverFunc(func(context.Context, string) ([]Credential, error) {
			return []Credential{{AppID: "app-a"}, {AppID: "app-b"}}, nil
		}),
		"resolver error": CredentialResolverFunc(func(context.Context, string) ([]Credential, error) {
			return nil, errors.New("lookup failed")
		}),
		"unknown shop": AppSecrets(map[string]map[string]string{
			"shop-b.cyberbiz.co": {"app-a": multiAppSecretA},
		}),
	}
	for name, res := range resolvers {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(context.Background(), multiAppRequest(t, multiAppSecretA, ""), res)
			if !errors.Is(err, ErrUnknownShop) {
				t.Fatalf("err = %v, want %v", err, ErrUnknownShop)
			}
		})
	}
}

func TestParseMultiAppIgnoresEmptySecrets(t *testing.T) {
	res := CredentialResolverFunc(func(context.Context, string) ([]Credential, error) {
		return []Credential{{AppID: "app-a"}, {AppID: "app-b", Secret: multiAppSecretB}}, nil
	})
	e, err := Parse(context.Background(), multiAppRequest(t, multiAppSecretB, ""), res)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.AppID != "app-b" {
		t.Errorf("AppID = %q, want app-b", e.AppID)
	}
}

func TestStaticSecretHasNoAppID(t *testing.T) {
	e, err := Parse(context.Background(), multiAppRequest(t, multiAppSecretA, ""), StaticSecret(multiAppSecretA))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.AppID != "" {
		t.Errorf("AppID = %q, want empty", e.AppID)
	}
}

func TestHandlerMultiApp(t *testing.T) {
	var got string
	h := Handler(twoApps(), func(_ context.Context, e *Event) error {
		got = e.AppID
		return nil
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, multiAppRequest(t, multiAppSecretB, ""))
	if w.Code != http.StatusOK || got != "app-b" {
		t.Errorf("status = %d, AppID = %q; want 200, app-b", w.Code, got)
	}
}

func TestHandlerAmbiguousSecretIs500(t *testing.T) {
	same := AppSecrets(map[string]map[string]string{
		multiAppShop: {"app-a": multiAppSecretA, "app-b": multiAppSecretA},
	})
	h := Handler(same, func(context.Context, *Event) error {
		t.Error("handler ran for an ambiguous secret")
		return nil
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, multiAppRequest(t, multiAppSecretA, ""))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestCredentialResolverFuncWebhookSecret(t *testing.T) {
	one := CredentialResolverFunc(func(context.Context, string) ([]Credential, error) {
		return []Credential{{AppID: "app-a", Secret: multiAppSecretA}}, nil
	})
	if s, err := one.WebhookSecret(context.Background(), multiAppShop); err != nil || s != multiAppSecretA {
		t.Errorf("WebhookSecret = %q, %v; want the single secret", s, err)
	}
	two := twoApps().(CredentialResolverFunc)
	if _, err := two.WebhookSecret(context.Background(), multiAppShop); err == nil {
		t.Error("WebhookSecret with two credentials returned no error")
	}
}

func TestCredentialStringHidesTheSecret(t *testing.T) {
	c := Credential{AppID: "app-a", Secret: multiAppSecretA}
	for _, s := range []string{fmt.Sprint(c), fmt.Sprintf("%+v", c), fmt.Sprintf("%#v", c)} {
		if bytes.Contains([]byte(s), []byte(multiAppSecretA)) {
			t.Errorf("formatted Credential %q shows the secret", s)
		}
	}
}
