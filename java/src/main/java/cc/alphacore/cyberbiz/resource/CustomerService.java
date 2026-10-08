package cc.alphacore.cyberbiz.resource;

import cc.alphacore.cyberbiz.CyberbizClient;
import cc.alphacore.cyberbiz.Request;
import cc.alphacore.cyberbiz.Response;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.model.Customer;
import cc.alphacore.cyberbiz.model.CustomerIdMatch;
import cc.alphacore.cyberbiz.model.CustomerMessagePost;
import cc.alphacore.cyberbiz.model.CustomerNameMatch;
import cc.alphacore.cyberbiz.model.CustomerSpendingOverview;
import cc.alphacore.cyberbiz.model.CustomerUidLookup;
import cc.alphacore.cyberbiz.model.CustomerVipInfo;
import cc.alphacore.cyberbiz.model.LineItem;
import cc.alphacore.cyberbiz.model.Order;
import cc.alphacore.cyberbiz.model.ProductVariant;
import cc.alphacore.cyberbiz.model.Tag;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.pagination.PageIterable;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * Customers, their orders, service threads, VIP state and external login identities, plus the v2
 * customer lookups. Obtained from {@link CyberbizClient#customers()}; stateless and thread-safe.
 *
 * <p>Request bodies and query options are maps shaped like the API's JSON (see {@code
 * docs/api/en/cyberbiz-openapi-v1.yaml} and {@code -v2.yaml}); use a {@code LinkedHashMap} to keep
 * the parameter order. In queries a collection is sent comma-separated and an {@code
 * OffsetDateTime} as a CYBERBIZ timestamp in Asia/Taipei (as a date for start_date and end_date);
 * in bodies a {@code Money} is sent exactly and a null value as JSON null.
 *
 * <p>Every method throws {@link cc.alphacore.cyberbiz.exception.ApiException} for an error
 * response, {@link cc.alphacore.cyberbiz.exception.TransportException} when no response arrived and
 * {@link cc.alphacore.cyberbiz.exception.DecodeException} when the reply does not fit the model.
 */
public final class CustomerService {
  private static final String CUSTOMERS = "/v1/customers";
  private static final String CUSTOMERS_V2 = "/v2/customers";

  private final Calls calls;

  /**
   * Creates the service; normally obtained from {@link CyberbizClient#customers()}.
   *
   * @param client the client to send through
   */
  public CustomerService(CyberbizClient client) {
    this.calls = new Calls(Objects.requireNonNull(client, "client"));
  }

  /**
   * One page of customers (GET /v1/customers).
   *
   * @param query page, per_page, offset, updated_at_start_time, updated_at_end_time; null for none
   * @return the page
   */
  public Page<Customer> list(Map<String, ?> query) {
    return calls.client().list(Params.query(Request.of("GET", CUSTOMERS), query), Customer.class);
  }

  /**
   * The first page of customers (GET /v1/customers).
   *
   * @return the page
   */
  public Page<Customer> list() {
    return list(null);
  }

  /**
   * Every customer, page by page, lazily (GET /v1/customers).
   *
   * @param query the filters of {@link #list(Map)}; null for none
   * @return the customers of every page
   */
  public PageIterable<Customer> all(Map<String, ?> query) {
    return calls.client().all(Params.query(Request.of("GET", CUSTOMERS), query), Customer.class);
  }

  /**
   * Every customer, page by page, lazily (GET /v1/customers).
   *
   * @return the customers of every page
   */
  public PageIterable<Customer> all() {
    return all(null);
  }

  /**
   * One customer; the response omits the id, which is filled in (GET /v1/customers/{id}).
   *
   * @param id the customer id
   * @return the customer
   * @throws NotFoundException also when the API answers with a null body
   */
  public Customer get(long id) {
    return calls.one(Request.of("GET", path(id, "")), Customer.class).withId(id);
  }

