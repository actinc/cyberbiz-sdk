<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Fake;

use Nyholm\Psr7\Response;
use Psr\Http\Client\ClientInterface;
use Psr\Http\Message\RequestInterface;
use Psr\Http\Message\ResponseInterface;

/** Replays scripted replies and records every request with the clock time it was sent. */
final class FakeHttpClient implements ClientInterface
{
    /** @var list<RequestInterface> */
    public array $requests = [];

    /** @var list<float> */
    public array $sentAt = [];

    /** @var list<ResponseInterface|'network'> */
    private array $replies;

    /** @param ResponseInterface|'network' ...$replies */
    public function __construct(private readonly ?FakeClock $clock = null, ResponseInterface|string ...$replies)
    {
        $this->replies = array_values($replies);
    }

    /** @param array<string, string> $headers */
    public static function json(int $status, string $body, array $headers = []): Response
    {
        return new Response($status, ['Content-Type' => 'application/json'] + $headers, $body);
    }

    public function sendRequest(RequestInterface $request): ResponseInterface
    {
        $this->requests[] = $request;
        $this->sentAt[] = $this->clock?->now() ?? 0.0;
        $reply = array_shift($this->replies) ?? self::json(200, '{}');
        if ($reply === 'network') {
            throw new NetworkError($request);
        }

        return $reply;
    }
}
