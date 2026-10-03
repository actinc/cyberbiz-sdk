<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Fake;

use Actinc\Cyberbiz\Http\Clock;

/** A clock that only moves when something sleeps. */
final class FakeClock implements Clock
{
    /** @var list<float> */
    public array $sleeps = [];

    public function __construct(private float $now = 1000.0) {}

    public function now(): float
    {
        return $this->now;
    }

    public function sleep(float $seconds): void
    {
        $this->sleeps[] = $seconds;
        $this->now += $seconds;
    }
}
