package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.ApiException;
import cc.alphacore.cyberbiz.exception.AuthenticationException;
import cc.alphacore.cyberbiz.exception.ForbiddenException;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.exception.RateLimitException;
import cc.alphacore.cyberbiz.exception.ServerException;
import cc.alphacore.cyberbiz.exception.TransportException;
import cc.alphacore.cyberbiz.exception.ValidationException;
import cc.alphacore.cyberbiz.http.Backoff;
import cc.alphacore.cyberbiz.http.TransportRequest;
import java.net.URI;
import java.time.Duration;
import java.util.List;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

class SendTest {
  private static final String TOKEN = "synthetic-token-0123456789";
  private static final Request GET_SHOP = Request.of("GET", "/v1/shop");

  private final FakeTransport transport = new FakeTransport();
  private final FakeClock clock = new FakeClock();

  /** No rate limit and no jitter, so sleeps are exactly the retry delays. */
  private CyberbizClient client() {
    return CyberbizClient.builder(TOKEN)
        .baseUrl(URI.create("https://api.example.test"))
        .transport(transport)
        .clock(clock)
        .rateLimit(0)
        .backoff(new Backoff(Duration.ofMillis(500), Duration.ofSeconds(8), () -> 0))
        .build();
  }

  @Test
  void sendsTheTokenAndJsonHeaders() {
    transport.reply(200, "{}");

    client().send(Request.of("POST", "v1/products").withBody("{\"title\":\"Synthetic Tea\"}"));

    TransportRequest sent = transport.requests.get(0);
    assertEquals("POST", sent.method());
    assertEquals(URI.create("https://api.example.test/v1/products"), sent.uri());
    assertEquals("Bearer " + TOKEN, sent.headers().get("Authorization"));
    assertEquals("application/json", sent.headers().get("Accept"));
    assertEquals("application/json", sent.headers().get("Content-Type"));
    assertEquals("cyberbiz-sdk-java/" + CyberbizClient.VERSION, sent.headers().get("User-Agent"));
    assertEquals("{\"title\":\"Synthetic Tea\"}", sent.body());
    assertFalse(sent.toString().contains(TOKEN));
  }

  @Test
  void encodesTheQueryLikeThePhpSdk() {
    transport.reply(200, "[]");

    client()
        .send(
            GET_SHOP
                .withQuery("ids", 1, 2)
                .withQuery("q", "綠茶 tea*")
                .withQuery("published", true)
                .withQuery("skip", (Object) null));

    assertEquals(
        "https://api.example.test/v1/shop?ids=1&ids=2&q=%E7%B6%A0%E8%8C%B6%20tea%2A&published=true",
        transport.requests.get(0).uri().toString());
  }

  @Test
  void retriesA429AfterRetryAfterSeconds() {
    transport.reply(429, "", "Retry-After", "2").reply(200, "{\"id\":1}");

    Response response = client().send(GET_SHOP);

    assertEquals(200, response.statusCode());
    assertEquals(List.of(Duration.ofSeconds(2)), clock.sleeps);
  }

  @Test
  void retriesAfterARetryAfterDate() {
    transport.reply(503, "", "Retry-After", "Wed, 07 Oct 2026 00:00:05 GMT").reply(200, "{}");

    client().send(GET_SHOP);

    assertEquals(List.of(Duration.ofSeconds(5)), clock.sleeps);
  }

  @Test
  void backsOffExponentiallyThenThrowsTheLastError() {
    for (int i = 0; i < 4; i++) {
      transport.reply(503, "{\"error\":\"維護中\"}", "X-Request-Id", "req-" + i);
    }

    ServerException e = assertThrows(ServerException.class, () -> client().send(GET_SHOP));

    assertEquals(4, transport.requests.size());
    assertEquals(
        List.of(Duration.ofMillis(500), Duration.ofSeconds(1), Duration.ofSeconds(2)),
        clock.sleeps);
    assertEquals(503, e.statusCode());
    assertEquals("req-3", e.requestId());
    assertEquals(List.of("維護中"), e.messages());
    assertEquals("cyberbiz: GET /v1/shop: 503 維護中", e.getMessage());
  }

