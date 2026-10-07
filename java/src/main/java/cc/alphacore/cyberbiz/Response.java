package cc.alphacore.cyberbiz;

import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * A response from the API, body fully read.
 *
 * @param statusCode the HTTP status
 * @param headers the response headers; names keep the case the server sent
 * @param body the body as UTF-8 text
 */
public record Response(int statusCode, Map<String, List<String>> headers, String body) {

  /** Freezes the headers. */
  public Response {
    headers = headers == null ? Map.of() : Map.copyOf(headers);
    body = Objects.requireNonNullElse(body, "");
  }

  /**
   * Returns the first value of a header, matched case-insensitively.
   *
   * @param name the header name
   * @return the value, or "" when absent
   */
  public String header(String name) {
    for (Map.Entry<String, List<String>> entry : headers.entrySet()) {
      if (entry.getKey().equalsIgnoreCase(name) && !entry.getValue().isEmpty()) {
        return entry.getValue().get(0);
      }
    }
    return "";
  }

  /** Returns the {@code X-Request-Id} header, useful when contacting CYBERBIZ, or "". */
  public String requestId() {
    return header("X-Request-Id");
  }

  /** Whether the status is 2xx. */
  public boolean isSuccessful() {
    return statusCode >= 200 && statusCode < 300;
  }

  /** Whether the body is the JSON literal null (some lookups return it instead of 404). */
  public boolean isNull() {
    return "null".equals(body.strip());
  }
}
