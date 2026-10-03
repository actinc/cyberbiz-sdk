<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Http;

/**
 * Spaces requests evenly so that no more than $perSecond start in any second
 * (a token bucket with a burst of one, like the Go SDK).
 */
final class RateLimiter
{
    private ?float $next = null;

    public function __construct(private readonly float $perSecond, private readonly Clock $clock)
    {
        if ($perSecond <= 0) {
            throw new \InvalidArgumentException('cyberbiz: rate limit must be positive');
        }
    }

    /** Blocks until the next request may start. */
    public function wait(): void
    {
        $now = $this->clock->now();
        if ($this->next !== null && $this->next > $now) {
            $this->clock->sleep($this->next - $now);
            $now = $this->next;
        }
        $this->next = $now + 1 / $this->perSecond;
    }
}
