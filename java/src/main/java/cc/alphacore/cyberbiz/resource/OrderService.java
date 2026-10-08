package cc.alphacore.cyberbiz.resource;

import cc.alphacore.cyberbiz.CyberbizClient;
import cc.alphacore.cyberbiz.Request;
import cc.alphacore.cyberbiz.Response;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.model.Fulfillment;
import cc.alphacore.cyberbiz.model.FulfillmentInfo;
import cc.alphacore.cyberbiz.model.Order;
import cc.alphacore.cyberbiz.model.OrderEticket;
import cc.alphacore.cyberbiz.model.OrderNumberId;
import cc.alphacore.cyberbiz.model.OrderReturn;
import cc.alphacore.cyberbiz.model.OrderTransaction;
import cc.alphacore.cyberbiz.model.SupportShippingBatchResult;
import cc.alphacore.cyberbiz.model.SupportShippingsResult;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.pagination.PageIterable;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;

/**
 * Orders, their fulfillments, payments, returns and e-tickets, plus the v2 shipping-label
 * endpoints. Obtained from {@link CyberbizClient#orders()}; stateless and thread-safe.
 *
 * <p>Request bodies and query options are maps shaped like the API's JSON (see {@code
 * docs/api/en/cyberbiz-openapi-v1.yaml} and {@code -v2.yaml}); use a {@code LinkedHashMap} to keep
 * the parameter order. In queries a collection is sent comma-separated and an {@code
 * OffsetDateTime} as a CYBERBIZ timestamp in Asia/Taipei; in bodies a {@code Money} is sent exactly
 * and a null value as JSON null. The writes that answer with an order return {@link
 * Optional#empty()} when the platform replies with an empty or null body.
 *
 * <p>Every method throws {@link cc.alphacore.cyberbiz.exception.ApiException} for an error
 * response, {@link cc.alphacore.cyberbiz.exception.TransportException} when no response arrived and
 * {@link cc.alphacore.cyberbiz.exception.DecodeException} when the reply does not fit the model.
 */
public final class OrderService {
  private static final String ORDERS = "/v1/orders";

  private final Calls calls;

  /**
   * Creates the service; normally obtained from {@link CyberbizClient#orders()}.
   *
   * @param client the client to send through
   */
  public OrderService(CyberbizClient client) {
    this.calls = new Calls(Objects.requireNonNull(client, "client"));
  }

  /**
   * One page of orders (GET /v1/orders). Filters include page, per_page, offset, the inclusive time
   * ranges start_time/end_time, updated_at_, closed_at_, refund_at_ and cancelled_at_
   * start_time/end_time, statuses, financial_statuses, fulfillment_statuses, return_statuses, tags,
   * excluded_tags, data_source ("ec" or "pos") and vendor.
   *
   * @param query the filters; null for none
   * @return the page
   */
  public Page<Order> list(Map<String, ?> query) {
    return calls.client().list(Params.query(Request.of("GET", ORDERS), query), Order.class);
  }

  /**
   * The first page of orders (GET /v1/orders).
   *
   * @return the page
   */
  public Page<Order> list() {
    return list(null);
  }

  /**
   * Every order, page by page, lazily (GET /v1/orders).
   *
   * @param query the filters of {@link #list(Map)}; null for none
   * @return the orders of every page
   */
  public PageIterable<Order> all(Map<String, ?> query) {
    return calls.client().all(Params.query(Request.of("GET", ORDERS), query), Order.class);
  }

  /**
   * Every order, page by page, lazily (GET /v1/orders).
   *
   * @return the orders of every page
   */
  public PageIterable<Order> all() {
    return all(null);
  }

  /**
   * One order (GET /v1/orders/{id}).
   *
   * @param id the order id
   * @return the order
   * @throws NotFoundException also when the API answers with a null body
   */
  public Order get(long id) {
    return calls.one(Request.of("GET", path(id, "")), Order.class);
  }

