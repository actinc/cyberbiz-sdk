package cc.alphacore.cyberbiz.json;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.DecodeException;
import com.google.gson.annotations.SerializedName;
import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;

class TypesTest {
  private static final ZoneOffset TAIPEI = ZoneOffset.ofHours(8);

  /** A synthetic model using every type the SDK configures. */
  record Line(
      long id,
      BigDecimal price,
      BigDecimal compareAtPrice,
      OffsetDateTime createdAt,
      @SerializedName("qc") String qualityCode,
      List<String> tags) {}

  record Amount(BigDecimal value) {}

  record Stamp(OffsetDateTime at) {}

  private static BigDecimal amount(String json) {
    return Json.decode("{\"value\":" + json + "}", Amount.class).value();
  }

  private static OffsetDateTime time(String text) {
    return Json.decode("{\"at\":\"" + text + "\"}", Stamp.class).at();
  }

  @Test
  void decodesSnakeCaseIntoARecordAndIgnoresUnknownFields() {
    Line line =
        Json.decode(
            "{\"id\":7,\"price\":200.0,\"compare_at_price\":\"250\",\"created_at\":\"2026-07-10"
                + " 20:22:05\",\"qc\":\"SKU-000001\",\"tags\":[\"a\"],\"new_field\":{\"x\":1}}",
            Line.class);

    assertEquals(7, line.id());
    assertEquals(new BigDecimal("200.0"), line.price());
    assertEquals(new BigDecimal("250"), line.compareAtPrice());
    assertEquals(OffsetDateTime.of(2026, 7, 10, 20, 22, 5, 0, TAIPEI), line.createdAt());
    assertEquals("SKU-000001", line.qualityCode());
    assertEquals(List.of("a"), line.tags());
  }

  @Test
  void zeroPointOneIsExactNotADouble() {
    BigDecimal decoded = amount("0.1");

    assertEquals(new BigDecimal("0.1"), decoded);
    assertNotEquals(0.3, 0.1 + 0.2);
    assertEquals(0, amount("0.1").add(amount("0.2")).compareTo(new BigDecimal("0.3")));
  }

  @ParameterizedTest
  @CsvSource({
    "120.50, 120.50",
    "'\"120.50\"', 120.50",
    "'\" 99 \"', 99",
    "3690.0, 3690.0",
    "-0.005, -0.005",
    "0.10000000000000000000001, 0.10000000000000000000001",
    "12345678901234567890.123456789, 12345678901234567890.123456789",
    "'\"98765432109876543210987654321.01\"', 98765432109876543210987654321.01",
    "123456789012345678901234567890, 123456789012345678901234567890",
  })
  void amountsAreExactFromNumbersAndStrings(String json, String expected) {
    BigDecimal decoded = amount(json);

    assertEquals(new BigDecimal(expected), decoded);
    assertEquals(new BigDecimal(expected).scale(), decoded.scale());
  }

  @ParameterizedTest
  @ValueSource(strings = {"null", "\"\"", "\"  \""})
  void aNullOrEmptyAmountIsNull(String json) {
    assertNull(amount(json));
  }

  @ParameterizedTest
  @ValueSource(strings = {"\"12 TWD\"", "\"NaN\"", "true", "[1]", "{}"})
  void aNonNumericAmountFailsWithItsPath(String json) {
    DecodeException e = assertThrows(DecodeException.class, () -> amount(json));

    assertTrue(e.getMessage().startsWith("cyberbiz: "), e.getMessage());
    assertTrue(e.getMessage().contains("$.value"), e.getMessage());
  }

  /** Values that parse cheaply but would exhaust memory or CPU in toPlainString or arithmetic. */
  static Stream<String> hugeDecimals() {
    String tenThousandDigits = "9".repeat(10_000);
    return Stream.of(
        "1e999999999",
        "1E-2147483647",
        "\"1e999999999\"",
        "\"1E-2147483647\"",
        "1e3",
        tenThousandDigits,
        "\"" + tenThousandDigits + "\"",
        "0." + "0".repeat(40) + "1",
        "1" + "0".repeat(41));
  }

  @ParameterizedTest
  @MethodSource("hugeDecimals")
  void aHugeOrExponentDecimalFailsClosed(String json) {
    DecodeException e = assertThrows(DecodeException.class, () -> amount(json));

    assertTrue(e.getMessage().contains("$.value"), e.getMessage());
    assertTrue(e.getMessage().length() < 300, "the message must not echo the value");
    if (!json.startsWith("\"")) {
      // Untyped values keep strings as strings; only JSON numbers become BigDecimal.
      assertThrows(
          DecodeException.class, () -> Json.decode("{\"value\":" + json + "}", Object.class));
    }
  }

  @Test
  void decimalsAtTheLimitsStillPass() {
    String fortyDigits = "1234567890".repeat(4);
    String thirtyTwoPlaces = "0." + "1".repeat(32);

    assertEquals(new BigDecimal(fortyDigits), amount(fortyDigits));
    assertEquals(new BigDecimal(thirtyTwoPlaces), amount("\"" + thirtyTwoPlaces + "\""));
    assertEquals(new BigDecimal("-9999.99"), amount("-9999.99"));
  }

