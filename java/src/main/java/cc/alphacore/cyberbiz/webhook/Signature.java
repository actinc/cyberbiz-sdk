package cc.alphacore.cyberbiz.webhook;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.MessageDigest;
import java.util.Base64;
import java.util.HexFormat;
import java.util.Locale;
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;

/**
 * HMAC-SHA256 webhook signatures (ADR-0001). CYBERBIZ's documentation says the digest is base64,
 * but every production delivery carries 64 lowercase hex characters, so verification accepts hex
 * (either case) and base64, both compared in constant time with {@link MessageDigest#isEqual}.
 */
public final class Signature {

  private static final String ALGORITHM = "HmacSHA256";

  private Signature() {}

  /**
   * Returns the Signature CYBERBIZ sends for a body: lowercase hex HMAC-SHA256 keyed by the App
   * Secret.
   *
   * @param body the raw body
   * @param secret the App Secret, not empty
   * @return 64 lowercase hex characters
   */
  public static String sign(byte[] body, String secret) {
    return HexFormat.of().formatHex(mac(body, secret));
  }

  /**
   * Returns the Domain Signature for a Shop Domain.
   *
   * @param shopDomain the Shop Domain
   * @param secret the App Secret, not empty
   * @return 64 lowercase hex characters
   */
  public static String signDomain(String shopDomain, String secret) {
    return sign(shopDomain.getBytes(StandardCharsets.UTF_8), secret);
  }

  /**
   * Whether a signature (hex in either case, or base64) is the HMAC of the body. An empty secret or
   * signature never verifies.
   *
   * @param body the raw body
   * @param signature the X-Cyberbiz-Hmac-Sha256 value
   * @param secret the App Secret
   * @return whether it matches
   */
  public static boolean verify(byte[] body, String signature, String secret) {
    if (secret.isEmpty() || signature.isBlank()) {
      return false;
    }
    return matches(mac(body, secret), signature.strip());
  }

  /**
   * Whether a signature is the HMAC of the Shop Domain, with the same encodings as {@link #verify}.
   *
   * @param shopDomain the Shop Domain
   * @param signature the X-Cyberbiz-Domain-Hmac-Sha256 value
   * @param secret the App Secret
   * @return whether it matches
   */
  public static boolean verifyDomain(String shopDomain, String signature, String secret) {
    return verify(shopDomain.getBytes(StandardCharsets.UTF_8), signature, secret);
  }

  /** Both encodings are always compared, so timing does not reveal which matched. */
  private static boolean matches(byte[] digest, String signature) {
    byte[] hex = HexFormat.of().formatHex(digest).getBytes(StandardCharsets.US_ASCII);
    byte[] base64 = Base64.getEncoder().encode(digest);
    boolean hexOk =
        MessageDigest.isEqual(
            hex, signature.toLowerCase(Locale.ROOT).getBytes(StandardCharsets.UTF_8));
    boolean base64Ok = MessageDigest.isEqual(base64, signature.getBytes(StandardCharsets.UTF_8));
    return hexOk || base64Ok;
  }

  private static byte[] mac(byte[] message, String secret) {
    try {
      Mac mac = Mac.getInstance(ALGORITHM);
      mac.init(new SecretKeySpec(secret.getBytes(StandardCharsets.UTF_8), ALGORITHM));
      return mac.doFinal(message);
    } catch (GeneralSecurityException e) {
      throw new IllegalStateException("HmacSHA256 is unavailable", e);
    }
  }
}
