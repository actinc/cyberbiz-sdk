package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.json.Json;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.Objects;

/**
 * An exact monetary amount in the shop currency (TWD for every CYBERBIZ shop today), kept with two
 * decimal places, as in the PHP and Go SDKs. CYBERBIZ sends amounts as numbers such as {@code
 * 9999.0} and sometimes as strings; both parse without passing through {@code double}. More than
 * two decimals round half away from zero, so {@code 1.005} becomes {@code 1.01}.
 *
 * <p>Two amounts are equal when their values are, whatever scale the API sent: {@code
 * Money.of("200.0").equals(Money.of("200"))} is true. Immutable and thread-safe.
 *
 * <p>Unlike the PHP SDK, exponent notation such as {@code "1e3"} is rejected: a value like {@code
 * "1e999999999"} would otherwise exhaust memory when rounded (see {@link Json#decimal(String)}).
 */
public final class Money implements Comparable<Money> {

  private static final int SCALE = 2;
  private static final int MAX_PRECISION = 40;
  private static final Money ZERO = new Money(BigDecimal.ZERO.setScale(SCALE));

  private final BigDecimal amount;

  private Money(BigDecimal amount) {
    this.amount = amount;
  }

  /**
   * Parses decimal text such as {@code "199.5"}.
   *
   * @param text plain decimal notation, at most 64 characters
   * @return the amount, rounded to two places
   * @throws cc.alphacore.cyberbiz.exception.DecodeException when the text is not a plain decimal
   *     within the SDK's limits
   */
  public static Money of(String text) {
    return of(Json.decimal(Objects.requireNonNull(text, "text").strip()));
  }

  /**
   * Converts a whole amount.
   *
   * @param value the amount in whole currency units
   * @return the amount with two decimal places
   */
  public static Money of(long value) {
    return new Money(BigDecimal.valueOf(value).setScale(SCALE));
  }

  /**
   * Converts an exact decimal.
   *
   * @param value the amount; at most 40 significant digits and 32 decimal places
   * @return the amount, rounded to two places half away from zero
   * @throws IllegalArgumentException when the value is outside those limits
   */
  public static Money of(BigDecimal value) {
    Objects.requireNonNull(value, "value");
    if (value.precision() > MAX_PRECISION || Math.abs(value.scale()) > 32) {
      throw new IllegalArgumentException(
          "amount exceeds " + MAX_PRECISION + " digits or 32 decimal places");
    }
    BigDecimal rounded = value.setScale(SCALE, RoundingMode.HALF_UP);
    return new Money(rounded.signum() == 0 ? ZERO.amount : rounded);
  }

  /** Returns zero, {@code 0.00}. */
  public static Money zero() {
    return ZERO;
  }

  /** Returns the amount as a decimal with exactly two places. */
  public BigDecimal amount() {
    return amount;
  }

  /**
   * Adds another amount.
   *
   * @param other the amount to add
   * @return the sum
   */
  public Money add(Money other) {
    return new Money(amount.add(other.amount));
  }

  /**
   * Subtracts another amount.
   *
   * @param other the amount to subtract
   * @return the difference
   */
  public Money subtract(Money other) {
    return new Money(amount.subtract(other.amount));
  }

  /** Whether the amount is below zero. */
  public boolean isNegative() {
    return amount.signum() < 0;
  }

  @Override
  public int compareTo(Money other) {
    return amount.compareTo(other.amount);
  }

  @Override
  public boolean equals(Object other) {
    return other instanceof Money money && amount.compareTo(money.amount) == 0;
  }

  @Override
  public int hashCode() {
    return amount.hashCode();
  }

  /** Returns the amount with two decimal places, e.g. {@code "199.00"}. */
  @Override
  public String toString() {
    return amount.toPlainString();
  }
}
