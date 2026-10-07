package cc.alphacore.cyberbiz.http;

import java.time.Duration;
import java.time.Instant;

/** Time as the client sees it; injectable so tests never really sleep. */
public interface Clock {

  /** Returns the wall-clock time, for {@code Retry-After} dates. */
  Instant now();

  /** Returns a monotonic reading in nanoseconds, for spacing requests. */
  long nanoTime();

  /**
   * Waits for the given time.
   *
   * @param duration how long to wait; zero or negative returns at once
   * @throws InterruptedException when the thread is interrupted while waiting
   */
  void sleep(Duration duration) throws InterruptedException;
}