  /**
   * Creates a customer (POST /v1/customers). birthday and other_accumulated_consumption_expired_at
   * are "YYYY-MM-DD"; tags_text is comma-separated; status is one of the {@link Customer#status()}
   * values.
   *
   * @param customer the new customer's fields
   * @return the created customer
   */
  public Customer create(Map<String, ?> customer) {
    return calls.object(Params.body(Request.of("POST", CUSTOMERS), customer), Customer.class);
  }

  /**
   * Changes a customer (PUT /v1/customers/{id}). Sending confirmed_at or mobile_sms_confirmed_at as
   * null clears that verification.
   *
   * @param id the customer id
   * @param changes the fields to change; a null value is sent as JSON null
   * @return the updated customer, with the id filled in
   */
  public Customer update(long id, Map<String, ?> changes) {
    Request request = Params.body(Request.of("PUT", path(id, "")), changes);
    return calls.object(request, Customer.class).withId(id);
  }

  /**
   * Finds customer ids by email or mobile; no match is an empty list (GET
   * /v1/customers/get_customer_id).
   *
   * @param query customer_emails and/or customer_mobiles, each a collection
   * @return the matches
   */
  public List<CustomerIdMatch> lookupIds(Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", CUSTOMERS + "/get_customer_id"), query);
    return calls.list(request, CustomerIdMatch.class);
  }

  /**
   * Finds customers whose name starts with a prefix (GET /v1/customers/get_customer_id_by_name).
   *
   * @param prefix the start of the name
   * @param limit the most hits (up to 50); 0 means the platform default
   * @return the matches
   */
  public List<CustomerNameMatch> lookupIdsByName(String prefix, int limit) {
    Request request =
        Request.of("GET", CUSTOMERS + "/get_customer_id_by_name")
            .withQuery("customer_name", prefix)
            .withQuery("limit", limit > 0 ? limit : null);
    return calls.list(request, CustomerNameMatch.class);
  }

  /**
   * Finds customers whose name starts with a prefix, with the platform's default limit (GET
   * /v1/customers/get_customer_id_by_name).
   *
   * @param prefix the start of the name
   * @return the matches
   */
  public List<CustomerNameMatch> lookupIdsByName(String prefix) {
    return lookupIdsByName(prefix, 0);
  }

  /**
   * The shop's extra gender options besides male and female (GET
   * /v1/customers/default_gender_options).
   *
   * @return the options
   */
  public List<String> defaultGenderOptions() {
    Request request = Request.of("GET", CUSTOMERS + "/default_gender_options");
    return calls.list(request, CustomerReplies.GenderOption.class).stream()
        .map(option -> Objects.requireNonNullElse(option.defaultGenderOption(), ""))
        .toList();
  }

  /**
   * One page of customer tags (GET /v1/customers/tags).
   *
   * @param query page, per_page, offset; null for none
   * @return the page
   */
  public Page<Tag> listTags(Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", CUSTOMERS + "/tags"), query);
    return calls.client().list(request, Tag.class);
  }

  /**
   * The first page of customer tags (GET /v1/customers/tags).
   *
   * @return the page
   */
  public Page<Tag> listTags() {
    return listTags(null);
  }

  /**
   * The customer's account activation link (GET /v1/customers/{id}/account_activation_url).
   *
   * @param id the customer id
   * @return the link
   * @throws NotFoundException also when the API answers with a null body
   */
  public String accountActivationUrl(long id) {
    Request request = Request.of("GET", path(id, "/account_activation_url"));
    String url = calls.one(request, CustomerReplies.ActivationUrl.class).accountActivationUrl();
    return Objects.requireNonNullElse(url, "");
  }

  /**
   * Deducts bonus points (POST /v1/customers/{id}/consume_bonus_points). consume_all spends every
   * remaining point and amount is then ignored.
   *
   * @param id the customer id
   * @param consume amount (a {@code Money}) or consume_all
   * @return the raw response
   */
  public Response consumeBonusPoints(long id, Map<String, ?> consume) {
    return calls.call(Params.body(Request.of("POST", path(id, "/consume_bonus_points")), consume));
  }

