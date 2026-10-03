<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Client;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class ClientTest extends TestCase
{
    public function testDefaultsToTheCyberbizApiHost(): void
    {
        $client = new Client('synthetic-token');

        self::assertSame(Client::DEFAULT_BASE_URI, $client->baseUri());
        self::assertSame('Bearer synthetic-token', $client->authorization());
    }

    public function testNormalisesTheBaseUriToEndInASlash(): void
    {
        $client = new Client('synthetic-token', 'http://localhost:8080');

        self::assertSame('http://localhost:8080/', $client->baseUri());
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

    public function testKeepsTheTokenOutOfDebugOutput(): void
    {
        $client = new Client('synthetic-token');

        self::assertStringNotContainsString('synthetic-token', print_r($client, true));
    }
}
