package cc.alphacore.cyberbiz.exception;

import java.util.List;

/** A 401, the token is missing, invalid or expired. */
public final class AuthenticationException extends ApiException {
  private static final long serialVersionUID = 1L;

  /**
   * Creates the exception; see {@link ApiException#ApiException}.
   *
   * @param statusCode the HTTP status
   * @param method the request method, upper case
   * @param path the request path
   * @param requestId the {@code X-Request-Id} header, or ""
   * @param messages messages from the body
   * @param body the raw response body
   */
  public AuthenticationException(
      int statusCode,
      String method,
      String path,
      String requestId,
      List<String> messages,
      String body) {
    super(statusCode, method, path, requestId, messages, body);
  }
}
