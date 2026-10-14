package cc.alphacore.cyberbiz.model;

import com.google.gson.JsonSyntaxException;
import com.google.gson.TypeAdapter;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;
import com.google.gson.stream.JsonWriter;
import java.io.IOException;

/**
 * Reads a latitude or longitude. The platform sends store coordinates as JSON numbers or as decimal
 * strings ("25.0330000"); JSON null and an empty string become null. Anything else fails.
 */
final class CoordinateAdapter extends TypeAdapter<Double> {

  @Override
  public Double read(JsonReader in) throws IOException {
    JsonToken token = in.peek();
    if (token == JsonToken.NULL) {
      in.nextNull();
      return null;
    }
    if (token != JsonToken.NUMBER && token != JsonToken.STRING) {
      throw new JsonSyntaxException("expected a coordinate at " + in.getPath() + ", got " + token);
    }
    String text = in.nextString().strip();
    if (text.isEmpty()) {
      return null;
    }
    try {
      double value = Double.parseDouble(text);
      if (Double.isNaN(value) || Double.isInfinite(value)) {
        throw new NumberFormatException("not a finite number");
      }
      return value;
    } catch (NumberFormatException e) {
      throw new JsonSyntaxException("invalid coordinate at " + in.getPreviousPath(), e);
    }
  }

  @Override
  public void write(JsonWriter out, Double value) throws IOException {
    if (value == null) {
      out.nullValue();
      return;
    }
    out.value(value);
  }
}
