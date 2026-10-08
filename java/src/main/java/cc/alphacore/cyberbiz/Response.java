package cc.alphacore.cyberbiz;

import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * A response from the API, body fully read. Immutable and thread-safe.
 *
 * <p>The body is kept as the raw bytes the server sent, so binary replies such as the
 * shipping-label zip archives survive intact: read them with {@link #bytes()}. {@link #body()}
 * returns the same bytes decoded as UTF-8 text, for JSON.
 *
 * <p>This is a final class rather than a record because a record cannot safely hold an array: the
 * byte array is copied on the way in and on the way out. The {@link #Response(int, Map, String)}
 * constructor of the former record still works, so custom {@link
 * cc.alphacore.cyberbiz.http.Transport}s written for it need no change; ones that read bytes should
 * use {@link #ofBytes} instead.
 */
public final class Response {
  private static final Pattern FILENAME =
      Pattern.compile("(?i)filename\\s*=\\s*(?:\"([^\"]*)\"|([^;\\s]+))");
  private static final Pattern FILENAME_EXTENDED =
      Pattern.compile("(?i)filename\\*\\s*=\\s*([^']*)'[^']*'([^;\\s]+)");

  private final int statusCode;
  private final Map<String, List<String>> headers;
  private final byte[] bytes;
  private final String body;

  /**
   * Creates a response from a text body, encoded as UTF-8.
   *
   * @param statusCode the HTTP status
   * @param headers the response headers; names keep the case the server sent
   * @param body the body as text; null means empty
   */
  public Response(int statusCode, Map<String, List<String>> headers, String body) {
    this.statusCode = statusCode;
    this.headers = headers == null ? Map.of() : Map.copyOf(headers);
    this.body = Objects.requireNonNullElse(body, "");
    this.bytes = this.body.getBytes(StandardCharsets.UTF_8);
  }

  private Response(int statusCode, Map<String, List<String>> headers, byte[] bytes) {
    this.statusCode = statusCode;
    this.headers = headers == null ? Map.of() : Map.copyOf(headers);
    this.bytes = bytes == null ? new byte[0] : bytes.clone();
    this.body = new String(this.bytes, StandardCharsets.UTF_8);
  }

  /**
   * Creates a response from the raw body bytes, as a transport reads them.
   *
   * @param statusCode the HTTP status
   * @param headers the response headers; names keep the case the server sent
   * @param bytes the body; copied, and null means empty
   * @return the response
   */
  public static Response ofBytes(int statusCode, Map<String, List<String>> headers, byte[] bytes) {
    return new Response(statusCode, headers, bytes);
  }

  /** Returns the HTTP status. */
  public int statusCode() {
    return statusCode;
  }

  /** Returns the response headers, unmodifiable; names keep the case the server sent. */
  public Map<String, List<String>> headers() {
    return headers;
  }

  /** Returns the body decoded as UTF-8 text; malformed bytes become U+FFFD. */
  public String body() {
    return body;
  }

  /** Returns a copy of the raw body bytes, e.g. a zip archive. */
  public byte[] bytes() {
    return bytes.clone();
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

  /** Returns the {@code Content-Type} header, e.g. "application/zip", or "". */
  public String contentType() {
    return header("Content-Type");
  }

  /**
   * Returns the file name the {@code Content-Disposition} header suggests, preferring the RFC 6266
   * {@code filename*} form when it is UTF-8.
   *
   * @return the file name, or empty when the header names none
   */
  public Optional<String> filename() {
    String disposition = header("Content-Disposition");
    Matcher extended = FILENAME_EXTENDED.matcher(disposition);
    if (extended.find() && "utf-8".equals(extended.group(1).toLowerCase(Locale.ROOT))) {
      // RFC 5987 percent-encoding: unlike a form, "+" is a literal plus.
      String encoded = extended.group(2).replace("+", "%2B");
      return Optional.of(URLDecoder.decode(encoded, StandardCharsets.UTF_8));
    }
    Matcher plain = FILENAME.matcher(disposition);
    if (plain.find()) {
      String name = plain.group(1) != null ? plain.group(1) : plain.group(2);
      return name.isEmpty() ? Optional.empty() : Optional.of(name);
    }
    return Optional.empty();
  }

  /** Whether the status is 2xx. */
  public boolean isSuccessful() {
    return statusCode >= 200 && statusCode < 300;
  }

  /** Whether the body is the JSON literal null (some lookups return it instead of 404). */
  public boolean isNull() {
    return "null".equals(body.strip());
  }

  /** Equal when the status, headers and body bytes are. */
  @Override
  public boolean equals(Object other) {
    return other instanceof Response that
        && statusCode == that.statusCode
        && headers.equals(that.headers)
        && Arrays.equals(bytes, that.bytes);
  }

  @Override
  public int hashCode() {
    return Objects.hash(statusCode, headers, Arrays.hashCode(bytes));
  }

  /** Describes the status, headers and body size; never the body, which may be binary. */
  @Override
  public String toString() {
    return "Response[statusCode="
        + statusCode
        + ", headers="
        + headers
        + ", body="
        + bytes.length
        + " bytes]";
  }
}
