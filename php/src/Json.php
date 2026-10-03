<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

use Actinc\Cyberbiz\Exception\DecodeException;

/**
 * Decodes API JSON without losing precision: a number with a fraction or an
 * exponent (3690.0, 1e3) is returned as its exact decimal text, never as a
 * float, and integers too large for PHP stay strings. Objects become
 * associative arrays.
 */
final class Json
{
    /** A JSON string literal, or a number token that is not a plain integer. */
    private const TOKEN = '/"(?:[^"\\\\]|\\\\.)*+"|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?/';

    /** @throws DecodeException for malformed JSON */
    public static function decode(string $json): mixed
    {
        $quoted = preg_replace_callback(self::TOKEN, static function (array $m): string {
            $token = $m[0];

            return $token[0] === '"' || strpbrk($token, '.eE') === false ? $token : '"' . $token . '"';
        }, $json);
        if ($quoted === null) {
            throw new DecodeException('cyberbiz: cannot scan JSON: ' . preg_last_error_msg());
        }
        try {
            return json_decode($quoted, true, 512, \JSON_THROW_ON_ERROR | \JSON_BIGINT_AS_STRING);
        } catch (\JsonException $e) {
            throw new DecodeException('cyberbiz: invalid JSON: ' . $e->getMessage(), 0, $e);
        }
    }
}
