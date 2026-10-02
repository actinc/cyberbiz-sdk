// Package cyberbiz is a Go client for the CYBERBIZ e-commerce platform API.
//
// The CYBERBIZ API is a single HTTP API served at
// https://app-store-api.cyberbiz.io and authenticated with a Bearer token that
// belongs to one shop. API versions are path prefixes (/v1, /v2) plus a few
// unprefixed app endpoints (/shop, /settings); this package exposes all of
// them through one Client whose methods are grouped by resource:
//
//	client, err := cyberbiz.New(os.Getenv("CYBERBIZ_API_TOKEN"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	page, err := client.Orders.List(ctx, &cyberbiz.OrderListOptions{
//		ListOptions: cyberbiz.ListOptions{PerPage: 50},
//	})
//
// Every list method returns a [Page] carrying the items of one page plus the
// pagination headers CYBERBIZ sends. The matching All method returns an
// iterator that walks every page for you:
//
//	for order, err := range client.Orders.All(ctx, nil) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(order.OrderNumber)
//	}
//
// The client enforces the platform's 5 requests/second limit and retries
// rate-limited and transient server failures with exponential backoff; both
// behaviours are configurable through [Option] values passed to [New].
//
// Errors returned by the API are *[APIError] values that also match the
// sentinel errors [ErrNotFound], [ErrUnauthorized], [ErrForbidden],
// [ErrValidation] and [ErrRateLimited] through errors.Is.
//
// Timestamps are decoded into [Time], which parses the platform's
// zone-less "2006-01-02 15:04:05" format in Asia/Taipei; monetary amounts
// are decoded into [Money], an exact fixed-point type.
//
// Inbound webhooks are handled by the sibling package webhook.
package cyberbiz
