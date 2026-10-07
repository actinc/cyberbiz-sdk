package cc.alphacore.cyberbiz;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.StringJoiner;

/**
 * Encodes query parameters the way CYBERBIZ expects and the PHP SDK sends them: percent-encoding
 * per RFC 3986 (a space is %20, not +), a key with several values repeated.
 */
final class QueryEncoder {
  private QueryEncoder() {}

  static String encode(Map<String, List<String>> query) {
    StringJoiner pairs = new StringJoiner("&");
    query.forEach(
        (key, values) -> {
          for (String value : values) {
            pairs.add(component(key) + "=" + component(value));
          }
        });
    return pairs.toString();
  }

  /** Percent-encodes everything except A-Z a-z 0-9 - _ . ~ (PHP's rawurlencode). */
  static String component(String text) {
    return URLEncoder.encode(text, StandardCharsets.UTF_8)
        .replace("+", "%20")
        .replace("*", "%2A")
        .replace("%7E", "~");
  }
}
