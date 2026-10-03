<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Http;

/** Retry delays: Retry-After when the server sent one, else exponential. */
final class Backoff
{
    /**
     * @param \Closure(): float $random returns a number in [0, 1); injectable for tests
     */
    public function __construct(
        private readonly float $base = 0.5,
        private readonly float $max = 8.0,
        private readonly ?\Closure $random = null,
    ) {}

    /** Doubles $base per attempt (1-based), caps at $max, adds up to 25% jitter. */
    public function delay(int $attempt): float
    {
        $delay = min($this->base * 2 ** max(0, $attempt - 1), $this->max);
        $random = $this->random ?? static fn(): float => mt_rand() / (mt_getrandmax() + 1);

        return $delay + $delay / 4 * $random();
    }

    /** Seconds to wait per a Retry-After header (seconds or HTTP date), or null. */
    public static function retryAfter(string $header, float $nowUnix): ?float
    {
        $header = trim($header);
        if ($header === '') {
            return null;
        }
        if (ctype_digit($header)) {
            return (float) $header;
        }
        $date = \DateTimeImmutable::createFromFormat(\DATE_RFC7231, $header);
        if ($date === false) {
            return null;
        }

        return max(0.0, $date->getTimestamp() - $nowUnix);
    }
}
