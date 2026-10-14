package cc.alphacore.cyberbiz.model;

/**
 * The credential pair of a registered customer-login OAuth application (POST /v2/customer_oauth).
 *
 * @param clientId the client id
 * @param clientSecret the client secret; keep it out of logs
 */
public record CustomerOAuthApp(String clientId, String clientSecret) {

  /** Leaves the secret out, so logging the record cannot leak it. */
  @Override
  public String toString() {
    return "CustomerOAuthApp[clientId=" + clientId + ", clientSecret=***]";
  }
}
