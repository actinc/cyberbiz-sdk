package cc.alphacore.cyberbiz.webhook;

import cc.alphacore.cyberbiz.exception.BodyTooLargeException;
import cc.alphacore.cyberbiz.exception.InvalidDomainSignatureException;
import cc.alphacore.cyberbiz.exception.InvalidSignatureException;
import cc.alphacore.cyberbiz.exception.MissingHeaderException;
import cc.alphacore.cyberbiz.exception.UnknownShopException;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;
import java.util.StringJoiner;

/**
 * Authenticates CYBERBIZ App webhooks, independent of any web framework: pass the request headers
 * and the raw body bytes. Every request carries X-Cyberbiz-Event, X-Cyberbiz-Domain (the Shop
 * Domain) and X-Cyberbiz-Hmac-Sha256 (the body Signature); X-Cyberbiz-Domain-Hmac-Sha256 is checked
 * when present. Immutable and safe to share between threads.
 */
public final class WebhookParser {

  /** The Event header, e.g. "orders/paid". */
  public static final String HEADER_EVENT = "X-Cyberbiz-Event";

  /** The Shop Domain header, e.g. "example.cyberbiz.co". */
  public static final String HEADER_DOMAIN = "X-Cyberbiz-Domain";

  /** The merchant's storefront hostname; informational only. */
  public static final String HEADER_SHOP_DOMAIN = "X-Cyberbiz-Shop-Domain";

  /** The body Signature header. */
  public static final String HEADER_SIGNATURE = "X-Cyberbiz-Hmac-Sha256";

  /** The Domain Signature header. */
  public static final String HEADER_DOMAIN_SIGNATURE = "X-Cyberbiz-Domain-Hmac-Sha256";

  /** The largest body accepted by default, as in the Go and PHP SDKs (2 MiB). */
  public static final int MAX_BODY_BYTES = 2 << 20;

  private final SecretResolver secrets;
  private final int maxBodyBytes;
  private final boolean checkDomainSignature;

  /**
   * Creates a parser with the default body limit that checks the Domain Signature when present.
   *
   * @param secrets resolves the App Secret for each Shop Domain
   */
  public WebhookParser(SecretResolver secrets) {
    this(secrets, MAX_BODY_BYTES, true);
  }

  /**
   * Creates a parser.
   *
   * @param secrets resolves the App Secret for each Shop Domain
   * @param maxBodyBytes the largest body accepted, at least 0
   * @param checkDomainSignature whether to verify X-Cyberbiz-Domain-Hmac-Sha256 when present; the
   *     body Signature is always verified
   */
  public WebhookParser(SecretResolver secrets, int maxBodyBytes, boolean checkDomainSignature) {
    if (maxBodyBytes < 0) {
      throw new IllegalArgumentException("maxBodyBytes must be at least 0");
    }
    this.secrets = Objects.requireNonNull(secrets, "secrets");
    this.maxBodyBytes = maxBodyBytes;
    this.checkDomainSignature = checkDomainSignature;
  }

  /**
   * Authenticates one request. Header names match ignoring case; a value may be a {@code String} or
   * a collection of strings (as in {@code Map<String, List<String>>}), joined with ", ".
   *
   * @param headers the request headers
   * @param body the raw body, byte for byte as received
   * @return the authenticated Event
   * @throws MissingHeaderException when X-Cyberbiz-Event, X-Cyberbiz-Domain or
   *     X-Cyberbiz-Hmac-Sha256 is absent or empty
   * @throws BodyTooLargeException when the body exceeds the limit
   * @throws UnknownShopException when the resolver has no App Secret for the Shop Domain
   * @throws InvalidSignatureException when the Signature does not match the body
   * @throws InvalidDomainSignatureException when the Domain Signature is present and wrong
   */
  public Event parse(Map<String, ?> headers, byte[] body) {
    Map<String, String> normalised = normalise(headers);
    String type = required(normalised, HEADER_EVENT);
    String shopDomain = required(normalised, HEADER_DOMAIN);
    String signature = required(normalised, HEADER_SIGNATURE);
    if (body.length > maxBodyBytes) {
      throw new BodyTooLargeException("webhook: body exceeds " + maxBodyBytes + " bytes");
    }
    Optional<String> resolved = secrets.secretFor(shopDomain);
    String secret = resolved == null ? "" : resolved.orElse("");
    if (secret.isEmpty()) {
      throw new UnknownShopException("webhook: unknown shop \"" + shopDomain + "\"");
    }
    if (!Signature.verify(body, signature, secret)) {
      throw new InvalidSignatureException(
          "webhook: invalid signature for " + type + " from \"" + shopDomain + "\"");
    }
    String domainSignature = normalised.getOrDefault(lower(HEADER_DOMAIN_SIGNATURE), "");
    if (checkDomainSignature
        && !domainSignature.isEmpty()
        && !Signature.verifyDomain(shopDomain, domainSignature, secret)) {
      throw new InvalidDomainSignatureException(
          "webhook: invalid domain signature for \"" + shopDomain + "\"");
    }
    String customDomain = normalised.getOrDefault(lower(HEADER_SHOP_DOMAIN), "");
    return new Event(
        type,
        shopDomain,
        customDomain,
        signature,
        domainSignature,
        normalised,
        new String(body, StandardCharsets.UTF_8));
  }

  /** Lower-cases names, joins repeated values and trims; drops values that are not strings. */
  private static Map<String, String> normalise(Map<String, ?> headers) {
    Map<String, String> out = new HashMap<>();
    headers.forEach(
        (name, value) -> {
          String text = text(value);
          if (name != null && text != null) {
            out.put(lower(name), text.strip());
          }
        });
    return out;
  }

  private static String text(Object value) {
    if (value instanceof String s) {
      return s;
    }
    if (value instanceof Iterable<?> values) {
      StringJoiner joined = new StringJoiner(", ");
      for (Object v : values) {
        if (v instanceof String s) {
          joined.add(s);
        }
      }
      return joined.toString();
    }
    return null;
  }

  private static String required(Map<String, String> headers, String name) {
    String value = headers.getOrDefault(lower(name), "");
    if (value.isEmpty()) {
      throw new MissingHeaderException("webhook: missing header " + name);
    }
    return value;
  }

  static String lower(String name) {
    return name.toLowerCase(Locale.ROOT);
  }

  /** Never shows secrets. */
  @Override
  public String toString() {
    return "WebhookParser[maxBodyBytes="
        + maxBodyBytes
        + ", checkDomainSignature="
        + checkDomainSignature
        + "]";
  }
}
