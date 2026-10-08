package cc.alphacore.cyberbiz.webhook;

import cc.alphacore.cyberbiz.exception.CyberbizException;
import com.google.gson.Gson;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParseException;
import com.google.gson.JsonParser;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;

/**
 * One authenticated webhook, as returned by {@link WebhookParser#parse}.
 *
 * @param type the X-Cyberbiz-Event value, e.g. "orders/paid"
 * @param shopDomain the Shop Domain (X-Cyberbiz-Domain) that identifies the Shop
 * @param customDomain the merchant's storefront hostname (X-Cyberbiz-Shop-Domain), informational
 *     only; "" when absent
 * @param signature the verified X-Cyberbiz-Hmac-Sha256 value
 * @param domainSignature the X-Cyberbiz-Domain-Hmac-Sha256 value, "" when absent
 * @param headers the request headers, names lower-cased, repeated values joined with ", "
 * @param body the raw body decoded as UTF-8; the signature was verified over the original bytes
 * @param appId the App whose secret verified the body, as named by {@link
 *     SecretResolver#credentialsFor}; "" for a resolver that only implements {@link
 *     SecretResolver#secretFor}
 */
public record Event(
    String type,
    String shopDomain,
    String customDomain,
    String signature,
    String domainSignature,
    Map<String, String> headers,
    String body,
    String appId) {

  private static final Gson GSON = new Gson();

  /** Rejects nulls and freezes the headers. */
  public Event {
    Objects.requireNonNull(type, "type");
    Objects.requireNonNull(shopDomain, "shopDomain");
    Objects.requireNonNull(customDomain, "customDomain");
    Objects.requireNonNull(signature, "signature");
    Objects.requireNonNull(domainSignature, "domainSignature");
    headers = Map.copyOf(headers);
    Objects.requireNonNull(body, "body");
    Objects.requireNonNull(appId, "appId");
  }

  /**
   * Creates an Event without an App ID, as before App IDs existed.
   *
   * @param type the X-Cyberbiz-Event value
   * @param shopDomain the Shop Domain
   * @param customDomain the merchant's storefront hostname, "" when absent
   * @param signature the verified body Signature
   * @param domainSignature the Domain Signature, "" when absent
   * @param headers the request headers, names lower-cased
   * @param body the raw body decoded as UTF-8
   */
  public Event(
      String type,
      String shopDomain,
      String customDomain,
      String signature,
      String domainSignature,
      Map<String, String> headers,
      String body) {
    this(type, shopDomain, customDomain, signature, domainSignature, headers, body, "");
  }

  /** Returns the documented Event, or empty for one this SDK does not know yet. */
  public Optional<EventType> eventType() {
    return EventType.of(type);
  }

  /**
   * Returns a header value, matched case-insensitively.
   *
   * @param name the header name
   * @return the value, or "" when absent
   */
  public String header(String name) {
    return headers.getOrDefault(WebhookParser.lower(name), "");
  }

  /**
   * Parses the body as a JSON object (every documented payload is one).
   *
   * @return the payload
   * @throws CyberbizException when the body is not a JSON object
   */
  public JsonObject payload() {
    try {
      JsonElement parsed = JsonParser.parseString(body);
      if (parsed.isJsonObject()) {
        return parsed.getAsJsonObject();
      }
    } catch (JsonParseException e) {
      throw new CyberbizException("webhook: body is not valid JSON", e);
    }
    throw new CyberbizException("webhook: body is not a JSON object");
  }

  /**
   * Decodes the body into your own class or record with Gson.
   *
   * @param <T> the target type
   * @param type the target class
   * @return the decoded payload
   * @throws CyberbizException when the body does not fit the type
   */
  public <T> T decode(Class<T> type) {
    T value;
    try {
      value = GSON.fromJson(body, type);
    } catch (JsonParseException e) {
      throw new CyberbizException("webhook: cannot decode body as " + type.getSimpleName(), e);
    }
    if (value == null) {
      throw new CyberbizException("webhook: body is empty");
    }
    return value;
  }
}
