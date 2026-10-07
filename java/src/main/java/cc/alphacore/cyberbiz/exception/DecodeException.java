package cc.alphacore.cyberbiz.exception;

/**
 * A response body (or a value in it) could not be decoded: malformed JSON, a field of the wrong
 * type, an amount that is not a decimal number, or a timestamp in no accepted format.
 */
public final class DecodeException extends CyberbizException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message what could not be decoded, and where
   */
  public DecodeException(String message) {
    super(message);
  }

  /**
   * Creates the exception with its cause.
   *
   * @param message what could not be decoded, and where
   * @param cause the parser failure
   */
  public DecodeException(String message, Throwable cause) {
    super(message, cause);
  }
}
