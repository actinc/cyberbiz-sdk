package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.http.Transport;
import com.google.gson.JsonObject;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;
import java.util.zip.ZipOutputStream;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/** Raw response bytes end to end: binary label zips, UTF-8 JSON and custom transports. */
class BinaryResponseTest {
  private static final String CVS_PATH = "/v2/orders/cvs_shipping_labels";
  private static final String SUPPORT_PATH = "/v2/orders/fulfillments/support_shipping_labels";
  private static final String LABEL_TEXT = "SYNTHETIC LABEL 合成標籤 #0001\n";
  private static final String CHINESE_JSON = "{\"name\":\"合成商店\",\"note\":\"中文測試\"}";

  private HttpServer server;
  private byte[] zip;

  @BeforeEach
  void start() throws IOException {
    zip = syntheticZip();
    server = HttpServer.create(new InetSocketAddress(InetAddress.getLoopbackAddress(), 0), 0);
    server.createContext("/", this::serve);
    server.start();
  }

  @AfterEach
  void stop() {
    server.stop(0);
  }

  /** Serves the zip on the two label paths and Chinese JSON everywhere else. */
  private void serve(HttpExchange exchange) throws IOException {
    exchange.getRequestBody().readAllBytes();
    String path = exchange.getRequestURI().getPath();
    byte[] reply;
    if (CVS_PATH.equals(path) || SUPPORT_PATH.equals(path)) {
      reply = zip;
      exchange.getResponseHeaders().add("Content-Type", "application/zip");
      exchange.getResponseHeaders().add("Content-Disposition", "attachment; filename=\"l.zip\"");
    } else {
      String json = path.equals("/v1/products") ? "[" + CHINESE_JSON + "]" : CHINESE_JSON;
      reply = json.getBytes(StandardCharsets.UTF_8);
      exchange.getResponseHeaders().add("Content-Type", "application/json; charset=utf-8");
    }
    exchange.sendResponseHeaders(200, reply.length);
    try (OutputStream out = exchange.getResponseBody()) {
      out.write(reply);
    }
  }

  /**
   * A zip whose bytes include every value 0-255 (stored, not deflated), so any text decoding of the
   * body would corrupt it.
   */
  private static byte[] syntheticZip() throws IOException {
    byte[] allBytes = new byte[256];
    for (int i = 0; i < allBytes.length; i++) {
      allBytes[i] = (byte) i;
    }
    ByteArrayOutputStream buffer = new ByteArrayOutputStream();
    try (ZipOutputStream out = new ZipOutputStream(buffer)) {
      out.putNextEntry(new ZipEntry("label-0001.txt"));
      out.write(LABEL_TEXT.getBytes(StandardCharsets.UTF_8));
      out.closeEntry();
      out.putNextEntry(new ZipEntry("all-bytes.bin"));
      out.write(allBytes);
      out.closeEntry();
    }
    return buffer.toByteArray();
  }

  private CyberbizClient client() {
    URI base = URI.create("http://127.0.0.1:" + server.getAddress().getPort());
    return CyberbizClient.builder("synthetic").baseUrl(base).build();
  }

  /** Reads the first entry of an archive back. */
  private static String firstEntry(byte[] archive) throws IOException {
    try (ZipInputStream in = new ZipInputStream(new ByteArrayInputStream(archive))) {
      ZipEntry entry = in.getNextEntry();
      assertEquals("label-0001.txt", entry.getName());
      return new String(in.readAllBytes(), StandardCharsets.UTF_8);
    }
  }

  @Test
  void cvsLabelsArriveByteForByte() throws IOException {
    Response labels =
        client()
            .orders()
            .printCvsShippingLabels(
                Map.of("shipping_type", "seven", "fulfillment_ids", List.of(101, 102)));

    assertArrayEquals(zip, labels.bytes());
    assertEquals(LABEL_TEXT, firstEntry(labels.bytes()));
    assertEquals("application/zip", labels.contentType());
    assertEquals(Optional.of("l.zip"), labels.filename());
  }

