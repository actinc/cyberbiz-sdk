package cc.alphacore.cyberbiz.http;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.net.http.HttpClient;
import org.junit.jupiter.api.Test;

class JdkTransportDefaultsTest {

  @Test
  void theDefaultClientSpeaksHttp11() {
    // Over HTTP/2 the JDK client cannot talk to the CYBERBIZ API host (CBSDK-42).
    assertEquals(HttpClient.Version.HTTP_1_1, JdkTransport.defaultClient().version());
  }

  @Test
  void theDefaultClientKeepsTheConnectTimeout() {
    assertEquals(
        JdkTransport.DEFAULT_CONNECT_TIMEOUT, JdkTransport.defaultClient().connectTimeout().get());
  }
}
