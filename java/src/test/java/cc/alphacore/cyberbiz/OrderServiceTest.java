package cc.alphacore.cyberbiz;

import static cc.alphacore.cyberbiz.ServiceHarness.map;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.model.Fulfillment;
import cc.alphacore.cyberbiz.model.Order;
import cc.alphacore.cyberbiz.model.OrderNumberId;
import cc.alphacore.cyberbiz.model.RelatedItem;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.resource.OrderService;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Function;
import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

/** {@code client.orders()}: Golden File decoding and the request shape of every method. */
class OrderServiceTest {

  private static OffsetDateTime taipei(String local) {
    return OffsetDateTime.parse(local + "+08:00");
  }

  @Test
  void listDecodesTheGoldenPage() {
    ServiceHarness h = ServiceHarness.golden("v1/GET_v1_orders.json");

    Page<Order> page = h.client.orders().list(map("page", 1, "per_page", 2));

    assertEquals("GET v1/orders?page=1&per_page=2", h.line());
    assertEquals(92, page.pagination().total());
    Order order = page.items().get(1);
    assertEquals(56944000L, order.id());
    assertEquals(1102L, order.orderNumber());
    assertEquals(Money.of("9999"), order.subtotalPrice());
    assertEquals(taipei("2026-09-07T12:39:33"), order.createdAt());
    assertEquals("pending", order.statuses().financialStatus());
    assertEquals(42058619L, order.customer().id());
    assertEquals("inclusive_tax", order.lineItems().get(0).taxTypeId());
    assertFalse(order.token().isEmpty());
  }

  @Test
  void getDecodesTheGoldenOrder() {
    ServiceHarness h = ServiceHarness.golden("v1/GET_v1_orders_{id}.json");

    Order order = h.client.orders().get(56943817);

    assertEquals("GET v1/orders/56943817", h.line());
    assertEquals(56943817L, order.id());
    assertEquals("#1101", order.orderName());
    assertEquals(taipei("1990-01-01T00:00"), order.customer().birthday());
    assertEquals(Money.of(5000), order.customer().bonusRemain());
    assertNull(order.customer().otherAccumulatedConsumptionExpiredAt());
    assertEquals("custom", order.shippingVendor().type());
    assertNull(order.deliveryDate());
    assertNull(order.einvoice());
    assertNull(order.timings().closedAt());
    assertEquals(taipei("2026-09-07T12:35:05"), order.timings().confirmedAt());
    assertEquals(ZoneOffset.ofHours(8), order.timings().confirmedAt().getOffset());
    assertEquals(Money.zero(), order.prices().discounts().vipDiscount());
    assertNull(order.prices().discounts().shopDiscount());
    RelatedItem component = order.lineItems().get(0).relatedItems().get(0).items().get(0);
    assertEquals(Money.of("9991"), component.comboProductPriceDifference());
    assertEquals(List.of(Money.of("9991")), component.comboProductPriceDiffDetails());
    assertEquals("非導購訂單", order.linkedOrderInfo().source());
    assertNull(order.posInfo().posUserId());
    assertNull(order.warehouseTypeId());
    assertNull(order.token(), "detail responses carry no token");
    assertThrows(UnsupportedOperationException.class, () -> order.lineItems().add(null));
  }

  @Test
  void getTurnsANullBodyIntoNotFound() {
    OrderService orders = ServiceHarness.json("null").client.orders();

    NotFoundException e = assertThrows(NotFoundException.class, () -> orders.get(5));

    assertEquals("/v1/orders/5", e.path());
  }

  @Test
  void lookupIdsDecodesTheGoldenFile() {
    ServiceHarness h = ServiceHarness.golden("v1/GET_v1_orders_get_order_id.json");

    List<OrderNumberId> ids = h.client.orders().lookupIds(List.of(1102L, 1101L));

    assertEquals("GET v1/orders/get_order_id?order_numbers=1102%2C1101", h.line());
    assertEquals(new OrderNumberId(1102, 56944000), ids.get(0));
  }

