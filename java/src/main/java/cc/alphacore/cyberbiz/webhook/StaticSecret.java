package cc.alphacore.cyberbiz.webhook;

import java.util.Objects;
import java.util.Optional;

/** One App Secret for every Shop (single-shop integrations). */
public final class StaticSecret implements SecretResolver {

  /** The App Secret; empty means every Shop is unknown. */
  private final String secret;

  /**
   * Creates the resolver.
   *
   * @param secret the App Secret
   */
  public StaticSecret(String secret) {
    this.secret = Objects.requireNonNull(secret, "secret");
  }

  @Override
  public Optional<String> secretFor(String shopDomain) {
    return secret.isEmpty() ? Optional.empty() : Optional.of(secret);
  }

  /** Never shows the secret. */
  @Override
  public String toString() {
    return "StaticSecret[***]";
  }
}
