package cc.alphacore.cyberbiz;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/**
 * One API call: method, path relative to the base URL, query, JSON body and extra headers.
 * Immutable; the {@code with*} methods return a copy.
 *
 * @param method the HTTP method, upper case
 * @param path the path relative to the base URL, e.g. {@code /v1/products}
 * @param query query parameters in order; a key with several values is repeated
 * @param body the JSON body, or null for none
 * @param headers extra headers, sent after the SDK's own
 */
public record Request(
    String method,
    String path,
    Map<String, List<String>> query,
    String body,
    Map<String, String> headers) {

  private static final Set<String> IDEMPOTENT = Set.of("GET", "HEAD", "PUT", "DELETE", "OPTIONS");

  /** Validates the request and freezes its maps. */
  public Request {
    Objects.requireNonNull(method, "method");
    Objects.requireNonNull(path, "path");
    method = method.toUpperCase(Locale.ROOT);
    query = freeze(query);
    headers = headers == null ? Map.of() : Map.copyOf(headers);
  }

  /**
   * Starts a request without query, body or extra headers.
   *
   * @param method the HTTP method
   * @param path the path relative to the base URL
   * @return the request
   */
  public static Request of(String method, String path) {
    return new Request(method, path, Map.of(), null, Map.of());
  }

  /**
   * Adds a query parameter. Null values are dropped; booleans become "true"/"false".
   *
   * @param name the parameter name
   * @param values one or more values; several repeat the key ({@code ids=1&ids=2})
   * @return a copy with the parameter added
   */
  public Request withQuery(String name, Object... values) {
    Objects.requireNonNull(name, "name");
    Map<String, List<String>> copy = new LinkedHashMap<>(query);
    List<String> list = new ArrayList<>(copy.getOrDefault(name, List.of()));
    for (Object value : values) {
      if (value != null) {
        list.add(String.valueOf(value));
      }
    }
    copy.put(name, list);
    return new Request(method, path, copy, body, headers);
  }

  /**
   * Sets the JSON body.
   *
   * @param json the encoded JSON
   * @return a copy with the body
   */
  public Request withBody(String json) {
    return new Request(method, path, query, json, headers);
  }

  /**
   * Adds a header, replacing one of the same name.
   *
   * @param name the header name
   * @param value the header value
   * @return a copy with the header
   */
  public Request withHeader(String name, String value) {
    Map<String, String> copy = new LinkedHashMap<>(headers);
    copy.put(Objects.requireNonNull(name, "name"), Objects.requireNonNull(value, "value"));
    return new Request(method, path, query, body, copy);
  }

  /** Whether repeating the request after a network failure is safe. */
  public boolean isIdempotent() {
    return IDEMPOTENT.contains(method);
  }

  private static Map<String, List<String>> freeze(Map<String, List<String>> query) {
    if (query == null || query.isEmpty()) {
      return Map.of();
    }
    Map<String, List<String>> copy = new LinkedHashMap<>();
    query.forEach((key, values) -> copy.put(key, List.copyOf(values)));
    return Collections.unmodifiableMap(copy);
  }
}
