package cc.alphacore.cyberbiz.webhook;

import java.util.Objects;

/**
 * One App's webhook secret for a Shop. CYBERBIZ sends no App identifier header, so the App behind a
 * webhook is identified only by the Credential whose secret verifies the body Signature.
 *
 * @param appId your own identifier for the App, reported as {@link Event#appId()}
 * @param secret the App Secret; an empty secret never verifies
 */
public record Credential(String appId, String secret) {

  /** Rejects nulls. */
  public Credential {
    Objects.requireNonNull(appId, "appId");
    Objects.requireNonNull(secret, "secret");
  }

  /** Never shows the secret. */
  @Override
  public String toString() {
    return "Credential[appId=" + appId + ", secret=***]";
  }
}
