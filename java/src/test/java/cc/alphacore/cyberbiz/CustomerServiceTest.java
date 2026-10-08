package cc.alphacore.cyberbiz;

import static cc.alphacore.cyberbiz.ServiceHarness.map;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.model.Customer;
import cc.alphacore.cyberbiz.model.CustomerNameMatch;
import cc.alphacore.cyberbiz.model.CustomerSpendingOverview;
import cc.alphacore.cyberbiz.model.CustomerUidLookup;
import cc.alphacore.cyberbiz.model.CustomerVipInfo;
import cc.alphacore.cyberbiz.model.LineItem;
import cc.alphacore.cyberbiz.model.Order;
import cc.alphacore.cyberbiz.model.ProductVariant;
import cc.alphacore.cyberbiz.model.Tag;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.resource.CustomerService;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.function.Function;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;

/** {@code client.customers()}: Golden File decoding and the request shape of every method. */
class CustomerServiceTest {

  private static OffsetDateTime taipei(String local) {
    return OffsetDateTime.parse(local + "+08:00");
  }

  @Test
  void listDecodesTheGoldenPage() {
    ServiceHarness h = ServiceHarness.golden("v1/GET_v1_customers.json");

    Page<Customer> page =
        h.client
            .customers()
            .list(map("per_page", 2, "updated_at_start_time", taipei("2026-01-01T00:00:00")));

    assertEquals(
        "GET v1/customers?per_page=2&updated_at_start_time=2026-01-01%2000%3A00%3A00", h.line());
    assertEquals(23296, page.pagination().total());
    Customer customer = page.items().get(0);
    assertEquals(23300813L, customer.id());
    assertEquals("enabled", customer.status());
    assertEquals(Money.of(10012), customer.bonusRemain());
    assertEquals(taipei("2023-10-03T12:04:02"), customer.createdAt());
    assertNull(customer.birthday());
    assertEquals("line", customer.uidProviders().get(0).providerType());
    assertNull(customer.vipInfo());
  }

  @Test
  void getDecodesTheGoldenCustomerAndSetsTheId() {
    ServiceHarness h = ServiceHarness.golden("v1/GET_v1_customers_{id}.json");

    Customer customer = h.client.customers().get(23300813);

    assertEquals("GET v1/customers/23300813", h.line());
    assertEquals(23300813L, customer.id());
    assertEquals(taipei("2023-11-30T22:30:58"), customer.confirmedAt());
    assertEquals(Money.zero(), customer.otherAccumulatedConsumption());
    assertNotNull(customer.address().detailAddress());
    assertEquals(List.of(), customer.tags());
  }

  @Test
  void getTurnsANullBodyIntoNotFound() {
    CustomerService customers = ServiceHarness.json("null").client.customers();

    assertThrows(NotFoundException.class, () -> customers.get(5));
  }

  @Test
  void lookupGoldenFiles() {
    ServiceHarness h =
        ServiceHarness.golden(
            "v1/GET_v1_customers_get_customer_id.json",
            "v1/GET_v1_customers_get_customer_id_by_name.json",
            "v1/GET_v1_customers_default_gender_options.json",
            "v1/GET_v1_customers_tags.json");
    CustomerService customers = h.client.customers();

    assertEquals(
        List.of(),
        customers.lookupIds(map("customer_emails", List.of("a@example.com", "b@example.com"))));
    List<CustomerNameMatch> byName = customers.lookupIdsByName("Syn", 10);
    assertEquals(List.of("", ""), customers.defaultGenderOptions());
    Page<Tag> tags = customers.listTags(map("per_page", 2));

    assertEquals(
        "GET v1/customers/get_customer_id?customer_emails=a%40example.com%2Cb%40example.com",
        h.line(0));
    assertEquals("GET v1/customers/get_customer_id_by_name?customer_name=Syn&limit=10", h.line(1));
    assertEquals(40401725L, byName.get(0).customerId());
    assertEquals("GET v1/customers/default_gender_options", h.line(2));
    assertEquals("REDACTED", tags.items().get(0).name());
    assertFalse(tags.hasNext());
  }

