package cc.alphacore.cyberbiz.webhook;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * The signed samples in docs/api/en/webhooks.md ("## Samples"), read in place as the PHP tests do:
 * one per Event, with the HTTP headers, the base64 form of the signature and the exact body (the
 * JSON block plus its trailing newline), all signed with the App Secret "example-app-secret".
 */
final class WebhookSamples {

  static final String SECRET = "example-app-secret";

  static final Path FILE = Path.of("..", "docs", "api", "en", "webhooks.md");

  private static final Pattern SAMPLE =
      Pattern.compile(
          "^### `([^`]+)`\\n.*?```http\\n(.*?)```.*?`([A-Za-z0-9+/=]{44})`.*?```json\\n(.*?)```",
          Pattern.MULTILINE | Pattern.DOTALL);

  /**
   * One signed sample.
   *
   * @param event the X-Cyberbiz-Event value
   * @param headers the headers in the order the document lists them
   * @param base64 the base64 form of X-Cyberbiz-Hmac-Sha256
   * @param body the body bytes as UTF-8 text
   */
  record Sample(String event, Map<String, String> headers, String base64, String body) {

    byte[] bytes() {
      return body.getBytes(StandardCharsets.UTF_8);
    }

    /** The headers with one value replaced, or removed when value is null. */
    Map<String, String> with(String name, String value) {
      Map<String, String> copy = new LinkedHashMap<>(headers);
      if (value == null) {
        copy.remove(name);
      } else {
        copy.put(name, value);
      }
      return copy;
    }
  }

  private WebhookSamples() {}

  static List<Sample> all() {
    String doc;
    try {
      doc = Files.readString(FILE, StandardCharsets.UTF_8);
    } catch (IOException e) {
      throw new UncheckedIOException(e);
    }
    Matcher m = SAMPLE.matcher(doc.substring(doc.indexOf("\n## Samples")));
    List<Sample> out = new ArrayList<>();
    while (m.find()) {
      out.add(new Sample(m.group(1), headers(m.group(2)), m.group(3), m.group(4)));
    }
    return out;
  }

  static Sample get(String event) {
    return all().stream().filter(s -> s.event().equals(event)).findFirst().orElseThrow();
  }

  private static Map<String, String> headers(String http) {
    Map<String, String> headers = new LinkedHashMap<>();
    List<String> lines = http.strip().lines().toList();
    for (String line : lines.subList(1, lines.size())) {
      int colon = line.indexOf(':');
      headers.put(line.substring(0, colon).strip(), line.substring(colon + 1).strip());
    }
    return headers;
  }
}
