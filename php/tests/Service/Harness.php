<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Service;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Tests\Fake\FakeHttpClient;
use Actinc\Cyberbiz\Tests\Golden;
use Nyholm\Psr7\Factory\Psr17Factory;
use Psr\Http\Message\RequestInterface;
use Psr\Http\Message\ResponseInterface;

/** A Client wired to a FakeHttpClient, plus helpers to read what it sent. */
final class Harness
{
    public readonly Client $client;

    public function __construct(public readonly FakeHttpClient $http)
    {
        $factory = new Psr17Factory();
        $this->client = new Client('synthetic-token', 'https://api.example.test/', $http, $factory, $factory, rateLimit: 0, maxRetries: 0);
    }

    /** Replies with the given Golden Files, in order. */
    public static function golden(string ...$names): self
    {
        return new self(new FakeHttpClient(null, ...array_map(self::goldenReply(...), $names)));
    }

    /** Replies with the given JSON bodies (status 200), in order. */
    public static function json(string ...$bodies): self
    {
        return new self(new FakeHttpClient(null, ...array_map(static fn(string $b): ResponseInterface => FakeHttpClient::json(200, $b), $bodies)));
    }

    /** The n-th request sent. */
    public function request(int $n = 0): RequestInterface
    {
        return $this->http->requests[$n] ?? throw new \OutOfRangeException('no request #' . $n);
    }

    /** "METHOD path?query" of the n-th request, relative to the base URI. */
    public function line(int $n = 0): string
    {
        $uri = $this->request($n)->getUri();
        $query = $uri->getQuery();

        return $this->request($n)->getMethod() . ' ' . ltrim($uri->getPath(), '/') . ($query === '' ? '' : '?' . $query);
    }

    /** The decoded JSON body of the n-th request. */
    public function body(int $n = 0): mixed
    {
        return json_decode((string) $this->request($n)->getBody(), true, 512, \JSON_THROW_ON_ERROR);
    }

    private static function goldenReply(string $name): ResponseInterface
    {
        $recorded = Golden::response($name);
        $headers = [];
        foreach ($recorded->headers as $key => $values) {
            $headers[$key] = implode(', ', $values);
        }

        return FakeHttpClient::json(200, $recorded->body, $headers);
    }
}
