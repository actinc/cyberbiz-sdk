package cc.alphacore.cyberbiz.exception;

/** The body exceeds the size limit (respond 413). */
public final class BodyTooLargeException extends WebhookException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message names the limit
   */
  public BodyTooLargeException(String message) {
    super(message, 413);
  }
}
