package cc.alphacore.cyberbiz.exception;

/** The request failed before a response arrived (DNS, TLS, timeout, interruption, ...). */
public final class TransportException extends CyberbizException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception.
   *
   * @param message what failed
   * @param cause the I/O failure, or null
   */
  public TransportException(String message, Throwable cause) {
    super(message, cause);
  }
}
