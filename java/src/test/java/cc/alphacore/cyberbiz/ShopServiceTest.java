package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.http.TransportRequest;
import cc.alphacore.cyberbiz.model.AppSettings;
import cc.alphacore.cyberbiz.model.ShopInfo;
import java.math.BigDecimal;
import java.net.URI;
import java.util.LinkedHashMap;
import java.util.Map;
import org.junit.jupiter.api.Test;

class ShopServiceTest {
  private final FakeTransport transport = new FakeTransport();

  private CyberbizClient client() {
    return CyberbizClient.builder("synthetic-token-0123456789")
        .baseUrl(URI.create("https://api.example.test"))
        .transport(transport)
        .rateLimit(0)
        .build();
  }

  @Test
  void infoGetsTheShopAndUnwrapsShopInfo() {
    transport.reply(200, Golden.body("app/GET_shop.json"));

    ShopInfo shop = client().shop().info();

    TransportRequest sent = transport.requests.get(0);
    assertEquals("GET", sent.method());
    assertEquals("https://api.example.test/shop", sent.uri().toString());
    assertNull(sent.body());
    assertEquals(26721, shop.id());
    assertEquals("TWD", shop.currency());
  }

  @Test
  void settingsGetsTheAppRecordAndUnwrapsShopAddOn() {
    transport.reply(200, Golden.body("app/GET_settings.json"));

    AppSettings settings = client().shop().settings();

    assertEquals("GET", transport.requests.get(0).method());
    assertEquals("https://api.example.test/settings", transport.requests.get(0).uri().toString());
    assertEquals(29384, settings.id());
  }

  @Test
  void updateSettingsPutsFieldDataPairsInOrder() {
    transport.reply(200, Golden.body("app/GET_settings.json"));
    Map<String, Object> values = new LinkedHashMap<>();
    values.put("greeting", "hello");
    values.put("threshold", new BigDecimal("120.50"));

    client().shop().updateSettings(values);

    TransportRequest sent = transport.requests.get(0);
    assertEquals("PUT", sent.method());
    assertEquals("https://api.example.test/settings", sent.uri().toString());
    assertEquals(
        "{\"settings\":[{\"field\":\"greeting\",\"data\":\"hello\"},"
            + "{\"field\":\"threshold\",\"data\":120.50}]}",
        sent.body());
  }

  @Test
  void aMissingEnvelopeIsADecodeErrorNamingTheRequest() {
    transport.reply(200, "{\"shop\":{}}");

    DecodeException e = assertThrows(DecodeException.class, () -> client().shop().info());

    assertEquals("cyberbiz: GET /shop: expected an object under \"shop_info\"", e.getMessage());
  }
}
