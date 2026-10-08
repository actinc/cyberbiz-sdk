# CYBERBIZ Java SDK

Java client for the [CYBERBIZ](https://www.cyberbiz.io) e-commerce platform
API, published on Maven Central as `cc.alphacore:cyberbiz-sdk`. Versions
before 1.0 may still change the API in a minor release.

Requires Java 17 or later. HTTP uses the JDK's `java.net.http.HttpClient`;
the only dependency is Gson.

## Install

From Maven Central, once `0.1.0` is published (releases are signed; the
sources and Javadoc jars are published alongside):

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
threads. Typed resources return records from `cc.alphacore.cyberbiz.model`;
`send(Request)` calls any endpoint without a wrapper and returns the raw
response.

### Shop and products

```java
ShopInfo shop = client.shop().info();               // GET /shop
AppSettings app = client.shop().settings();         // GET /settings
client.shop().updateSettings(Map.of("greeting", "hello"));

Page<Product> page = client.products().list(Map.of("page", 2));
for (Product p : client.products().all()) { ... }    // every page, lazily
Product tea = client.products().get(42);
Product created = client.products().create(Map.of(
    "title", "Green Tea", "handle", "green-tea", "published", true,
    "price", Money.of("120.50")));               // sent as 120.50, exactly
client.products().update(42, Map.of("title", "Oolong Tea"));
client.products().delete(42);
```

`client.products()` also covers search (`search`, `searchCollection`),
variants (`listVariants`, `getVariant`, `createVariant`, `updateVariant`,
`deleteVariant`, `listVariantsBySku`, `allVariantsBySku`), options
(`listOptions`, `getOption`, `createOption`, `updateOption`, `deleteOption`),
tags (`listTags`, `addTags`, `removeTags`), shipping bindings
(`listBindableShippings`, `getBindShippings`, `bindShippings`), SEO fields
(`updateSeoMetaTags`), `createForPosShops` and `listDescriptionSettingNames`.
Request bodies and query parameters are maps shaped like the API's JSON
(`docs/api/en/cyberbiz-openapi-v1.yaml`). A read by id whose response is
`null` throws `NotFoundException`; reads and updates by id set the record's
`id`, which the API leaves out of detail responses.

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

`Response` keeps the body as the raw bytes the server sent (`bytes()`, a
copy) next to its UTF-8 text (`body()`).

> **Custom transports**: `new Response(status, headers, text)` still compiles
> and behaves as before, so existing `Transport` implementations keep working
> for JSON. To pass binary replies such as label zips through intact, read the
> body as bytes and return `Response.ofBytes(status, headers, bytes)` instead.
> `Response` is now a final class rather than a record; its accessors
> (`statusCode()`, `headers()`, `body()`), `equals` and `hashCode` are kept,
> and `toString()` no longer prints the body.

### Orders and customers

`client.orders()` and `client.customers()` wrap the same endpoints as the
PHP SDK's `orders()` and `customers()` and decode into the records in
`cc.alphacore.cyberbiz.model` (`Order`, `Fulfillment`, `Customer`, ...).
Queries and bodies are maps shaped like the API's JSON (a `LinkedHashMap`
keeps their order); a collection is sent comma-separated in a query and an
`OffsetDateTime` as a timestamp in Asia/Taipei.

```java
import cc.alphacore.cyberbiz.model.Order;
import java.util.Map;

Page<Order> open = client.orders().list(Map.of("statuses", List.of("open"), "per_page", 50));
Order order = client.orders().get(56943817);   // NotFoundException when missing
client.orders().updateTags(order.id(), List.of("vip"));
for (Order o : client.customers().allOrders(42)) { ... }
```

Writes that answer with the changed order (`updateStatus`, `markPreparing`,
shipping bookings, ...) return `Optional`, empty when the platform sends no
body or, for a booking, a 202 while the carrier is still assigning a number.

`printCvsShippingLabels` and `printSupportShippingLabels` return the raw
`Response`, its body a zip archive, as the PHP SDK does. Read the archive with
`bytes()`, never `body()` (the UTF-8 text view, for JSON):

```java
Response labels = client.orders().printCvsShippingLabels(
    Map.of("shipping_type", "seven", "fulfillment_ids", List.of(101, 102)));
Files.write(Path.of(labels.filename().orElse("labels.zip")), labels.bytes());
labels.contentType();   // e.g. "application/zip", or "" when not sent
```

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

Model amounts are `cc.alphacore.cyberbiz.Money`, as in the PHP and Go SDKs:
an exact value with two decimal places (more round half away from zero), so
`Money.of("200.0").equals(Money.of("200"))` holds and `200.0` prints as
`200.00`. `amount()` gives the `BigDecimal`; `add`, `subtract`, `compareTo`
and `isNegative` cover the usual arithmetic. Unlike PHP, exponent notation
(`1e3`) is rejected, for the same safety reason as above.

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
  // answer e.httpStatus() (400, 401, 413, or 500 for a configuration error) and ignore the body
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

### Several Apps on one Shop

One receiver can serve several Apps installed on the same Shop. CYBERBIZ
sends no App identifier header, so the App is the one whose secret verifies
the body; `event.appId()` names it.

```java
import cc.alphacore.cyberbiz.webhook.AppSecrets;

WebhookParser webhooks = new WebhookParser(new AppSecrets(Map.of(
    "shop-a.cyberbiz.co", Map.of(
        "app-a", System.getenv("CYBERBIZ_APP_A_SECRET"),
        "app-b", System.getenv("CYBERBIZ_APP_B_SECRET")))));

Event event = webhooks.parse(headers, body);
switch (event.appId()) {
  case "app-a" -> handleAppA(event);
  case "app-b" -> handleAppB(event);
  default -> { }
}
```

To look the secrets up in your own store, use
`SecretResolver.ofCredentials(shopDomain -> List.of(new Credential("app-a", secretA), ...))`
or override `SecretResolver.credentialsFor`. Every candidate is checked in
constant time and exactly one must verify: two Apps sharing a secret throw
`AmbiguousSecretException`, more than `WebhookParser.MAX_CREDENTIALS` (16)
candidates `TooManyCredentialsException`. Both are `WebhookException`s with
`httpStatus()` 500, because the receiver is misconfigured. The Domain
Signature, when present, is verified with the matched secret. `StaticSecret`,
`SecretMap` and `secretFor` lambdas keep working and leave `appId()` empty.


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
