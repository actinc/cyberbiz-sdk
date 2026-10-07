package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;

import cc.alphacore.cyberbiz.http.JdkTransport;
import cc.alphacore.cyberbiz.http.TransportRequest;
import com.sun.net.httpserver.HttpServer;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.concurrent.atomic.AtomicReference;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/** The default transport against a real local HTTP server. */
class JdkTransportTest {
  private HttpServer server;
  private final AtomicReference<String> seen = new AtomicReference<>();

  @BeforeEach
  void start() throws IOException {
    server = HttpServer.create(new InetSocketAddress(InetAddress.getLoopbackAddress(), 0), 0);
    server.createContext(
        "/",
        exchange -> {
          String body =
              new String(exchange.getRequestBody().readAllBytes(), StandardCharsets.UTF_8);
          seen.set(
              exchange.getRequestMethod()
                  + " "
                  + exchange.getRequestURI()
                  + " "
                  + exchange.getRequestHeaders().getFirst("Authorization")
                  + " "
                  + body);
          byte[] reply = "{\"name\":\"合成商店\"}".getBytes(StandardCharsets.UTF_8);
          exchange.getResponseHeaders().add("X-Request-Id", "req-42");
          exchange.sendResponseHeaders(201, reply.length);
          try (OutputStream out = exchange.getResponseBody()) {
            out.write(reply);
          }
        });
    server.start();
  }

  @AfterEach
  void stop() {
    server.stop(0);
  }

  @Test
  void sendsAndReadsTheWholeExchange() throws IOException {
    URI uri = URI.create("http://127.0.0.1:" + server.getAddress().getPort() + "/v1/shop?a=1");

    Response response =
        new JdkTransport()
            .send(
                new TransportRequest(
                    "POST", uri, Map.of("Authorization", "Bearer synthetic"), "{\"x\":1}"));

    assertEquals("POST /v1/shop?a=1 Bearer synthetic {\"x\":1}", seen.get());
    assertEquals(201, response.statusCode());
    assertEquals("req-42", response.requestId());
    assertEquals("{\"name\":\"合成商店\"}", response.body());
  }

  @Test
  void worksThroughTheClient() {
    URI base = URI.create("http://127.0.0.1:" + server.getAddress().getPort());
    CyberbizClient client = CyberbizClient.builder("synthetic").baseUrl(base).build();

    assertEquals(201, client.send(Request.of("GET", "/v1/shop")).statusCode());
    assertEquals("GET /v1/shop Bearer synthetic ", seen.get());
  }
}
