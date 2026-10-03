<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Request;
use Actinc\Cyberbiz\Tests\Fake\FakeHttpClient;
use Nyholm\Psr7\Factory\Psr17Factory;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class ClientTest extends TestCase
{
    public function testDefaultsToTheCyberbizApiHost(): void
    {
        self::assertSame(Client::DEFAULT_BASE_URI, (new Client('synthetic-token'))->baseUri());
    }

    public function testNormalisesTheBaseUriToEndInASlash(): void
    {
        self::assertSame('http://localhost:8080/', (new Client('synthetic-token', 'http://localhost:8080'))->baseUri());
    }

    public function testRejectsAnEmptyToken(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new Client('  ');
    }

    #[DataProvider('invalidBaseUris')]
    public function testRejectsANonHttpBaseUri(string $baseUri): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new Client('synthetic-token', $baseUri);
    }

    /** @return iterable<string, array{string}> */
    public static function invalidBaseUris(): iterable
    {
        yield 'relative' => ['/v1'];
        yield 'ftp' => ['ftp://example.com/'];
        yield 'no host' => ['https://'];
    }

    public function testRejectsNegativeLimits(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        new Client('synthetic-token', maxRetries: -1);
    }

    public function testKeepsTheTokenOutOfDebugOutput(): void
    {
        self::assertStringNotContainsString('synthetic-token', print_r(new Client('synthetic-token'), true));
    }

    public function testDiscoversAnHttpClientWhenNoneIsInjected(): void
    {
        // Discovery finds the PSR-18 client and PSR-17 factories installed
        // for development (symfony/http-client, nyholm/psr7).
        $client = new Client('synthetic-token');

        self::assertInstanceOf(Client::class, $client);
    }

    public function testUsesTheInjectedHttpClient(): void
    {
        $http = new FakeHttpClient(null, FakeHttpClient::json(200, '{"id":1}'));
        $factory = new Psr17Factory();
        $client = new Client('synthetic-token', httpClient: $http, requestFactory: $factory, streamFactory: $factory, rateLimit: 0);

        $client->send(new Request('GET', '/shop'));

        self::assertCount(1, $http->requests);
    }
}
