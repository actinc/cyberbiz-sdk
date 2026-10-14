package cc.alphacore.cyberbiz.webhook;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.webhook.WebhookSamples.Sample;
import java.nio.charset.StandardCharsets;
import java.util.Locale;
import java.util.Optional;
import org.junit.jupiter.api.Test;

class SignatureTest {

  private static final String SECRET = WebhookSamples.SECRET;

  @Test
  void signsLikeTheDocumentedSamples() {
    Sample s = WebhookSamples.get("orders/paid");

    assertEquals(
        s.headers().get(WebhookParser.HEADER_SIGNATURE), Signature.sign(s.bytes(), SECRET));
    assertEquals(
        s.headers().get(WebhookParser.HEADER_DOMAIN_SIGNATURE),
        Signature.signDomain("example.cyberbiz.co", SECRET));
  }

  @Test
  void verifiesHexBase64AndPaddedValues() {
    Sample s = WebhookSamples.get("bonus_points/create");
    String hex = s.headers().get(WebhookParser.HEADER_SIGNATURE);

    assertTrue(Signature.verify(s.bytes(), hex, SECRET));
    assertTrue(Signature.verify(s.bytes(), " " + hex + "\n", SECRET));
    assertTrue(Signature.verify(s.bytes(), s.base64(), SECRET));
    assertFalse(Signature.verify(s.bytes(), s.base64().toLowerCase(Locale.ROOT), SECRET));
    assertFalse(Signature.verify(s.bytes(), hex.substring(1), SECRET));
  }

  @Test
  void neverVerifiesWithAnEmptySecretOrSignature() {
    byte[] body = "{}".getBytes(StandardCharsets.UTF_8);

    assertFalse(Signature.verify(body, "", SECRET));
    assertFalse(Signature.verify(body, "  ", SECRET));
    assertFalse(Signature.verify(body, Signature.sign(body, SECRET), ""));
  }

  @Test
  void namesEventsAsDocumented() {
    assertEquals("orders", EventType.ORDERS_PARTIAL_RETURN.resource());
    assertEquals("partial_return", EventType.ORDERS_PARTIAL_RETURN.action());
    assertEquals("send_sms_finish", EventType.SEND_SMS_FINISH.resource());
    assertEquals("", EventType.SEND_SMS_FINISH.action());
    assertEquals("stock_receipts", EventType.STOCK_RECEIPTS_DONE.resource());
    assertEquals(Optional.of(EventType.APPS_UNINSTALL), EventType.of("apps/uninstall"));
    assertEquals(Optional.of(EventType.ORDERS_UPDATE), EventType.of("orders/update"));
    assertEquals(Optional.empty(), EventType.of("orders/teleported"));
  }

  /**
   * The platform signs hex(HMAC-SHA256(secret, base64(domain))); the vector is shared with the Go
   * and PHP SDKs.
   */
  @Test
  void domainSignatureSignsTheBase64Domain() {
    String want = "db34ece1aebadba93c127e010f3f913cc5541e0ec91376152efff0f224e96a40";
    String raw = "97e6b3fd6e226afc59eee357b9649ed557414abd6b84d69aefef43ae6fea50d5";
    assertEquals(want, Signature.signDomain("example.cyberbiz.co", "app-secret"));
    assertTrue(Signature.verifyDomain("example.cyberbiz.co", want, "app-secret"));
    assertFalse(Signature.verifyDomain("example.cyberbiz.co", raw, "app-secret"));
  }
}
