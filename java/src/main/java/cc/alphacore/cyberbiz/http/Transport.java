package cc.alphacore.cyberbiz.http;

import cc.alphacore.cyberbiz.Response;
import java.io.IOException;

/**
 * Sends one HTTP request. The default, {@link JdkTransport}, uses the JDK's {@code HttpClient};
 * implement this to use another HTTP library or to fake the API in tests. Implementations must be
 * thread-safe and must not retry: the client does that.
 */
@FunctionalInterface
public interface Transport {

  /**
   * Sends the request and reads the whole response.
   *
   * @param request the request to send
   * @return the response, whatever its status
   * @throws IOException when no response arrived
   */
  Response send(TransportRequest request) throws IOException;
}
