package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.http.TransportRequest;
import cc.alphacore.cyberbiz.model.Product;
import cc.alphacore.cyberbiz.model.ProductOption;
import cc.alphacore.cyberbiz.model.ProductVariant;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.resource.ProductService;
import java.math.BigDecimal;
import java.net.URI;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Function;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.CsvSource;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;

class ProductServiceTest {
  private static final String BASE = "https://api.example.test/";

  private final FakeTransport transport = new FakeTransport();

  private ProductService products() {
    return CyberbizClient.builder("synthetic-token-0123456789")
        .baseUrl(URI.create("https://api.example.test"))
        .transport(transport)
        .rateLimit(0)
        .build()
        .products();
  }

  private TransportRequest sent() {
    return transport.requests.get(0);
  }

  /** Every service method: what it sends (method, path and query, body) for a canned reply. */
  static Stream<Arguments> requestShapes() {
    return Stream.of(productShapes(), extraShapes(), variantShapes(), optionShapes())
        .flatMap(shapes -> shapes);
  }

  private static Stream<Arguments> productShapes() {
    return Stream.of(
        shape("list", s -> s.list(), "[]", "GET v1/products", null),
        shape("list page", s -> s.list(Map.of("page", 2)), "[]", "GET v1/products?page=2", null),
        shape(
            "all", s -> s.all().stream().count(), "[]", "GET v1/products?page=1&per_page=50", null),
        shape("get", s -> s.get(3), "{}", "GET v1/products/3", null),
        shape(
            "create",
            s -> s.create(Map.of("title", "T")),
            "{}",
            "POST v1/products",
            "{\"title\":\"T\"}"),
        shape(
            "update",
            s -> s.update(3, Map.of("title", "N")),
            "{}",
            "PUT v1/products/3",
            "{\"title\":\"N\"}"),
        shape("delete", s -> run(() -> s.delete(3)), "", "DELETE v1/products/3", null),
        shape(
            "pos shop batch",
            s -> s.createForPosShops(Map.of("pos_shop_ids", List.of(1, 2))),
            "[]",
            "POST v1/products/pos_shop_batch",
            "{\"pos_shop_ids\":[1,2]}"),
        shape(
            "search",
            s -> s.search(Map.of("q", "綠茶")),
            "[]",
            "GET v1/products/search?q=%E7%B6%A0%E8%8C%B6",
            null),
        shape(
            "search collection",
            s -> s.searchCollection(Map.of("collection_handle", "tea")),
            "[]",
            "GET v1/products/search/collection?collection_handle=tea",
            null));
  }

  private static Stream<Arguments> extraShapes() {
    return Stream.of(
        shape(
            "seo meta tags",
            s -> s.updateSeoMetaTags(3, Map.of("title", "T")),
            "{}",
            "PUT v1/products/3/seo_meta_tags",
            "{\"title\":\"T\"}"),
        shape("list tags", s -> s.listTags(3), "[]", "GET v1/products/3/product_tags", null),
        shape(
            "add tags",
            s -> s.addTags(3, List.of("a", "b")),
            "[]",
            "PUT v1/products/3/product_tags/add",
            "{\"tags\":\"a,b\"}"),
        shape(
            "remove tags",
            s -> s.removeTags(3, List.of("a")),
            "[]",
            "PUT v1/products/3/product_tags/remove",
            "{\"tags\":\"a\"}"),
        shape(
            "bindable shippings",
            s -> s.listBindableShippings(),
            "{}",
            "GET v1/products/bind_shippings",
            null),
        shape(
            "bound shippings",
            s -> s.getBindShippings(3),
            "{}",
            "GET v1/products/3/bind_shippings",
            null),
        shape(
            "bind shippings",
            s -> s.bindShippings(3, List.of("宅配")),
            "{}",
            "POST v1/products/3/bind_shippings",
            "{\"shipping_names\":[\"宅配\"]}"),
        shape(
            "description setting names",
            s -> s.listDescriptionSettingNames(),
            "[]",
            "GET v1/products/get_product_description_setting_names",
            null));
  }

