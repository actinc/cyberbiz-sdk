<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Http;

use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\AuthenticationException;
use Actinc\Cyberbiz\Exception\ForbiddenException;
use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Exception\RateLimitException;
use Actinc\Cyberbiz\Exception\ServerException;
use Actinc\Cyberbiz\Exception\ValidationException;
use Actinc\Cyberbiz\Request;
use Actinc\Cyberbiz\Response;

/**
 * Turns error responses into typed exceptions.
 *
 * @internal
 */
final class ErrorMapper
{
    private const MESSAGE_KEYS = ['error', 'errors', 'message', 'messages'];

    /** @throws ApiException for a non-2xx status or a 2xx that is only an error object */
    public static function check(Request $request, Response $response): void
    {
        $ok = $response->statusCode >= 200 && $response->statusCode < 300;
        if ($ok && !self::isBareErrorObject($response->body)) {
            return;
        }

        throw self::exception($request, $response);
    }

    private static function exception(Request $request, Response $response): ApiException
    {
        $status = $response->statusCode;
        $args = [
            $status,
            strtoupper($request->method),
            '/' . ltrim($request->path, '/'),
            $response->requestId(),
            self::messages($response->body),
            $response->body,
        ];

        return match (true) {
            $status === 401 => new AuthenticationException(...$args),
            $status === 403 => new ForbiddenException(...$args),
            $status === 404 => new NotFoundException(...$args),
            $status === 422 => new ValidationException(...$args),
            $status === 429 => new RateLimitException(...$args),
            $status >= 500 => new ServerException(...$args),
            default => new ApiException(...$args),
        };
    }

    /**
     * Messages from the shapes CYBERBIZ uses: {"error": ...}, {"errors": ...},
     * {"message": ...} and {"messages": ...}, each a string, a list or a map.
     *
     * @return list<string>
     */
    public static function messages(string $body): array
    {
        $decoded = json_decode($body, true);
        if (!\is_array($decoded)) {
            return [];
        }
        $out = [];
        foreach (self::MESSAGE_KEYS as $key) {
            array_push($out, ...self::flatten($decoded[$key] ?? null));
        }

        return $out;
    }

    /** @return list<string> */
    private static function flatten(mixed $value): array
    {
        if (\is_string($value)) {
            return $value === '' ? [] : [$value];
        }
        if (!\is_array($value)) {
            return [];
        }
        $out = [];
        foreach ($value as $field => $item) {
            foreach (self::flatten($item) as $text) {
                $out[] = \is_string($field) ? $field . ': ' . $text : $text;
            }
        }

        return $out;
    }

    /** A body that is exactly one of the error keys and nothing else. */
    private static function isBareErrorObject(string $body): bool
    {
        $decoded = json_decode($body, true);
        if (!\is_array($decoded) || array_is_list($decoded) || \count($decoded) !== 1) {
            return false;
        }

        return \in_array(array_key_first($decoded), ['error', 'errors', 'messages'], true);
    }
}