  @Test
  void fulfillmentGoldenFiles() {
    ServiceHarness h =
        ServiceHarness.golden(
            "v1/GET_v1_orders_{id}_fulfillments.json",
            "v1/GET_v1_orders_{id}_fulfillments_{id}.json");
    OrderService orders = h.client.orders();

    assertEquals(List.of(), orders.listFulfillments(7, map("per_page", 2)).items());
    Fulfillment fulfillment = orders.getFulfillment(7, 23064624);

    assertEquals("GET v1/orders/7/fulfillments?per_page=2", h.line(0));
    assertEquals("GET v1/orders/7/fulfillments/23064624", h.line(1));
    assertEquals(23064624L, fulfillment.id());
    assertEquals("other", fulfillment.trackingCompany());
    assertEquals("fulfilled", fulfillment.status());
    assertEquals(taipei("2026-06-03T21:39:47"), fulfillment.fulfilledAt());
    assertNull(fulfillment.receivedAt());
    assertEquals(Money.of(888), fulfillment.lineItems().get(0).price());
    RelatedItem component = fulfillment.lineItems().get(0).relatedItems().get(0).items().get(0);
    assertEquals(Money.of(100), component.cost());
  }

  @Test
  void returnAndTransactionGoldenFiles() {
    ServiceHarness h =
        ServiceHarness.golden(
            "v1/GET_v1_orders_{id}_returns.json", "v1/GET_v1_orders_{id}_transactions.json");
    OrderService orders = h.client.orders();

    assertEquals(List.of(), orders.listReturns(7));
    assertEquals(List.of(), orders.listTransactions(7).items());
    assertEquals("GET v1/orders/7/returns", h.line(0));
    assertEquals("GET v1/orders/7/transactions", h.line(1));
  }

  @Test
  void listEncodesFiltersCommaSeparatedAndTimesInTaipei() {
    ServiceHarness h = ServiceHarness.json("[]");

    h.client
        .orders()
        .list(
            map(
                "start_time",
                OffsetDateTime.parse("2026-01-01T00:00:00Z"),
                "end_time",
                "2026-01-31 23:59:59",
                "statuses",
                List.of("open", "closed"),
                "tags",
                List.of("vip"),
                "vendor",
                null,
                "data_source",
                "ec"));

    assertEquals(
        "GET v1/orders?start_time=2026-01-01%2008%3A00%3A00&end_time=2026-01-31%2023%3A59%3A59"
            + "&statuses=open%2Cclosed&tags=vip&data_source=ec",
        h.line());
  }

  @Test
  void allFollowsTheNextPage() {
    ServiceHarness h = new ServiceHarness();
    h.transport.reply(200, "[{\"id\":1}]", "X-Next-Page", "2");
    h.transport.reply(200, "[{\"id\":2}]");

    List<Long> ids = new ArrayList<>();
    h.client.orders().all(map("statuses", List.of("open"))).forEach(o -> ids.add(o.id()));

    assertEquals(List.of(1L, 2L), ids);
    assertEquals("GET v1/orders?page=1&statuses=open&per_page=50", h.line(0));
    assertEquals("GET v1/orders?page=2&statuses=open&per_page=50", h.line(1));
  }

  @Test
  void writesReturnEmptyForAnEmptyOrNullReply() {
    OrderService orders =
        ServiceHarness.json("", "null", "{\"id\":9,\"order_number\":1009}").client.orders();

    assertTrue(orders.markPreparing(9).isEmpty());
    assertTrue(orders.markArrived(9).isEmpty());
    assertEquals(1009L, orders.updateStatus(9, "closed").orElseThrow().orderNumber());
  }

  @Test
  void shipmentsReturnEmptyWhileTheCarrierIsPending() {
    ServiceHarness h = new ServiceHarness();
    h.transport.reply(202, "{\"id\":1}");
    h.transport.reply(200, "{\"id\":5,\"cvs_shipping_type\":\"seven\"}");
    OrderService orders = h.client.orders();

    assertTrue(
        orders
            .createSupportShipping(9, map("line_item_ids", List.of(1, 2), "source", "ezcat"))
            .isEmpty());
    assertEquals("seven", orders.createCvsShipping(9).orElseThrow().cvsShippingType());
    assertEquals("{\"line_item_ids\":\"1,2\",\"source\":\"ezcat\"}", h.body(0));
    assertEquals("POST v2/orders/9/cvs_shipping", h.line(1));
    assertEquals("{}", h.body(1));
  }

  @Test
  void cancelSendsMoneyExactly() {
    ServiceHarness h = ServiceHarness.json("{\"id\":9}");

    h.client
        .orders()
        .cancel(
            9,
            map("cancel_reason", "other", "email", false, "refund_shopdotcom", Money.of("10.5")));

    assertEquals("PUT v1/orders/9/cancelled", h.line());
    assertEquals(
        "{\"cancel_reason\":\"other\",\"email\":false,\"refund_shopdotcom\":10.50}", h.body(0));
  }