  private static Stream<Arguments> variantShapes() {
    return Stream.of(
        shape(
            "list variants",
            s -> s.listVariants(3),
            "[]",
            "GET v1/products/3/product_variants",
            null),
        shape(
            "get variant",
            s -> s.getVariant(3, 4),
            "{}",
            "GET v1/products/3/product_variants/4",
            null),
        shape(
            "create variant",
            s -> s.createVariant(3, Map.of("sku", "S1")),
            "{}",
            "POST v1/products/3/product_variants",
            "{\"sku\":\"S1\"}"),
        shape(
            "update variant",
            s -> s.updateVariant(3, 4, Map.of("sku", "S2")),
            "{}",
            "PUT v1/products/3/product_variants/4",
            "{\"sku\":\"S2\"}"),
        shape(
            "delete variant",
            s -> run(() -> s.deleteVariant(3, 4)),
            "",
            "DELETE v1/products/3/product_variants/4",
            null),
        shape(
            "variants by sku",
            s -> s.listVariantsBySku("A/B C"),
            "[]",
            "GET v1/products/sku/A%2FB%20C/product_variants",
            null),
        shape(
            "variants by sku page",
            s -> s.listVariantsBySku("S1", Map.of("per_page", 10)),
            "[]",
            "GET v1/products/sku/S1/product_variants?per_page=10",
            null),
        shape(
            "all variants by sku",
            s -> s.allVariantsBySku("S1").stream().count(),
            "[]",
            "GET v1/products/sku/S1/product_variants?page=1&per_page=50",
            null),
        shape(
            "all variants by sku from a page",
            s -> s.allVariantsBySku("S1", Map.of("page", 3)).stream().count(),
            "[]",
            "GET v1/products/sku/S1/product_variants?page=3&per_page=50",
            null));
  }

