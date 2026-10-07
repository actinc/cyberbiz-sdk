package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;

import cc.alphacore.cyberbiz.http.Backoff;
import java.time.Duration;
import java.time.Instant;
import java.util.Optional;
import org.junit.jupiter.api.Test;

class BackoffTest {
  private static final Instant NOW = Instant.parse("2026-10-07T00:00:00Z");

  @Test
  void doublesUpToTheCap() {
    Backoff backoff = new Backoff(Duration.ofMillis(500), Duration.ofSeconds(8), () -> 0);

    assertEquals(Duration.ofMillis(500), backoff.delay(1));
    assertEquals(Duration.ofSeconds(1), backoff.delay(2));
    assertEquals(Duration.ofSeconds(4), backoff.delay(4));
    assertEquals(Duration.ofSeconds(8), backoff.delay(10));
    assertEquals(Duration.ofSeconds(8), backoff.delay(Integer.MAX_VALUE));
  }

  @Test
  void addsUpToAQuarterOfJitter() {
    Backoff backoff = new Backoff(Duration.ofSeconds(4), Duration.ofSeconds(8), () -> 0.999);

    assertEquals(Duration.ofMillis(4999), backoff.delay(1));
  }

  @Test
  void readsRetryAfterSecondsAndDates() {
    assertEquals(Optional.of(Duration.ofSeconds(3)), Backoff.retryAfter(" 3 ", NOW));
    assertEquals(
        Optional.of(Duration.ofSeconds(90)),
        Backoff.retryAfter("Wed, 07 Oct 2026 00:01:30 GMT", NOW));
    assertEquals(
        Optional.of(Duration.ZERO), Backoff.retryAfter("Tue, 06 Oct 2026 23:00:00 GMT", NOW));
  }

  @Test
  void ignoresMissingOrUnreadableRetryAfter() {
    assertEquals(Optional.empty(), Backoff.retryAfter("", NOW));
    assertEquals(Optional.empty(), Backoff.retryAfter(null, NOW));
    assertEquals(Optional.empty(), Backoff.retryAfter("soon", NOW));
    assertEquals(Optional.empty(), Backoff.retryAfter("-1", NOW));
    assertEquals(Optional.empty(), Backoff.retryAfter("99999999999999999999", NOW));
  }
}
