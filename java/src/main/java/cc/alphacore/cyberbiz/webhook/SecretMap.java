package cc.alphacore.cyberbiz.webhook;

import java.util.HashMap;
import java.util.Locale;
import java.util.Map;
import java.util.Optional;

/** A fixed Shop Domain to App Secret map (multi-shop integrations). Domains match ignoring case. */
public final class SecretMap implements SecretResolver {

  /** App Secrets keyed by lower-cased Shop Domain. */
  private final Map<String, String> secrets;

  /**
   * Creates the resolver from a copy of the map.
   *
   * @param secrets App Secrets keyed by Shop Domain, e.g. "example.cyberbiz.co"
   */
  public SecretMap(Map<String, String> secrets) {
    Map<String, String> lower = new HashMap<>();
    secrets.forEach((domain, secret) -> lower.put(domain.toLowerCase(Locale.ROOT), secret));
    this.secrets = Map.copyOf(lower);
  }

  @Override
  public Optional<String> secretFor(String shopDomain) {
    String secret = secrets.get(shopDomain.toLowerCase(Locale.ROOT));
    return secret == null || secret.isEmpty() ? Optional.empty() : Optional.of(secret);
  }

  /** Lists the Shop Domains, never the secrets. */
  @Override
  public String toString() {
    return "SecretMap" + secrets.keySet();
  }
}
