package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.http.Clock;
import java.time.Duration;

/**
 * Spaces requests evenly so that no more than {@code perSecond} start in any second: a token bucket
 * with a burst of one, as in the Go and PHP SDKs. Thread-safe: each caller reserves the next free
 * slot under a lock and sleeps outside it.
 */
final class RateLimiter {
  private final long intervalNanos;
  private final Clock clock;
  private long nextNanos;
  private boolean started;

  RateLimiter(double perSecond, Clock clock) {
    if (!(perSecond > 0) || Double.isInfinite(perSecond)) {
      throw new IllegalArgumentException("rate limit must be a positive number: " + perSecond);
    }
    this.intervalNanos = (long) Math.ceil(1_000_000_000L / perSecond);
    this.clock = clock;
  }

  /** Blocks until the caller's slot comes. */
  void acquire() throws InterruptedException {
    long wait = reserve();
    if (wait > 0) {
      clock.sleep(Duration.ofNanos(wait));
    }
  }

  /** Reserves the next slot and returns how long to wait for it, in nanoseconds. */
  synchronized long reserve() {
    long now = clock.nanoTime();
    long slot = started && nextNanos - now > 0 ? nextNanos : now;
    started = true;
    nextNanos = slot + intervalNanos;
    return slot - now;
  }
}