  /**
   * Records the referrer code and marketing consent of a customer who registered through a
   * third-party login (POST /v1/customers/{id}/update_register_code_and_accepts_marketing).
   *
   * @param id the customer id
   * @param registration secret_key, register_code and optionally accepts_marketing
   * @return the raw response
   */
  public Response updateRegisterCode(long id, Map<String, ?> registration) {
    String path = path(id, "/update_register_code_and_accepts_marketing");
    return calls.call(Params.body(Request.of("POST", path), registration));
  }

  /**
   * The customer's external UID for a provider (GET
   * /v1/customers/{id}/uid_providers/{provider_type}).
   *
   * @param id the customer id
   * @param providerType "line", "line_at" or "facebook"; one path segment, "/" is encoded
   * @return the lookup
   * @throws NotFoundException also when the API answers with a null body
   * @throws IllegalArgumentException when providerType is empty, "." or "..", or over 256
   *     characters
   */
  public CustomerUidLookup getUidProvider(long id, String providerType) {
    return calls.one(Request.of("GET", uidPath(id, providerType)), CustomerUidLookup.class);
  }

  /**
   * Creates or replaces the customer's external UID for a provider (PUT
   * /v1/customers/{id}/uid_providers/{provider_type}).
   *
   * @param id the customer id
   * @param providerType "line", "line_at" or "facebook"; one path segment, "/" is encoded
   * @param uid the external UID
   * @return the raw response
   * @throws IllegalArgumentException when providerType is empty, "." or "..", or over 256
   *     characters
   */
  public Response setUidProvider(long id, String providerType, String uid) {
    Request request = Request.of("PUT", uidPath(id, providerType));
    return calls.call(Params.body(request, Map.of("uid", uid)));
  }

  /**
   * The customer's VIP membership state (GET /v1/customers/{id}/vip_info).
   *
   * @param id the customer id
   * @return the VIP state
   * @throws NotFoundException also when the API answers with a null body
   */
  public CustomerVipInfo vipInfo(long id) {
    return calls.one(Request.of("GET", path(id, "/vip_info")), CustomerVipInfo.class);
  }

  /**
   * The customer's paid and valid orders between two dates, both required (GET
   * /v1/customers/{id}/spending_overview).
   *
   * @param id the customer id
   * @param query start_date and end_date: "YYYY-MM-DD", a {@code LocalDate} or an {@code
   *     OffsetDateTime}
   * @return the overview
   * @throws NotFoundException also when the API answers with a null body
   */
  public CustomerSpendingOverview spendingOverview(long id, Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", path(id, "/spending_overview")), query, true);
    return calls.one(request, CustomerSpendingOverview.class);
  }

  /**
   * One page of the customer's service threads (GET /v1/customers/{id}/message_posts).
   *
   * @param id the customer id
   * @param query page, per_page, offset; null for none
   * @return the page
   */
  public Page<CustomerMessagePost> listMessagePosts(long id, Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", path(id, "/message_posts")), query);
    return calls.client().list(request, CustomerMessagePost.class);
  }

  /**
   * The first page of the customer's service threads (GET /v1/customers/{id}/message_posts).
   *
   * @param id the customer id
   * @return the page
   */
  public Page<CustomerMessagePost> listMessagePosts(long id) {
    return listMessagePosts(id, null);
  }

  /**
   * One page of the customer's orders, as full orders (GET /v1/customers/{id}/orders).
   *
   * @param id the customer id
   * @param query page, per_page, offset; null for none
   * @return the page
   */
  public Page<Order> listOrders(long id, Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", path(id, "/orders")), query);
    return calls.client().list(request, Order.class);
  }

  /**
   * The first page of the customer's orders (GET /v1/customers/{id}/orders).
   *
   * @param id the customer id
   * @return the page
   */
  public Page<Order> listOrders(long id) {
    return listOrders(id, null);
  }