  /**
   * Maps shop-facing order numbers to API ids (GET /v1/orders/get_order_id).
   *
   * @param orderNumbers the order numbers
   * @return one entry per order found
   */
  public List<OrderNumberId> lookupIds(List<Long> orderNumbers) {
    Request request =
        Request.of("GET", ORDERS + "/get_order_id")
            .withQuery("order_numbers", Params.join(orderNumbers));
    return calls.list(request, OrderNumberId.class);
  }

  /**
   * Changes an order's note and delivery preferences (PUT /v1/orders/{id}). delivery_date
   * ("YYYY-MM-DD") and delivery_time (0-3) are custom features CYBERBIZ must enable.
   *
   * @param id the order id
   * @param changes note, delivery_date, delivery_time
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> update(long id, Map<String, ?> changes) {
    return order("PUT", path(id, ""), changes);
  }

  /**
   * Replaces an order's tags (PUT /v1/orders/{id}/tags).
   *
   * @param id the order id
   * @param tags the new tags
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> updateTags(long id, List<String> tags) {
    return order("PUT", path(id, "/tags"), Map.of("tags", List.copyOf(tags)));
  }

  /**
   * Sets the logistics message shown for an order (PUT /v1/orders/{id}/update_warehouse_note).
   *
   * @param id the order id
   * @param note the message
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> updateWarehouseNote(long id, String note) {
    return order("PUT", path(id, "/update_warehouse_note"), Map.of("warehouse_note", note));
  }

  /**
   * Opens or closes an order (PUT /v1/orders/{id}/update_status).
   *
   * @param id the order id
   * @param status "open" or "closed"
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> updateStatus(long id, String status) {
    return order("PUT", path(id, "/update_status"), Map.of("status", status));
  }

  /**
   * Asks the platform to refresh an order's payment state (PUT
   * /v1/orders/{id}/update_financial_status).
   *
   * @param id the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> updateFinancialStatus(long id) {
    return order("PUT", path(id, "/update_financial_status"), null);
  }

  /**
   * Sets the fulfillment status back to unshipped (PUT /v1/orders/{id}/unshipped).
   *
   * @param id the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> markUnshipped(long id) {
    return order("PUT", path(id, "/unshipped"), null);
  }

  /**
   * Sets the fulfillment status to preparing (PUT /v1/orders/{id}/preparing).
   *
   * @param id the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> markPreparing(long id) {
    return order("PUT", path(id, "/preparing"), null);
  }

  /**
   * Cancels an order (PUT /v1/orders/{id}/cancelled). cancel_reason is required: "customer",
   * "duplicate", "not_pay", "fraud", "inventory", "forgot", "not_shipped_yet", "pos_sp_return",
   * "card_paid_fail" or "other". cancel_reason_detail is the customer-side reason: "wait_too_long",
   * "want_to_use_other_discount", "modify_shipment_location", "have_concern_about_product",
   * "price_too_high", "operation_mistake" or "other".
   *
   * @param id the order id
   * @param cancel the cancellation
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> cancel(long id, Map<String, ?> cancel) {
    return order("PUT", path(id, "/cancelled"), cancel);
  }

  /**
   * Moves an order's return status by hand (PUT /v1/orders/{id}/manual_return).
   *
   * @param id the order id
   * @param operation "manual_returning", "manual_check_goods", "manual_return_refuse" or
   *     "manual_return_done"
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> manualReturn(long id, String operation) {
    return order("PUT", path(id, "/manual_return"), Map.of("operation", operation));
  }

  /**
   * Switches an order to merchant-arranged shipping (PUT
   * /v1/orders/{id}/change_to_custom_shipping).
   *
   * @param id the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> changeToCustomShipping(long id) {
    return order("PUT", path(id, "/change_to_custom_shipping"), null);
  }

  /**
   * One page of an order's payments (GET /v1/orders/{id}/transactions).
   *
   * @param id the order id
   * @param query page, per_page, offset; null for none
   * @return the page
   */
  public Page<OrderTransaction> listTransactions(long id, Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", path(id, "/transactions")), query);
    return calls.client().list(request, OrderTransaction.class);
  }

