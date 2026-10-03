<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

/** A response from the API, body fully read. */
final class Response
{
    /**
     * @param array<string, list<string>> $headers as PSR-7 returns them
     */
    public function __construct(
        public readonly int $statusCode,
        public readonly array $headers,
        public readonly string $body,
    ) {}

    /** The first value of a header, matched case-insensitively, or "". */
    public function header(string $name): string
    {
        foreach ($this->headers as $key => $values) {
            if (strcasecmp($key, $name) === 0) {
                return $values[0] ?? '';
            }
        }

        return '';
    }

    /** The X-Request-Id header, useful when contacting CYBERBIZ. */
    public function requestId(): string
    {
        return $this->header('X-Request-Id');
    }

    /** Whether the body is the JSON literal null (some lookups return it instead of 404). */
    public function isNull(): bool
    {
        return trim($this->body) === 'null';
    }
}
