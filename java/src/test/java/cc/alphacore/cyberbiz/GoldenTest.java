package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.json.Json;
import cc.alphacore.cyberbiz.json.Times;
import cc.alphacore.cyberbiz.model.AppSettings;
import cc.alphacore.cyberbiz.model.Customer;
import cc.alphacore.cyberbiz.model.CustomerIdMatch;
import cc.alphacore.cyberbiz.model.CustomerMessagePost;
import cc.alphacore.cyberbiz.model.CustomerNameMatch;
import cc.alphacore.cyberbiz.model.CustomerSpendingOverview;
import cc.alphacore.cyberbiz.model.CustomerUidLookup;
import cc.alphacore.cyberbiz.model.CustomerVipInfo;
import cc.alphacore.cyberbiz.model.Fulfillment;
import cc.alphacore.cyberbiz.model.LineItem;
import cc.alphacore.cyberbiz.model.Order;
import cc.alphacore.cyberbiz.model.OrderNumberId;
import cc.alphacore.cyberbiz.model.OrderReturn;
import cc.alphacore.cyberbiz.model.OrderTransaction;
import cc.alphacore.cyberbiz.model.Product;
import cc.alphacore.cyberbiz.model.ProductDescriptionSettingName;
import cc.alphacore.cyberbiz.model.ProductOption;
import cc.alphacore.cyberbiz.model.ProductPhoto;
import cc.alphacore.cyberbiz.model.ProductTag;
import cc.alphacore.cyberbiz.model.ProductVariant;
import cc.alphacore.cyberbiz.model.ShopInfo;
import cc.alphacore.cyberbiz.model.Tag;
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
 * decode into its model. A resource ticket adds one line per model and file to {@link #MODELS}.
 */
class GoldenTest {

  /**
   * One Golden File and the type it decodes into; {@code list} for a JSON array body, {@code
   * envelope} for an object wrapped under one key (null when unwrapped).
   */
  record Case(String file, Class<?> type, boolean list, String envelope) {
    static Case object(String file, Class<?> type) {
      return new Case(file, type, false, null);
    }

    static Case wrapped(String file, String envelope, Class<?> type) {
      return new Case(file, type, false, envelope);
    }

    static Case list(String file, Class<?> itemType) {
      return new Case(file, itemType, true, null);
    }

    Object decode() {
      if (list) {
        return Golden.list(file, type);
      }
      return envelope == null ? Golden.object(file, type) : Golden.wrapped(file, envelope, type);
    }

    @Override
    public String toString() {
      return file + " -> " + (list ? "List<" + type.getSimpleName() + ">" : type.getSimpleName());
    }
  }

  /** Every Golden File with a model; add one line per model. */
  static final List<Case> MODELS =
      List.of(
          Case.wrapped("app/GET_shop.json", "shop_info", ShopInfo.class),
          Case.wrapped("app/GET_settings.json", "shop_add_on", AppSettings.class),
          Case.list("v1/GET_v1_products.json", Product.class),
          Case.object("v1/GET_v1_products_{id}.json", Product.class),
          Case.list("v1/GET_v1_products_search_query.json", Product.class),
          Case.list("v1/GET_v1_products_search_collection.json", Product.class),
          Case.list("v1/GET_v1_products_{id}_product_variants.json", ProductVariant.class),
          Case.object("v1/GET_v1_products_{id}_product_variants_{id}.json", ProductVariant.class),
          Case.list("v1/GET_v1_products_sku_{id}_product_variants.json", ProductVariant.class),
          Case.list("v1/GET_v1_products_{id}_product_options.json", ProductOption.class),
          Case.object("v1/GET_v1_products_{id}_product_options_{id}.json", ProductOption.class),
          Case.list("v1/GET_v1_products_{id}_product_tags.json", ProductTag.class),
          Case.list("v1/GET_v1_products_{id}_product_photos.json", ProductPhoto.class),
          Case.list(
              "v1/GET_v1_products_get_product_description_setting_names.json",
              ProductDescriptionSettingName.class),
          Case.list("v1/GET_v1_orders.json", Order.class),
          Case.object("v1/GET_v1_orders_{id}.json", Order.class),
          Case.list("v1/GET_v1_orders_get_order_id.json", OrderNumberId.class),
          Case.list("v1/GET_v1_orders_{id}_fulfillments.json", Fulfillment.class),
          Case.object("v1/GET_v1_orders_{id}_fulfillments_{id}.json", Fulfillment.class),
          Case.list("v1/GET_v1_orders_{id}_returns.json", OrderReturn.class),
          Case.list("v1/GET_v1_orders_{id}_transactions.json", OrderTransaction.class),
          Case.list("v1/GET_v1_customers.json", Customer.class),
          Case.object("v1/GET_v1_customers_{id}.json", Customer.class),
          Case.list("v1/GET_v1_customers_get_customer_id.json", CustomerIdMatch.class),
          Case.list("v1/GET_v1_customers_get_customer_id_by_name.json", CustomerNameMatch.class),
          Case.list("v1/GET_v1_customers_tags.json", Tag.class),
          Case.list("v1/GET_v1_customers_{id}_message_posts.json", CustomerMessagePost.class),
          Case.list("v1/GET_v1_customers_{id}_orders.json", Order.class),
          Case.list("v1/GET_v1_customers_{id}_recent_purchases.json", LineItem.class),
          Case.list("v1/GET_v1_customers_{id}_customer_cart_items.json", ProductVariant.class),
          Case.object(
              "v1/GET_v1_customers_{id}_spending_overview.json", CustomerSpendingOverview.class),
          Case.object("v1/GET_v1_customers_{id}_uid_providers_line.json", CustomerUidLookup.class),
          Case.object("v1/GET_v1_customers_{id}_vip_info.json", CustomerVipInfo.class),
          Case.list("v2/GET_v2_customers.json", Customer.class),
          Case.list("v2/GET_v2_customers_include.json", Customer.class));

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
    Object decoded = c.decode();

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
  void decodesAProductWithExactMoneyAndTaipeiTimes() {
    Product product = Golden.list("v1/GET_v1_products.json", Product.class).get(0);
    ProductVariant variant = product.productVariants().get(0);

    assertEquals(Money.of("200"), variant.price());
    assertEquals(Money.of("190.00"), variant.cost());
    assertEquals(Money.of("200"), product.price());
    assertEquals(OffsetDateTime.parse("2026-07-10T20:22:05+08:00"), variant.createdAt());
    assertNull(product.englishTitle());
    assertTrue(product.productCustomFields().isJsonNull());
    assertThrows(UnsupportedOperationException.class, () -> product.tags().add(null));
  }

  @Test
  void decodesTheShopAndTheAppSettings() {
    ShopInfo shop = Golden.wrapped("app/GET_shop.json", "shop_info", ShopInfo.class);
    AppSettings settings =
        Golden.wrapped("app/GET_settings.json", "shop_add_on", AppSettings.class);

    assertEquals(26721, shop.id());
    assertTrue(shop.shopLine().loginEnable());
    assertNull(shop.shopLine().liffId());
    assertNull(shop.shopLineChatBot());
    assertTrue(settings.settings().isJsonObject());
    assertNull(settings.startAt());
    assertTrue(settings.addOnVersion().manifest().webhookEvents().contains("orders/paid"));
    assertTrue(settings.toString().contains("token=***"), settings.toString());
  }

  @Test
  void readsTheRecordedPaginationHeaders() {
    Pagination pagination = Pagination.from(Golden.response("v1/GET_v1_products.json"));

    assertEquals(new Pagination(1, 2, 0, 195, 98, 2, 0), pagination);
    assertTrue(pagination.hasNext());
  }

  @Test
  void aMismatchNamesTheFileAndTheField() {
    String fixture =
        Path.of("src/test/resources/golden-fixtures/wrong-type.json").toAbsolutePath().toString();

    DecodeException e =
        assertThrows(DecodeException.class, () -> Golden.object(fixture, Order.class));

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