  @Test
  void supportLabelsArriveByteForByte() throws IOException {
    Response labels =
        client()
            .orders()
            .printSupportShippingLabels(
                Map.of("shipping_type", "hct", "print_type", "normal", "order_ids", List.of(7)));

    assertArrayEquals(zip, labels.bytes());
    assertEquals(LABEL_TEXT, firstEntry(labels.bytes()));
  }

  @Test
  void chineseJsonRoundTripsThroughJdkTransport() {
    CyberbizClient client = client();

    Response shop = client.send(Request.of("GET", "/v1/shop"));
    JsonObject product =
        client.list(Request.of("GET", "/v1/products"), JsonObject.class).items().get(0);
    JsonObject first =
        client.all(Request.of("GET", "/v1/products"), JsonObject.class).iterator().next();

    assertEquals(CHINESE_JSON, shop.body());
    assertArrayEquals(CHINESE_JSON.getBytes(StandardCharsets.UTF_8), shop.bytes());
    assertEquals("合成商店", product.get("name").getAsString());
    assertEquals("中文測試", first.get("note").getAsString());
  }

  @Test
  void customTransportBuiltFromTextStillWorks() {
    // Written against the former record: builds every reply from a String.
    Transport legacy =
        request -> {
          String text =
              request.uri().getPath().endsWith("/products")
                  ? "[" + CHINESE_JSON + "]"
                  : CHINESE_JSON;
          return new Response(200, Map.of("X-Request-Id", List.of("req-7")), text);
        };
    CyberbizClient client = CyberbizClient.builder("synthetic").transport(legacy).build();

    Response shop = client.send(Request.of("GET", "/v1/shop"));
    JsonObject decoded =
        client.list(Request.of("GET", "/v1/products"), JsonObject.class).items().get(0);

    assertEquals(CHINESE_JSON, shop.body());
    assertArrayEquals(CHINESE_JSON.getBytes(StandardCharsets.UTF_8), shop.bytes());
    assertEquals("req-7", shop.requestId());
    assertEquals("合成商店", decoded.get("name").getAsString());
  }

  @Test
  void bytesAreCopiedInAndOut() {
    byte[] source = {1, 2, 3};
    Response response = Response.ofBytes(200, Map.of(), source);

    source[0] = 9;
    response.bytes()[1] = 9;

    assertArrayEquals(new byte[] {1, 2, 3}, response.bytes());
  }

  @Test
  void textAndBytesConstructionsAgree() {
    byte[] utf8 = CHINESE_JSON.getBytes(StandardCharsets.UTF_8);

    Response fromText = new Response(200, Map.of(), CHINESE_JSON);
    Response fromBytes = Response.ofBytes(200, Map.of(), utf8);

    assertEquals(fromText, fromBytes);
    assertEquals(fromText.hashCode(), fromBytes.hashCode());
    assertEquals(CHINESE_JSON, fromBytes.body());
    assertNotEquals(fromText, Response.ofBytes(201, Map.of(), utf8));
    assertFalse(fromText.toString().contains("合成"));
  }

  @Test
  void nullBodiesAreEmpty() {
    assertEquals(0, Response.ofBytes(204, null, null).bytes().length);
    assertEquals("", new Response(204, null, null).body());
  }

  @Test
  void filenameReadsBothDispositionForms() {
    assertEquals(
        Optional.of("標籤 a+b.zip"),
        disposition("attachment; filename*=UTF-8''%E6%A8%99%E7%B1%A4%20a+b.zip"));
    assertEquals(Optional.of("labels.zip"), disposition("attachment; filename=labels.zip"));
    assertEquals(Optional.empty(), disposition("attachment"));
    assertTrue(new Response(200, Map.of(), "").filename().isEmpty());
    assertNull(new Response(200, Map.of(), "").headers().get("Content-Disposition"));
  }

  private static Optional<String> disposition(String value) {
    return new Response(200, Map.of("Content-Disposition", List.of(value)), "").filename();
  }
}
