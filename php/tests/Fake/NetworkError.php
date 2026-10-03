<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Fake;

use Psr\Http\Client\NetworkExceptionInterface;
use Psr\Http\Message\RequestInterface;

final class NetworkError extends \RuntimeException implements NetworkExceptionInterface
{
    public function __construct(private readonly RequestInterface $request)
    {
        parent::__construct('connection reset');
    }

    public function getRequest(): RequestInterface
    {
        return $this->request;
    }
}
