package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.exception.ApiException;
import cc.alphacore.cyberbiz.exception.AuthenticationException;
import cc.alphacore.cyberbiz.exception.ForbiddenException;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.exception.RateLimitException;
import cc.alphacore.cyberbiz.exception.ServerException;
import cc.alphacore.cyberbiz.exception.ValidationException;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParseException;
import com.google.gson.JsonParser;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

/** Turns error responses into typed exceptions, as the PHP SDK's ErrorMapper does. */
final class ErrorMapper {
  private static final List<String> MESSAGE_KEYS =
      List.of("error", "errors", "message", "messages");
  private static final Set<String> BARE_ERROR_KEYS = Set.of("error", "errors", "messages");

  private ErrorMapper() {}

  /** Throws for a non-2xx status, or a 2xx whose body is only an error object. */
  static void check(Request request, Response response) {
    if (response.isSuccessful() && !isBareErrorObject(response.body())) {
      return;
    }
    throw exception(request, response);
  }

  private static ApiException exception(Request request, Response response) {
    int status = response.statusCode();
    String path = "/" + request.path().replaceFirst("^/+", "");
    String id = response.requestId();
    List<String> messages = messages(response.body());
    String body = response.body();
    String method = request.method();
    if (status == 401) {
      return new AuthenticationException(status, method, path, id, messages, body);
    } else if (status == 403) {
      return new ForbiddenException(status, method, path, id, messages, body);
    } else if (status == 404) {
      return new NotFoundException(status, method, path, id, messages, body);
    } else if (status == 422) {
      return new ValidationException(status, method, path, id, messages, body);
    } else if (status == 429) {
      return new RateLimitException(status, method, path, id, messages, body);
    } else if (status >= 500) {
      return new ServerException(status, method, path, id, messages, body);
    }
    return new ApiException(status, method, path, id, messages, body);
  }

  /**
   * Messages from the shapes CYBERBIZ uses: {"error": ...}, {"errors": ...}, {"message": ...} and
   * {"messages": ...}, each a string, a list or a map ("field: text").
   */
  static List<String> messages(String body) {
    Optional<JsonObject> object = parseObject(body);
    if (object.isEmpty()) {
      return List.of();
    }
    List<String> out = new ArrayList<>();
    for (String key : MESSAGE_KEYS) {
      flatten(object.get().get(key), "", out);
    }
    return List.copyOf(out);
  }

  private static void flatten(JsonElement value, String prefix, List<String> out) {
    if (value == null || value.isJsonNull()) {
      return;
    }
    if (value.isJsonPrimitive()) {
      // Like the PHP SDK, only strings count as messages.
      if (value.getAsJsonPrimitive().isString() && !value.getAsString().isEmpty()) {
        out.add(prefix + value.getAsString());
      }
    } else if (value.isJsonArray()) {
      value.getAsJsonArray().forEach(item -> flatten(item, prefix, out));
    } else {
      for (Map.Entry<String, JsonElement> entry : value.getAsJsonObject().entrySet()) {
        flatten(entry.getValue(), prefix + entry.getKey() + ": ", out);
      }
    }
  }

  /** A body that is exactly one of the error keys and nothing else. */
  private static boolean isBareErrorObject(String body) {
    Optional<JsonObject> object = parseObject(body);
    return object.isPresent()
        && object.get().size() == 1
        && BARE_ERROR_KEYS.contains(object.get().keySet().iterator().next());
  }

  private static Optional<JsonObject> parseObject(String body) {
    try {
      JsonElement parsed = JsonParser.parseString(body);
      return parsed.isJsonObject() ? Optional.of(parsed.getAsJsonObject()) : Optional.empty();
    } catch (JsonParseException e) {
      // Not JSON (an HTML error page, say): no messages to extract.
      return Optional.empty();
    }
  }
}
