package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.exception.ApiException;
import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.exception.TransportException;
import cc.alphacore.cyberbiz.http.Backoff;
import cc.alphacore.cyberbiz.http.Clock;
import cc.alphacore.cyberbiz.http.JdkTransport;
import cc.alphacore.cyberbiz.http.SystemClock;
import cc.alphacore.cyberbiz.http.Transport;
import cc.alphacore.cyberbiz.http.TransportRequest;
import cc.alphacore.cyberbiz.json.Json;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.pagination.PageIterable;
import cc.alphacore.cyberbiz.pagination.Pagination;
import cc.alphacore.cyberbiz.resource.CustomerService;
import cc.alphacore.cyberbiz.resource.OrderService;
import cc.alphacore.cyberbiz.resource.ProductService;
import cc.alphacore.cyberbiz.resource.ShopService;
import java.io.IOException;
import java.net.URI;
import java.time.Duration;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/**
 * Talks to the CYBERBIZ API on behalf of exactly one Shop. It keeps to the platform's rate limit,
 * retries 429 and gateway errors, and turns error responses into typed exceptions. Immutable and
 * safe to share between threads.
 *
 * <pre>{@code
 * CyberbizClient client = CyberbizClient.builder(System.getenv("CYBERBIZ_API_TOKEN")).build();
 * Response shop = client.send(Request.of("GET", "/v1/shop"));
 * }</pre>
 */
public final class CyberbizClient {

  /** The SDK version; equal to {@code java/VERSION} and the Maven artifact version. */
  public static final String VERSION = "0.1.1";

  /** The CYBERBIZ API host shared by every Shop, used when the builder sets none. */
  public static final URI DEFAULT_BASE_URL = URI.create("https://app-store-api.cyberbiz.io/");

  /** The platform limit of requests per second. */
  public static final double DEFAULT_RATE_LIMIT = 5.0;

  /** Retries after a 429, or a 502/503/504 or network error on an idempotent request. */
  public static final int DEFAULT_MAX_RETRIES = 3;

  private static final Set<Integer> RETRYABLE_STATUS = Set.of(429, 502, 503, 504);
  private static final String USER_AGENT = "cyberbiz-sdk-java/" + VERSION;

  private final String token;
  private final URI baseUrl;
  private final Transport transport;
  private final Clock clock;
  private final RateLimiter limiter;
  private final int maxRetries;
  private final Backoff backoff;

  private CyberbizClient(Builder b) {
    this.token = b.token;
    this.baseUrl = b.baseUrl;
    this.transport = b.transport;
    this.clock = b.clock;
    this.limiter = b.rateLimit > 0 ? new RateLimiter(b.rateLimit, b.clock) : null;
    this.maxRetries = b.maxRetries;
    this.backoff = b.backoff;
  }

  /**
   * Starts building a client for one Shop.
   *
   * @param token the Shop's API token; must not be blank
   * @return a builder with the defaults
   */
  public static Builder builder(String token) {
    return new Builder(token);
  }

  /** Returns the base URL every request is sent to; it always ends in "/". */
  public URI baseUrl() {
    return baseUrl;
  }

  /** Returns the Shop profile and this App's settings on it. */
  public ShopService shop() {
    return new ShopService(this);
  }

  /** Returns products, their variants, options, tags and shipping bindings. */
  public ProductService products() {
    return new ProductService(this);
  }

  /**
   * Sends a request and returns the 2xx response. Retries 429, honouring {@code Retry-After}; for
   * idempotent methods (GET, HEAD, PUT, DELETE, OPTIONS) also 502/503/504 and network errors. POST
   * and PATCH are never repeated after a gateway error or a network failure, since the server may
   * already have acted on them.
   *
   * @param request the request
   * @return the successful response
   * @throws ApiException for an error response that survived every retry
   * @throws TransportException when the last attempt failed before a response
   */
  public Response send(Request request) {
    Objects.requireNonNull(request, "request");
    TransportRequest built = build(request);
    Response response = null;
    IOException failure = null;
    for (int attempt = 0; ; attempt++) {
      failure = null;
      try {
        acquireSlot(request);
        response = transport.send(built);
      } catch (IOException e) {
        response = null;
        failure = e;
      }
      if (!shouldRetry(request, response, attempt)) {
        break;
      }
      pause(request, retryDelay(response, attempt + 1));
    }
    if (response == null) {
      throw new TransportException(describe(request) + ": " + failure.getMessage(), failure);
    }
    ErrorMapper.check(request, response);
    return response;
  }

