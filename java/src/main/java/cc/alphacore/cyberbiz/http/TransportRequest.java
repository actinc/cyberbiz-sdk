package cc.alphacore.cyberbiz.http;

import java.net.URI;
import java.util.Map;
import java.util.Objects;

/**
 * A fully built HTTP request, handed to a {@link Transport}.
 *
 * @param method the HTTP method, upper case
 * @param uri the absolute URI, query included
 * @param headers the headers to send, Authorization included
 * @param body the JSON body, or null for none
 */
public record TransportRequest(String method, URI uri, Map<String, String> headers, String body) {

  /** Validates the request and freezes the headers. */
  public TransportRequest {
    Objects.requireNonNull(method, "method");
    Objects.requireNonNull(uri, "uri");
    headers = Map.copyOf(headers);
  }

  /** Describes the request without its headers, so the token never reaches a log. */
  @Override
  public String toString() {
    return "TransportRequest{" + method + " " + uri + "}";
  }
}
