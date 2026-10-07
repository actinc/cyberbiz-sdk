package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.json.Json;
import cc.alphacore.cyberbiz.json.Times;
import cc.alphacore.cyberbiz.pagination.Pagination;
import com.google.gson.JsonElement;
import com.google.gson.JsonPrimitive;
import java.math.BigDecimal;
import java.nio.file.Path;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.regex.Pattern;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

/**
 * The Golden File contract test (ADR-0002): every registered file in {@code testdata/golden/} must
 * decode into its model. A resource ticket adds one line per model to {@link #MODELS}; the probe
 * records below only prove the harness and are replaced by the real models.
 */
class GoldenTest {

  /** One Golden File and the type it decodes into; {@code list} for a JSON array body. */
  record Case(String file, Class<?> type, boolean list) {
    static Case object(String file, Class<?> type) {
      return new Case(file, type, false);
    }

    static Case list(String file, Class<?> itemType) {
      return new Case(file, itemType, true);
    }

    @Override
    public String toString() {
      return file + " -> " + (list ? "List<" + type.getSimpleName() + ">" : type.getSimpleName());
    }
  }

  record ShopProbe(ShopInfoProbe shopInfo) {}

  record ShopInfoProbe(long id, String currency, String language) {}

  record ProductProbe(
      long id,
      String title,
      BigDecimal price,
      OffsetDateTime createdAt,
      List<VariantProbe> productVariants) {}

  record VariantProbe(
      long id,
      BigDecimal price,
      BigDecimal cost,
      OffsetDateTime createdAt,
      OffsetDateTime updatedAt) {}

  /** Every Golden File with a model; add one line per model. */
  static final List<Case> MODELS =
      List.of(
          Case.object("app/GET_shop.json", ShopProbe.class),
          Case.list("v1/GET_v1_products.json", ProductProbe.class));

  /** A CYBERBIZ timestamp anywhere in a string value. */
  private static final Pattern TIMESTAMP =
      Pattern.compile("\\d{4}-\\d{2}-\\d{2}[ T]\\d{2}:\\d{2}.*");

  static Stream<Case> models() {
    return MODELS.stream();
  }

  static Stream<String> files() {
    return Golden.files().stream();
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("models")
  void everyRegisteredFileDecodesIntoItsModel(Case c) {
    Object decoded = c.list() ? Golden.list(c.file(), c.type()) : Golden.object(c.file(), c.type());

    assertNotNull(decoded, c.file());
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("files")
  void everyGoldenFileIsStrictJsonWithExactNumbersAndTaipeiTimes(String file) {
    List<JsonPrimitive> values = new ArrayList<>();
    collect(Json.decode(Golden.body(file), JsonElement.class), values);

    for (JsonPrimitive value : values) {
      if (value.isNumber()) {
        BigDecimal exact = Json.decode(value.toString(), BigDecimal.class);
        assertEquals(new BigDecimal(value.getAsString()), exact, file);
      } else if (value.isString() && TIMESTAMP.matcher(value.getAsString()).matches()) {
        OffsetDateTime time = Times.parse(value.getAsString());
        assertEquals(ZoneOffset.ofHours(8), time.getOffset(), file);
        assertEquals(value.getAsString(), Times.format(time), file);
      }
    }
  }

  @Test
  void theHarnessReadsTheRepositoryRootNotACopy() {
    assertTrue(Golden.ROOT.endsWith(Path.of("testdata", "golden")), Golden.ROOT.toString());
    assertFalse(Golden.ROOT.startsWith(Path.of("").toAbsolutePath()), Golden.ROOT.toString());
    assertTrue(Golden.files().size() > 100, "Golden Files found: " + Golden.files().size());
  }

  @Test
  void decodesAProbeModelWithExactMoneyAndTaipeiTimes() {
    ProductProbe product = Golden.list("v1/GET_v1_products.json", ProductProbe.class).get(0);
    VariantProbe variant = product.productVariants().get(0);

    assertEquals(new BigDecimal("200.0"), variant.price());
    assertEquals(new BigDecimal("190.0"), variant.cost());
    assertEquals(OffsetDateTime.parse("2026-07-10T20:22:05+08:00"), variant.createdAt());
    assertEquals(26721, Golden.object("app/GET_shop.json", ShopProbe.class).shopInfo().id());
  }

  @Test
  void readsTheRecordedPaginationHeaders() {
    Pagination pagination = Pagination.from(Golden.response("v1/GET_v1_products.json"));

    assertEquals(new Pagination(1, 2, 0, 195, 98, 2, 0), pagination);
    assertTrue(pagination.hasNext());
  }

  record OrderProbe(long id, List<LineProbe> lineItems) {}

  record LineProbe(BigDecimal price) {}

  @Test
  void aMismatchNamesTheFileAndTheField() {
    String fixture =
        Path.of("src/test/resources/golden-fixtures/wrong-type.json").toAbsolutePath().toString();

    DecodeException e =
        assertThrows(DecodeException.class, () -> Golden.object(fixture, OrderProbe.class));

    assertTrue(e.getMessage().startsWith("wrong-type.json: "), e.getMessage());
    assertTrue(e.getMessage().contains("$.line_items[1].price"), e.getMessage());
  }

  /** Collects every scalar of a JSON tree. */
  private static void collect(JsonElement element, List<JsonPrimitive> out) {
    if (element == null || element.isJsonNull()) {
      return;
    }
    if (element.isJsonPrimitive()) {
      out.add(element.getAsJsonPrimitive());
    } else if (element.isJsonArray()) {
      element.getAsJsonArray().forEach(child -> collect(child, out));
    } else {
      element.getAsJsonObject().entrySet().stream()
          .map(Map.Entry::getValue)
          .forEach(child -> collect(child, out));
    }
  }

  @Test
  void anUntypedGoldenFileKeepsNumbersAsBigDecimal() {
    Object products = Json.decode(Golden.body("v1/GET_v1_products.json"), Object.class);

    Map<?, ?> first = (Map<?, ?>) ((List<?>) products).get(0);
    assertInstanceOf(BigDecimal.class, first.get("price"));
  }
}
