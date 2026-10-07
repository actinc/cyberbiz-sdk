package cc.alphacore.cyberbiz.exception;

/**
 * X-Cyberbiz-Domain-Hmac-Sha256 does not match the Shop Domain (respond 401). The body signature
 * had already verified.
 */
public final class InvalidDomainSignatureException extends WebhookException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message names the Shop Domain
   */
  public InvalidDomainSignatureException(String message) {
    super(message, 401);
  }
}
