<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Service;

use Actinc\Cyberbiz\Time;

/**
 * Shapes caller options the way the order and customer endpoints expect:
 * multi-value filters and id lists comma-separated, times in the platform
 * format.
 *
 * @internal
 */
final class Options
{
    /** Date-only format used by start_date / end_date filters. */
    public const DATE = 'Y-m-d';

    /**
     * Encodes a query: a list becomes "a,b,c", a DateTimeInterface becomes
     * a CYBERBIZ timestamp in Asia/Taipei (or $timeFormat), anything else is
     * passed through.
     *
     * @param array<string, scalar|list<scalar>|\DateTimeInterface|null> $params
     *
     * @return array<string, scalar|null>
     */
    public static function query(array $params, string $timeFormat = Time::FORMAT): array
    {
        $out = [];
        foreach ($params as $key => $value) {
            $out[$key] = match (true) {
                $value instanceof \DateTimeInterface => self::time($value, $timeFormat),
                \is_array($value) => self::join($value),
                default => $value,
            };
        }

        return $out;
    }

    /**
     * Replaces a list under "line_item_ids" with the comma-separated string
     * the v1 fulfillment endpoints take; a string is kept as given.
     *
     * @param array<string, mixed> $body
     *
     * @return array<string, mixed>
     *
     * @throws \InvalidArgumentException when the list holds a non-scalar
     */
    public static function lineItemIds(array $body): array
    {
        $ids = $body['line_item_ids'] ?? null;
        if (!\is_array($ids)) {
            return $body;
        }
        $scalars = [];
        foreach ($ids as $id) {
            if (!\is_int($id) && !\is_string($id)) {
                throw new \InvalidArgumentException(\sprintf('cyberbiz: line_item_ids must hold ints or strings, got %s', get_debug_type($id)));
            }
            $scalars[] = $id;
        }
        $body['line_item_ids'] = self::join($scalars);

        return $body;
    }

    /** @param list<scalar> $values */
    public static function join(array $values): string
    {
        return implode(',', array_map(static fn(bool|float|int|string $v): string => \is_bool($v) ? ($v ? 'true' : 'false') : (string) $v, $values));
    }

    private static function time(\DateTimeInterface $time, string $format): string
    {
        return \DateTimeImmutable::createFromInterface($time)->setTimezone(new \DateTimeZone(Time::ZONE))->format($format);
    }
}
