<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Request;
use Actinc\Cyberbiz\Tests\Fake\FakeHttpClient;
use Nyholm\Psr7\Factory\Psr17Factory;
use PHPUnit\Framework\TestCase;
use Psr\Http\Message\ResponseInterface;

final class PaginationTest extends TestCase
{
    private FakeHttpClient $http;

    private function client(ResponseInterface ...$replies): Client
    {
        $this->http = new FakeHttpClient(null, ...$replies);
        $factory = new Psr17Factory();

        return new Client('synthetic-token', 'https://api.example.test/', $this->http, $factory, $factory, rateLimit: 0);
    }

    private static function page(string $body, string $page, string $next): ResponseInterface
    {
        return FakeHttpClient::json(200, $body, ['X-Page' => $page, 'X-Next-Page' => $next, 'X-Per-Page' => '50']);
    }

    /** @return array<string, string> */
    private function query(int $i): array
    {
        parse_str($this->http->requests[$i]->getUri()->getQuery(), $query);
        $out = [];
        foreach ($query as $key => $value) {
            if (\is_string($value)) {
                $out[(string) $key] = $value;
            }
        }
        ksort($out);

        return $out;
    }

    public function testWalksEveryPageWithExactlyOneRequestEach(): void
    {
        $client = $this->client(self::page('[{"id":1},{"id":2}]', '1', '2'), self::page('[{"id":3}]', '2', '3'), self::page('[{"id":4}]', '3', ''));

        $ids = iterator_to_array($client->each(new Request('GET', '/v1/products', ['offset' => 10]), static fn(mixed $p): mixed => \is_array($p) ? $p['id'] : null), false);

        self::assertSame([1, 2, 3, 4], $ids);
        self::assertCount(3, $this->http->requests);
        foreach ([1, 2, 3] as $i => $page) {
            self::assertSame(['offset' => '10', 'page' => (string) $page, 'per_page' => '50'], $this->query($i));
        }
    }

    public function testStartsAtTheRequestedPageAndKeepsAnExplicitPerPage(): void
    {
        $client = $this->client(self::page('[{"id":9}]', '4', ''));

        iterator_to_array($client->each(new Request('GET', '/v1/orders', ['page' => 4, 'per_page' => 10])));

        self::assertSame(['page' => '4', 'per_page' => '10'], $this->query(0));
    }

    public function testStopsAtAnEmptyPageEvenIfANextPageIsAdvertised(): void
    {
        $client = $this->client(self::page('[]', '1', '2'));

        self::assertSame([], iterator_to_array($client->each(new Request('GET', '/v1/orders'))));
        self::assertCount(1, $this->http->requests);
    }

    public function testPageRequiresAJsonArray(): void
    {
        $client = $this->client(self::page('{"id":1}', '1', ''));

        $this->expectException(DecodeException::class);
        $client->page(new Request('GET', '/v1/orders'));
    }
}
