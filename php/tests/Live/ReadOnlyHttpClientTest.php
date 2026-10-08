<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Live;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Tests\Fake\FakeHttpClient;
use Actinc\Cyberbiz\Tests\Golden;
use Nyholm\Psr7\Factory\Psr17Factory;
use PHPUnit\Framework\TestCase;

/** The guard LiveTest relies on: only GET reaches the network. */
final class ReadOnlyHttpClientTest extends TestCase
{
    private FakeHttpClient $fake;

    private Client $client;

    protected function setUp(): void
    {
        $this->fake = new FakeHttpClient(null, FakeHttpClient::json(200, Golden::response('app/GET_shop.json')->body));
        $factory = new Psr17Factory();
        $this->client = new Client('synthetic-token', 'https://api.example.test/', new ReadOnlyHttpClient($this->fake), $factory, $factory, rateLimit: 0, maxRetries: 0);
    }

    public function testPassesGetThrough(): void
    {
        $this->client->shop()->info();

        self::assertCount(1, $this->fake->requests);
        self::assertSame('GET', $this->fake->requests[0]->getMethod());
    }

    public function testRefusesWritesBeforeSending(): void
    {
        try {
            $this->client->shop()->updateSettings(['greeting' => 'hello']);
            self::fail('a write was allowed');
        } catch (\LogicException $e) {
            self::assertStringContainsString('refusing PUT', $e->getMessage());
        }
        self::assertSame([], $this->fake->requests);
    }
}
