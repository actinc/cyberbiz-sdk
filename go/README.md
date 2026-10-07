# CYBERBIZ Go SDK

Go client for the [CYBERBIZ](https://www.cyberbiz.io) e-commerce platform API,
plus a local **Console** for exercising the API and inspecting webhooks.

- Covers the whole API: every `/v1` resource, the `/v2` additions, and the app
  endpoints (`/shop`, `/settings`), all through one client.
- Zero third-party dependencies (standard library plus `golang.org/x`).
- Built-in rate limiting (5 req/s), retries with `Retry-After`, typed errors,
  auto-paging iterators, exact money and correctly-zoned time types.
- Webhook parsing and signature verification for every documented event.
- Models are verified against real API responses (redacted Golden Files in
  `testdata/golden` at the repo root), not just the published documentation.

Requires Go 1.27 or later.

## Install

```sh
go get github.com/actinc/cyberbiz-sdk/go
```

Releases are tagged `go/vX.Y.Z`, because the module lives in the `go/`
subdirectory.

## Quick start

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

func main() {
	client, err := cyberbiz.New(os.Getenv("CYBERBIZ_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// One page.
	page, err := client.Orders.List(ctx, &cyberbiz.OrderListOptions{
		ListOptions: cyberbiz.ListOptions{PerPage: 50},
		Statuses:    []cyberbiz.OrderStatus{cyberbiz.OrderStatusOpen},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("open orders:", page.Pagination.Total)

	// Every page, lazily.
	for order, err := range client.Orders.All(ctx, nil) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(order.OrderNumber, order.Prices.TotalPrice, order.CreatedAt)
	}

	// Typed errors.
	_, _, err = client.Orders.Get(ctx, 1)
	if errors.Is(err, cyberbiz.ErrNotFound) {
		fmt.Println("no such order")
	}
	var apiErr *cyberbiz.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Messages, apiErr.RequestID)
	}
}
```

## Configuration

```go
client, err := cyberbiz.New(token,
	cyberbiz.WithRateLimit(5),                 // requests per second; 0 disables
	cyberbiz.WithMaxRetries(3),                // see below; 0 disables
	cyberbiz.WithBackoff(cyberbiz.ExponentialBackoff(500*time.Millisecond, 8*time.Second)),
	cyberbiz.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
	cyberbiz.WithTransport(myRoundTripper),   // logging, tracing, mocking
	cyberbiz.WithLogger(slog.Default()),       // requests at Debug, retries at Warn
	cyberbiz.WithStrictJSON(),                 // reject invalid UTF-8 / duplicate keys
	cyberbiz.WithBaseURL("http://localhost:8080/"), // tests and proxies
)
```

Retries: a 429 is retried for every method (honouring `Retry-After`); 502,
503, 504 and network errors are retried only for idempotent methods (GET,
HEAD, PUT, DELETE, OPTIONS). POST and PATCH are retried only on 429, so a
write is never sent twice after the server may have acted on it.

A client belongs to exactly one shop (one token). For several shops, create
several clients.

## Values you should know about

| Type | Why |
|---|---|
| `cyberbiz.Time` | CYBERBIZ sends `"2026-07-10 20:22:05"` with no zone. `Time` parses it in Asia/Taipei (fixed +08:00, no tzdata needed), also accepts ISO 8601, and marshals back to the platform format. |
| `cyberbiz.Date` | Date-only fields such as birthdays. |
| `cyberbiz.Money` | Amounts in hundredths (`Money(19950)` is 199.50), exact under addition, decoded from JSON numbers or strings. `String()` prints `199.5`, `Float64()` when you need one. |
| `cyberbiz.Page[T]` | One page of a list plus the `X-Total`, `X-Total-Pages`, `X-Next-Page` headers as `Pagination`. |
| `*cyberbiz.APIError` | Every non-2xx (and the platform's "200 with an error body" quirk). Matches `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrValidation`, `ErrRateLimited`, `ErrServer` via `errors.Is`. Note that CYBERBIZ answers 401 both for a bad token and for a feature the shop has not licensed; read `Messages`. |

Response models use value types: a JSON `null` decodes to the zero value.
Timestamps whose absence is meaningful (`ClosedAt`, `CancelledAt`,
`ConfirmedAt`, ...) are `*cyberbiz.Time`. Request structs use pointers and
`omitzero` so you can send an explicit empty string or `false`.

## Any endpoint

Typed methods exist for every documented operation. If you need something
they do not cover, `Do` goes through the same auth, rate limiting and retries:

```go
var out []map[string]any
resp, err := client.Do(ctx, &cyberbiz.Request{
	Method: http.MethodGet,
	Path:   "v1/orders",
	Query:  url.Values{"per_page": {"10"}},
}, &out)
```

## Webhooks

```go
import "github.com/actinc/cyberbiz-sdk/go/webhook"

http.Handle("/webhooks/cyberbiz", webhook.Handler(
	webhook.StaticSecret(os.Getenv("CYBERBIZ_APP_SECRET")),
	func(ctx context.Context, e *webhook.Event) error {
		switch e.Type {
		case webhook.EventOrdersPaid:
			order, err := e.Order()
			if err != nil {
				return err
			}
			return markPaid(ctx, order.ID)
		}
		return nil
	},
))
```

`webhook.Parse` verifies `X-Cyberbiz-Hmac-Sha256` (HMAC-SHA256 of the raw
body keyed by the App Secret). The platform documents base64 but sends hex;
both are accepted. For several shops, implement
`webhook.SecretResolver` to look the secret up by the `X-Cyberbiz-Domain`
header. CYBERBIZ retries deliveries and can send several events for one
customer within the same second, so handlers must be idempotent.

To exercise your receiver without waiting for real traffic, import
`docs/api/en/cyberbiz-webhooks.postman_collection.json` into Postman: it has
one request per event with the full sample payload, and a pre-request script
that signs the body with your App Secret exactly the way CYBERBIZ does.

## Console

`console/` is a separate Go module (so the SDK stays dependency-free) with a
single-binary web app: an API tester for every operation, a webhook receiver,
searchable logs of everything sent and received, and one-click export of a
real response as a redacted Golden File. See [console/README.md](console/README.md).

## Documentation

- `docs/api/en/` (repo root, shared by every language SDK): OpenAPI 3.1
  specs, Postman collections (API v1, API v2, and a webhook test collection)
  and the webhook reference, corrected against real responses. `docs/api/zh-TW/` is the
  Traditional Chinese edition.
- `CONTEXT.md` (repo root): the vocabulary used throughout (Shop, Outbound, Inbound,
  Event, Signature, Golden File, Sample, Console).

## Development

Run these from `go/`:

```sh
make test          # unit tests, SDK and Console
make lint          # gofmt + go vet
make integration   # read-only calls against the live API (CYBERBIZ_API_TOKEN)
make console       # run the Console
```

Unit tests never touch the network; they decode the Golden Files and drive
the client against `httptest` servers. To hack on the SDK and the Console
together, copy `go.work.example` to `go.work`.

## Status

v0.x: the API may still change between minor versions. It becomes v1.0 once
the first production integrations have migrated.

## License

MIT, see [LICENSE](../LICENSE).
