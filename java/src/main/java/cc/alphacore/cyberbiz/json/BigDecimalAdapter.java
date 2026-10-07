package cc.alphacore.cyberbiz.json;

import com.google.gson.JsonSyntaxException;
import com.google.gson.TypeAdapter;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;
import com.google.gson.stream.JsonWriter;
import java.io.IOException;
import java.math.BigDecimal;

/**
 * Reads amounts exactly: a JSON number or a numeric string becomes a {@link BigDecimal} built from
 * the literal text, never through {@code double}, so {@code 0.1} stays 0.1 and {@code 120.50} keeps
 * its scale. JSON null and an empty string become null. Untrusted input is bounded by {@link
 * Decimals}: at most 64 characters, no exponent notation, at most 40 digits and 32 decimal places;
 * anything else fails. Writes a plain JSON number (no exponent).
 */
final class BigDecimalAdapter extends TypeAdapter<BigDecimal> {

  @Override
  public BigDecimal read(JsonReader in) throws IOException {
    JsonToken token = in.peek();
    if (token == JsonToken.NULL) {
      in.nextNull();
      return null;
    }
    if (token != JsonToken.NUMBER && token != JsonToken.STRING) {
      throw new JsonSyntaxException("expected a decimal at " + in.getPath() + ", got " + token);
    }
    String text = in.nextString().strip();
    if (text.isEmpty()) {
      return null;
    }
    try {
      return Decimals.parse(text);
    } catch (NumberFormatException e) {
      throw new JsonSyntaxException(e.getMessage() + " at " + in.getPreviousPath(), e);
    }
  }

  @Override
  public void write(JsonWriter out, BigDecimal value) throws IOException {
    if (value == null) {
      out.nullValue();
      return;
    }
    out.jsonValue(value.toPlainString());
  }
}
