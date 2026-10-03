<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

use Actinc\Cyberbiz\Exception\DecodeException;

/**
 * An exact monetary amount in the shop currency (TWD for every CYBERBIZ shop
 * today), kept as a decimal string with two places and computed with bcmath.
 * CYBERBIZ sends amounts as numbers such as 9999.0 and sometimes as strings;
 * both parse without passing through float. More than two decimals round
 * half away from zero, as in the Go SDK.
 */
final class Money implements \Stringable
{
    private const SCALE = 2;

    /** @param numeric-string $amount */
    private function __construct(public readonly string $amount) {}

    /**
     * @param int|string $value an integer or decimal text such as "199.5"
     *
     * @throws DecodeException when the value is not a decimal number
     */
    public static function of(int|string $value): self
    {
        $text = self::numeric(trim((string) $value));
        if (stripos($text, 'e') !== false) {
            $text = self::expand($text);
        }

        return new self(self::round($text));
    }

    public static function zero(): self
    {
        return new self('0.00');
    }

    public function add(self $other): self
    {
        return new self(bcadd($this->amount, $other->amount, self::SCALE));
    }

    public function sub(self $other): self
    {
        return new self(bcsub($this->amount, $other->amount, self::SCALE));
    }

    /** -1, 0 or 1. */
    public function compare(self $other): int
    {
        return bccomp($this->amount, $other->amount, self::SCALE);
    }

    public function equals(self $other): bool
    {
        return $this->compare($other) === 0;
    }

    public function isNegative(): bool
    {
        return bccomp($this->amount, '0', self::SCALE) < 0;
    }

    public function __toString(): string
    {
        return $this->amount;
    }

    /**
     * @return numeric-string
     *
     * @throws DecodeException
     */
    private static function numeric(string $text): string
    {
        if (preg_match('/^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$/', $text) !== 1 || !is_numeric($text)) {
            throw new DecodeException(\sprintf('cyberbiz: cannot parse "%s" as Money', $text));
        }

        return $text;
    }

    /**
     * Rounds to two places, half away from zero.
     *
     * @param numeric-string $decimal
     *
     * @return numeric-string
     */
    private static function round(string $decimal): string
    {
        $half = str_starts_with(ltrim($decimal), '-') ? '-0.005' : '0.005';
        $rounded = bcadd(bcadd($decimal, '0', 10), $half, self::SCALE);

        return $rounded === '-0.00' ? '0.00' : $rounded;
    }

    /**
     * Exponent notation is rare; expand it exactly with bcmath.
     *
     * @param numeric-string $text
     *
     * @return numeric-string
     */
    private static function expand(string $text): string
    {
        [$mantissa, $exponent] = preg_split('/[eE]/', $text) ?: [$text, '0'];
        $mantissa = self::numeric($mantissa);
        $power = (int) $exponent;
        $factor = bcpow('10', (string) abs($power), 0);

        return $power >= 0 ? bcmul($mantissa, $factor, 10) : bcdiv($mantissa, $factor, 10);
    }
}
