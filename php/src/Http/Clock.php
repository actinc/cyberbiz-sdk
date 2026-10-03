<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Http;

/** Time source for rate limiting and retries; tests inject a fake. */
interface Clock
{
    /** Seconds since an arbitrary fixed point, with sub-second precision. */
    public function now(): float;

    public function sleep(float $seconds): void;
}
