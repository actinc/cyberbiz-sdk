package webhook_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/actinc/cyberbiz-sdk/go/webhook"
)

// ExampleHandler mounts a webhook endpoint that dispatches on the event
// type. In production the secret comes from configuration and the handler
// is registered on a real server; here a recorded-style request is sent
// through httptest so the example runs offline.
func ExampleHandler() {
	const secret = "app-secret-from-config"
	secrets := webhook.StaticSecret(secret)

	handler := webhook.Handler(secrets, func(ctx context.Context, e *webhook.Event) error {
		switch e.Type {
		case webhook.EventBonusPointsCreate, webhook.EventBonusPointsUpdate:
			bp, err := e.BonusPoint()
			if err != nil {
				return err
			}
			fmt.Printf("%s: customer %d now has %.0f unused points\n", e.Type, bp.CustomerID, bp.UnusedPoints)
		case webhook.EventOrdersPaid:
			order, err := e.Order()
			if err != nil {
				return err
			}
			fmt.Printf("%s: order %s paid\n", e.Type, order.OrderName)
		default:
			// Unknown or unhandled events still get a 200 so CYBERBIZ does
			// not retry them.
			fmt.Printf("%s: ignored\n", e.Type)
		}
		return nil
	})
	mux := http.NewServeMux()
	mux.Handle("POST /webhooks/cyberbiz", handler)

	// Simulate one delivery from CYBERBIZ.
	body := []byte(`{"id":100001,"title":"Welcome bonus","points":100.0,"unused_points":100.0,` +
		`"consumption_price":0.0,"deadline":"2027-01-09 22:00:37","customer_id":900001,` +
		`"source":"signup","order_id":null}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/cyberbiz", bytes.NewReader(body))
	req.Header.Set("User-Agent", webhook.UserAgent)
	req.Header.Set(webhook.HeaderDomain, "example.cyberbiz.co")
	req.Header.Set(webhook.HeaderShopDomain, "www.example-shop.com")
	req.Header.Set(webhook.HeaderEvent, string(webhook.EventBonusPointsCreate))
	req.Header.Set(webhook.HeaderSignature, webhook.Sign(body, secret))
	req.Header.Set(webhook.HeaderDomainHMAC, webhook.SignDomain("example.cyberbiz.co", secret))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	fmt.Println("status:", rec.Code)

	// Output:
	// bonus_points/create: customer 900001 now has 100 unused points
	// status: 200
}

// ExampleParse shows the lower-level entry point for callers who own the
// http.Handler, for example to store every Inbound before acting on it.
func ExampleParse() {
	const secret = "app-secret-from-config"
	secrets := webhook.SecretResolverFunc(func(ctx context.Context, shopDomain string) (string, error) {
		// Look the App Secret up by Shop Domain; one integration can serve
		// many shops.
		if shopDomain == "example.cyberbiz.co" {
			return secret, nil
		}
		return "", fmt.Errorf("no credentials for %s", shopDomain)
	})

	body := []byte(`{"app_uuid":"00000000-0000-4000-8000-000000000001","app_version_uuid":"00000000-0000-4000-8000-000000000002","app_client_id":"sample-client-id"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/cyberbiz", bytes.NewReader(body))
	req.Header.Set(webhook.HeaderDomain, "example.cyberbiz.co")
	req.Header.Set(webhook.HeaderEvent, string(webhook.EventAppsUninstall))
	req.Header.Set(webhook.HeaderSignature, webhook.Sign(body, secret))

	e, err := webhook.Parse(req.Context(), req, secrets)
	if err != nil {
		fmt.Println("rejected:", err)
		return
	}
	app, err := e.App()
	if err != nil {
		fmt.Println("decode:", err)
		return
	}
	fmt.Printf("%s from %s: client %s\n", e.Type, e.ShopDomain, app.AppClientID)

	// Output:
	// apps/uninstall from example.cyberbiz.co: client sample-client-id
}

// ExampleAppSecrets serves several Apps installed on the same Shop from one
// endpoint. CYBERBIZ sends no App identifier, so the App is the one whose
// secret verifies the body; Event.AppID reports it.
func ExampleAppSecrets() {
	secrets := webhook.AppSecrets(map[string]map[string]string{
		"shop-a.cyberbiz.co": {
			"app-a": "app-a-secret-from-config",
			"app-b": "app-b-secret-from-config",
		},
	})

	handler := webhook.Handler(secrets, func(ctx context.Context, e *webhook.Event) error {
		switch e.AppID {
		case "app-a":
			fmt.Printf("app-a handles %s\n", e.Type)
		case "app-b":
			fmt.Printf("app-b handles %s\n", e.Type)
		}
		return nil
	})

	// Simulate one delivery signed by App B.
	body := []byte(`{"id":1001}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/cyberbiz", bytes.NewReader(body))
	req.Header.Set(webhook.HeaderDomain, "shop-a.cyberbiz.co")
	req.Header.Set(webhook.HeaderEvent, string(webhook.EventOrdersPaid))
	req.Header.Set(webhook.HeaderSignature, webhook.Sign(body, "app-b-secret-from-config"))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	fmt.Println("status:", rec.Code)

	// Output:
	// app-b handles orders/paid
	// status: 200
}