  /**
   * Every order of the customer, page by page, lazily (GET /v1/customers/{id}/orders).
   *
   * @param id the customer id
   * @param query per_page and the start page; null for none
   * @return the orders of every page
   */
  public PageIterable<Order> allOrders(long id, Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", path(id, "/orders")), query);
    return calls.client().all(request, Order.class);
  }

  /**
   * Every order of the customer, page by page, lazily (GET /v1/customers/{id}/orders).
   *
   * @param id the customer id
   * @return the orders of every page
   */
  public PageIterable<Order> allOrders(long id) {
    return allOrders(id, null);
  }

  /**
   * The variants waiting in the customer's web-shop cart (GET
   * /v1/customers/{id}/customer_cart_items).
   *
   * @param id the customer id
   * @return the variants; empty when the cart is
   */
  public List<ProductVariant> cartItems(long id) {
    return calls.list(Request.of("GET", path(id, "/customer_cart_items")), ProductVariant.class);
  }

  /**
   * The line items of the customer's recent valid orders in a date range; every option is required
   * (GET /v1/customers/{id}/recent_purchases).
   *
   * @param id the customer id
   * @param query start_date, end_date (as for {@link #spendingOverview}) and max_products
   * @return the line items
   */
  public List<LineItem> recentPurchases(long id, Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", path(id, "/recent_purchases")), query, true);
    return calls.list(request, LineItem.class);
  }

  /**
   * One page of customers through the v2 endpoint, which can select by id (at most 100; paging is
   * then ignored) and include "uid_providers", "tags" or "vip_info" (GET /v2/customers).
   *
   * @param query page, per_page, offset, ids and include_params (collections); null for none
   * @return the page
   */
  public Page<Customer> listV2(Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", CUSTOMERS_V2), query);
    return calls.client().list(request, Customer.class);
  }

  /**
   * The first page of customers through the v2 endpoint (GET /v2/customers).
   *
   * @return the page
   */
  public Page<Customer> listV2() {
    return listV2(null);
  }

  /**
   * Every customer through the v2 endpoint, page by page, lazily (GET /v2/customers).
   *
   * @param query the options of {@link #listV2(Map)}; null for none
   * @return the customers of every page
   */
  public PageIterable<Customer> allV2(Map<String, ?> query) {
    Request request = Params.query(Request.of("GET", CUSTOMERS_V2), query);
    return calls.client().all(request, Customer.class);
  }

  /**
   * Every customer through the v2 endpoint, page by page, lazily (GET /v2/customers).
   *
   * @return the customers of every page
   */
  public PageIterable<Customer> allV2() {
    return allV2(null);
  }

  /**
   * The customer linked to an external identity. The query parameter is "provider" (verified live;
   * the Notion reference says "provider_type") (GET /v2/customers/by_uid_provider).
   *
   * @param providerType "line", "line_at" or "facebook"
   * @param uid the external UID
   * @return the customer
   * @throws NotFoundException also for an unknown uid (the API answers 200 null)
   */
  public Customer getByUidProvider(String providerType, String uid) {
    Request request =
        Request.of("GET", CUSTOMERS_V2 + "/by_uid_provider")
            .withQuery("uid", uid)
            .withQuery("provider", providerType);
    return calls.one(request, Customer.class);
  }

  /**
   * Authorises a customer through an external identity and returns the matching customer; shapes
   * follow the v2 Postman collection (POST /v2/customer_oauth).
   *
   * @param identity uid and provider
   * @return the customer
   */
  public Customer oauth(Map<String, ?> identity) {
    return calls.object(
        Params.body(Request.of("POST", "/v2/customer_oauth"), identity), Customer.class);
  }

  private static String path(long customerId, String suffix) {
    return CUSTOMERS + "/" + customerId + suffix;
  }

  private static String uidPath(long customerId, String providerType) {
    return path(customerId, "/uid_providers/" + Params.segment(providerType));
  }
}