  @Test
  void numbersInUntypedValuesStayExact() {
    Map<?, ?> map = (Map<?, ?>) Json.decode("{\"a\":0.1,\"b\":12345678901234567890}", Object.class);

    assertEquals(new BigDecimal("0.1"), map.get("a"));
    assertEquals(new BigDecimal("12345678901234567890"), map.get("b"));
  }

  @Test
  void amountsAreEncodedAsPlainExactNumbers() {
    assertEquals("{\"value\":1000}", Json.encode(new Amount(new BigDecimal("1E+3"))));
    assertEquals("{\"value\":0.10}", Json.encode(new Amount(new BigDecimal("0.10"))));
  }

  @ParameterizedTest
  @CsvSource({
    "2026-09-01 10:00:00, 2026-09-01T10:00+08:00",
    "2026-09-01T10:00:00, 2026-09-01T10:00+08:00",
    "2026-09-01 10:00, 2026-09-01T10:00+08:00",
    "2026-09-01 10:00:00.5, 2026-09-01T10:00:00.5+08:00",
    "2026-09-01, 2026-09-01T00:00+08:00",
    "2026-09-01T02:00:00Z, 2026-09-01T10:00+08:00",
    "2026-09-01T10:00:00+08:00, 2026-09-01T10:00+08:00",
    "2026-09-01T10:00:00.123+08:00, 2026-09-01T10:00:00.123+08:00",
    "2026-09-01 10:00:00 +0800, 2026-09-01T10:00+08:00",
    "2026-09-01 10:00:00+0800, 2026-09-01T10:00+08:00",
    "2026-09-01T03:00:00+01:00, 2026-09-01T10:00+08:00",
    "2026-08-31T21:00:00-05:00, 2026-09-01T10:00+08:00",
    "2026-09-01 10:00+08, 2026-09-01T10:00+08:00",
  })
  void timesWithAndWithoutAZoneAreInTaipei(String text, String expected) {
    OffsetDateTime parsed = time(text);

    assertEquals(OffsetDateTime.parse(expected), parsed);
    assertEquals(TAIPEI, parsed.getOffset());
    assertEquals(parsed, Times.parse(text));
  }

  @ParameterizedTest
  @ValueSource(strings = {"", "  "})
  void anEmptyTimeIsNull(String text) {
    assertNull(time(text));
    assertNull(Times.parse(text));
    assertNull(Times.parse(null));
    assertNull(Json.decode("{\"at\":null}", Stamp.class).at());
  }

  @ParameterizedTest
  @ValueSource(strings = {"next tuesday", "2026-02-30 10:00:00", "2026-09-01 25:00:00", "1693"})
  void anUnknownTimeFormatFails(String text) {
    DecodeException e = assertThrows(DecodeException.class, () -> time(text));
    assertTrue(e.getMessage().contains("$.at"), e.getMessage());

    assertThrows(DecodeException.class, () -> Times.parse(text));
  }

  @Test
  void aNumericTimeFails() {
    assertThrows(DecodeException.class, () -> Json.decode("{\"at\":1693}", Stamp.class));
  }

  @Test
  void timesAreEncodedInThePlatformLayoutInTaipei() {
    OffsetDateTime utc = OffsetDateTime.of(2026, 9, 1, 2, 0, 0, 0, ZoneOffset.UTC);

    assertEquals("{\"at\":\"2026-09-01 10:00:00\"}", Json.encode(new Stamp(utc)));
    assertEquals("2026-09-01 10:00:00", Times.format(utc));
    assertEquals("Asia/Taipei", Times.ZONE.getId());
  }

  @ParameterizedTest
  @ValueSource(strings = {"{\"id\":", "{id:1}", "{\"id\":1} trailing", "[1,]"})
  void malformedJsonFails(String json) {
    assertThrows(DecodeException.class, () -> Json.decode(json, Line.class));
  }

  @Test
  void aWrongTypeNamesTheField() {
    DecodeException e =
        assertThrows(DecodeException.class, () -> Json.decode("{\"id\":\"seven\"}", Line.class));

    assertTrue(e.getMessage().contains("$.id"), e.getMessage());
  }

  @Test
  void decodeListAcceptsNullAndEmptyAndRejectsObjects() {
    assertEquals(List.of(), Json.decodeList("null", Amount.class));
    assertEquals(List.of(), Json.decodeList("", Amount.class));
    assertEquals(
        List.of(new Amount(new BigDecimal("1.50"))),
        Json.decodeList("[{\"value\":\"1.50\"}]", Amount.class));
    DecodeException e =
        assertThrows(DecodeException.class, () -> Json.decodeList("7", Amount.class));
    assertEquals("cyberbiz: expected a JSON array, got a scalar", e.getMessage());
  }

  @Test
  void anEmptyDocumentDecodesToNull() {
    assertNull(Json.decode("", Line.class));
    assertNull(Json.decode("null", Line.class));
  }
}
