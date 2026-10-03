<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/**
 * An error response from the CYBERBIZ API: any non-2xx status, or a 2xx whose
 * body is nothing but an error object. Subclasses narrow it by status.
 */
class ApiException extends \RuntimeException implements CyberbizException
{
    /**
     * @param list<string> $messages human-readable messages from the body,
     *                               usually Traditional Chinese
     */
    public function __construct(
        public readonly int $statusCode,
        public readonly string $method,
        public readonly string $path,
        public readonly string $requestId,
        public readonly array $messages,
        public readonly string $body,
    ) {
        $text = $messages === [] ? 'HTTP ' . $statusCode : implode('; ', $messages);
        parent::__construct(\sprintf('cyberbiz: %s %s: %d %s', $method, $path, $statusCode, $text), $statusCode);
    }
}
