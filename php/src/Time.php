<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

use Actinc\Cyberbiz\Exception\DecodeException;

/**
 * CYBERBIZ timestamps: "2006-01-02 15:04:05" with no zone, always meaning
 * Asia/Taipei. ISO 8601 and the other forms the platform occasionally emits
 * are accepted too; anything without a zone is read in Taipei.
 */
final class Time
{
    public const ZONE = 'Asia/Taipei';

    public const FORMAT = 'Y-m-d H:i:s';

    /** Accepted formats, most common first ("!" zeroes unspecified fields). */
    private const FORMATS = [
        '!Y-m-d H:i:s',
        '!Y-m-d\TH:i:sP',
        '!Y-m-d\TH:i:s.uP',
        '!Y-m-d H:i:s O',
        '!Y-m-d H:i:s P',
        '!Y-m-d H:i:sO',
        '!Y-m-d H:iO',
        '!Y-m-d H:i',
        '!Y-m-d\TH:i:s',
        '!Y-m-d H:i:s.u',
        '!Y-m-d',
    ];

    /**
     * @return \DateTimeImmutable|null null for null or an empty string
     *
     * @throws DecodeException when no accepted format matches
     */
    public static function parse(?string $value): ?\DateTimeImmutable
    {
        $value = trim((string) $value);
        if ($value === '') {
            return null;
        }
        $zone = new \DateTimeZone(self::ZONE);
        foreach (self::FORMATS as $format) {
            $parsed = \DateTimeImmutable::createFromFormat($format, $value, $zone);
            $errors = \DateTimeImmutable::getLastErrors();
            if ($parsed !== false && ($errors === false || $errors['warning_count'] === 0)) {
                return $parsed->setTimezone($zone);
            }
        }

        throw new DecodeException(\sprintf('cyberbiz: cannot parse "%s" as a CYBERBIZ timestamp', $value));
    }

    /** Formats in the platform layout, in Taipei. */
    public static function format(\DateTimeInterface $time): string
    {
        return \DateTimeImmutable::createFromInterface($time)->setTimezone(new \DateTimeZone(self::ZONE))->format(self::FORMAT);
    }
}
