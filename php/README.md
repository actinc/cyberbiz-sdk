# CYBERBIZ PHP SDK

PHP client for the [CYBERBIZ](https://www.cyberbiz.io) e-commerce platform
API. Versions before 1.0 may still change the API in a minor release.

Requires PHP 8.2 or later with `ext-bcmath`. HTTP goes through any PSR-18
client you already use (Guzzle, Symfony HttpClient, ...); the SDK does not
pin one.

## Install

```sh
composer require actinc/cyberbiz-sdk
```

## Usage

```php
use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Money;

$client = new Client(getenv('CYBERBIZ_API_TOKEN'));

try {
    $shop = $client->shop()->info();

    foreach ($client->products()->all() as $product) {
        echo $product->title, ' ', $product->price, PHP_EOL; // Money: exact decimal
    }

    $product = $client->products()->create([
        'title' => 'Synthetic Tea',
        'handle' => 'synthetic-tea',
        'published' => false,
        'price' => Money::of('120'),
    ]);
} catch (ApiException $e) {
    // $e->statusCode, $e->requestId and $e->messages describe the failure.
}
```

Resources so far: `shop()` (profile and app settings), `products()`
(products, variants, options, tags, shipping bindings), `orders()` (orders,
fulfillments and shipping labels, payments, returns, e-tickets) and
`customers()` (customers, their orders, cart, VIP state, login identities). `list()` returns one
`Page` with the pagination headers; `all()` walks every page. For an
endpoint without a wrapper yet, `$client->send(new Request(...))` returns the
raw response.

The client keeps to the platform limit of 5 requests per second, retries
429/502/503/504 (honouring `Retry-After`) up to 3 times, and throws a typed
exception per status: `AuthenticationException` (401),
`ForbiddenException` (403), `NotFoundException` (404),
`ValidationException` (422), `RateLimitException` (429) and
`ServerException` (5xx), all extending `ApiException`. Network failures
throw `TransportException`.

## Webhooks

```php
use Actinc\Cyberbiz\Exception\WebhookException;
use Actinc\Cyberbiz\Webhook\EventType;
use Actinc\Cyberbiz\Webhook\Parser;
use Actinc\Cyberbiz\Webhook\StaticSecret;

$parser = new Parser(new StaticSecret(getenv('CYBERBIZ_APP_SECRET')));

try {
    // Any framework: the raw body and $_SERVER-style or plain headers.
    $event = $parser->parseRaw(file_get_contents('php://input'), $_SERVER);
    // With PSR-7: $event = $parser->parse($serverRequest);
} catch (WebhookException $e) {
    http_response_code(401); // MissingHeader: 400, BodyTooLarge: 413
    exit;
}

if ($event->eventType() === EventType::OrdersPaid) {
    $order = $event->fields(); // typed reads, amounts exact
}
```

The signature (`X-Cyberbiz-Hmac-Sha256`) is verified as hex, which CYBERBIZ
sends, or base64, which its documentation describes, in constant time; the
Domain Signature is checked when present. Multi-shop integrations pass a
`SecretMap` or their own `SecretResolver`. CYBERBIZ retries deliveries, so
handlers must be idempotent.

## Development

The package's `composer.json` is at the repository root, so run Composer
from there:

```sh
composer install
composer check    # PHP-CS-Fixer (dry run), PHPStan, PHPUnit
composer cs-fix   # apply the coding style
```

Ground rules it shares with the other SDKs in this repository:

- Endpoints and webhook payloads come from `../docs/api/en/`.
- Model tests decode the shared Golden Files in `../testdata/golden/`; never
  copy them into this directory.
- Test data is synthetic: no real tokens, customers or shop names.