  @Test
  void throwsRateLimitExceptionWhen429Persists() {
    for (int i = 0; i < 4; i++) {
      transport.reply(429, "", "Retry-After", "1");
    }

    assertThrows(RateLimitException.class, () -> client().send(GET_SHOP));
    assertEquals(Duration.ofSeconds(3), clock.totalSleep());
  }

  @Test
  void doesNotRetryOtherErrors() {
    transport.reply(400, "{\"message\":\"bad\"}");

    ApiException e = assertThrows(ApiException.class, () -> client().send(GET_SHOP));

    assertEquals(1, transport.requests.size());
    assertEquals(400, e.statusCode());
  }

  @ParameterizedTest
  @CsvSource({"401", "403", "404", "422", "500"})
  void mapsStatusesToTypedExceptions(int status) {
    transport.reply(status, "{\"errors\":{\"title\":[\"can't be blank\"]}}", "X-Request-Id", "r1");

    ApiException e = assertThrows(ApiException.class, () -> client().send(GET_SHOP));

    Class<?> expected =
        switch (status) {
          case 401 -> AuthenticationException.class;
          case 403 -> ForbiddenException.class;
          case 404 -> NotFoundException.class;
          case 422 -> ValidationException.class;
          default -> ServerException.class;
        };
    assertInstanceOf(expected, e);
    assertEquals(List.of("title: can't be blank"), e.messages());
    assertEquals("r1", e.requestId());
    assertFalse(e.getMessage().contains(TOKEN));
    assertFalse(e.toString().contains(TOKEN));
  }

  @Test
  void treatsA200ThatIsOnlyAnErrorObjectAsAnError() {
    transport.reply(200, "{\"error\":\"找不到商品\"}");

    ApiException e = assertThrows(ApiException.class, () -> client().send(GET_SHOP));

    assertEquals(200, e.statusCode());
    assertEquals(List.of("找不到商品"), e.messages());
  }

  @Test
  void retriesANetworkErrorForAnIdempotentRequest() {
    transport.fail("connection reset").reply(200, "{}");

    assertEquals(200, client().send(GET_SHOP).statusCode());
    assertEquals(2, transport.requests.size());
  }

  @Test
  void doesNotRetryANetworkErrorForAPost() {
    transport.fail("connection reset");

    TransportException e =
        assertThrows(TransportException.class, () -> client().send(Request.of("POST", "/v1/x")));

    assertEquals(1, transport.requests.size());
    assertTrue(e.getMessage().contains("connection reset"));
  }

  @ParameterizedTest
  @CsvSource({"POST, 502", "POST, 503", "POST, 504", "PATCH, 504"})
  void doesNotRepeatAWriteAfterAGatewayError(String method, int status) {
    transport.reply(status, "");

    ServerException e =
        assertThrows(ServerException.class, () -> client().send(Request.of(method, "/v1/orders")));

    assertEquals(status, e.statusCode());
    assertEquals(1, transport.requests.size());
    assertTrue(clock.sleeps.isEmpty());
  }

  @Test
  void repeatsAWriteRejectedBy429() {
    transport.reply(429, "", "Retry-After", "1").reply(201, "{}");

    assertEquals(201, client().send(Request.of("POST", "/v1/orders")).statusCode());
    assertEquals(2, transport.requests.size());
  }

  @Test
  void spacesRequestsToTheRateLimit() {
    for (int i = 0; i < 10; i++) {
      transport.reply(200, "{}");
    }
    CyberbizClient limited =
        CyberbizClient.builder(TOKEN).transport(transport).clock(clock).build();

    for (int i = 0; i < 10; i++) {
      limited.send(GET_SHOP);
    }

    // The first request goes at once; the other nine wait 0.2 s each: 5 per second.
    assertEquals(9, clock.sleeps.size());
    assertTrue(clock.sleeps.stream().allMatch(d -> d.equals(Duration.ofMillis(200))));
  }

  @Test
  void stopsWhenInterruptedWhileWaiting() throws InterruptedException {
    transport.reply(503, "", "Retry-After", "1").reply(200, "{}");
    CyberbizClient real = CyberbizClient.builder(TOKEN).transport(transport).rateLimit(0).build();
    Thread.currentThread().interrupt();

    try {
      assertThrows(TransportException.class, () -> real.send(GET_SHOP));
      assertTrue(Thread.currentThread().isInterrupted());
    } finally {
      Thread.interrupted();
    }
  }
}
