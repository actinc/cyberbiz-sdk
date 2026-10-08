package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.json.Json;
import java.math.BigDecimal;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

/** Mirrors the PHP SDK's Money tests (php/tests/TypesTest.php). */
class MoneyTest {

  /** A model with an amount, as CBSDK-33/34 models will have. */
  record Priced(Money price, Money compareAtPrice) {}

  @Test
  void isExactWhereDoubleIsNot() {
    assertEquals(Money.of("0.3"), Money.of("0.1").add(Money.of("0.2")));
    assertEquals("3690.00", Money.of("3690.0").toString());
    assertEquals("199.00", Money.of(199).toString());
    assertEquals("12345678901234567.89", Money.of("12345678901234567.89").toString());
  }

  @Test
  void roundsHalfAwayFromZero() {
    assertEquals("1.01", Money.of("1.005").toString());
    assertEquals("-1.01", Money.of("-1.005").toString());
    assertEquals("0.00", Money.of("-0.001").toString());
    assertTrue(Money.of("9.99").compareTo(Money.of("10")) < 0);
    assertTrue(Money.of("-0.5").isNegative());
    assertFalse(Money.zero().isNegative());
  }

  @Test
  void equalsIgnoresTheScaleTheApiSent() {
    assertEquals(Money.of("200"), Money.of("200.0"));
    assertEquals(Money.of("200").hashCode(), Money.of("200.000").hashCode());
    assertEquals(Money.of(new BigDecimal("200.0")), Money.of(200));
    assertEquals(Money.of(5), Money.of(8).subtract(Money.of(3)));
  }

  @ParameterizedTest
  @ValueSource(strings = {"12 TWD", "", "abc", "1e3", "1e999999999", "1.2.3"})
  void rejectsWhatIsNotAPlainDecimal(String text) {
    assertThrows(DecodeException.class, () -> Money.of(text));
  }

  @Test
  void rejectsAnUnboundedBigDecimal() {
    assertThrows(IllegalArgumentException.class, () -> Money.of(new BigDecimal("1e999999999")));
    assertThrows(IllegalArgumentException.class, () -> Money.of(new BigDecimal("1".repeat(41))));
  }

  @Test
  void decodesNumbersStringsAndNull() {
    Priced p = Json.decode("{\"price\":9999.0,\"compare_at_price\":\"120.5\"}", Priced.class);

    assertEquals("9999.00", p.price().toString());
    assertEquals("120.50", p.compareAtPrice().toString());
    assertNull(Json.decode("{\"price\":null,\"compare_at_price\":\"\"}", Priced.class).price());
  }

  @Test
  void decodingFailsClosedOnHugeOrExponentAmounts() {
    assertThrows(DecodeException.class, () -> Json.decode("{\"price\":1e999999999}", Priced.class));
    assertThrows(
        DecodeException.class,
        () -> Json.decode("{\"price\":\"" + "9".repeat(10_000) + "\"}", Priced.class));
  }

  @Test
  void encodesAPlainNumberWithTwoPlaces() {
    assertEquals(
        "{\"price\":120.50,\"compare_at_price\":0.00}",
        Json.encode(new Priced(Money.of("120.5"), Money.zero())));
  }
}
