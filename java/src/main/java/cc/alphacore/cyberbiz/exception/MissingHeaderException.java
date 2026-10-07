package cc.alphacore.cyberbiz.exception;

/** A required X-Cyberbiz-* header is absent (respond 400). */
public final class MissingHeaderException extends WebhookException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message names the missing header
   */
  public MissingHeaderException(String message) {
    super(message, 400);
  }
}