  private static Stream<Arguments> optionShapes() {
    return Stream.of(
        shape(
            "list options", s -> s.listOptions(3), "[]", "GET v1/products/3/product_options", null),
        shape(
            "get option",
            s -> s.getOption(3, 5),
            "{}",
            "GET v1/products/3/product_options/5",
            null),
        shape(
            "create option",
            s -> s.createOption(3, Map.of("name", "Size")),
            "{}",
            "POST v1/products/3/product_options",
            "{\"name\":\"Size\"}"),
        shape(
            "update option",
            s -> s.updateOption(3, 5, Map.of("name", "Color")),
            "{}",
            "PUT v1/products/3/product_options/5",
            "{\"name\":\"Color\"}"),
        shape(
            "delete option",
            s -> run(() -> s.deleteOption(3, 5)),
            "",
            "DELETE v1/products/3/product_options/5",
            null));
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("requestShapes")
  void sendsTheDocumentedRequest(
      String name, Function<ProductService, Object> call, String reply, String line, String body) {
    transport.reply(200, reply);

    var unused = call.apply(products());

    assertEquals(line, sent().method() + " " + sent().uri().toString().substring(BASE.length()));
    assertEquals(body, sent().body(), name);
  }

  @Test
  void createSendsAmountsAsExactDecimals() {
    transport.reply(200, "{\"id\":9,\"title\":\"Tea\",\"price\":120.50}");
    Map<String, Object> product = new LinkedHashMap<>();
    product.put("title", "Tea");
    product.put("handle", "tea");
    product.put("published", true);
    product.put("price", Money.of("120.5"));
    product.put("tiny", new BigDecimal("0.1"));

    Product created = products().create(product);

    assertEquals(
        "{\"title\":\"Tea\",\"handle\":\"tea\",\"published\":true,\"price\":120.50,\"tiny\":0.1}",
        sent().body());
    assertEquals(9, created.id());
    assertEquals(Money.of("120.50"), created.price());
  }

  @Test
  void listDecodesTheGoldenPageWithItsPagination() {
    transport.reply(
        200, Golden.body("v1/GET_v1_products.json"), "X-Total", "195", "X-Next-Page", "2");

    Page<Product> page = products().list();

    assertEquals(2, page.items().size());
    assertEquals(195, page.pagination().total());
    assertTrue(page.hasNext());
  }

  @Test
  void getSetsTheIdTheDetailResponseOmits() {
    transport.reply(200, Golden.body("v1/GET_v1_products_{id}.json"));

    assertEquals(42, products().get(42).id());
  }

  @Test
  void readsAndWritesByIdKeepTheRequestedId() {
    transport
        .reply(200, Golden.body("v1/GET_v1_products_{id}_product_variants_{id}.json"))
        .reply(200, Golden.body("v1/GET_v1_products_{id}_product_options_{id}.json"))
        .reply(200, "{}")
        .reply(200, "{}")
        .reply(200, "{}")
        .reply(200, "{}");
    ProductService s = products();

    ProductVariant variant = s.getVariant(3, 4);
    ProductOption option = s.getOption(3, 5);

    assertEquals(4, variant.id());
    assertEquals(5, option.id());
    assertEquals(3, s.update(3, Map.of()).id());
    assertEquals(3, s.updateSeoMetaTags(3, Map.of()).id());
    assertEquals(4, s.updateVariant(3, 4, Map.of()).id());
    assertEquals(5, s.updateOption(3, 5, Map.of()).id());
  }

  @Test
  void aNullBodyForAReadByIdIsNotFound() {
    transport.reply(200, "null", "X-Request-Id", "req-1");

    NotFoundException e = assertThrows(NotFoundException.class, () -> products().get(3));

    assertEquals(404, e.statusCode());
    assertEquals("/v1/products/3", e.path());
    assertEquals("req-1", e.requestId());
  }

  @Test
  void decodesTheShippingNamesGoldenFiles() {
    transport
        .reply(200, Golden.body("v1/GET_v1_products_bind_shippings.json"))
        .reply(200, Golden.body("v1/GET_v1_products_{id}_bind_shippings.json"));

    assertEquals(4, products().listBindableShippings().size());
    assertEquals("門市取貨（預設）", products().getBindShippings(3).get(0));
  }

  @ParameterizedTest(name = "\"{0}\" -> {1}")
  @CsvSource({
    "a/b, a%2Fb",
    "%2e%2e, %252e%252e",
    "..., ...",
    "SKU-f05512, SKU-f05512",
  })
  void encodesTheSkuAsOneSegment(String sku, String encoded) {
    transport.reply(200, "[]");

    var unused = products().listVariantsBySku(sku);

    assertEquals(
        BASE + "v1/products/sku/" + encoded + "/product_variants", sent().uri().toString());
  }

  @ParameterizedTest(name = "\"{0}\"")
  @ValueSource(strings = {"", ".", ".."})
  void rejectsASkuThatWouldAddressAnotherPath(String sku) {
    ProductService s = products();

    assertThrows(IllegalArgumentException.class, () -> s.listVariantsBySku(sku));
    assertThrows(IllegalArgumentException.class, () -> s.allVariantsBySku(sku));
    assertTrue(transport.requests.isEmpty());
  }

  @Test
  void rejectsAnOverlongSkuWithoutEchoingIt() {
    String sku = "x".repeat(257);

    IllegalArgumentException e =
        assertThrows(IllegalArgumentException.class, () -> products().listVariantsBySku(sku));

    assertFalse(e.getMessage().contains(sku.substring(0, 20)), e.getMessage());
    assertTrue(transport.requests.isEmpty());
  }

  @Test
  void acceptsASkuOfExactlyTheLimit() {
    transport.reply(200, "[]");

    var unused = products().listVariantsBySku("x".repeat(256));

    assertEquals(1, transport.requests.size());
  }

  private static Arguments shape(
      String name, Function<ProductService, Object> call, String reply, String line, String body) {
    return Arguments.of(name, call, reply, line, body);
  }

  private static Object run(Runnable action) {
    action.run();
    return null;
  }
}
