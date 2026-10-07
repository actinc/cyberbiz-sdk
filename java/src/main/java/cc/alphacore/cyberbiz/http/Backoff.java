package cc.alphacore.cyberbiz.http;

import java.time.Duration;
import java.time.Instant;
import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.util.Objects;
import java.util.Optional;
import java.util.concurrent.ThreadLocalRandom;
import java.util.function.DoubleSupplier;

/** Retry delays: {@code Retry-After} when the server sent one, otherwise exponential. */
public final class Backoff {

  private final Duration base;
  private final Duration max;
  private final DoubleSupplier random;

  /** Creates the default backoff: 0.5 s doubling per attempt, at most 8 s, plus up to 25%. */
  public Backoff() {
    this(
        Duration.ofMillis(500),
        Duration.ofSeconds(8),
        () -> ThreadLocalRandom.current().nextDouble());
  }

  /**
   * Creates a backoff.
   *
   * @param base the delay before the first retry
   * @param max the longest delay before jitter
   * @param random returns a number in [0, 1) for the jitter; injectable for tests
   */
  public Backoff(Duration base, Duration max, DoubleSupplier random) {
    this.base = Objects.requireNonNull(base, "base");
    this.max = Objects.requireNonNull(max, "max");
    this.random = Objects.requireNonNull(random, "random");
  }

  /**
   * Returns the delay before a retry: base doubled per attempt, capped, plus up to 25% jitter.
   *
   * @param attempt the retry number, starting at 1
   * @return the delay
   */
  public Duration delay(int attempt) {
    long cap = max.toNanos();
    long nanos = base.toNanos();
    for (int i = 1; i < attempt && nanos < cap; i++) {
      nanos = nanos > Long.MAX_VALUE / 2 ? Long.MAX_VALUE : nanos * 2;
    }
    nanos = Math.min(nanos, cap);
    return Duration.ofNanos(nanos + (long) (nanos / 4.0 * random.getAsDouble()));
  }

  /**
   * Reads a {@code Retry-After} header: whole seconds or an HTTP date.
   *
   * @param header the header value, possibly empty
   * @param now the current time, for the date form
   * @return the wait, never negative, or empty when the header is absent or unreadable
   */
  public static Optional<Duration> retryAfter(String header, Instant now) {
    String value = header == null ? "" : header.strip();
    if (value.isEmpty()) {
      return Optional.empty();
    }
    if (value.chars().allMatch(c -> c >= '0' && c <= '9')) {
      // More than 9 digits is over 30 years: treat it as unreadable.
      return value.length() > 9
          ? Optional.empty()
          : Optional.of(Duration.ofSeconds(Long.parseLong(value)));
    }
    try {
      Instant at = ZonedDateTime.parse(value, DateTimeFormatter.RFC_1123_DATE_TIME).toInstant();
      Duration wait = Duration.between(now, at);
      return Optional.of(wait.isNegative() ? Duration.ZERO : wait);
    } catch (DateTimeParseException e) {
      return Optional.empty();
    }
  }
}
