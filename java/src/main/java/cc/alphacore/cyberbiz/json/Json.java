package cc.alphacore.cyberbiz.json;

import cc.alphacore.cyberbiz.Money;
import cc.alphacore.cyberbiz.exception.DecodeException;
import com.google.gson.FieldNamingPolicy;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonParseException;
import com.google.gson.Strictness;
import com.google.gson.reflect.TypeToken;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;
import java.io.IOException;
import java.io.StringReader;
import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.Collections;
import java.util.List;
import java.util.Objects;

/**
 * The SDK's single JSON configuration, shared by the client and every model.
 *
 * <ul>
 *   <li>Record components in camelCase map to the API's snake_case ({@code createdAt} reads {@code
 *       created_at}); {@code @SerializedName} overrides a name.
 *   <li>{@link BigDecimal} is read exactly from a JSON number or a numeric string, never through
 *       {@code double}; numbers read into {@code Object} or {@code Map} become {@code BigDecimal}
 *       too. Decimal input is bounded against resource exhaustion: at most 64 characters, no
 *       exponent notation, at most 40 digits and 32 decimal places; anything else fails.
 *   <li>{@link OffsetDateTime} is read in Asia/Taipei, with or without an offset ({@link Times}).
 *   <li>Unknown fields are ignored; malformed JSON is rejected.
 * </ul>
 *
 * Every decoding failure becomes an unchecked {@link DecodeException}.
 */
public final class Json {

  private static final Gson GSON =
      new GsonBuilder()
          .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
          .registerTypeAdapter(BigDecimal.class, new BigDecimalAdapter())
          .registerTypeAdapter(Money.class, new MoneyAdapter())
          .registerTypeAdapter(OffsetDateTime.class, new OffsetDateTimeAdapter())
          .setObjectToNumberStrategy(Decimals.STRATEGY)
          .setNumberToNumberStrategy(Decimals.STRATEGY)
          .setStrictness(Strictness.STRICT)
          .disableHtmlEscaping()
          .create();

  private Json() {}

  /**
   * Parses decimal text with the same bounds as decoding: plain notation (no exponent), at most 64
   * characters, 40 significant digits and 32 decimal places.
   *
   * @param text the decimal text, already stripped
   * @return the exact value
   * @throws DecodeException when the text is not a plain decimal within the limits
   */
  public static BigDecimal decimal(String text) {
    try {
      return Decimals.parse(Objects.requireNonNull(text, "text"));
    } catch (NumberFormatException e) {
      throw new DecodeException("cyberbiz: " + e.getMessage(), e);
    }
  }

  /**
   * Returns the shared, thread-safe Gson instance with the SDK's configuration.
   *
   * @return the Gson instance
   */
  public static Gson gson() {
    return GSON;
  }

  /**
   * Decodes a JSON document.
   *
   * @param json the JSON text
   * @param type the class to decode into, e.g. a record
   * @param <T> the result type
   * @return the value, or null for the JSON literal null or empty text
   * @throws DecodeException when the JSON is malformed or does not fit the type
   */
  public static <T> T decode(String json, Class<T> type) {
    Objects.requireNonNull(type, "type");
    return decode(json, TypeToken.get(type));
  }

  /**
   * Decodes a JSON document into a generic type, e.g. {@code new TypeToken<Map<String, Product>>()
   * {}}.
   *
   * @param json the JSON text
   * @param type the type to decode into
   * @param <T> the result type
   * @return the value, or null for the JSON literal null or empty text
   * @throws DecodeException when the JSON is malformed or does not fit the type
   */
  public static <T> T decode(String json, TypeToken<T> type) {
    Objects.requireNonNull(type, "type");
    if (json == null || json.isBlank()) {
      return null;
    }
    return read(json, type, false);
  }

  /**
   * Decodes a JSON array, as list endpoints return it. Empty text and the JSON literal null count
   * as an empty list.
   *
   * @param json the JSON text
   * @param itemType the class of each item
   * @param <T> the item type
   * @return the items in order; unmodifiable
   * @throws DecodeException when the JSON is malformed, not an array, or an item does not fit
   */
  public static <T> List<T> decodeList(String json, Class<T> itemType) {
    Objects.requireNonNull(itemType, "itemType");
    if (json == null || json.isBlank()) {
      return List.of();
    }
    @SuppressWarnings("unchecked") // getParameterized(List.class, T) is exactly List<T>
    TypeToken<List<T>> listType =
        (TypeToken<List<T>>) TypeToken.getParameterized(List.class, itemType);
    List<T> items = read(json, listType, true);
    return items == null ? List.of() : Collections.unmodifiableList(items);
  }

  /**
   * Encodes a value as a request body: amounts as plain JSON numbers, times in the platform layout
   * in Asia/Taipei, camelCase names as snake_case; null fields are left out.
   *
   * @param value the value
   * @return the JSON text
   */
  public static String encode(Object value) {
    return GSON.toJson(value);
  }

  /** Decodes one strict JSON document; failures name the JSON path where they occurred. */
  private static <T> T read(String json, TypeToken<T> type, boolean array) {
    JsonReader reader = new JsonReader(new StringReader(json));
    reader.setStrictness(Strictness.STRICT);
    try {
      JsonToken first = reader.peek();
      if (array && first != JsonToken.BEGIN_ARRAY && first != JsonToken.NULL) {
        String kind = first == JsonToken.BEGIN_OBJECT ? "an object" : "a scalar";
        throw new DecodeException("cyberbiz: expected a JSON array, got " + kind);
      }
      T value = GSON.fromJson(reader, type);
      if (reader.peek() != JsonToken.END_DOCUMENT) {
        throw new DecodeException(
            "cyberbiz: unexpected data after the JSON value at " + reader.getPath());
      }
      return value;
    } catch (JsonParseException | IOException e) {
      String message = e.getMessage();
      if (!message.contains("$")) {
        message += " at path " + reader.getPreviousPath();
      }
      throw new DecodeException("cyberbiz: " + message, e);
    }
  }
}
