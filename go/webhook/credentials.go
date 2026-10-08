package webhook

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// MaxCredentials is the most candidate App Secrets [Parse] accepts for one
// Shop Domain. Every candidate costs one HMAC over the body, so the cap
// bounds the CPU a single Inbound can consume.
const MaxCredentials = 16

// Configuration errors. They mean the receiver is set up wrongly, not that
// the request is bad, so [Handler] answers 500 and CYBERBIZ retries the
// delivery once the configuration is fixed.
var (
	// ErrAmbiguousSecret means the body Signature verified under more than
	// one candidate App Secret, i.e. two Apps of the Shop are configured with
	// the same secret. Parse never picks one of them.
	ErrAmbiguousSecret = errors.New("webhook: ambiguous app secret")
	// ErrTooManyCredentials means a [CredentialResolver] returned more than
	// [MaxCredentials] candidates for one Shop Domain.
	ErrTooManyCredentials = errors.New("webhook: too many app credentials")
)

// Credential is one App's webhook secret for a Shop. CYBERBIZ sends no App
// identifier header, so the App behind an Inbound is identified only by the
// Credential whose Secret verifies the body Signature.
type Credential struct {
	// AppID is your own identifier for the App, reported as [Event.AppID].
	AppID string
	// Secret is the App Secret. An empty Secret never verifies.
	Secret string
}

// String never shows the secret.
func (c Credential) String() string {
	return fmt.Sprintf("Credential{AppID:%q}", c.AppID)
}

// GoString never shows the secret, so %#v is safe to log too.
func (c Credential) GoString() string {
	return fmt.Sprintf("webhook.Credential{AppID:%q, Secret:\"***\"}", c.AppID)
}

// CredentialResolver returns every App Secret configured for a Shop Domain,
// for a receiver that serves several Apps installed on the same Shop.
//
// [Parse] accepts a [SecretResolver]; when the value also implements
// CredentialResolver, Parse calls WebhookCredentials instead of
// WebhookSecret. Use [CredentialResolverFunc] or [AppSecrets] to build one.
// Return an error, or no non-empty secret, for a Shop the integration does
// not know; Parse maps both to [ErrUnknownShop].
type CredentialResolver interface {
	WebhookCredentials(ctx context.Context, shopDomain string) ([]Credential, error)
}

// CredentialResolverFunc adapts a function to both [CredentialResolver] and
// [SecretResolver], so it can be passed to [Parse] and [Handler].
type CredentialResolverFunc func(ctx context.Context, shopDomain string) ([]Credential, error)

// WebhookCredentials calls f.
func (f CredentialResolverFunc) WebhookCredentials(ctx context.Context, shopDomain string) ([]Credential, error) {
	return f(ctx, shopDomain)
}

// WebhookSecret returns the secret when f yields exactly one Credential,
// and an error otherwise. Parse never calls it; it exists so the function
// satisfies [SecretResolver].
func (f CredentialResolverFunc) WebhookSecret(ctx context.Context, shopDomain string) (string, error) {
	creds, err := f(ctx, shopDomain)
	if err != nil {
		return "", err
	}
	if len(creds) != 1 {
		return "", fmt.Errorf("webhook: %d credentials for %q, want exactly one", len(creds), shopDomain)
	}
	return creds[0].Secret, nil
}

// AppSecrets returns a resolver for a fixed configuration: App Secrets keyed
// by Shop Domain, then by App ID. Shop Domains match ignoring case. The maps
// are copied.
//
//	webhook.AppSecrets(map[string]map[string]string{
//		"shop-a.cyberbiz.co": {"app-a": secretA, "app-b": secretB},
//	})
func AppSecrets(secrets map[string]map[string]string) SecretResolver {
	byShop := make(map[string][]Credential, len(secrets))
	for shop, apps := range secrets {
		creds := make([]Credential, 0, len(apps))
		for appID, secret := range apps {
			creds = append(creds, Credential{AppID: appID, Secret: secret})
		}
		sort.Slice(creds, func(i, j int) bool { return creds[i].AppID < creds[j].AppID })
		key := strings.ToLower(shop)
		byShop[key] = append(byShop[key], creds...)
	}
	return CredentialResolverFunc(func(_ context.Context, shopDomain string) ([]Credential, error) {
		creds, ok := byShop[strings.ToLower(shopDomain)]
		if !ok {
			return nil, errors.New("no app secrets configured")
		}
		return append([]Credential(nil), creds...), nil
	})
}

// resolveCredentials asks the resolver for the candidates of shopDomain. A
// plain [SecretResolver] yields one Credential with an empty AppID. Empty
// secrets are dropped; no candidate left means [ErrUnknownShop].
func resolveCredentials(ctx context.Context, secrets SecretResolver, shopDomain string) ([]Credential, error) {
	var creds []Credential
	if cr, ok := secrets.(CredentialResolver); ok {
		resolved, err := cr.WebhookCredentials(ctx, shopDomain)
		if err != nil {
			return nil, fmt.Errorf("%w %q: %w", ErrUnknownShop, shopDomain, err)
		}
		if len(resolved) > MaxCredentials {
			return nil, fmt.Errorf("%w: %d for %q, at most %d", ErrTooManyCredentials, len(resolved), shopDomain, MaxCredentials)
		}
		creds = resolved
	} else {
		secret, err := secrets.WebhookSecret(ctx, shopDomain)
		if err != nil {
			return nil, fmt.Errorf("%w %q: %w", ErrUnknownShop, shopDomain, err)
		}
		creds = []Credential{{Secret: secret}}
	}
	usable := make([]Credential, 0, len(creds))
	for _, c := range creds {
		if c.Secret != "" {
			usable = append(usable, c)
		}
	}
	if len(usable) == 0 {
		return nil, fmt.Errorf("%w %q: resolver returned no secret", ErrUnknownShop, shopDomain)
	}
	return usable, nil
}

// matchCredential returns the one candidate whose secret verifies the body
// Signature. Every candidate is checked, each in constant time, so the time
// taken does not depend on which one matched.
func matchCredential(body []byte, signature string, creds []Credential) (Credential, error) {
	var matched Credential
	matches := 0
	for _, c := range creds {
		if Verify(body, signature, c.Secret) {
			matched = c
			matches++
		}
	}
	switch matches {
	case 0:
		return Credential{}, ErrInvalidSignature
	case 1:
		return matched, nil
	}
	return Credential{}, fmt.Errorf("%w: %d apps share the secret that signed this body", ErrAmbiguousSecret, matches)
}