  @Test
  void detailGoldenFiles() {
    ServiceHarness h =
        ServiceHarness.golden(
            "v1/GET_v1_customers_{id}_account_activation_url.json",
            "v1/GET_v1_customers_{id}_uid_providers_line.json",
            "v1/GET_v1_customers_{id}_vip_info.json",
            "v1/GET_v1_customers_{id}_spending_overview.json",
            "v1/GET_v1_customers_{id}_message_posts.json");
    CustomerService customers = h.client.customers();

    assertTrue(
        customers
            .accountActivationUrl(1)
            .startsWith("http://example.cyberbiz.co/account/customer/activate"));
    CustomerUidLookup uid = customers.getUidProvider(1, "line");
    CustomerVipInfo vip = customers.vipInfo(1);
    CustomerSpendingOverview spending =
        customers.spendingOverview(
            1, map("start_date", LocalDate.of(2026, 1, 1), "end_date", "2026-06-30"));

    assertEquals("GET v1/customers/1/uid_providers/line", h.line(1));
    assertEquals(23300813L, uid.customerId());
    assertEquals("查詢成功", uid.message());
    assertEquals(23300813L, vip.customerId());
    assertNull(vip.currentGroup());
    assertNull(vip.extraInfo().startAt());
    assertNull(vip.extraInfo().differenceOfTotalSpentForUpgrade());
    assertEquals(
        "GET v1/customers/1/spending_overview?start_date=2026-01-01&end_date=2026-06-30",
        h.line(3));
    assertEquals(Money.of(1398), spending.paidAndValidTotalSpent());
    assertEquals(3, spending.paidAndValidOrdersCount());
    assertEquals(Money.of(466), spending.paidAndValidAverageSpent());
    assertEquals(List.of(), customers.listMessagePosts(1).items());
  }

  @Test
  void orderGoldenFiles() {
    ServiceHarness h =
        ServiceHarness.golden(
            "v1/GET_v1_customers_{id}_orders.json",
            "v1/GET_v1_customers_{id}_recent_purchases.json");
    CustomerService customers = h.client.customers();

    Page<Order> orders = customers.listOrders(42, map("per_page", 2));
    List<LineItem> recent =
        customers.recentPurchases(
            42,
            map(
                "start_date",
                OffsetDateTime.parse("2025-12-31T16:00:00Z"),
                "end_date",
                "2026-03-31",
                "max_products",
                5));

    assertEquals("GET v1/customers/42/orders?per_page=2", h.line(0));
    assertEquals(31, orders.pagination().total());
    Order first = orders.items().get(0);
    assertEquals(49492440L, first.id());
    assertEquals(1068L, first.orderNumber());
    assertEquals(Money.of(200), first.subtotalPrice());
    assertEquals(Money.of(866), first.lineItems().get(0).discounts().get(0).discount());
    assertEquals("bundle_discount", first.lineItems().get(0).discounts().get(0).code());
    assertEquals(
        "GET v1/customers/42/recent_purchases?start_date=2026-01-01&end_date=2026-03-31"
            + "&max_products=5",
        h.line(1));
    assertEquals(3, recent.size());
    assertEquals(109940716L, recent.get(0).id());
    assertEquals(Money.of(500), recent.get(0).price());
    assertEquals("inclusive_tax", recent.get(0).taxTypeId());
    assertEquals(taipei("2026-03-10T11:34:31"), recent.get(0).createdAt());
  }

  @Test
  void v2GoldenFiles() {
    ServiceHarness h =
        ServiceHarness.golden(
            "v2/GET_v2_customers.json",
            "v2/GET_v2_customers_include.json",
            "v2/GET_v2_customers_by_uid_provider.json");
    CustomerService customers = h.client.customers();

    Page<Customer> page = customers.listV2(map("per_page", 2));
    Page<Customer> selected =
        customers.listV2(
            map("ids", List.of(23300813, 24077804), "include_params", List.of("tags", "vip_info")));

    assertEquals(23300813L, page.items().get(0).id());
    assertEquals(2, page.pagination().nextPage());
    assertEquals(
        "GET v2/customers?ids=23300813%2C24077804&include_params=tags%2Cvip_info", h.line(1));
    assertEquals(24077804L, selected.items().get(1).id());
    assertThrows(NotFoundException.class, () -> customers.getByUidProvider("line", "U-synthetic"));
    assertEquals("GET v2/customers/by_uid_provider?uid=U-synthetic&provider=line", h.line(2));
  }

  @Test
  void decodesSyntheticVipAndMessages() {
    CustomerService customers =
        ServiceHarness.json(
                "{\"customer_id\":1,\"current_group\":{\"id\":2,\"name\":\"Gold\","
                    + "\"customer_tags\":[\"vip\"]},\"current_level\":{\"id\":3,"
                    + "\"bonus_point_enabled\":true,\"upgrade_condition_total_spent\":5000.0},"
                    + "\"next_level\":null,\"extra_info\":{\"start_at\":\"2026-01-01 00:00:00\","
                    + "\"difference_of_total_spent_for_upgrade\":1200}}",
                "[{\"id\":4,\"title\":\"Where is my parcel\",\"status\":\"replied\","
                    + "\"comments\":[{\"id\":5,\"role\":\"admin\",\"content\":\"Shipped\","
                    + "\"created_at\":\"2026-02-01 10:00:00\"}],"
                    + "\"created_at\":\"2026-02-01 09:00:00\"}]")
            .client
            .customers();

    CustomerVipInfo vip = customers.vipInfo(1);
    assertEquals(List.of("vip"), vip.currentGroup().customerTags());
    assertEquals(Money.of(5000), vip.currentLevel().upgradeConditionTotalSpent());
    assertTrue(vip.currentLevel().bonusPointEnabled());
    assertNull(vip.nextLevel());
    assertEquals(Money.of(1200), vip.extraInfo().differenceOfTotalSpentForUpgrade());
    assertEquals(
        "Shipped", customers.listMessagePosts(1).items().get(0).comments().get(0).content());
  }

