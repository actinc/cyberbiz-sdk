package cc.alphacore.cyberbiz.webhook;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

import cc.alphacore.cyberbiz.exception.CyberbizException;
import cc.alphacore.cyberbiz.exception.InvalidDomainSignatureException;
import cc.alphacore.cyberbiz.exception.InvalidSignatureException;
import cc.alphacore.cyberbiz.exception.MissingHeaderException;
import cc.alphacore.cyberbiz.exception.UnknownShopException;
import cc.alphacore.cyberbiz.webhook.WebhookSamples.Sample;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

/** Regression tests from the security review: every path fails closed, one shop per secret. */
class WebhookSecurityTest {

  private static final String SECRET = WebhookSamples.SECRET;
  private static final String SHOP_A = "example.cyberbiz.co";
  private static final String SHOP_B = "other.cyberbiz.co";
  private static final WebhookParser PARSER = new WebhookParser(new StaticSecret(SECRET));
  private static final WebhookParser TWO_SHOPS =
      new WebhookParser(new SecretMap(Map.of(SHOP_A, SECRET, SHOP_B, "other-secret")));

  private static Sample paid() {
    return WebhookSamples.get("orders/paid");
  }

  @SuppressWarnings("NullOptional") // a misbehaving resolver must still fail closed
  @Test
  void failsClosedWhenTheResolverHasNoSecret() {
    Sample s = paid();
    List<SecretResolver> resolvers =
        List.of(
            domain -> Optional.empty(),
            domain -> Optional.of(""),
            domain -> null,
            new StaticSecret(""),
            new SecretMap(Map.of(SHOP_A, "")));

    for (SecretResolver resolver : resolvers) {
      WebhookParser parser = new WebhookParser(resolver);
      assertThrows(UnknownShopException.class, () -> parser.parse(s.headers(), s.bytes()));
    }
  }

  @Test
  void neverFallsBackToAnotherShopsSecret() {
    Sample s = paid();
    WebhookParser onlyB = new WebhookParser(new SecretMap(Map.of(SHOP_B, SECRET)));

    assertThrows(UnknownShopException.class, () -> onlyB.parse(s.headers(), s.bytes()));
  }

  @ParameterizedTest
  @ValueSource(strings = {"", " ", "\t"})
  void rejectsAnEmptySignatureHeader(String value) {
    Sample s = paid();
    Map<String, String> headers = s.with(WebhookParser.HEADER_SIGNATURE, value);

    assertThrows(MissingHeaderException.class, () -> PARSER.parse(headers, s.bytes()));
  }

  @ParameterizedTest
  @ValueSource(
      strings = {
        "zz",
        "not-hex-at-all!",
        "====",
        "äöü",
        "a4e1773d94c3cbfcabdaf7218dc1133445417a8cc31f007d2ea41f19ea203cd", // one char short
        "a4e1773d94c3cbfcabdaf7218dc1133445417a8cc31f007d2ea41f19ea203cd6a4", // two chars long
        "a4e1773d94c3cbfcabdaf7218dc1133445417a8cc31f007d2ea41f19ea203cd6 x",
        "pOF3PZTDy/yr2vchjcETNEVBeozDHwB9LqQfGeogPN", // base64 truncated
      })
  void rejectsAGarbageSignature(String value) {
    Sample s = paid();
    Map<String, String> headers = s.with(WebhookParser.HEADER_SIGNATURE, value);

    assertThrows(InvalidSignatureException.class, () -> PARSER.parse(headers, s.bytes()));
  }

  @Test
  void verifiesBeforeDecodingAndNeverSwallowsADecodingError() {
    Sample s = paid();
    byte[] broken = "{not json".getBytes(StandardCharsets.UTF_8);

    assertThrows(InvalidSignatureException.class, () -> PARSER.parse(s.headers(), broken));

    Map<String, String> signed =
        s.with(WebhookParser.HEADER_SIGNATURE, Signature.sign(broken, SECRET));
    Event event = PARSER.parse(signed, broken);
    assertThrows(CyberbizException.class, event::payload);
    assertThrows(CyberbizException.class, () -> event.decode(Map.class));
  }

  @Test
  void rejectsShopASignaturePresentedWithShopBDomain() {
    Sample s = paid();
    Map<String, String> asB = s.with(WebhookParser.HEADER_DOMAIN, SHOP_B);
    Map<String, String> asBWithoutDomainSignature = s.with(WebhookParser.HEADER_DOMAIN, SHOP_B);
    asBWithoutDomainSignature.remove(WebhookParser.HEADER_DOMAIN_SIGNATURE);

    assertThrows(InvalidSignatureException.class, () -> TWO_SHOPS.parse(asB, s.bytes()));
    assertThrows(
        InvalidSignatureException.class,
        () -> TWO_SHOPS.parse(asBWithoutDomainSignature, s.bytes()));
  }

  @Test
  void reportsExactlyTheShopWhoseSecretVerified() {
    Sample s = paid();
    List<String> asked = new ArrayList<>();
    WebhookParser parser =
        new WebhookParser(
            domain -> {
              asked.add(domain);
              return Optional.of(SECRET);
            });
    Map<String, String> headers =
        s.with(WebhookParser.HEADER_SHOP_DOMAIN, "www.someone-else.example");

    Event event = parser.parse(headers, s.bytes());

    assertEquals(List.of(event.shopDomain()), asked);
    assertEquals(SHOP_A, event.shopDomain());
    assertEquals("www.someone-else.example", event.customDomain());
  }

  @Test
  void caseVariantsOfTheDomainMapToOneShop() {
    Sample s = paid();
    Map<String, String> upper = s.with(WebhookParser.HEADER_DOMAIN, "EXAMPLE.Cyberbiz.co");

    // The Domain Signature covers the exact header bytes, so a re-cased domain fails it.
    assertThrows(InvalidDomainSignatureException.class, () -> TWO_SHOPS.parse(upper, s.bytes()));

    upper.remove(WebhookParser.HEADER_DOMAIN_SIGNATURE);
    assertEquals("EXAMPLE.Cyberbiz.co", TWO_SHOPS.parse(upper, s.bytes()).shopDomain());
    Map<String, String> upperB = s.with(WebhookParser.HEADER_DOMAIN, "OTHER.cyberbiz.co");
    upperB.remove(WebhookParser.HEADER_DOMAIN_SIGNATURE);
    assertThrows(InvalidSignatureException.class, () -> TWO_SHOPS.parse(upperB, s.bytes()));
  }
}
