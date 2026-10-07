package cc.alphacore.cyberbiz.webhook;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.BodyTooLargeException;
import cc.alphacore.cyberbiz.exception.CyberbizException;
import cc.alphacore.cyberbiz.exception.InvalidDomainSignatureException;
import cc.alphacore.cyberbiz.exception.InvalidSignatureException;
import cc.alphacore.cyberbiz.exception.MissingHeaderException;
import cc.alphacore.cyberbiz.exception.UnknownShopException;
import cc.alphacore.cyberbiz.exception.WebhookException;
import cc.alphacore.cyberbiz.webhook.WebhookSamples.Sample;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Optional;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;

class WebhookParserTest {

  private static final WebhookParser PARSER =
      new WebhookParser(new StaticSecret(WebhookSamples.SECRET));

  static List<Sample> samples() {
    return WebhookSamples.all();
  }

  private static Sample paid() {
    return WebhookSamples.get("orders/paid");
  }

  @Test
  void coversEveryDocumentedEvent() {
    List<String> documented = samples().stream().map(Sample::event).toList();

    assertEquals(Arrays.stream(EventType.values()).map(EventType::value).toList(), documented);
  }

  // AC6: hex and base64 pass, a wrong signature fails.

  @ParameterizedTest
  @MethodSource("samples")
  void authenticatesTheHexSample(Sample sample) {
    Event event = PARSER.parse(sample.headers(), sample.bytes());

    assertEquals(sample.event(), event.type());
    assertEquals(Optional.of(sample.event()), event.eventType().map(EventType::value));
    assertEquals("example.cyberbiz.co", event.shopDomain());
    assertEquals("www.example-shop.com", event.customDomain());
    assertEquals(sample.body(), event.body());
    assertTrue(event.payload().size() > 0);
  }

  @ParameterizedTest
  @MethodSource("samples")
  void authenticatesTheBase64Form(Sample sample) {
    Map<String, String> headers = sample.with(WebhookParser.HEADER_SIGNATURE, sample.base64());

    assertEquals(sample.event(), PARSER.parse(headers, sample.bytes()).type());
  }

  @Test
  void acceptsUppercaseHex() {
    Sample s = paid();
    String upper = s.headers().get(WebhookParser.HEADER_SIGNATURE).toUpperCase(Locale.ROOT);

    assertEquals(
        "orders/paid",
        PARSER.parse(s.with(WebhookParser.HEADER_SIGNATURE, upper), s.bytes()).type());
  }

  @Test
  void rejectsAWrongSignature() {
    Sample s = paid();
    String wrong = Signature.sign(s.bytes(), "another-secret");

    InvalidSignatureException e =
        assertThrows(
            InvalidSignatureException.class,
            () -> PARSER.parse(s.with(WebhookParser.HEADER_SIGNATURE, wrong), s.bytes()));
    assertEquals(401, e.httpStatus());
    assertFalse(e.getMessage().contains(WebhookSamples.SECRET));
  }

  @Test
  void rejectsATamperedBody() {
    Sample s = paid();
    byte[] tampered = s.body().replace("\"id\": 1", "\"id\": 2").getBytes(StandardCharsets.UTF_8);

    assertThrows(InvalidSignatureException.class, () -> PARSER.parse(s.headers(), tampered));
  }

  @Test
  void rejectsAWrongDomainSignatureSeparately() {
    Sample s = paid();
    Map<String, String> headers =
        s.with(
            WebhookParser.HEADER_DOMAIN_SIGNATURE,
            Signature.signDomain("other.cyberbiz.co", WebhookSamples.SECRET));

    assertThrows(InvalidDomainSignatureException.class, () -> PARSER.parse(headers, s.bytes()));
    WebhookParser lenient =
        new WebhookParser(
            new StaticSecret(WebhookSamples.SECRET), WebhookParser.MAX_BODY_BYTES, false);
    assertEquals("orders/paid", lenient.parse(headers, s.bytes()).type());
  }

  @Test
  void acceptsAnAbsentDomainSignature() {
    Sample s = paid();
    Event event = PARSER.parse(s.with(WebhookParser.HEADER_DOMAIN_SIGNATURE, null), s.bytes());

    assertEquals("", event.domainSignature());
  }

  // AC-J6a: missing header, unknown shop, body too large.

  @ParameterizedTest
  @ValueSource(
      strings = {
        WebhookParser.HEADER_EVENT,
        WebhookParser.HEADER_DOMAIN,
        WebhookParser.HEADER_SIGNATURE
      })
  void requiresTheCyberbizHeaders(String name) {
    Sample s = paid();

    MissingHeaderException absent =
        assertThrows(
            MissingHeaderException.class, () -> PARSER.parse(s.with(name, null), s.bytes()));
    assertEquals("webhook: missing header " + name, absent.getMessage());
    assertEquals(400, absent.httpStatus());
    assertThrows(MissingHeaderException.class, () -> PARSER.parse(s.with(name, " "), s.bytes()));
  }