  /**
   * Sends a request to a list endpoint and decodes one page. The request is sent as given; set
   * {@code page} and {@code per_page} with {@link Request#withQuery} to choose the page. An empty
   * body or the JSON literal null counts as an empty page.
   *
   * <pre>{@code
   * Request request = Request.of("GET", "/v1/products").withQuery("page", 2);
   * Page<Product> page = client.list(request, Product.class);
   * }</pre>
   *
   * @param request the request
   * @param itemType the class each item decodes into, e.g. a record or {@code JsonObject}
   * @param <T> the item type
   * @return the items with the pagination headers
   * @throws ApiException for an error response that survived every retry
   * @throws TransportException when the last attempt failed before a response
   * @throws DecodeException when the body is not a JSON array of {@code itemType}
   */
  public <T> Page<T> list(Request request, Class<T> itemType) {
    Objects.requireNonNull(itemType, "itemType");
    Response response = send(request);
    List<T> items;
    try {
      items = Json.decodeList(response.body(), itemType);
    } catch (DecodeException e) {
      String detail = e.getMessage().replaceFirst("^cyberbiz: ", "");
      throw new DecodeException(describe(request) + ": " + detail, e);
    }
    return new Page<>(items, Pagination.from(response), response);
  }

  /**
   * Walks every page of a list endpoint lazily, as an {@link Iterable} or a {@code Stream}. It
   * starts at the request's {@code page} (default 1) with {@code per_page} 50 unless the request
   * sets it, follows {@code X-Next-Page}, and stops at the last page or an empty page, so it sends
   * exactly one request per page. Nothing is sent until iteration starts.
   *
   * <pre>{@code
   * for (Product p : client.all(Request.of("GET", "/v1/products"), Product.class)) { ... }
   * }</pre>
   *
   * @param request the request; its {@code page} parameter is the first page
   * @param itemType the class each item decodes into
   * @param <T> the item type
   * @return the items of every page in order; iterating again starts over
   */
  public <T> PageIterable<T> all(Request request, Class<T> itemType) {
    Objects.requireNonNull(request, "request");
    Objects.requireNonNull(itemType, "itemType");
    return new PageIterable<>(
        page -> list(PageRequests.withPage(request, page), itemType),
        PageRequests.startPage(request));
  }

  private void acquireSlot(Request request) {
    if (limiter != null) {
      try {
        limiter.acquire();
      } catch (InterruptedException e) {
        throw interrupted(request, e);
      }
    }
  }

  private void pause(Request request, Duration delay) {
    try {
      clock.sleep(delay);
    } catch (InterruptedException e) {
      throw interrupted(request, e);
    }
  }

  private static TransportException interrupted(Request request, InterruptedException e) {
    Thread.currentThread().interrupt();
    return new TransportException(describe(request) + ": interrupted", e);
  }

  private boolean shouldRetry(Request request, Response response, int attempt) {
    if (attempt >= maxRetries) {
      return false;
    }
    if (response == null) {
      return request.isIdempotent();
    }
    int status = response.statusCode();
    // A 502/503/504 may arrive after the server already acted on the request, so only a 429
    // (rejected before processing) is safe to repeat for POST and PATCH.
    return request.isIdempotent() ? RETRYABLE_STATUS.contains(status) : status == 429;
  }

  private Duration retryDelay(Response response, int attempt) {
    if (response != null) {
      var retryAfter = Backoff.retryAfter(response.header("Retry-After"), clock.now());
      if (retryAfter.isPresent()) {
        return retryAfter.get();
      }
    }
    return backoff.delay(attempt);
  }

  private TransportRequest build(Request request) {
    Map<String, String> headers = new LinkedHashMap<>();
    headers.put("Authorization", "Bearer " + token);
    headers.put("Accept", "application/json");
    headers.put("User-Agent", USER_AGENT);
    if (request.body() != null) {
      headers.put("Content-Type", "application/json");
    }
    headers.putAll(request.headers());
    return new TransportRequest(request.method(), uri(request), headers, request.body());
  }