  @Test
  void decodesSyntheticSubResources() {
    OrderService orders =
        ServiceHarness.json(
                "[{\"id\":1,\"amount\":100.0,\"kind_name\":\"已收款\",\"paid_type_name\":\"手動\"}]",
                "[{\"id\":2,\"created_at\":\"2026-01-02 03:04:05\",\"tracking_company\":\"ezcat\","
                    + "\"line_items\":[{\"id\":3,\"price\":\"50\"}],\"return_suda5\":\"100\"}]",
                "[{\"title\":\"Ticket\",\"ticket_number\":\"T-1\",\"available_quantity\":2,"
                    + "\"enabled\":true,\"separate\":true,\"transactions\":[{\"ticket_number\":"
                    + "\"T-1-1\",\"used_at\":\"2026-01-02 03:04:05\"}]}]",
                "[{\"id\":4,\"id_verification\":true,\"amount\":60,\"store_no\":\"S1\"}]",
                "{\"request_id\":\"r-1\",\"results\":[{\"order_id\":5,\"tracking_numbers\":"
                    + "{\"tracking_company\":\"hct\",\"tracking_number\":\"N1\"},"
                    + "\"line_items\":\"1,2\"}]}",
                "{\"failed_orders\":[{\"order_id\":6,\"message\":\"no stock\"}],"
                    + "\"fulfillments\":[{\"id\":7,\"order_id\":8,\"tracking_number\":\"\"}]}")
            .client
            .orders();

    assertEquals(Money.of(100), orders.listTransactions(1).items().get(0).amount());
    assertEquals("100", orders.listReturns(1).get(0).returnSuda5());
    assertEquals(
        taipei("2026-01-02T03:04:05"),
        orders.listEtickets().items().get(0).transactions().get(0).usedAt());
    assertTrue(orders.getFulfillmentInfos(List.of(4L)).get(0).idVerification());
    assertEquals(
        "hct",
        orders
            .createSupportShippingBatch(
                map("support_shippings", List.of(), "email", "ops@example.com"))
            .results()
            .get(0)
            .trackingNumbers()
            .trackingCompany());
    assertEquals(6L, orders.createSupportShippings(map()).failedOrders().get(0).orderId());
  }

  /** One method call, the reply it gets, and the request it must send. */
  record Shape(
      String name, Function<OrderService, ?> call, String reply, String line, String body) {
    @Override
    public String toString() {
      return name;
    }
  }

  static Stream<Arguments> requestShapes() {
    return Stream.of(shapes1(), shapes2(), shapes3(), shapes4(), shapes5(), shapes6(), shapes7())
        .flatMap(s -> s)
        .map(Arguments::of);
  }

  private static Stream<Shape> shapes1() {
    return Stream.of(
        new Shape(
            "update",
            s -> s.update(3, map("note", "n", "delivery_time", 0)),
            "{}",
            "PUT v1/orders/3",
            "{\"note\":\"n\",\"delivery_time\":0}"),
        new Shape(
            "update tags",
            s -> s.updateTags(3, List.of("a", "b")),
            "{}",
            "PUT v1/orders/3/tags",
            "{\"tags\":[\"a\",\"b\"]}"),
        new Shape(
            "warehouse note",
            s -> s.updateWarehouseNote(3, "fragile"),
            "{}",
            "PUT v1/orders/3/update_warehouse_note",
            "{\"warehouse_note\":\"fragile\"}"),
        new Shape(
            "update status",
            s -> s.updateStatus(3, "open"),
            "{}",
            "PUT v1/orders/3/update_status",
            "{\"status\":\"open\"}"),
        new Shape(
            "financial status",
            s -> s.updateFinancialStatus(3),
            "{}",
            "PUT v1/orders/3/update_financial_status",
            null),
        new Shape("unshipped", s -> s.markUnshipped(3), "{}", "PUT v1/orders/3/unshipped", null),
        new Shape("preparing", s -> s.markPreparing(3), "{}", "PUT v1/orders/3/preparing", null),
        new Shape(
            "manual return",
            s -> s.manualReturn(3, "manual_return_done"),
            "{}",
            "PUT v1/orders/3/manual_return",
            "{\"operation\":\"manual_return_done\"}"));
  }

