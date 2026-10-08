package cc.alphacore.cyberbiz.webhook;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

import cc.alphacore.cyberbiz.exception.InvalidSignatureException;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import org.junit.jupiter.api.Test;

/**
 * Cross-shop replay (CBSDK-39): every shop has its own App Secret, so shop A's delivery resent with
 * shop B's X-Cyberbiz-Domain fails the body Signature under B's secret, whatever Domain Signature
 * it carries. The Domain Signature stays optional (checked only when present).
 */
class CrossShopReplayTest {

  private static final String SHOP_A = "shop-a.cyberbiz.co";
  private static final String SHOP_B = "shop-b.cyberbiz.co";
  private static final String SECRET_A = "shop-a-test-secret";
  private static final String SECRET_B = "shop-b-test-secret";
  private static final byte[] BODY =
      "{\"id\":1001,\"name\":\"#1001\"}".getBytes(StandardCharsets.UTF_8);
  private static final WebhookParser PARSER =
      new WebhookParser(new SecretMap(Map.of(SHOP_A, SECRET_A, SHOP_B, SECRET_B)));

  /** Shop A's body and body Signature, labelled with {@code domain}. */
  private static Map<String, String> shopADelivery(String domain) {
    Map<String, String> headers = new HashMap<>();
    headers.put(WebhookParser.HEADER_EVENT, "orders/paid");
    headers.put(WebhookParser.HEADER_DOMAIN, domain);
    headers.put(WebhookParser.HEADER_SIGNATURE, Signature.sign(BODY, SECRET_A));
    return headers;
  }

  private static void assertRejectedAsInvalidSignature(Map<String, String> headers) {
    InvalidSignatureException e =
        assertThrows(InvalidSignatureException.class, () -> PARSER.parse(headers, BODY));
    assertEquals(401, e.httpStatus());
  }

  @Test
  void rejectsShopADeliveryAsShopBWithoutDomainSignature() {
    assertRejectedAsInvalidSignature(shopADelivery(SHOP_B));
  }

  @Test
  void rejectsShopADeliveryAsShopBWithShopBDomainSignature() {
    Map<String, String> headers = shopADelivery(SHOP_B);
    headers.put(WebhookParser.HEADER_DOMAIN_SIGNATURE, Signature.signDomain(SHOP_B, SECRET_B));

    assertRejectedAsInvalidSignature(headers);
  }

  @Test
  void rejectsShopADeliveryAsShopBWithShopADomainSignature() {
    Map<String, String> headers = shopADelivery(SHOP_B);
    headers.put(WebhookParser.HEADER_DOMAIN_SIGNATURE, Signature.signDomain(SHOP_A, SECRET_A));

    assertRejectedAsInvalidSignature(headers);
  }

  @Test
  void acceptsShopADeliveryAsShopA() {
    Map<String, String> headers = shopADelivery(SHOP_A);
    headers.put(WebhookParser.HEADER_DOMAIN_SIGNATURE, Signature.signDomain(SHOP_A, SECRET_A));

    assertEquals(SHOP_A, PARSER.parse(headers, BODY).shopDomain());
  }
}
