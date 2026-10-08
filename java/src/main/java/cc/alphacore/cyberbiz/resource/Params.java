package cc.alphacore.cyberbiz.resource;

import cc.alphacore.cyberbiz.Request;
import cc.alphacore.cyberbiz.json.Json;
import cc.alphacore.cyberbiz.json.Times;
import com.google.gson.Gson;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.Collection;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.StringJoiner;

/**
 * Shapes caller options the way the order and customer endpoints expect, as the PHP SDK's Options:
 * multi-value filters and id lists comma-separated, times in the platform format, bodies as JSON.
 */
final class Params {

  /** The body encoder: the SDK's JSON configuration, but a null value in a map is sent as null. */
  private static final Gson BODY = Json.gson().newBuilder().serializeNulls().create();

  /** The longest caller-supplied path segment accepted. */
  static final int MAX_SEGMENT = 256;

  private Params() {}

  /**
   * Adds query parameters in the map's order. A collection becomes "a,b,c", an {@link
   * OffsetDateTime} a CYBERBIZ timestamp in Asia/Taipei (or a date when {@code dates}), a {@link
   * LocalDate} "YYYY-MM-DD"; null values are left out.
   */
  static Request query(Request request, Map<String, ?> query, boolean dates) {
    Request out = request;
    if (query == null) {
      return out;
    }
    for (Map.Entry<String, ?> entry : query.entrySet()) {
      if (entry.getValue() != null) {
        out = out.withQuery(entry.getKey(), value(entry.getValue(), dates));
      }
    }
    return out;
  }

  /** Adds query parameters; times are sent as timestamps. */
  static Request query(Request request, Map<String, ?> query) {
    return query(request, query, false);
  }

  /** Encodes a map as the JSON body; amounts exactly, times in the platform format. */
  static Request body(Request request, Map<String, ?> body) {
    return request.withBody(BODY.toJson(body == null ? Map.of() : body));
  }

  /** Joins values with commas, as the API takes id lists and multi-value filters. */
  static String join(Collection<?> values) {
    StringJoiner joined = new StringJoiner(",");
    for (Object value : values) {
      joined.add(value(value, false));
    }
    return joined.toString();
  }

  /**
   * Replaces a collection under "line_item_ids" with the comma-separated string the v1 fulfillment
   * endpoints take; a string is kept as given.
   *
   * @throws IllegalArgumentException when the collection holds something other than numbers or
   *     strings
   */
  static Map<String, Object> lineItemIds(Map<String, ?> body) {
    Map<String, Object> copy = new LinkedHashMap<>(body == null ? Map.of() : body);
    if (copy.get("line_item_ids") instanceof Collection<?> ids) {
      for (Object id : ids) {
        if (!(id instanceof Number || id instanceof String)) {
          String type = id == null ? "null" : id.getClass().getSimpleName();
          throw new IllegalArgumentException(
              "cyberbiz: line_item_ids must hold numbers or strings, got " + type);
        }
      }
      copy.put("line_item_ids", join(ids));
    }
    return copy;
  }

  /** Applies {@link #lineItemIds} to every map in a list. */
  static List<Object> lineItemIdsOfEach(Object entries) {
    if (!(entries instanceof Collection<?> list)) {
      throw new IllegalArgumentException("cyberbiz: support_shippings must be a list");
    }
    return list.stream().map(Params::lineItemIdsOfMap).toList();
  }

  /**
   * Percent-encodes one caller-supplied path segment (PHP's rawurlencode), so "/" becomes %2F and
   * never a separator.
   *
   * @throws IllegalArgumentException for an empty segment, "." or "..", which encoding leaves alone
   *     and which would address another API path, or one over {@value #MAX_SEGMENT} characters
   */
  static String segment(String text) {
    if (text == null || text.isEmpty() || ".".equals(text) || "..".equals(text)) {
      throw new IllegalArgumentException(
          "cyberbiz: a path segment must not be empty, \".\" or \"..\"");
    }
    if (text.length() > MAX_SEGMENT) {
      throw new IllegalArgumentException(
          "cyberbiz: a path segment must not exceed " + MAX_SEGMENT + " characters");
    }
    return URLEncoder.encode(text, StandardCharsets.UTF_8)
        .replace("+", "%20")
        .replace("*", "%2A")
        .replace("%7E", "~");
  }

  private static Object lineItemIdsOfMap(Object entry) {
    if (!(entry instanceof Map<?, ?> map)) {
      return entry;
    }
    Map<String, Object> named = new LinkedHashMap<>();
    map.forEach((key, value) -> named.put(String.valueOf(key), value));
    return lineItemIds(named);
  }

  private static String value(Object value, boolean dates) {
    if (value instanceof Collection<?> values) {
      return join(values);
    }
    if (value instanceof OffsetDateTime time) {
      return dates
          ? time.atZoneSameInstant(Times.ZONE).toLocalDate().toString()
          : Times.format(time);
    }
    return String.valueOf(value);
  }
}
