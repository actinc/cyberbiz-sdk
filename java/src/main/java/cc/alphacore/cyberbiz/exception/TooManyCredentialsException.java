package cc.alphacore.cyberbiz.exception;

/**
 * The SecretResolver returned more candidate App Secrets for one Shop than the parser accepts. A
 * configuration error on the receiver (respond 500).
 */
public final class TooManyCredentialsException extends WebhookException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message names the Shop Domain, the count and the limit
   */
  public TooManyCredentialsException(String message) {
    super(message, 500);
  }
}