  /**
   * The first page of an order's payments (GET /v1/orders/{id}/transactions).
   *
   * @param id the order id
   * @return the page
   */
  public Page<OrderTransaction> listTransactions(long id) {
    return listTransactions(id, null);
  }

  /**
   * Records a manual payment against an order (POST /v1/orders/{id}/transactions).
   *
   * @param id the order id
   * @param transaction kind "capture", paid_type "manual"
   * @return the recorded payment
   */
  public OrderTransaction createTransaction(long id, Map<String, ?> transaction) {
    Request request = Params.body(Request.of("POST", path(id, "/transactions")), transaction);
    return calls.object(request, OrderTransaction.class);
  }

  /**
   * An order's return shipments. Not paginated; the swagger documents an object but the platform
   * sends an array (GET /v1/orders/{id}/returns).
   *
   * @param id the order id
   * @return the returns
   */
  public List<OrderReturn> listReturns(long id) {
    return calls.list(Request.of("GET", path(id, "/returns")), OrderReturn.class);
  }

  /**
   * One page of e-tickets; the shop needs the e-ticket feature or the platform answers 403 (GET
   * /v1/order_etickets).
   *
   * @param query page, per_page, offset, search_column ("phone", "ticket_number", "title", "name"
   *     or "order_name") and q; null for none
   * @return the page
   */
  public Page<OrderEticket> listEtickets(Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", "/v1/order_etickets"), query);
    return calls.client().list(request, OrderEticket.class);
  }

  /**
   * The first page of e-tickets (GET /v1/order_etickets).
   *
   * @return the page
   */
  public Page<OrderEticket> listEtickets() {
    return listEtickets(null);
  }

  /**
   * Redeems units of an e-ticket at a branch store (POST /v1/order_etickets/submit_redeem).
   *
   * @param redeem ticket_number, redeem_quantity, user_id, branch_store_id
   * @return the e-ticket after redemption
   */
  public OrderEticket redeemEticket(Map<String, ?> redeem) {
    Request request = Params.body(Request.of("POST", "/v1/order_etickets/submit_redeem"), redeem);
    return calls.object(request, OrderEticket.class);
  }

  /**
   * One page of an order's fulfillments (GET /v1/orders/{id}/fulfillments).
   *
   * @param orderId the order id
   * @param query page, per_page, offset; null for none
   * @return the page
   */
  public Page<Fulfillment> listFulfillments(long orderId, Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", path(orderId, "/fulfillments")), query);
    return calls.client().list(request, Fulfillment.class);
  }

  /**
   * The first page of an order's fulfillments (GET /v1/orders/{id}/fulfillments).
   *
   * @param orderId the order id
   * @return the page
   */
  public Page<Fulfillment> listFulfillments(long orderId) {
    return listFulfillments(orderId, null);
  }

  /**
   * One fulfillment (GET /v1/orders/{id}/fulfillments/{fulfillment_id}).
   *
   * @param orderId the order id
   * @param fulfillmentId the fulfillment id
   * @return the fulfillment
   * @throws NotFoundException also when the API answers with a null body
   */
  public Fulfillment getFulfillment(long orderId, long fulfillmentId) {
    Request request = Request.of("GET", path(orderId, "/fulfillments/" + fulfillmentId));
    return calls.one(request, Fulfillment.class);
  }

  /**
   * Ships line items with a merchant-arranged carrier (POST
   * /v1/orders/{id}/fulfillments/custom_shipping). line_item_ids may be a collection; it is sent
   * comma-separated.
   *
   * @param orderId the order id
   * @param shipping line_item_ids, tracking_number, tracking_company, notify_customer
   * @return the fulfillment, if the platform sent it
   */
  public Optional<Fulfillment> createCustomShipping(long orderId, Map<String, ?> shipping) {
    return fulfillment(
        path(orderId, "/fulfillments/custom_shipping"), Params.lineItemIds(shipping));
  }

