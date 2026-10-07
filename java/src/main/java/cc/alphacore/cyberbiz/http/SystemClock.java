package cc.alphacore.cyberbiz.http;

import java.time.Duration;
import java.time.Instant;

/** The real clock. */
public final class SystemClock implements Clock {

  /** The shared instance. */
  public static final SystemClock INSTANCE = new SystemClock();

  private SystemClock() {}

  @Override
  public Instant now() {
    return Instant.now();
  }

  @Override
  public long nanoTime() {
    return System.nanoTime();
  }

  @Override
  public void sleep(Duration duration) throws InterruptedException {
    if (!duration.isNegative() && !duration.isZero()) {
      Thread.sleep(duration.toMillis(), duration.toNanosPart() % 1_000_000);
    }
  }
}
