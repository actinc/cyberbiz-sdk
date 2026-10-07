package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import org.junit.jupiter.api.Test;

class RateLimiterTest {

  @Test
  void givesConcurrentCallersDistinctSlotsFiveASecond() throws Exception {
    // Time stands still: every thread arrives at t=0 and must be given its own slot.
    FakeClock clock = new FakeClock(false);
    RateLimiter limiter = new RateLimiter(5, clock);
    int threads = 20;
    CountDownLatch start = new CountDownLatch(1);
    ExecutorService pool = Executors.newFixedThreadPool(threads);
    try {
      List<Future<Long>> waits = new ArrayList<>();
      for (int i = 0; i < threads; i++) {
        waits.add(
            pool.submit(
                () -> {
                  start.await();
                  return limiter.reserve();
                }));
      }
      start.countDown();
      List<Long> got = new ArrayList<>();
      for (Future<Long> wait : waits) {
        got.add(wait.get());
      }
      got.sort(null);

      // Slots 0, 0.2 s, 0.4 s, ...: never more than 5 starts in any one-second window.
      for (int i = 0; i < threads; i++) {
        assertEquals(Duration.ofMillis(200L * i).toNanos(), got.get(i));
      }
    } finally {
      pool.shutdownNow();
    }
  }

  @Test
  void doesNotWaitAfterAQuietPeriod() {
    FakeClock clock = new FakeClock();
    RateLimiter limiter = new RateLimiter(5, clock);

    assertEquals(0, limiter.reserve());
    clock.sleep(Duration.ofSeconds(1));
    assertEquals(0, limiter.reserve());
  }

  @Test
  void rejectsANonPositiveRate() {
    FakeClock clock = new FakeClock();

    assertThrows(IllegalArgumentException.class, () -> new RateLimiter(0, clock));
    assertThrows(IllegalArgumentException.class, () -> new RateLimiter(Double.NaN, clock));
  }
}
