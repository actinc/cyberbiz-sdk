package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.pagination.Pagination;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** The page and per_page query parameters of a walk over every page, as the PHP SDK sets them. */
final class PageRequests {
  private PageRequests() {}

  /** The request's "page" parameter, at least 1; 1 when absent or not a number. */
  static int startPage(Request request) {
    List<String> values = request.query().getOrDefault("page", List.of());
    String value = values.isEmpty() ? "" : values.get(0).strip();
    return value.matches("-?\\d{1,9}") ? Math.max(1, Integer.parseInt(value)) : 1;
  }

  /** A copy with "page" first and set to {@code page}, and per_page 50 unless the caller set it. */
  static Request withPage(Request request, int page) {
    Map<String, List<String>> query = new LinkedHashMap<>();
    query.put("page", List.of(String.valueOf(page)));
    request
        .query()
        .forEach(
            (name, values) -> {
              if (!"page".equals(name)) {
                query.put(name, values);
              }
            });
    query.putIfAbsent("per_page", List.of(String.valueOf(Pagination.MAX_PER_PAGE)));
    return new Request(request.method(), request.path(), query, request.body(), request.headers());
  }
}