  private static Stream<Shape> shapes2() {
    return Stream.of(
        new Shape(
            "custom shipping switch",
            s -> s.changeToCustomShipping(3),
            "{}",
            "PUT v1/orders/3/change_to_custom_shipping",
            null),
        new Shape(
            "transactions page",
            s -> s.listTransactions(3, map("page", 2)),
            "[]",
            "GET v1/orders/3/transactions?page=2",
            null),
        new Shape(
            "create transaction",
            s -> s.createTransaction(3, map("kind", "capture", "paid_type", "manual")),
            "{}",
            "POST v1/orders/3/transactions",
            "{\"kind\":\"capture\",\"paid_type\":\"manual\"}"),
        new Shape(
            "etickets",
            s -> s.listEtickets(map("search_column", "phone", "q", "0900")),
            "[]",
            "GET v1/order_etickets?search_column=phone&q=0900",
            null),
        new Shape(
            "redeem eticket",
            s ->
                s.redeemEticket(
                    map(
                        "ticket_number", "T",
                        "redeem_quantity", 1,
                        "user_id", 2,
                        "branch_store_id", 3)),
            "{}",
            "POST v1/order_etickets/submit_redeem",
            "{\"ticket_number\":\"T\",\"redeem_quantity\":1,\"user_id\":2,"
                + "\"branch_store_id\":3}"));
  }

  private static Stream<Shape> shapes3() {
    return Stream.of(
        new Shape(
            "custom shipping",
            s ->
                s.createCustomShipping(
                    3,
                    map(
                        "line_item_ids",
                        List.of(4, 5),
                        "tracking_number",
                        "N",
                        "tracking_company",
                        "ezcat",
                        "notify_customer",
                        false)),
            "{}",
            "POST v1/orders/3/fulfillments/custom_shipping",
            "{\"line_item_ids\":\"4,5\",\"tracking_number\":\"N\","
                + "\"tracking_company\":\"ezcat\",\"notify_customer\":false}"),
        new Shape(
            "support shipping",
            s ->
                s.createSupportShipping(
                    3, map("line_item_ids", "4", "size", 60, "is_fragile", false)),
            "{}",
            "POST v1/orders/3/fulfillments/support_shipping",
            "{\"line_item_ids\":\"4\",\"size\":60,\"is_fragile\":false}"),
        new Shape(
            "cvs shipping v1",
            s -> s.createCvsShippingV1(3, map("measurement", "S60")),
            "{}",
            "POST v1/orders/3/fulfillments/cvs_shipping",
            "{\"measurement\":\"S60\"}"),
        new Shape(
            "partial cvs",
            s ->
                s.createPartialCvsShipping(
                    3, map("items", List.of(map("line_item_id", 4, "quantity", 1)))),
            "{}",
            "POST v1/orders/3/fulfillments/partial_cvs_shipping",
            "{\"items\":[{\"line_item_id\":4,\"quantity\":1}]}"));
  }

  private static Stream<Shape> shapes4() {
    return Stream.of(
        new Shape(
            "conclude partial cvs",
            s -> s.concludePartialCvsShipping(3),
            "{}",
            "POST v1/orders/3/fulfillments/partial_cvs_shipping/conclude",
            null),
        new Shape(
            "express delivery",
            s -> s.createExpressDeliveryShipping(3, List.of(4L, 5L)),
            "{}",
            "POST v1/orders/3/fulfillments/express_delivery_shipping",
            "{\"line_item_ids\":\"4,5\"}"),
        new Shape(
            "arrived", s -> s.markArrived(3), "{}", "POST v1/orders/3/fulfillments/arrived", null),
        new Shape(
            "received",
            s -> s.markReceived(3),
            "{}",
            "POST v1/orders/3/fulfillments/received",
            null),
        new Shape(
            "expired", s -> s.markExpired(3), "{}", "POST v1/orders/3/fulfillments/expired", null),
        new Shape(
            "fulfillment infos",
            s -> s.getFulfillmentInfos(List.of(4L, 5L)),
            "[]",
            "POST v1/orders/fulfillment_infos",
            "{\"fulfillment_ids\":[4,5]}"));
  }

