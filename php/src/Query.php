<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

/**
 * Encodes query parameters the way CYBERBIZ expects: a list value repeats
 * the key (ids=1&ids=2), booleans are "true"/"false", nulls are dropped.
 *
 * @internal
 */
final class Query
{
    /** @param array<string, scalar|list<scalar>|null> $params */
    public static function encode(array $params): string
    {
        $pairs = [];
        foreach ($params as $key => $value) {
            foreach (\is_array($value) ? $value : [$value] as $item) {
                if ($item === null) {
                    continue;
                }
                $pairs[] = rawurlencode($key) . '=' . rawurlencode(self::scalar($item));
            }
        }

        return implode('&', $pairs);
    }

    private static function scalar(bool|float|int|string $value): string
    {
        return \is_bool($value) ? ($value ? 'true' : 'false') : (string) $value;
    }
}
