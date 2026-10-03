<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/**
 * Shared readers for nested values the models decode: optional objects and
 * lists of amounts.
 *
 * @internal
 */
final class Nested
{
    /**
     * Builds a model from a nested object; absent or null is null.
     *
     * @template T
     *
     * @param callable(Fields): T $build
     *
     * @return T|null
     *
     * @throws DecodeException when the value is not an object
     */
    public static function object(Fields $f, string $key, callable $build): mixed
    {
        $object = $f->objectOrNull($key);

        return $object === null ? null : $build($object);
    }

    /**
     * A JSON array of amounts; absent or null is an empty list.
     *
     * @return list<Money>
     *
     * @throws DecodeException when the value is not an array of numbers
     */
    public static function moneys(Fields $f, string $key): array
    {
        $value = $f->raw($key) ?? [];
        $path = $f->path . '.' . $key;
        if (!\is_array($value) || !array_is_list($value)) {
            throw new DecodeException(\sprintf('%s: expected array of money, got %s', $path, get_debug_type($value)));
        }
        $out = [];
        foreach ($value as $i => $item) {
            $out[] = self::money($item, $path . '[' . $i . ']');
        }

        return $out;
    }

    /** @throws DecodeException */
    private static function money(mixed $item, string $path): Money
    {
        if (!\is_int($item) && !\is_string($item)) {
            throw new DecodeException(\sprintf('%s: expected money, got %s', $path, get_debug_type($item)));
        }
        try {
            return Money::of($item);
        } catch (DecodeException $e) {
            throw new DecodeException($path . ': ' . $e->getMessage(), 0, $e);
        }
    }
}