  private URI uri(Request request) {
    String url = baseUrl + request.path().replaceFirst("^/+", "");
    String query = QueryEncoder.encode(request.query());
    if (!query.isEmpty()) {
      url += (url.contains("?") ? "&" : "?") + query;
    }
    return URI.create(url);
  }

  private static String describe(Request request) {
    return "cyberbiz: " + request.method() + " /" + request.path().replaceFirst("^/+", "");
  }

  /**
   * Orders, their fulfillments, payments, returns and e-tickets, and shipping labels.
   *
   * @return the order service
   */
  public OrderService orders() {
    return new OrderService(this);
  }

  /**
   * Customers, their orders, service threads, VIP state and login identities.
   *
   * @return the customer service
   */
  public CustomerService customers() {
    return new CustomerService(this);
  }

  /** Describes the client without the token, so it is safe to log. */
  @Override
  public String toString() {
    return "CyberbizClient{baseUrl=" + baseUrl + ", version=" + VERSION + "}";
  }

  /** Configures a {@link CyberbizClient}. Not thread-safe; the built client is. */
  public static final class Builder {
    private final String token;
    private URI baseUrl = DEFAULT_BASE_URL;
    private Transport transport;
    private Clock clock = SystemClock.INSTANCE;
    private double rateLimit = DEFAULT_RATE_LIMIT;
    private int maxRetries = DEFAULT_MAX_RETRIES;
    private Backoff backoff = new Backoff();

    private Builder(String token) {
      Objects.requireNonNull(token, "token");
      if (token.isBlank()) {
        throw new IllegalArgumentException("token must not be blank");
      }
      this.token = token;
    }

    /**
     * Overrides the API host; only for tests or a proxy.
     *
     * @param baseUrl an absolute http or https URL; a trailing "/" is added when missing
     * @return this builder
     */
    public Builder baseUrl(URI baseUrl) {
      Objects.requireNonNull(baseUrl, "baseUrl");
      String scheme = baseUrl.getScheme();
      if (!("https".equals(scheme) || "http".equals(scheme)) || baseUrl.getHost() == null) {
        throw new IllegalArgumentException("baseUrl must be an absolute http(s) URL: " + baseUrl);
      }
      String text = baseUrl.toString();
      this.baseUrl = text.endsWith("/") ? baseUrl : URI.create(text + "/");
      return this;
    }

    /**
     * Sets how requests are sent; the default is a {@link JdkTransport}.
     *
     * @param transport the transport
     * @return this builder
     */
    public Builder transport(Transport transport) {
      this.transport = Objects.requireNonNull(transport, "transport");
      return this;
    }

    /**
     * Sets the clock; for tests.
     *
     * @param clock the clock
     * @return this builder
     */
    public Builder clock(Clock clock) {
      this.clock = Objects.requireNonNull(clock, "clock");
      return this;
    }

    /**
     * Sets the most requests started per second; 0 turns limiting off.
     *
     * @param perSecond requests per second, not negative
     * @return this builder
     */
    public Builder rateLimit(double perSecond) {
      if (!(perSecond >= 0) || Double.isInfinite(perSecond)) {
        throw new IllegalArgumentException("rateLimit must not be negative: " + perSecond);
      }
      this.rateLimit = perSecond;
      return this;
    }

    /**
     * Sets how many times a request is retried; 0 turns retries off.
     *
     * @param maxRetries retries, not negative
     * @return this builder
     */
    public Builder maxRetries(int maxRetries) {
      if (maxRetries < 0) {
        throw new IllegalArgumentException("maxRetries must not be negative: " + maxRetries);
      }
      this.maxRetries = maxRetries;
      return this;
    }

    /**
     * Sets the delays between retries when the server sends no {@code Retry-After}.
     *
     * @param backoff the backoff
     * @return this builder
     */
    public Builder backoff(Backoff backoff) {
      this.backoff = Objects.requireNonNull(backoff, "backoff");
      return this;
    }

    /** Builds the client. */
    public CyberbizClient build() {
      if (transport == null) {
        transport = new JdkTransport();
      }
      return new CyberbizClient(this);
    }
  }
}
