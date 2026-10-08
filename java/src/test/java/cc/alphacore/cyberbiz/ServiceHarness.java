package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.http.TransportRequest;
import java.net.URI;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** A client wired to a {@link FakeTransport}, plus helpers to read what it sent. */
final class ServiceHarness {
  final FakeTransport transport = new FakeTransport();
  final CyberbizClient client =
      CyberbizClient.builder("synthetic-token-0123456789")
          .baseUrl(URI.create("https://api.example.test"))
          .transport(transport)
          .clock(new FakeClock())
          .rateLimit(0)
          .maxRetries(0)
          .build();

  /** Replies with the given Golden Files and their recorded headers, in order. */
  static ServiceHarness golden(String... names) {
    ServiceHarness h = new ServiceHarness();
    for (String name : names) {
      Response recorded = Golden.response(name);
      List<String> pairs = new ArrayList<>();
      recorded
          .headers()
          .forEach(
              (key, values) -> {
                pairs.add(key);
                pairs.add(String.join(", ", values));
              });
      h.transport.reply(200, recorded.body(), pairs.toArray(String[]::new));
    }
    return h;
  }

  /** Replies with the given bodies, status 200, in order. */
  static ServiceHarness json(String... bodies) {
    ServiceHarness h = new ServiceHarness();
    for (String body : bodies) {
      h.transport.reply(200, body);
    }
    return h;
  }

  /** "METHOD path?query" of the n-th request, relative to the base URL. */
  String line(int n) {
    TransportRequest request = transport.requests.get(n);
    URI uri = request.uri();
    String query = uri.getRawQuery();
    return request.method()
        + " "
        + uri.getRawPath().replaceFirst("^/", "")
        + (query == null ? "" : "?" + query);
  }

  /** The first request's line. */
  String line() {
    return line(0);
  }

  /** The JSON body of the n-th request, or null. */
  String body(int n) {
    return transport.requests.get(n).body();
  }

  /** Builds an ordered map from alternating keys and values; values may be null. */
  static Map<String, Object> map(Object... pairs) {
    Map<String, Object> out = new LinkedHashMap<>();
    for (int i = 0; i + 1 < pairs.length; i += 2) {
      out.put((String) pairs[i], pairs[i + 1]);
    }
    return out;
  }
}
