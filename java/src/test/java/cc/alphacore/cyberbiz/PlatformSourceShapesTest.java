package cc.alphacore.cyberbiz;

import static cc.alphacore.cyberbiz.ServiceHarness.map;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import cc.alphacore.cyberbiz.exception.ApiException;
import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.json.Json;
import cc.alphacore.cyberbiz.model.CustomerMessageComment;
import cc.alphacore.cyberbiz.model.CustomerMessagePost;
import cc.alphacore.cyberbiz.model.CustomerOAuthApp;
import cc.alphacore.cyberbiz.model.OrderBranchStore;
import cc.alphacore.cyberbiz.model.SupportShippingTask;
import cc.alphacore.cyberbiz.resource.ProductService;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import org.junit.jupiter.api.Test;

/**
 * Request and response shapes confirmed against the platform source rather than the recorded Golden
 * Files (CBSDK-44/45 in Go, ported by CBSDK-52).
 */
class PlatformSourceShapesTest {

  @Test
  void getByUidProviderSendsProviderTypeAndProvider() {
    ServiceHarness h = ServiceHarness.json("{\"id\":1}");

    h.client.customers().getByUidProvider("line", "U1");

    assertEquals(
        "GET v2/customers/by_uid_provider?uid=U1&provider_type=line&provider=line", h.line());
  }

  @Test
  void createOAuthAppSendsTheAppAndReturnsItsCredentials() {
    ServiceHarness h = new ServiceHarness();
    h.transport.reply(201, "{\"client_id\":\"cid\",\"client_secret\":\"sec\"}");

    CustomerOAuthApp app =
        h.client
            .customers()
            .createOAuthApp(
                map(
                    "app_name", "會員登入",
                    "scopes_description", "讀取會員資料",
                    "description", "d",
                    "redirect_uri", "https://example.com/cb",
                    "display_shop_name", true));

    assertEquals("POST v2/customer_oauth", h.line());
    assertEquals(
        "{\"app_name\":\"會員登入\",\"scopes_description\":\"讀取會員資料\",\"description\":\"d\","
            + "\"redirect_uri\":\"https://example.com/cb\",\"display_shop_name\":true}",
        h.body(0));
    assertEquals("cid", app.clientId());
    assertEquals("sec", app.clientSecret());
    assertFalse(app.toString().contains("sec"));
  }

  @Test
  void createOAuthAppThrowsForASuccessFalseReply() {
    ServiceHarness h =
        ServiceHarness.json("{\"success\":false,\"errors\":[\"redirect uri invalid\"]}");

    ApiException e =
        assertThrows(
            ApiException.class, () -> h.client.customers().createOAuthApp(map("app_name", "a")));

    assertEquals(200, e.statusCode());
    assertEquals("POST", e.method());
    assertEquals("/v2/customer_oauth", e.path());
    assertEquals(List.of("redirect uri invalid"), e.messages());
  }

  @Test
  void productCreateAndPosBatchSendTheSku() {
    ServiceHarness h = ServiceHarness.json("{\"id\":1}", "[{\"id\":2}]");
    ProductService products = h.client.products();

    products.create(
        map("title", "T", "handle", "t", "published", false, "price", 100, "sku", "SKU-1"));
    products.createForPosShops(map("title", "T", "pos_shop_ids", "3", "sku", "SKU-1"));

    assertEquals(
        "{\"title\":\"T\",\"handle\":\"t\",\"published\":false,\"price\":100,\"sku\":\"SKU-1\"}",
        h.body(0));
    assertEquals("POST v1/products/pos_shop_batch", h.line(1));
    assertEquals("{\"title\":\"T\",\"pos_shop_ids\":\"3\",\"sku\":\"SKU-1\"}", h.body(1));
  }

  @Test
  void productUpdateKeepsTheCurrentTaxAndTemperature() {
    ServiceHarness h =
        ServiceHarness.json(
            "{\"title\":\"old\",\"tax_type_id\":\"zero_tax\",\"temperature_types\":[\"冷凍\"]}",
            "{\"title\":\"new\"}");
    Map<String, Object> changes = map("title", "new");

    h.client.products().update(7, changes);

    assertEquals("GET v1/products/7", h.line(0));
    assertEquals("PUT v1/products/7", h.line(1));
    assertEquals(
        "{\"title\":\"new\",\"tax_type_id\":\"zero_tax\",\"temperature_types\":[\"冷凍\"]}",
        h.body(1));
    assertEquals(Map.of("title", "new"), changes);
  }

