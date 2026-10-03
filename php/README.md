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

$client = new Client(getenv('CYBERBIZ_API_TOKEN'));
```

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
