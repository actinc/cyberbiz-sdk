package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.http.Transport;
import cc.alphacore.cyberbiz.http.TransportRequest;
import java.io.IOException;
import java.util.Objects;

/**
 * Passes GET requests to another transport and refuses every other method before anything is sent,
 * so the live tests ({@link LiveTest}) cannot write to a real shop even by mistake.
 */
final class ReadOnlyTransport implements Transport {
  private final Transport next;

  ReadOnlyTransport(Transport next) {
    this.next = Objects.requireNonNull(next, "next");
  }

  @Override
  public Response send(TransportRequest request) throws IOException {
    if (!"GET".equals(request.method())) {
      throw new IllegalStateException(
          "read-only transport: refusing " + request.method() + " " + request.uri().getPath());
    }
    return next.send(request);
  }
}
