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
    assertEquals(Optional.of(EventType.APPS_UNINSTALL), EventType.of("apps/uninstall"));
    assertEquals(Optional.empty(), EventType.of("orders/update"));
  }
}
