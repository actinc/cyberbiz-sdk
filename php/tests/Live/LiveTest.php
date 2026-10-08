<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Live;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\CyberbizException;
use Http\Discovery\Psr18ClientDiscovery;
use PHPUnit\Framework\Attributes\Group;
use PHPUnit\Framework\TestCase;

/**
 * Read-only calls against the live CYBERBIZ API with the token in
 * CYBERBIZ_API_TOKEN, the PHP side of Go's TestLive*. The "live" group is
 * excluded by php/phpunit.xml.dist; the main pipeline runs it with
 * `composer test -- --group live`, and the tests skip without a token.
 *
 * Only GET methods are called, and ReadOnlyHttpClient refuses anything else.
 * The tests check that each call succeeds and decodes into SDK types; they
 * never compare or print values, and a failure reports only the exception
 * type, status and request id, so no shop data reaches the log.
 */
#[Group('live')]
final class LiveTest extends TestCase
{
    private Client $client;

    protected function setUp(): void
    {
        $token = getenv('CYBERBIZ_API_TOKEN');
        if (!\is_string($token) || $token === '') {
            self::markTestSkipped('CYBERBIZ_API_TOKEN not set');
        }
        $this->client = new Client($token, httpClient: new ReadOnlyHttpClient(Psr18ClientDiscovery::find()));
    }

    public function testShopInfo(): void
    {
        $info = self::call('GET shop', fn() => $this->client->shop()->info());

        self::assertGreaterThan(0, $info->id, 'shop id is not positive');
        self::assertNotSame('', $info->primaryDomain, 'empty primary domain');
    }

    public function testAppSettings(): void
    {
        $settings = self::call('GET settings', fn() => $this->client->shop()->settings());

        self::assertGreaterThan(0, $settings->id, 'settings id is not positive');
        self::assertNotNull($settings->addOnVersion, 'no add_on_version');
    }

    public function testProductsFirstPage(): void
    {
        $page = self::call('GET v1/products', fn() => $this->client->products()->list(['page' => 1, 'per_page' => 1]));

        self::assertLessThanOrEqual(1, \count($page->items), 'more items than per_page');
        foreach ($page->items as $product) {
            self::assertGreaterThan(0, $product->id, 'product id is not positive');
        }
    }

    /**
     * Runs one API call. On failure it reports the exception type, HTTP
     * status and request id but not the message or the previous exception,
     * which can quote the response body.
     *
     * @template T
     *
     * @param callable(): T $call
     *
     * @return T
     */
    private static function call(string $what, callable $call): mixed
    {
        try {
            return $call();
        } catch (ApiException $e) {
            self::fail(\sprintf('%s failed: %s, HTTP %d, X-Request-Id %s', $what, self::type($e), $e->statusCode, $e->requestId));
        } catch (CyberbizException $e) {
            self::fail(\sprintf('%s failed: %s', $what, self::type($e)));
        }
    }

    private static function type(\Throwable $e): string
    {
        $class = $e::class;
        $slash = strrpos($class, '\\');

        return $slash === false ? $class : substr($class, $slash + 1);
    }
}
