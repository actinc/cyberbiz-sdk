<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Live;

use Psr\Http\Client\ClientInterface;
use Psr\Http\Message\RequestInterface;
use Psr\Http\Message\ResponseInterface;

/**
 * Passes GET requests to another PSR-18 client and refuses every other
 * method before anything is sent, so the live tests (LiveTest) cannot write
 * to a real shop even by mistake.
 */
final class ReadOnlyHttpClient implements ClientInterface
{
    public function __construct(private readonly ClientInterface $next) {}

    public function sendRequest(RequestInterface $request): ResponseInterface
    {
        if ($request->getMethod() !== 'GET') {
            // Not a PSR-18 exception, so the SDK neither retries nor wraps it.
            throw new \LogicException(\sprintf('read-only client: refusing %s %s', $request->getMethod(), $request->getUri()->getPath()));
        }

        return $this->next->sendRequest($request);
    }
}
