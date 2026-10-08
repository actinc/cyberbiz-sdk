package cc.alphacore.cyberbiz.http;

import cc.alphacore.cyberbiz.Response;
import java.io.IOException;
import java.io.InterruptedIOException;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Objects;

/** A {@link Transport} on the JDK's {@link HttpClient}. Thread-safe. */
public final class JdkTransport implements Transport {

  /** The connect timeout used by {@link #JdkTransport()}. */
  public static final Duration DEFAULT_CONNECT_TIMEOUT = Duration.ofSeconds(10);

  /** The per-request timeout used by {@link #JdkTransport()}. */
  public static final Duration DEFAULT_REQUEST_TIMEOUT = Duration.ofSeconds(30);

  private final HttpClient http;
  private final Duration requestTimeout;

  /** Creates a transport with the default timeouts, speaking HTTP/1.1. */
  public JdkTransport() {
    this(defaultClient(), DEFAULT_REQUEST_TIMEOUT);
  }

  /**
   * The client {@link #JdkTransport()} uses. It is pinned to HTTP/1.1: over HTTP/2 the JDK client
   * fails every request to the CYBERBIZ API host with "EOF reached while reading" (found by the
   * live tests, CBSDK-42), while HTTP/1.1, curl and the Go and PHP SDKs work.
   */
  static HttpClient defaultClient() {
    return HttpClient.newBuilder()
        .version(HttpClient.Version.HTTP_1_1)
        .connectTimeout(DEFAULT_CONNECT_TIMEOUT)
        .build();
  }

  /**
   * Creates a transport on a configured client, e.g. one with a proxy.
   *
   * @param http the client to send with
   * @param requestTimeout how long one request may take
   */
  public JdkTransport(HttpClient http, Duration requestTimeout) {
    this.http = Objects.requireNonNull(http, "http");
    this.requestTimeout = Objects.requireNonNull(requestTimeout, "requestTimeout");
  }

  @Override
  public Response send(TransportRequest request) throws IOException {
    HttpRequest.BodyPublisher body =
        request.body() == null
            ? HttpRequest.BodyPublishers.noBody()
            : HttpRequest.BodyPublishers.ofString(request.body(), StandardCharsets.UTF_8);
    HttpRequest.Builder builder =
        HttpRequest.newBuilder(request.uri())
            .timeout(requestTimeout)
            .method(request.method(), body);
    request.headers().forEach(builder::header);
    try {
      // Bytes, not a String: binary replies such as label zips must arrive intact.
      HttpResponse<byte[]> reply =
          http.send(builder.build(), HttpResponse.BodyHandlers.ofByteArray());
      return Response.ofBytes(reply.statusCode(), reply.headers().map(), reply.body());
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
      InterruptedIOException io = new InterruptedIOException("interrupted while sending");
      io.initCause(e);
      throw io;
    }
  }
}
