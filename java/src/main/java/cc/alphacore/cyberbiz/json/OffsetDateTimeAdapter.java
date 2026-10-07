package cc.alphacore.cyberbiz.json;

import com.google.gson.JsonSyntaxException;
import com.google.gson.TypeAdapter;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;
import com.google.gson.stream.JsonWriter;
import java.io.IOException;
import java.time.DateTimeException;
import java.time.OffsetDateTime;

/**
 * Reads CYBERBIZ timestamps into {@link OffsetDateTime} in Asia/Taipei (see {@link Times}) and
 * writes them back in the platform layout. JSON null and an empty string become null.
 */
final class OffsetDateTimeAdapter extends TypeAdapter<OffsetDateTime> {

  @Override
  public OffsetDateTime read(JsonReader in) throws IOException {
    JsonToken token = in.peek();
    if (token == JsonToken.NULL) {
      in.nextNull();
      return null;
    }
    if (token != JsonToken.STRING) {
      throw new JsonSyntaxException(
          "expected a timestamp string at " + in.getPath() + ", got " + token);
    }
    String text = in.nextString();
    if (text.isBlank()) {
      return null;
    }
    try {
      return Times.parseStrict(text);
    } catch (DateTimeException e) {
      throw new JsonSyntaxException(
          "cannot parse \"" + text + "\" as a CYBERBIZ timestamp at " + in.getPreviousPath(), e);
    }
  }

  @Override
  public void write(JsonWriter out, OffsetDateTime value) throws IOException {
    if (value == null) {
      out.nullValue();
      return;
    }
    out.value(Times.format(value));
  }
}
