package cc.alphacore.cyberbiz.json;

import cc.alphacore.cyberbiz.Money;
import com.google.gson.JsonSyntaxException;
import com.google.gson.TypeAdapter;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;
import com.google.gson.stream.JsonWriter;
import java.io.IOException;

/**
 * Reads an amount (a JSON number or numeric string) into {@link Money}, with the same bounds as
 * {@link BigDecimalAdapter}; JSON null and an empty string become null. Writes a plain JSON number
 * with two decimal places.
 */
final class MoneyAdapter extends TypeAdapter<Money> {

  @Override
  public Money read(JsonReader in) throws IOException {
    JsonToken token = in.peek();
    if (token == JsonToken.NULL) {
      in.nextNull();
      return null;
    }
    if (token != JsonToken.NUMBER && token != JsonToken.STRING) {
      throw new JsonSyntaxException("expected an amount at " + in.getPath() + ", got " + token);
    }
    String text = in.nextString().strip();
    if (text.isEmpty()) {
      return null;
    }
    try {
      return Money.of(Decimals.parse(text));
    } catch (NumberFormatException e) {
      throw new JsonSyntaxException(e.getMessage() + " at " + in.getPreviousPath(), e);
    }
  }

  @Override
  public void write(JsonWriter out, Money value) throws IOException {
    if (value == null) {
      out.nullValue();
      return;
    }
    out.jsonValue(value.toString());
  }
}
