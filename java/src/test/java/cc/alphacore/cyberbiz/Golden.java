package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.json.Json;
import com.google.gson.JsonElement;
import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

/**
 * Reads the shared Golden Files in {@code testdata/golden/} at the repository root (never copied
 * into {@code java/}) and decodes them with the SDK's JSON configuration. A decoding failure names
 * the file as well as the field.
 */
public final class Golden {

  /** {@code testdata/golden/}; Maven runs the tests in {@code java/}. */
  public static final Path ROOT = Path.of("..", "testdata", "golden").toAbsolutePath().normalize();

  private Golden() {}

  /**
   * Resolves a Golden File.
   *
   * @param name a path relative to {@link #ROOT}, e.g. {@code v1/GET_v1_products.json}, or absolute
   * @return the absolute path
   */
  public static Path path(String name) {
    return ROOT.resolve(name);
  }

  /** Every response body under {@link #ROOT} (not the .headers.json files), relative and sorted. */
  public static List<String> files() {
    try (Stream<Path> paths = Files.walk(ROOT)) {
      return paths
          .filter(Files::isRegularFile)
          .map(path -> ROOT.relativize(path).toString().replace('\\', '/'))
          .filter(name -> name.endsWith(".json") && !name.endsWith(".headers.json"))
          .sorted()
          .toList();
    } catch (IOException e) {
      throw new UncheckedIOException("cannot list Golden Files in " + ROOT, e);
    }
  }

  /** The raw body of a Golden File. */
  public static String body(String name) {
    try {
      return Files.readString(path(name), StandardCharsets.UTF_8);
    } catch (IOException e) {
      throw new UncheckedIOException("cannot read Golden File " + path(name), e);
    }
  }

  /** The recorded response: status 200, the body, and the headers saved next to it. */
  public static Response response(String name) {
    String headersName = name.replaceFirst("\\.json$", ".headers.json");
    Map<String, List<String>> headers = new LinkedHashMap<>();
    if (Files.isRegularFile(path(headersName))) {
      JsonElement saved = Json.decode(body(headersName), JsonElement.class);
      saved
          .getAsJsonObject()
          .entrySet()
          .forEach(
              entry ->
                  headers.put(
                      entry.getKey(),
                      entry.getValue().getAsJsonArray().asList().stream()
                          .map(JsonElement::getAsString)
                          .toList()));
    }
    return new Response(200, headers, body(name));
  }

  /**
   * Decodes a Golden File holding one JSON object.
   *
   * @throws DecodeException naming the file and the field
   */
  public static <T> T object(String name, Class<T> type) {
    try {
      return Json.decode(body(name), type);
    } catch (DecodeException e) {
      throw named(name, e);
    }
  }

  /**
   * Decodes the object under {@code envelope} in a Golden File, e.g. {@code shop_info}.
   *
   * @throws DecodeException naming the file and the field
   */
  public static <T> T wrapped(String name, String envelope, Class<T> type) {
    try {
      JsonElement inner =
          Json.decode(body(name), JsonElement.class).getAsJsonObject().get(envelope);
      if (inner == null) {
        throw new DecodeException("cyberbiz: no \"" + envelope + "\" object");
      }
      return Json.decode(inner.toString(), type);
    } catch (DecodeException e) {
      throw named(name, e);
    }
  }

  /**
   * Decodes a Golden File holding a JSON array, as a list endpoint returns it.
   *
   * @throws DecodeException naming the file and the field
   */
  public static <T> List<T> list(String name, Class<T> itemType) {
    try {
      return Json.decodeList(body(name), itemType);
    } catch (DecodeException e) {
      throw named(name, e);
    }
  }

  private static DecodeException named(String name, DecodeException e) {
    String file = path(name).getFileName().toString();
    return new DecodeException(file + ": " + e.getMessage().replaceFirst("^cyberbiz: ", ""), e);
  }
}
