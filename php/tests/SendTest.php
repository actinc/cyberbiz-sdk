<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\AuthenticationException;
use Actinc\Cyberbiz\Exception\ForbiddenException;
use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Exception\RateLimitException;
use Actinc\Cyberbiz\Exception\ServerException;
use Actinc\Cyberbiz\Exception\TransportException;
use Actinc\Cyberbiz\Exception\ValidationException;
use Actinc\Cyberbiz\Http\Backoff;
use Actinc\Cyberbiz\Request;
use Actinc\Cyberbiz\Tests\Fake\FakeClock;
use Actinc\Cyberbiz\Tests\Fake\FakeHttpClient;
use Nyholm\Psr7\Factory\Psr17Factory;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;
use Psr\Http\Message\ResponseInterface;

final class SendTest extends TestCase
{
    private const TOKEN = 'synthetic-token-do-not-use';

    private FakeClock $clock;

    private FakeHttpClient $http;

    protected function setUp(): void
    {
        $this->clock = new FakeClock();
    }

    /** @param ResponseInterface|'network' ...$replies */
    private function client(float $rateLimit = 0, ResponseInterface|string ...$replies): Client
    {
        $this->http = new FakeHttpClient($this->clock, ...$replies);
        $factory = new Psr17Factory();

        return new Client(
            self::TOKEN,
            'https://api.example.test/',
            $this->http,
            $factory,
            $factory,
            rateLimit: $rateLimit,
            backoff: new Backoff(random: static fn(): float => 0.0),
            clock: $this->clock
        );
    }

    public function testSendsAuthJsonAndQuery(): void
    {
        $client = $this->client(0, FakeHttpClient::json(201, '{"id":7}'));

        $response = $client->send(new Request('post', 'v1/products', ['ids' => [1, 2], 'draft' => false, 'skip' => null], ['title' => '範例商品']));

        $sent = $this->http->requests[0];
        self::assertSame('POST', $sent->getMethod());
        self::assertSame('https://api.example.test/v1/products?ids=1&ids=2&draft=false', (string) $sent->getUri());
        self::assertSame('Bearer ' . self::TOKEN, $sent->getHeaderLine('Authorization'));
        self::assertSame('application/json', $sent->getHeaderLine('Accept'));
        self::assertSame('application/json', $sent->getHeaderLine('Content-Type'));
        self::assertStringStartsWith('cyberbiz-sdk-php/', $sent->getHeaderLine('User-Agent'));
        self::assertSame('{"title":"範例商品"}', (string) $sent->getBody());
        self::assertSame(201, $response->statusCode);
    }

    public function testHonoursRetryAfterOn429(): void
    {
        $client = $this->client(0, FakeHttpClient::json(429, '{}', ['Retry-After' => '2']), FakeHttpClient::json(200, '{"ok":true}'));

        self::assertSame('{"ok":true}', $client->send(new Request('GET', '/shop'))->body);
        self::assertSame([2.0], $this->clock->sleeps);
        self::assertCount(2, $this->http->requests);
    }

    public function testBacksOffExponentiallyThenThrowsServerException(): void
    {
        $busy = FakeHttpClient::json(503, '{"error":"busy"}', ['X-Request-Id' => 'req-503']);
        $client = $this->client(0, $busy, $busy, $busy, $busy);

        try {
            $client->send(new Request('GET', '/v1/orders'));
            self::fail('expected ServerException');
        } catch (ServerException $e) {
            self::assertSame(503, $e->statusCode);
            self::assertSame('req-503', $e->requestId);
            self::assertSame(['busy'], $e->messages);
        }
        self::assertSame([0.5, 1.0, 2.0], $this->clock->sleeps);
        self::assertCount(1 + Client::DEFAULT_MAX_RETRIES, $this->http->requests);
    }

    public function testThrowsRateLimitExceptionWhenRetriesRunOut(): void
    {
        $limited = FakeHttpClient::json(429, '{"message":"too many"}', ['Retry-After' => '1', 'X-Request-Id' => 'req-429']);
        $client = $this->client(0, $limited, $limited, $limited, $limited);

        $this->expectException(RateLimitException::class);
        $client->send(new Request('GET', '/v1/orders'));
    }

    /** @param class-string<ApiException> $class */
    #[DataProvider('clientErrors')]
    public function testMapsClientErrorsWithoutRetryingOrLeakingTheToken(int $status, string $class): void
    {
        $client = $this->client(0, FakeHttpClient::json($status, '{"error":["權限不足"]}', ['X-Request-Id' => 'req-x']));

        try {
            $client->send(new Request('GET', '/v1/customers/1'));
            self::fail('expected ' . $class);
        } catch (ApiException $e) {
            self::assertInstanceOf($class, $e);
            self::assertSame('req-x', $e->requestId);
            self::assertSame('cyberbiz: GET /v1/customers/1: ' . $status . ' 權限不足', $e->getMessage());
            self::assertStringNotContainsString(self::TOKEN, (string) $e);
        }
        self::assertCount(1, $this->http->requests);
        self::assertSame([], $this->clock->sleeps);
    }

    /** @return iterable<string, array{int, class-string<ApiException>}> */
    public static function clientErrors(): iterable
    {
        yield '401' => [401, AuthenticationException::class];
        yield '403' => [403, ForbiddenException::class];
        yield '404' => [404, NotFoundException::class];
        yield '422' => [422, ValidationException::class];
    }

    public function testTreatsA2xxErrorObjectAsAnError(): void
    {
        $client = $this->client(0, FakeHttpClient::json(200, '{"errors":{"title":["不可空白"]}}'));

        $this->expectExceptionMessage('title: 不可空白');
        $client->send(new Request('GET', '/v1/products/1'));
    }

    public function testRetriesNetworkErrorsOnlyForIdempotentMethods(): void
    {
        $client = $this->client(0, 'network', FakeHttpClient::json(200, '{}'));
        $client->send(new Request('GET', '/shop'));
        self::assertCount(2, $this->http->requests);

        $client = $this->client(0, 'network', FakeHttpClient::json(200, '{}'));
        $this->expectException(TransportException::class);
        $client->send(new Request('POST', '/v1/orders', body: ['x' => 1]));
    }

    public function testNeverStartsMoreThanFiveRequestsPerSecond(): void
    {
        $client = $this->client(Client::DEFAULT_RATE_LIMIT);

        for ($i = 0; $i < 12; ++$i) {
            $client->send(new Request('GET', '/shop'));
        }

        foreach ($this->http->sentAt as $i => $start) {
            $inWindow = array_filter($this->http->sentAt, static fn(float $t): bool => $t >= $start && $t < $start + 1.0);
            self::assertLessThanOrEqual(5, \count($inWindow), 'requests starting within 1s of #' . $i);
        }
        self::assertEqualsWithDelta(11 * 0.2, end($this->http->sentAt) - $this->http->sentAt[0], 1e-9);
    }
}