  private static Stream<Shape> shapes5() {
    return Stream.of(
        new Shape(
            "support shipping batch",
            s ->
                s.createSupportShippingBatch(
                    map(
                        "support_shippings",
                        List.of(map("order_id", 3, "line_item_ids", List.of(4))),
                        "email",
                        "ops@example.com")),
            "{}",
            "POST v1/orders/fulfillments/support_shipping",
            "{\"support_shippings\":[{\"order_id\":3,\"line_item_ids\":\"4\"}],"
                + "\"email\":\"ops@example.com\"}"),
        new Shape(
            "cvs shipping v2",
            s -> s.createCvsShipping(3, map("size", 60)),
            "{}",
            "POST v2/orders/3/cvs_shipping",
            "{\"size\":60}"),
        new Shape(
            "cvs labels",
            s ->
                s.printCvsShippingLabels(
                    map("shipping_type", "seven", "fulfillment_ids", List.of(6))),
            "PK",
            "POST v2/orders/cvs_shipping_labels",
            "{\"shipping_type\":\"seven\",\"fulfillment_ids\":[6]}"),
        new Shape(
            "support shippings v2",
            s ->
                s.createSupportShippings(
                    map(
                        "shipping_type",
                        "hct",
                        "shipping_orders",
                        List.of(map("order_id", 3, "line_item_ids", List.of(4))))),
            "{}",
            "POST v2/orders/fulfillments/support_shippings",
            "{\"shipping_type\":\"hct\",\"shipping_orders\":[{\"order_id\":3,"
                + "\"line_item_ids\":[4]}]}"));
  }

  private static Stream<Shape> shapes6() {
    return Stream.of(
        new Shape(
            "support labels",
            s ->
                s.printSupportShippingLabels(
                    map("shipping_type", "hct", "print_type", "thermal", "order_ids", List.of(3))),
            "PK",
            "POST v2/orders/fulfillments/support_shipping_labels",
            "{\"shipping_type\":\"hct\",\"print_type\":\"thermal\",\"order_ids\":[3]}"),
        new Shape("list", OrderService::list, "[]", "GET v1/orders", null),
        new Shape("get", s -> s.get(3), "{\"id\":3}", "GET v1/orders/3", null),
        new Shape(
            "lookup ids",
            s -> s.lookupIds(List.of(7L)),
            "[]",
            "GET v1/orders/get_order_id?order_numbers=7",
            null),
        new Shape(
            "cancel",
            s -> s.cancel(3, map("cancel_reason", "fraud")),
            "{}",
            "PUT v1/orders/3/cancelled",
            "{\"cancel_reason\":\"fraud\"}"),
        new Shape("returns", s -> s.listReturns(3), "[]", "GET v1/orders/3/returns", null),
        new Shape(
            "fulfillments",
            s -> s.listFulfillments(3),
            "[]",
            "GET v1/orders/3/fulfillments",
            null));
  }

  private static Stream<Shape> shapes7() {
    return Stream.of(
        new Shape(
            "fulfillment",
            s -> s.getFulfillment(3, 4),
            "{\"id\":4}",
            "GET v1/orders/3/fulfillments/4",
            null));
  }

  @ParameterizedTest(name = "{0}")
  @MethodSource("requestShapes")
  void requestShape(Shape shape) {
    ServiceHarness h = ServiceHarness.json(shape.reply());

    var unused = shape.call().apply(h.client.orders());

    assertEquals(shape.line(), h.line());
    assertEquals(shape.body(), h.body(0));
    assertEquals(1, h.transport.requests.size());
  }

  @Test
  void allSendsOneRequestPerPage() {
    ServiceHarness h = ServiceHarness.json("[]");

    assertFalse(h.client.orders().all().iterator().hasNext());
    assertEquals("GET v1/orders?page=1&per_page=50", h.line());
  }

  @Test
  void labelsReturnTheRawResponse() {
    ServiceHarness h = ServiceHarness.json("PK-zip-bytes");

    Response labels =
        h.client
            .orders()
            .printCvsShippingLabels(map("shipping_type", "seven", "fulfillment_ids", List.of(1)));

    assertEquals("PK-zip-bytes", labels.body());
  }

  @Test
  void rejectsNonScalarLineItemIds() {
    OrderService orders = ServiceHarness.json("{}").client.orders();

    assertThrows(
        IllegalArgumentException.class,
        () -> orders.createSupportShipping(1, map("line_item_ids", List.of(List.of(1)))));
  }

  @Test
  void anEmptyReplyToAnObjectWriteIsADecodeError() {
    OrderService orders = ServiceHarness.json("").client.orders();

    DecodeException e =
        assertThrows(
            DecodeException.class, () -> orders.createTransaction(1, map("kind", "capture")));

    assertEquals("cyberbiz: POST /v1/orders/1/transactions: empty response body", e.getMessage());
  }

  @Test
  void aMistypedReplyNamesTheRequestAndField() {
    OrderService orders = ServiceHarness.json("{\"id\":\"x\"}").client.orders();

    DecodeException e = assertThrows(DecodeException.class, () -> orders.get(1));

    assertTrue(e.getMessage().startsWith("cyberbiz: GET /v1/orders/1: "), e.getMessage());
    assertTrue(e.getMessage().contains("$.id"), e.getMessage());
  }
}
