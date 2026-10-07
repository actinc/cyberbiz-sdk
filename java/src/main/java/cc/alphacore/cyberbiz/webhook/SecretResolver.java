package cc.alphacore.cyberbiz.webhook;

import java.util.Optional;

/**
 * Maps a Shop Domain (X-Cyberbiz-Domain) to that Shop's App Secret. Use {@link StaticSecret} for a
 * single-shop integration, {@link SecretMap} for a fixed set of Shops, or a lambda that looks the
 * Shop up in your own store.
 */
@FunctionalInterface
public interface SecretResolver {

  /**
   * Returns the App Secret for a Shop.
   *
   * @param shopDomain the Shop Domain, e.g. "example.cyberbiz.co"
   * @return the App Secret, or empty when the Shop is unknown; an empty string counts as unknown
   */
  Optional<String> secretFor(String shopDomain);
}