  /**
   * Books a home-delivery shipment through CYBERBIZ. The platform may answer 202 while the tracking
   * number is being issued; the result is then empty and the fulfillment can be read later with
   * {@link #listFulfillments} (POST /v1/orders/{id}/fulfillments/support_shipping). source:
   * "ezcat", "pelican", "sf" or "hct"; temperature: "normal" or "cold"; fridge_or_frozen: "none",
   * "fridge" or "frozen".
   *
   * @param orderId the order id
   * @param shipping the booking; line_item_ids may be a collection
   * @return the fulfillment, if the platform sent it
   */
  public Optional<Fulfillment> createSupportShipping(long orderId, Map<String, ?> shipping) {
    return fulfillment(
        path(orderId, "/fulfillments/support_shipping"), Params.lineItemIds(shipping));
  }

  /**
   * Books a convenience-store shipment through the v1 endpoint; prefer {@link #createCvsShipping},
   * whose reply carries the label type (POST /v1/orders/{id}/fulfillments/cvs_shipping).
   *
   * @param orderId the order id
   * @param shipping measurement ("S60" or "S105") and size; null for none
   * @return the fulfillment, if the platform sent it
   */
  public Optional<Fulfillment> createCvsShippingV1(long orderId, Map<String, ?> shipping) {
    return fulfillment(path(orderId, "/fulfillments/cvs_shipping"), shipping);
  }

  /**
   * Ships part of a Hi-Life order (POST /v1/orders/{id}/fulfillments/partial_cvs_shipping).
   *
   * @param orderId the order id
   * @param shipping items (a list of line_item_id and quantity) and an optional charge
   * @return the fulfillment, if the platform sent it
   */
  public Optional<Fulfillment> createPartialCvsShipping(long orderId, Map<String, ?> shipping) {
    return fulfillment(path(orderId, "/fulfillments/partial_cvs_shipping"), shipping);
  }

  /**
   * Closes partial shipping on a Hi-Life order (POST
   * /v1/orders/{id}/fulfillments/partial_cvs_shipping/conclude).
   *
   * @param orderId the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> concludePartialCvsShipping(long orderId) {
    return order("POST", path(orderId, "/fulfillments/partial_cvs_shipping/conclude"), null);
  }

  /**
   * Ships line items by express delivery from a branch store (POST
   * /v1/orders/{id}/fulfillments/express_delivery_shipping).
   *
   * @param orderId the order id
   * @param lineItemIds the line items to ship
   * @return the fulfillment, if the platform sent it
   */
  public Optional<Fulfillment> createExpressDeliveryShipping(long orderId, List<Long> lineItemIds) {
    return fulfillment(
        path(orderId, "/fulfillments/express_delivery_shipping"),
        Map.of("line_item_ids", Params.join(lineItemIds)));
  }

  /**
   * Marks a CVS order as arrived at the store (POST /v1/orders/{id}/fulfillments/arrived).
   *
   * @param orderId the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> markArrived(long orderId) {
    return order("POST", path(orderId, "/fulfillments/arrived"), null);
  }

  /**
   * Marks an order as received by the customer (POST /v1/orders/{id}/fulfillments/received).
   *
   * @param orderId the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> markReceived(long orderId) {
    return order("POST", path(orderId, "/fulfillments/received"), null);
  }

  /**
   * Marks a CVS order as not collected in time (POST /v1/orders/{id}/fulfillments/expired).
   *
   * @param orderId the order id
   * @return the updated order, if the platform sent it
   */
  public Optional<Order> markExpired(long orderId) {
    return order("POST", path(orderId, "/fulfillments/expired"), null);
  }

  /**
   * The CVS label data of the given fulfillments (POST /v1/orders/fulfillment_infos).
   *
   * @param fulfillmentIds the fulfillments
   * @return one entry per fulfillment
   */
  public List<FulfillmentInfo> getFulfillmentInfos(List<Long> fulfillmentIds) {
    Request request =
        Params.body(
            Request.of("POST", ORDERS + "/fulfillment_infos"),
            Map.of("fulfillment_ids", List.copyOf(fulfillmentIds)));
    return calls.list(request, FulfillmentInfo.class);
  }

