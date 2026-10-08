package cc.alphacore.cyberbiz.exception;

/**
 * The body Signature verified under more than one App's secret: two Apps of the Shop are configured
 * with the same secret, so the App cannot be identified. A configuration error on the receiver
 * (respond 500, so CYBERBIZ retries once it is fixed).
 */
public final class AmbiguousSecretException extends WebhookException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message names the Shop Domain and how many candidates matched
   */
  public AmbiguousSecretException(String message) {
    super(message, 500);
  }
}
