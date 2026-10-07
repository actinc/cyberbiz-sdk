package cc.alphacore.cyberbiz.exception;

/**
 * Every webhook rejection extends this; answer the request with {@link #httpStatus()} and do not
 * process the body.
 */
public abstract class WebhookException extends CyberbizException {
  private static final long serialVersionUID = 1L;

  /** The status to answer the request with. */
  private final int httpStatus;

  /**
   * Creates the exception.
   *
   * @param message what was wrong with the request, never including the App Secret
   * @param httpStatus the status to answer the request with
   */
  protected WebhookException(String message, int httpStatus) {
    super(message);
    this.httpStatus = httpStatus;
  }

  /** Returns the HTTP status to answer the request with: 400, 401 or 413. */
  public int httpStatus() {
    return httpStatus;
  }
}