  /**
   * Books home-delivery shipments for several orders through the v1 endpoint (POST
   * /v1/orders/fulfillments/support_shipping). Each entry's line_item_ids may be a collection.
   *
   * @param batch support_shippings (a list of maps) and email
   * @return the outcome per order
   */
  public SupportShippingBatchResult createSupportShippingBatch(Map<String, ?> batch) {
    Map<String, Object> body = new LinkedHashMap<>(batch);
    body.put("support_shippings", Params.lineItemIdsOfEach(batch.get("support_shippings")));
    Request request =
        Params.body(Request.of("POST", ORDERS + "/fulfillments/support_shipping"), body);
    return calls.object(request, SupportShippingBatchResult.class);
  }

  /**
   * Books a convenience-store shipment; only a reply with a non-empty cvsShippingType can be
   * printed with {@link #printCvsShippingLabels} (POST /v2/orders/{id}/cvs_shipping).
   *
   * @param orderId the order id
   * @param shipping measurement and size; null for none
   * @return the fulfillment, if the platform sent it
   */
  public Optional<Fulfillment> createCvsShipping(long orderId, Map<String, ?> shipping) {
    return fulfillment("/v2/orders/" + orderId + "/cvs_shipping", shipping);
  }

  /**
   * Books a convenience-store shipment with the default size (POST /v2/orders/{id}/cvs_shipping).
   *
   * @param orderId the order id
   * @return the fulfillment, if the platform sent it
   */
  public Optional<Fulfillment> createCvsShipping(long orderId) {
    return createCvsShipping(orderId, null);
  }

  /**
   * CVS shipping labels as a zip archive (POST /v2/orders/cvs_shipping_labels), as the PHP SDK
   * returns them: read the archive with {@link Response#bytes()}, and {@link
   * Response#contentType()} and {@link Response#filename()} when the platform sends them.
   *
   * @param labels shipping_type and fulfillment_ids
   * @return the raw response, its body the zip archive
   */
  public Response printCvsShippingLabels(Map<String, ?> labels) {
    return calls.call(Params.body(Request.of("POST", "/v2/orders/cvs_shipping_labels"), labels));
  }

  /**
   * Books home-delivery labels for several orders (POST /v2/orders/fulfillments/support_shippings).
   *
   * @param shippings shipping_type, size, temperature, ..., shipping_orders (a list of order_id and
   *     line_item_ids)
   * @return the created fulfillments and the failed orders
   */
  public SupportShippingsResult createSupportShippings(Map<String, ?> shippings) {
    Request request =
        Params.body(Request.of("POST", "/v2/orders/fulfillments/support_shippings"), shippings);
    return calls.object(request, SupportShippingsResult.class);
  }

  /**
   * Home-delivery labels as a zip archive (POST /v2/orders/fulfillments/support_shipping_labels);
   * read it as for {@link #printCvsShippingLabels}.
   *
   * @param labels shipping_type, print_type ("normal" or "thermal", HCT only) and order_ids
   * @return the raw response, its body the zip archive
   */
  public Response printSupportShippingLabels(Map<String, ?> labels) {
    Request request =
        Params.body(Request.of("POST", "/v2/orders/fulfillments/support_shipping_labels"), labels);
    return calls.call(request);
  }

  /** Sends a write whose reply is the updated order, or nothing; a null body sends no body. */
  private Optional<Order> order(String method, String path, Map<String, ?> body) {
    Request request = Request.of(method, path);
    return calls.optional(body == null ? request : Params.body(request, body), Order.class);
  }

  /** Posts a shipment; a 202, empty or null reply yields an empty result. */
  private Optional<Fulfillment> fulfillment(String path, Map<String, ?> body) {
    Request request = Params.body(Request.of("POST", path), body);
    Response response = calls.call(request);
    if (response.statusCode() == 202) {
      return Optional.empty();
    }
    return Optional.ofNullable(Calls.decode(request, response, Fulfillment.class));
  }

  private static String path(long orderId, String suffix) {
    return ORDERS + "/" + orderId + suffix;
  }
}
