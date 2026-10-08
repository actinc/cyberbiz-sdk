package cc.alphacore.cyberbiz.resource;

import cc.alphacore.cyberbiz.CyberbizClient;
import cc.alphacore.cyberbiz.Request;
import cc.alphacore.cyberbiz.Response;
import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.json.Json;
import com.google.gson.JsonElement;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;

/** The request-and-decode steps the resource services share, as the PHP SDK's Endpoint. */
final class Calls {
  private final CyberbizClient client;

  Calls(CyberbizClient client) {
    this.client = Objects.requireNonNull(client, "client");
  }

  CyberbizClient client() {
    return client;
  }

  /** Sends a request whose reply is one JSON object; an empty or null reply fails to decode. */
  <T> T object(Request request, Class<T> type) {
    Response response = client.send(request);
    T value = decode(request, response, type);
    if (value == null) {
      throw new DecodeException(describe(request) + ": empty response body");
    }
    return value;
  }

  /**
   * Like {@link #object} for a GET by id: the API answers some missing resources with 200 and a
   * null body, which becomes a {@link NotFoundException}.
   */
  <T> T one(Request request, Class<T> type) {
    Response response = sendForOne(request);
    T value = decode(request, response, type);
    if (value == null) {
      throw new DecodeException(describe(request) + ": empty response body");
    }
    return value;
  }

  /** Like {@link #one}, setting {@code id} on the result: detail replies omit it. */
  <T> T one(Request request, Class<T> type, long id) {
    return tree(request, sendForOne(request), type, null, id);
  }

  /** Like {@link #object} after unwrapping the object under {@code envelope}, e.g. shop_info. */
  <T> T object(Request request, Class<T> type, String envelope) {
    return tree(request, client.send(request), type, Objects.requireNonNull(envelope), null);
  }

  /** Like {@link #object}, setting {@code id} on the result: update replies may omit it. */
  <T> T withId(Request request, Class<T> type, long id) {
    return tree(request, client.send(request), type, null, id);
  }

  /** Sends a GET by id; a 200 with a null body becomes a {@link NotFoundException}. */
  private Response sendForOne(Request request) {
    Response response = client.send(request);
    if (response.isNull()) {
      throw new NotFoundException(
          404,
          request.method(),
          "/" + request.path().replaceFirst("^/+", ""),
          response.requestId(),
          List.of("resource is null"),
          response.body());
    }
    return response;
  }

  /**
   * Decodes a reply object, optionally unwrapping {@code envelope} first and setting {@code id}. An
   * empty reply, or a missing envelope, fails to decode; failures name the request.
   */
  private static <T> T tree(
      Request request, Response response, Class<T> type, String envelope, Long id) {
    JsonElement element = decode(request, response, JsonElement.class);
    if (element == null || element.isJsonNull()) {
      throw new DecodeException(describe(request) + ": empty response body");
    }
    if (envelope != null) {
      JsonElement inner = element.isJsonObject() ? element.getAsJsonObject().get(envelope) : null;
      if (inner == null || !inner.isJsonObject()) {
        throw new DecodeException(
            describe(request) + ": expected an object under \"" + envelope + "\"");
      }
      element = inner;
    }
    if (id != null && element.isJsonObject()) {
      element.getAsJsonObject().addProperty("id", id);
    }
    return decode(request, new Response(200, Map.of(), element.toString()), type);
  }

  /** Sends a write whose reply is the changed resource, or an empty or null body. */
  <T> Optional<T> optional(Request request, Class<T> type) {
    return Optional.ofNullable(decode(request, client.send(request), type));
  }

  /** Sends a request whose reply is an unpaginated JSON array. */
  <T> List<T> list(Request request, Class<T> itemType) {
    return client.list(request, itemType).items();
  }

  /** Sends a request and returns the raw reply. */
  Response call(Request request) {
    return client.send(request);
  }

  /** Decodes a reply body; null for an empty body or JSON null. Failures name the request. */
  static <T> T decode(Request request, Response response, Class<T> type) {
    try {
      return Json.decode(response.body(), type);
    } catch (DecodeException e) {
      String detail = e.getMessage().replaceFirst("^cyberbiz: ", "");
      throw new DecodeException(describe(request) + ": " + detail, e);
    }
  }

  private static String describe(Request request) {
    return "cyberbiz: " + request.method() + " /" + request.path().replaceFirst("^/+", "");
  }
}