  @Test
  void productUpdateSkipsTheReadWhenBothAreSet() {
    ServiceHarness h = ServiceHarness.json("{}");

    h.client
        .products()
        .update(7, map("tax_type_id", "exclusive_tax", "temperature_types", List.of("冷藏")));

    assertEquals(1, h.transport.requests.size());
    assertEquals("PUT v1/products/7", h.line());
    assertEquals("{\"tax_type_id\":\"exclusive_tax\",\"temperature_types\":[\"冷藏\"]}", h.body(0));
  }

  @Test
  void productUpdateFailsWhenTheReadFails() {
    ServiceHarness h = ServiceHarness.json("null");
    ProductService products = h.client.products();

    assertThrows(
        NotFoundException.class,
        () -> products.update(7, map("title", "new", "tax_type_id", "zero_tax")));
    assertEquals(1, h.transport.requests.size());
    assertEquals("GET v1/products/7", h.line());
  }

  @Test
  void customerVerificationTimesAreSentAsRfc3339() {
    ServiceHarness h = ServiceHarness.json("{\"id\":1}", "{\"id\":1}");
    OffsetDateTime at = OffsetDateTime.parse("2026-01-01T19:04:05Z");

    h.client.customers().create(map("name", "n", "confirmed_at", at));
    h.client
        .customers()
        .update(
            1, map("confirmed_at", null, "mobile_sms_confirmed_at", at, "birthday", "2000-01-02"));

    assertEquals("{\"name\":\"n\",\"confirmed_at\":\"2026-01-02T03:04:05+08:00\"}", h.body(0));
    assertEquals(
        "{\"confirmed_at\":null,\"mobile_sms_confirmed_at\":\"2026-01-02T03:04:05+08:00\","
            + "\"birthday\":\"2000-01-02\"}",
        h.body(1));
  }

  @Test
  void messagePostOrderAndAdminAreObjects() {
    ServiceHarness h =
        ServiceHarness.json(
            "[{\"id\":1,\"order\":{\"id\":9,\"type\":\"Order\",\"name\":\"#1001\"},"
                + "\"comments\":[{\"id\":2,\"role\":\"admin\",\"admin\":{\"id\":3,"
                + "\"name\":\"店長\",\"email\":\"admin@example.com\"}},"
                + "{\"id\":4,\"role\":\"customer\",\"admin\":null}]},"
                + "{\"id\":5,\"order\":null}]");

    List<CustomerMessagePost> posts = h.client.customers().listMessagePosts(1).items();

    CustomerMessagePost post = posts.get(0);
    assertEquals(9L, post.order().id());
    assertEquals("Order", post.order().type());
    assertEquals("#1001", post.order().name());
    CustomerMessageComment admin = post.comments().get(0);
    assertEquals(3L, admin.admin().id());
    assertEquals("店長", admin.admin().name());
    assertEquals("admin@example.com", admin.admin().email());
    assertNull(post.comments().get(1).admin());
    assertNull(posts.get(1).order());
  }

  @Test
  void supportShippingTrackingNumbersAreAList() {
    SupportShippingTask task =
        Json.decode(
            "{\"order_id\":1,\"tracking_numbers\":[{\"tracking_company\":\"tcat\","
                + "\"tracking_number\":\"T1\"},{\"tracking_company\":\"tcat\","
                + "\"tracking_number\":\"T2\"}]}",
            SupportShippingTask.class);
    SupportShippingTask empty = Json.decode("{\"order_id\":2}", SupportShippingTask.class);

    assertEquals(2, task.trackingNumbers().size());
    assertEquals("T2", task.trackingNumbers().get(1).trackingNumber());
    assertEquals(List.of(), empty.trackingNumbers());
  }

  @Test
  void branchStoreCoordinatesAcceptDecimalStrings() {
    OrderBranchStore text =
        Json.decode("{\"lat\":\"25.0330000\",\"lng\":\"121.5654000\"}", OrderBranchStore.class);
    OrderBranchStore number =
        Json.decode("{\"lat\":25.033,\"lng\":121.5654}", OrderBranchStore.class);
    OrderBranchStore blank = Json.decode("{\"lat\":null,\"lng\":\"\"}", OrderBranchStore.class);

    assertEquals(25.033, text.lat());
    assertEquals(121.5654, text.lng());
    assertEquals(25.033, number.lat());
    assertNull(blank.lat());
    assertNull(blank.lng());
    assertThrows(
        DecodeException.class, () -> Json.decode("{\"lat\":\"north\"}", OrderBranchStore.class));
  }
}
