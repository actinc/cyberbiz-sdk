package cc.alphacore.cyberbiz.exception;

/** Every exception this SDK throws extends this class. All are unchecked. */
public class CyberbizException extends RuntimeException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates an exception with a message.
   *
   * @param message what went wrong, never including the token
   */
  public CyberbizException(String message) {
    super(message);
  }

  /**
   * Creates an exception with a message and its cause.
   *
   * @param message what went wrong, never including the token
   * @param cause the underlying failure
   */
  public CyberbizException(String message, Throwable cause) {
    super(message, cause);
  }
}
