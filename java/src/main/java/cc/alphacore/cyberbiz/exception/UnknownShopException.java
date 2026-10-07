package cc.alphacore.cyberbiz.exception;

/** No App Secret is known for the Shop Domain (respond 401). */
public final class UnknownShopException extends WebhookException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message names the Shop Domain
   */
  public UnknownShopException(String message) {
    super(message, 401);
  }
}
