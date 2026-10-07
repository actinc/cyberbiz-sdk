package cc.alphacore.cyberbiz.exception;

import java.util.List;

/**
 * An error response from the CYBERBIZ API: any non-2xx status, or a 2xx whose body is nothing but
 * an error object. Subclasses narrow it by status.
 */
public class ApiException extends CyberbizException {
  private static final long serialVersionUID = 1L;

  private final int statusCode;
  private final String method;
  private final String path;
  private final String requestId;
  private final List<String> messages;
  private final String body;

  /**
   * Creates the exception.
   *
   * @param statusCode the HTTP status
   * @param method the request method, upper case
   * @param path the request path, starting with "/"
   * @param requestId the {@code X-Request-Id} response header, or ""
   * @param messages human-readable messages from the body, usually Traditional Chinese
   * @param body the raw response body
   */
  public ApiException(
      int statusCode,
      String method,
      String path,
      String requestId,
      List<String> messages,
      String body) {
    super(describe(statusCode, method, path, messages));
    this.statusCode = statusCode;
    this.method = method;
    this.path = path;
    this.requestId = requestId;
    this.messages = List.copyOf(messages);
    this.body = body;
  }

  private static String describe(int status, String method, String path, List<String> messages) {
    String text = messages.isEmpty() ? "HTTP " + status : String.join("; ", messages);
    return "cyberbiz: " + method + " " + path + ": " + status + " " + text;
  }

  /** Returns the HTTP status. */
  public int statusCode() {
    return statusCode;
  }

  /** Returns the request method, upper case. */
  public String method() {
    return method;
  }

  /** Returns the request path, starting with "/". */
  public String path() {
    return path;
  }

  /** Returns the {@code X-Request-Id} header, useful when contacting CYBERBIZ, or "". */
  public String requestId() {
    return requestId;
  }

  /** Returns the messages found in the body; empty when it had none. */
  public List<String> messages() {
    return messages;
  }

  /** Returns the raw response body. */
  public String body() {
    return body;
  }
}
