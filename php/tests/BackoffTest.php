<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Http\Backoff;
use Actinc\Cyberbiz\Query;
use PHPUnit\Framework\TestCase;

final class BackoffTest extends TestCase
{
    public function testDoublesFromHalfASecondAndCapsAtEight(): void
    {
        $backoff = new Backoff(random: static fn(): float => 0.0);

        self::assertSame([0.5, 1.0, 2.0, 4.0, 8.0, 8.0], array_map($backoff->delay(...), [1, 2, 3, 4, 5, 6]));
    }

    public function testAddsAtMostAQuarterOfJitter(): void
    {
        $backoff = new Backoff(random: static fn(): float => 0.999);

        self::assertEqualsWithDelta(2.0 * 1.25, $backoff->delay(3), 0.01);
    }

    public function testParsesRetryAfterSecondsAndDates(): void
    {
        $now = (new \DateTimeImmutable('2026-10-03 00:00:00 UTC'))->getTimestamp();

        self::assertSame(120.0, Backoff::retryAfter('120', $now));
        self::assertSame(30.0, Backoff::retryAfter('Sat, 03 Oct 2026 00:00:30 GMT', $now));
        self::assertSame(0.0, Backoff::retryAfter('Fri, 02 Oct 2026 23:00:00 GMT', $now));
        self::assertNull(Backoff::retryAfter('', $now));
        self::assertNull(Backoff::retryAfter('soon', $now));
    }

    public function testEncodesQueriesTheWayCyberbizExpects(): void
    {
        self::assertSame('ids=1&ids=2&q=%E7%AF%84%E4%BE%8B&on=true', Query::encode(['ids' => [1, 2], 'q' => '範例', 'on' => true, 'off' => null]));
        self::assertSame('', Query::encode([]));
    }
}
