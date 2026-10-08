package cc.alphacore.cyberbiz.webhook;

import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.function.Function;

/**
 * Maps a Shop Domain (X-Cyberbiz-Domain) to that Shop's App Secret. Use {@link StaticSecret} for a
 * single-shop integration, {@link SecretMap} for a fixed set of Shops, or a lambda that looks the
 * Shop up in your own store.
 *
 * <p>A Shop can install several Apps, each with its own secret. To serve them from one receiver,
 * override {@link #credentialsFor}, or use {@link AppSecrets} or {@link #ofCredentials}; {@link
 * Event#appId()} then names the App whose secret verified the body.
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

  /**
   * Returns every candidate App Secret for a Shop, at most {@link WebhookParser#MAX_CREDENTIALS}.
   * {@link WebhookParser} calls only this method. The default wraps {@link #secretFor} in one
   * Credential with an empty App ID.
   *
   * @param shopDomain the Shop Domain, e.g. "example.cyberbiz.co"
   * @return the candidates, empty when the Shop is unknown; empty secrets are ignored
   */
  default List<Credential> credentialsFor(String shopDomain) {
    Optional<String> secret = secretFor(shopDomain);
    if (secret == null || secret.isEmpty()) {
      return List.of();
    }
    return List.of(new Credential("", secret.get()));
  }

  /**
   * Adapts a lookup that returns several Apps' secrets for a Shop, e.g. from your own store.
   *
   * @param lookup returns the candidates for a Shop Domain, empty when the Shop is unknown
   * @return a resolver whose {@link #credentialsFor} calls {@code lookup}, and whose {@link
   *     #secretFor} answers only when exactly one candidate exists
   */
  static SecretResolver ofCredentials(Function<String, List<Credential>> lookup) {
    Objects.requireNonNull(lookup, "lookup");
    return new SecretResolver() {
      @Override
      public Optional<String> secretFor(String shopDomain) {
        List<Credential> credentials = credentialsFor(shopDomain);
        return credentials.size() == 1
            ? Optional.of(credentials.get(0).secret())
            : Optional.empty();
      }

      @Override
      public List<Credential> credentialsFor(String shopDomain) {
        return lookup.apply(shopDomain);
      }

      @Override
      public String toString() {
        return "SecretResolver.ofCredentials";
      }
    };
  }
}
