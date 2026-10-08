package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assumptions.assumeTrue;

import cc.alphacore.cyberbiz.exception.ApiException;
import cc.alphacore.cyberbiz.exception.CyberbizException;
import cc.alphacore.cyberbiz.exception.TransportException;
import cc.alphacore.cyberbiz.http.JdkTransport;
import cc.alphacore.cyberbiz.model.AppSettings;
import cc.alphacore.cyberbiz.model.Product;
import cc.alphacore.cyberbiz.model.ShopInfo;
import cc.alphacore.cyberbiz.pagination.Page;
import java.util.Map;
import java.util.function.Supplier;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

/**
 * Read-only calls against the live CYBERBIZ API with the token in {@code CYBERBIZ_API_TOKEN}, the
 * Java side of Go's {@code TestLive*}. Tagged {@code live}, which {@code mvn verify} excludes; the
 * main pipeline runs them with {@code mvn -Plive test}, and they skip without a token.
 *
 * <p>Only GET methods are called, and {@link ReadOnlyTransport} refuses anything else. The tests
 * check that each call succeeds and decodes into SDK types; they never compare or print values, and
 * a failure reports only the exception type, status and request id, so no shop data reaches the
 * log.
 */
@Tag("live")
class LiveTest {
  private CyberbizClient client;

  @BeforeEach
  void liveClient() {
    String token = System.getenv("CYBERBIZ_API_TOKEN");
    assumeTrue(token != null && !token.isEmpty(), "CYBERBIZ_API_TOKEN not set");
    client =
        CyberbizClient.builder(token).transport(new ReadOnlyTransport(new JdkTransport())).build();
  }

  @Test
  void shopInfo() {
    ShopInfo info = call("GET shop", () -> client.shop().info());
    assertTrue(info.id() > 0, "shop id is not positive");
    assertNotNull(info.primaryDomain(), "no primary domain");
    assertFalse(info.primaryDomain().isEmpty(), "empty primary domain");
  }

  @Test
  void appSettings() {
    AppSettings settings = call("GET settings", () -> client.shop().settings());
    assertTrue(settings.id() > 0, "settings id is not positive");
    assertNotNull(settings.addOnVersion(), "no add_on_version");
  }

  @Test
  void productsFirstPage() {
    Page<Product> page =
        call("GET v1/products", () -> client.products().list(Map.of("page", 1, "per_page", 1)));
    assertNotNull(page.pagination(), "no pagination");
    for (Product product : page.items()) {
      assertTrue(product.id() > 0, "product id is not positive");
    }
  }

  /**
   * Runs one API call. On failure it reports the exception type, HTTP status and request id but not
   * the message or the cause, which can quote the response body.
   */
  @SuppressWarnings("UnusedException") // the cause is dropped on purpose, see above
  private static <T> T call(String what, Supplier<T> call) {
    try {
      T result = call.get();
      assertNotNull(result, what + " returned null");
      return result;
    } catch (ApiException e) {
      throw new AssertionError(
          what
              + " failed: "
              + e.getClass().getSimpleName()
              + ", HTTP "
              + e.statusCode()
              + ", X-Request-Id "
              + e.requestId());
    } catch (TransportException e) {
      // No response arrived, so the cause holds no response data: show it.
      throw new AssertionError(what + " failed: TransportException" + causes(e));
    } catch (CyberbizException e) {
      throw new AssertionError(what + " failed: " + e.getClass().getSimpleName());
    }
  }

  /** The class and message of each cause, e.g. " | EOFException: EOF reached while reading". */
  private static String causes(Throwable e) {
    StringBuilder out = new StringBuilder();
    for (Throwable t = e.getCause(); t != null; t = t.getCause()) {
      out.append(" | ").append(t.getClass().getSimpleName()).append(": ").append(t.getMessage());
    }
    return out.toString();
  }
}
