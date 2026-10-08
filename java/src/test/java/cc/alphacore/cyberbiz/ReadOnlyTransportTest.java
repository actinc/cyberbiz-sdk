package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.net.URI;
import java.util.Map;
import org.junit.jupiter.api.Test;

/** The guard {@link LiveTest} relies on: only GET reaches the network. */
class ReadOnlyTransportTest {
  private final FakeTransport fake = new FakeTransport();

  private CyberbizClient client() {
    return CyberbizClient.builder("synthetic-token-0123456789")
        .baseUrl(URI.create("https://api.example.test"))
        .transport(new ReadOnlyTransport(fake))
        .rateLimit(0)
        .maxRetries(0)
        .build();
  }

  @Test
  void passesGetThrough() {
    fake.reply(200, Golden.body("app/GET_shop.json"));

    client().shop().info();

    assertEquals(1, fake.requests.size());
    assertEquals("GET", fake.requests.get(0).method());
  }

  @Test
  void refusesWritesBeforeSending() {
    IllegalStateException e =
        assertThrows(
            IllegalStateException.class,
            () -> client().shop().updateSettings(Map.of("greeting", "hello")));

    assertTrue(e.getMessage().contains("PUT"), e.getMessage());
    assertTrue(fake.requests.isEmpty(), "a write reached the transport");
  }
}