  @Test
  void cartItemsDecodesTheVariants() {
    CustomerService empty =
        ServiceHarness.golden("v1/GET_v1_customers_{id}_customer_cart_items.json")
            .client
            .customers();
    CustomerService one =
        ServiceHarness.json(
                "[{\"id\":11,\"product_id\":7,\"name\":\"Synthetic Tea / 500g\","
                    + "\"price\":120.5,\"sku\":\"SKU-SYN-1\"}]")
            .client
            .customers();

    assertEquals(List.of(), empty.cartItems(1));
    ProductVariant variant = one.cartItems(1).get(0);
    assertEquals(7L, variant.productId());
    assertEquals(Money.of("120.50"), variant.price());
    assertEquals("SKU-SYN-1", variant.sku());
  }

  @Test
  void updateKeepsTheIdAndSendsExplicitNulls() {
    ServiceHarness h = ServiceHarness.json("{\"name\":\"Synthetic\"}");

    Customer customer =
        h.client
            .customers()
            .update(8, map("confirmed_at", null, "other_accumulated_consumption", Money.of(0)));

    assertEquals(8L, customer.id());
    assertEquals("Synthetic", customer.name());
    assertEquals("PUT v1/customers/8", h.line());
    assertEquals("{\"confirmed_at\":null,\"other_accumulated_consumption\":0.00}", h.body(0));
  }

  @ParameterizedTest
  @ValueSource(strings = {"", ".", ".."})
  void rejectsAProviderTypeThatWouldLeaveItsPathSegment(String providerType) {
    ServiceHarness h = ServiceHarness.json("{}");
    CustomerService customers = h.client.customers();

    assertThrows(IllegalArgumentException.class, () -> customers.getUidProvider(3, providerType));
    assertThrows(
        IllegalArgumentException.class, () -> customers.setUidProvider(3, providerType, "U1"));
    assertEquals(0, h.transport.requests.size());
  }

  @Test
  void rejectsAnOverlongProviderType() {
    CustomerService customers = ServiceHarness.json("{}").client.customers();

    assertThrows(
        IllegalArgumentException.class, () -> customers.getUidProvider(3, "x".repeat(257)));
  }

  @Test
  void encodesASlashInTheProviderType() {
    ServiceHarness h = ServiceHarness.json("{}", "");

    h.client.customers().getUidProvider(3, "a/b");
    h.client.customers().setUidProvider(3, "../x", "U1");

    assertEquals("GET v1/customers/3/uid_providers/a%2Fb", h.line(0));
    assertEquals("PUT v1/customers/3/uid_providers/..%2Fx", h.line(1));
  }

  /** One method call, the reply it gets, and the request it must send. */
  record Shape(
      String name, Function<CustomerService, ?> call, String reply, String line, String body) {
    @Override
    public String toString() {
      return name;
    }
  }

  static Stream<Arguments> requestShapes() {
    return Stream.of(shapes1(), shapes2(), shapes3(), shapes4()).flatMap(s -> s).map(Arguments::of);
  }

  private static Stream<Shape> shapes1() {
    return Stream.of(
        new Shape(
            "all",
            s -> s.all().stream().toList(),
            "[]",
            "GET v1/customers?page=1&per_page=50",
            null),
        new Shape("list", CustomerService::list, "[]", "GET v1/customers", null),
        new Shape("get", s -> s.get(3), "{}", "GET v1/customers/3", null),
        new Shape(
            "create",
            s ->
                s.create(
                    map(
                        "name",
                        "Synthetic",
                        "password",
                        "not-a-real-secret",
                        "enable_cvs_pickup",
                        false,
                        "accepts_marketing",
                        false)),
            "{\"id\":1}",
            "POST v1/customers",
            "{\"name\":\"Synthetic\",\"password\":\"not-a-real-secret\","
                + "\"enable_cvs_pickup\":false,\"accepts_marketing\":false}"),
        new Shape(
            "update",
            s -> s.update(3, map("note", "n")),
            "{}",
            "PUT v1/customers/3",
            "{\"note\":\"n\"}"),
        new Shape(
            "lookup by mobile",
            s -> s.lookupIds(map("customer_mobiles", List.of("0900000000"))),
            "[]",
            "GET v1/customers/get_customer_id?customer_mobiles=0900000000",
            null));
  }

