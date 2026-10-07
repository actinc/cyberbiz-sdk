package cc.alphacore.cyberbiz.exception;

/** X-Cyberbiz-Hmac-Sha256 does not match the body (respond 401). */
public final class InvalidSignatureException extends WebhookException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message names the Event and the Shop Domain
   */
  public InvalidSignatureException(String message) {
    super(message, 401);
  }
}
