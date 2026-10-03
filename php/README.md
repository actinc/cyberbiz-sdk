# CYBERBIZ PHP SDK

PHP client for the [CYBERBIZ](https://www.cyberbiz.io) e-commerce platform
API. **In development: not yet released**, so the API below may still change.

Requires PHP 8.2 or later with `ext-bcmath`. HTTP goes through any PSR-18
client you already use (Guzzle, Symfony HttpClient, ...); the SDK does not
pin one.

## Install

Once the first version is published:

```sh
composer require actinc/cyberbiz-sdk
```

## Usage

```php
use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Request;

$client = new Client(getenv('CYBERBIZ_API_TOKEN'));

try {
    $response = $client->send(new Request('GET', '/v1/products', ['page' => 1]));
    $products = json_decode($response->body, true);
} catch (ApiException $e) {
    // $e->statusCode, $e->requestId and $e->messages describe the failure.
}
```

The client keeps to the platform limit of 5 requests per second, retries
429/502/503/504 (honouring `Retry-After`) up to 3 times, and throws a typed
exception per status: `AuthenticationException` (401),
`ForbiddenException` (403), `NotFoundException` (404),
`ValidationException` (422), `RateLimitException` (429) and
`ServerException` (5xx), all extending `ApiException`. Network failures
throw `TransportException`.

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