  private static Stream<Shape> shapes2() {
    return Stream.of(
        new Shape(
            "lookup by name without limit",
            s -> s.lookupIdsByName("A B"),
            "[]",
            "GET v1/customers/get_customer_id_by_name?customer_name=A%20B",
            null),
        new Shape(
            "gender options",
            CustomerService::defaultGenderOptions,
            "[]",
            "GET v1/customers/default_gender_options",
            null),
        new Shape("tags", CustomerService::listTags, "[]", "GET v1/customers/tags", null),
        new Shape(
            "activation url",
            s -> s.accountActivationUrl(3),
            "{\"account_activation_url\":\"https://example.com/a\"}",
            "GET v1/customers/3/account_activation_url",
            null),
        new Shape(
            "consume bonus",
            s -> s.consumeBonusPoints(3, map("consume_all", true)),
            "",
            "POST v1/customers/3/consume_bonus_points",
            "{\"consume_all\":true}"),
        new Shape(
            "consume bonus amount",
            s -> s.consumeBonusPoints(3, map("amount", Money.of("12.5"))),
            "",
            "POST v1/customers/3/consume_bonus_points",
            "{\"amount\":12.50}"),
        new Shape(
            "register code",
            s ->
                s.updateRegisterCode(
                    3, map("secret_key", "k", "register_code", "R1", "accepts_marketing", false)),
            "",
            "POST v1/customers/3/update_register_code_and_accepts_marketing",
            "{\"secret_key\":\"k\",\"register_code\":\"R1\",\"accepts_marketing\":false}"));
  }

  private static Stream<Shape> shapes3() {
    return Stream.of(
        new Shape(
            "get uid",
            s -> s.getUidProvider(3, "facebook"),
            "{}",
            "GET v1/customers/3/uid_providers/facebook",
            null),
        new Shape(
            "set uid",
            s -> s.setUidProvider(3, "line_at", "U1"),
            "",
            "PUT v1/customers/3/uid_providers/line_at",
            "{\"uid\":\"U1\"}"),
        new Shape("vip info", s -> s.vipInfo(3), "{}", "GET v1/customers/3/vip_info", null),
        new Shape(
            "spending overview",
            s -> s.spendingOverview(3, map("start_date", "2026-01-01", "end_date", "2026-01-31")),
            "{}",
            "GET v1/customers/3/spending_overview?start_date=2026-01-01&end_date=2026-01-31",
            null),
        new Shape(
            "message posts page",
            s -> s.listMessagePosts(3, map("page", 2)),
            "[]",
            "GET v1/customers/3/message_posts?page=2",
            null),
        new Shape("orders", s -> s.listOrders(3), "[]", "GET v1/customers/3/orders", null),
        new Shape(
            "all orders",
            s -> s.allOrders(3, map("per_page", 10)).stream().toList(),
            "[]",
            "GET v1/customers/3/orders?page=1&per_page=10",
            null));
  }

  private static Stream<Shape> shapes4() {
    return Stream.of(
        new Shape(
            "cart items",
            s -> s.cartItems(3),
            "[]",
            "GET v1/customers/3/customer_cart_items",
            null),
        new Shape(
            "recent purchases",
            s ->
                s.recentPurchases(
                    3,
                    map("start_date", "2026-01-01", "end_date", "2026-01-31", "max_products", 5)),
            "[]",
            "GET v1/customers/3/recent_purchases?start_date=2026-01-01&end_date=2026-01-31"
                + "&max_products=5",
            null),
        new Shape("list v2", CustomerService::listV2, "[]", "GET v2/customers", null),
        new Shape(
            "all v2",
            s -> s.allV2(map("include_params", List.of("tags"))).stream().toList(),
            "[]",
            "GET v2/customers?page=1&include_params=tags&per_page=50",
            null),
        new Shape(
            "by uid provider",
            s -> s.getByUidProvider("facebook", "U 1"),
            "{\"id\":1}",
            "GET v2/customers/by_uid_provider?uid=U%201&provider=facebook",
            null),
        new Shape(
            "oauth",
            s -> s.oauth(map("uid", "U1", "provider", "line")),
            "{\"id\":1}",
            "POST v2/customer_oauth",
            "{\"uid\":\"U1\",\"provider\":\"line\"}"));
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("requestShapes")
  void requestShape(Shape shape) {
    ServiceHarness h = ServiceHarness.json(shape.reply());

    var unused = shape.call().apply(h.client.customers());

    assertEquals(shape.line(), h.line());
    assertEquals(shape.body(), h.body(0));
    assertEquals(1, h.transport.requests.size());
  }
}
