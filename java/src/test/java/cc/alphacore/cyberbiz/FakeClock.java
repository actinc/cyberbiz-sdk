package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.http.Clock;
import java.time.Duration;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/** A clock that never really sleeps: sleeping advances it and is recorded. */
final class FakeClock implements Clock {
  private Instant now = Instant.parse("2026-10-07T00:00:00Z");
  private long nanos;
  private final boolean advanceOnSleep;
  final List<Duration> sleeps = new ArrayList<>();

  FakeClock() {
    this(true);
  }

  /** With advanceOnSleep false, time stands still, as if every caller started at once. */
  FakeClock(boolean advanceOnSleep) {
    this.advanceOnSleep = advanceOnSleep;
  }

  @Override
  public synchronized Instant now() {
    return now;
  }

  @Override
  public synchronized long nanoTime() {
    return nanos;
  }

  @Override
  public synchronized void sleep(Duration duration) {
    sleeps.add(duration);
    if (advanceOnSleep) {
      now = now.plus(duration);
      nanos += duration.toNanos();
    }
  }

  synchronized Duration totalSleep() {
    return sleeps.stream().reduce(Duration.ZERO, Duration::plus);
  }
}
