# CYBERBIZ Java SDK

Java client for the [CYBERBIZ](https://www.cyberbiz.io) e-commerce platform
API. In development: the version is `0.1.0-SNAPSHOT`, nothing is published
to Maven Central yet, and versions before 1.0 may still change the API in a
minor release.

Requires Java 17 or later. HTTP uses the JDK's `java.net.http.HttpClient`;
the only dependency is Gson.

## Install

Once published:

```xml
<dependency>
  <groupId>cc.alphacore</groupId>
  <artifactId>cyberbiz-sdk</artifactId>
  <version>0.1.0</version>
</dependency>
```

```kotlin
implementation("cc.alphacore:cyberbiz-sdk:0.1.0")
```

## Usage

```java
import cc.alphacore.cyberbiz.CyberbizClient;
import cc.alphacore.cyberbiz.Request;
import cc.alphacore.cyberbiz.Response;
import cc.alphacore.cyberbiz.exception.ApiException;

CyberbizClient client = CyberbizClient.builder(System.getenv("CYBERBIZ_API_TOKEN")).build();

try {
  Response shop = client.send(Request.of("GET", "/v1/shop"));
  System.out.println(shop.body());
} catch (ApiException e) {
  // e.statusCode(), e.requestId() and e.messages() describe the failure.
}
```

One client per Shop token; it is immutable and safe to share between
threads. Typed resources (shop, products, orders, customers) and webhook
verification are being added; until then `send(Request)` calls any endpoint
and returns the raw response.

The client keeps to the platform limit of 5 requests per second (across
threads) and retries up to 3 times: 429 for every method (honouring
`Retry-After`), and 502/503/504 and network errors only for idempotent
methods, so a POST is never sent twice after the server may have acted on
it. Errors become a typed, unchecked
exception per status: `AuthenticationException` (401), `ForbiddenException`
(403), `NotFoundException` (404), `ValidationException` (422),
`RateLimitException` (429) and `ServerException` (5xx), all extending
`ApiException`. Network failures throw `TransportException`. Messages never
include the token.

HTTP goes through `java.net.http.HttpClient` by default; pass your own
`Transport` to the builder to use another HTTP library or a proxy.

### Pagination

List endpoints return one page at a time. `list` decodes one page into a
`Page<T>` with the pagination headers; `all` walks every page lazily as an
`Iterable` or a `Stream`:

```java
import cc.alphacore.cyberbiz.pagination.Page;
import com.google.gson.JsonObject;

Request products = Request.of("GET", "/v1/products");

Page<JsonObject> page = client.list(products.withQuery("page", 2), JsonObject.class);
page.items();                     // the decoded items
page.pagination().total();        // X-Total; also page, perPage, totalPages, nextPage...
page.hasNext();

for (JsonObject product : client.all(products, JsonObject.class)) {
  // every product, page after page
}
long count = client.all(products, JsonObject.class).stream().count();
```

`all` starts at the request's `page` (default 1) with `per_page` 50 (the
platform maximum) unless the request sets it, follows `X-Next-Page`, and
stops at the last page or an empty page: exactly one request per page, and
nothing is sent until iteration starts. An empty body or `null` counts as an
empty page.

### JSON types

Items decode with one shared Gson configuration (`cc.alphacore.cyberbiz.json.Json`),
usable for your own records too: camelCase record components read the API's
snake_case fields, unknown fields are ignored, amounts are exact `BigDecimal`
(from JSON numbers or numeric strings, never through `double`; plain
notation of at most 64 characters, 40 digits and 32 decimal places, so a
hostile `1e999999999` fails instead of exhausting memory), and
timestamps are `OffsetDateTime` in Asia/Taipei, with or without a zone in the
input (`Times` in the same package parses and formats the platform layout,
`2026-07-10 20:22:05`). A body that does not fit throws `DecodeException`.

An amount keeps the exact scale the API sent, so `200.0` arrives with one
decimal place and `200.00` with two. Compare amounts with
`compareTo(other) == 0`, not `equals`, which also compares the scale. The
PHP and Go SDKs normalise to two decimal places instead.

## Webhooks

`WebhookParser` authenticates CYBERBIZ App webhooks without tying you to a
web framework: pass the request headers and the raw body bytes, before
anything parses the JSON.

```java
import cc.alphacore.cyberbiz.exception.WebhookException;
import cc.alphacore.cyberbiz.webhook.Event;
import cc.alphacore.cyberbiz.webhook.EventType;
import cc.alphacore.cyberbiz.webhook.StaticSecret;
import cc.alphacore.cyberbiz.webhook.WebhookParser;

WebhookParser webhooks = new WebhookParser(new StaticSecret(System.getenv("CYBERBIZ_APP_SECRET")));

// headers: Map<String, String> or Map<String, List<String>>; body: byte[]
try {
  Event event = webhooks.parse(headers, body);
  if (event.eventType().orElse(null) == EventType.ORDERS_PAID) {
    long orderId = event.payload().get("id").getAsLong();
  }
  // answer 200
} catch (WebhookException e) {
  // answer e.httpStatus() (400, 401 or 413) and ignore the body
}
```

The parser requires `X-Cyberbiz-Event`, `X-Cyberbiz-Domain` and
`X-Cyberbiz-Hmac-Sha256` (`MissingHeaderException`), rejects bodies over
2 MiB (`BodyTooLargeException`; configurable), asks the `SecretResolver` for
the Shop's App Secret (`UnknownShopException` when there is none) and
verifies the HMAC-SHA256 signature in constant time, as hex or base64
(`InvalidSignatureException`). `X-Cyberbiz-Domain-Hmac-Sha256` is checked
too when present (`InvalidDomainSignatureException`). All of them extend
`WebhookException`.

For several Shops, use `new SecretMap(Map.of("example.cyberbiz.co", secret, ...))`
or a lambda `shopDomain -> Optional.ofNullable(lookup(shopDomain))`.
`event.decode(MyPayload.class)` decodes the body into your own class with
Gson. Payloads and signed samples of every Event are in
[`docs/api/en/webhooks.md`](../docs/api/en/webhooks.md).


## Development

```sh
cd java
./mvnw verify           # compile (-Xlint:all -Werror, Error Prone), JUnit, format check
./mvnw spotless:apply   # fix formatting
```

`./mvnw` downloads the Maven version CI uses; an installed `mvn` 3.9 works
too. Builds and tests run on JDK 17, the lowest supported version. Model
tests read the shared Golden Files in `../testdata/golden/` in place
(`Golden` in the tests); register a new model with one line in
`GoldenTest.MODELS`.