  @Test
  void rejectsAnUnknownShop() {
    Sample s = paid();
    WebhookParser otherShop =
        new WebhookParser(new SecretMap(Map.of("other.cyberbiz.co", "other-secret")));
    WebhookParser noSecret = new WebhookParser(new StaticSecret(""));

    UnknownShopException e =
        assertThrows(UnknownShopException.class, () -> otherShop.parse(s.headers(), s.bytes()));
    assertEquals("webhook: unknown shop \"example.cyberbiz.co\"", e.getMessage());
    assertEquals(401, e.httpStatus());
    assertThrows(UnknownShopException.class, () -> noSecret.parse(s.headers(), s.bytes()));
  }

  @Test
  void limitsTheBodySize() {
    Sample s = paid();
    WebhookParser small = new WebhookParser(new StaticSecret(WebhookSamples.SECRET), 10, true);

    BodyTooLargeException e =
        assertThrows(BodyTooLargeException.class, () -> small.parse(s.headers(), s.bytes()));
    assertEquals(413, e.httpStatus());

    byte[] atLimit = new byte[WebhookParser.MAX_BODY_BYTES];
    byte[] overLimit = new byte[WebhookParser.MAX_BODY_BYTES + 1];
    assertEquals("orders/paid", PARSER.parse(signed(s, atLimit), atLimit).type());
    assertThrows(BodyTooLargeException.class, () -> PARSER.parse(signed(s, overLimit), overLimit));
  }

  @Test
  void rejectsANegativeLimit() {
    StaticSecret secret = new StaticSecret(WebhookSamples.SECRET);

    assertThrows(IllegalArgumentException.class, () -> new WebhookParser(secret, -1, true));
  }

  // AC-J6b: a multi-shop resolver picks the right secret.

  @Test
  void resolvesTheSecretPerShop() {
    Sample s = paid();
    WebhookParser parser =
        new WebhookParser(
            new SecretMap(
                Map.of(
                    "Example.cyberbiz.co",
                    WebhookSamples.SECRET,
                    "other.cyberbiz.co",
                    "other-secret")));

    assertEquals("example.cyberbiz.co", parser.parse(s.headers(), s.bytes()).shopDomain());

    Map<String, String> other = s.with(WebhookParser.HEADER_DOMAIN, "other.cyberbiz.co");
    other.put(WebhookParser.HEADER_SIGNATURE, Signature.sign(s.bytes(), "other-secret"));
    other.put(
        WebhookParser.HEADER_DOMAIN_SIGNATURE,
        Signature.signDomain("other.cyberbiz.co", "other-secret"));
    assertEquals("other.cyberbiz.co", parser.parse(other, s.bytes()).shopDomain());

    WebhookParser swapped =
        new WebhookParser(new SecretMap(Map.of("example.cyberbiz.co", "other-secret")));
    assertThrows(InvalidSignatureException.class, () -> swapped.parse(s.headers(), s.bytes()));
  }

  @Test
  void asksTheResolverForTheShopDomain() {
    Sample s = paid();
    List<String> asked = new ArrayList<>();
    WebhookParser parser =
        new WebhookParser(
            domain -> {
              asked.add(domain);
              return Optional.of(WebhookSamples.SECRET);
            });

    parser.parse(s.headers(), s.bytes());

    assertEquals(List.of("example.cyberbiz.co"), asked);
  }

  // Input shapes and the Event.

  @Test
  void acceptsHeaderNamesInAnyCaseAndListValues() {
    Sample s = paid();
    Map<String, List<String>> headers = new HashMap<>();
    s.headers()
        .forEach((name, value) -> headers.put(name.toLowerCase(Locale.ROOT), List.of(value)));

    Event event = PARSER.parse(headers, s.bytes());

    assertEquals("orders/paid", event.type());
    assertEquals("CyberbizAppWebhook/1.0", event.header("USER-AGENT"));
    assertEquals("", event.header("X-Missing"));
  }

  /** A caller-defined payload type. */
  record Paid(long id, String token) {}

  @Test
  void decodesThePayload() {
    Sample s = paid();
    Event event = PARSER.parse(s.headers(), s.bytes());

    assertEquals(1, event.payload().get("id").getAsLong());
    assertEquals(new Paid(1, "synthetic-token-do-not-use"), event.decode(Paid.class));
    assertEquals(EventType.ORDERS_PAID, event.eventType().orElseThrow());
  }

  @Test
  void reportsABodyThatIsNotAJsonObject() {
    Sample s = paid();
    byte[] array = "[1]".getBytes(StandardCharsets.UTF_8);
    Event event = PARSER.parse(signed(s, array), array);

    assertThrows(CyberbizException.class, event::payload);
  }

  @Test
  void everyRejectionIsAWebhookException() {
    Sample s = paid();

    assertThrows(WebhookException.class, () -> PARSER.parse(Map.of(), s.bytes()));
  }

  @Test
  void neverShowsTheSecret() {
    assertFalse(new StaticSecret("top-secret-value").toString().contains("top-secret-value"));
    assertFalse(
        new SecretMap(Map.of("a.cyberbiz.co", "top-secret-value"))
            .toString()
            .contains("top-secret-value"));
  }

  private static Map<String, String> signed(Sample s, byte[] body) {
    return s.with(WebhookParser.HEADER_SIGNATURE, Signature.sign(body, WebhookSamples.SECRET));
  }
}
