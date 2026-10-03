<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

/** One API call: method, path relative to the base URI, query and body. */
final class Request
{
    /**
     * @param array<string, scalar|list<scalar>|null> $query   null values are dropped
     * @param mixed                                   $body    JSON-encoded unless null
     * @param array<string, string>                   $headers extra headers
     */
    public function __construct(
        public readonly string $method,
        public readonly string $path,
        public readonly array $query = [],
        public readonly mixed $body = null,
        public readonly array $headers = [],
    ) {}

    /** Whether repeating the request after a network failure is safe. */
    public function isIdempotent(): bool
    {
        return \in_array(strtoupper($this->method), ['GET', 'HEAD', 'PUT', 'DELETE', 'OPTIONS'], true);
    }
}
