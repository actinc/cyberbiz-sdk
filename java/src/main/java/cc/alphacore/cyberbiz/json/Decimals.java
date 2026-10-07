package cc.alphacore.cyberbiz.json;

import com.google.gson.JsonSyntaxException;
import com.google.gson.ToNumberStrategy;
import com.google.gson.stream.JsonReader;
import java.io.IOException;
import java.math.BigDecimal;
import java.util.regex.Pattern;

/**
 * Parses untrusted decimal text into a bounded {@link BigDecimal}. A value such as {@code
 * "1e999999999"} or a 100,000-digit string parses cheaply but makes {@code toPlainString()}, {@code
 * setScale()} or arithmetic by the caller exhaust memory or CPU, so the SDK fails closed:
 *
 * <ul>
 *   <li>at most {@value #MAX_LENGTH} characters of text;
 *   <li>plain decimal notation only ({@code -?digits[.digits]}); exponent notation is rejected (the
 *       Golden Files never use it);
 *   <li>at most {@value #MAX_PRECISION} significant digits and {@value #MAX_SCALE} decimal places.
 * </ul>
 *
 * The largest number in the Golden Files has 10 characters and 2 decimal places, so the limits
 * leave ample room for real amounts and ids.
 */
final class Decimals {

  /** The longest accepted text. */
  static final int MAX_LENGTH = 64;

  /** The most significant digits accepted. */
  static final int MAX_PRECISION = 40;

  /** The most decimal places accepted. */
  static final int MAX_SCALE = 32;

  private static final Pattern PLAIN = Pattern.compile("[+-]?(\\d+(\\.\\d*)?|\\.\\d+)");

  /** Reads JSON numbers into {@code Object}, {@code Map} or {@code Number} values, bounded. */
  static final ToNumberStrategy STRATEGY =
      new ToNumberStrategy() {
        @Override
        public Number readNumber(JsonReader in) throws IOException {
          String text = in.nextString();
          try {
            return parse(text);
          } catch (NumberFormatException e) {
            throw new JsonSyntaxException(e.getMessage() + " at " + in.getPreviousPath(), e);
          }
        }
      };

  private Decimals() {}

  /**
   * Parses stripped, non-empty decimal text within the limits.
   *
   * @throws NumberFormatException naming the limit that was exceeded; the message never includes an
   *     over-long value
   */
  static BigDecimal parse(String text) {
    if (text.length() > MAX_LENGTH) {
      throw new NumberFormatException(
          "decimal of " + text.length() + " characters exceeds the limit of " + MAX_LENGTH);
    }
    if (!PLAIN.matcher(text).matches()) {
      throw new NumberFormatException("cannot parse \"" + text + "\" as a plain decimal");
    }
    BigDecimal value = new BigDecimal(text);
    if (value.precision() > MAX_PRECISION || value.scale() > MAX_SCALE) {
      throw new NumberFormatException(
          "decimal \""
              + text
              + "\" exceeds "
              + MAX_PRECISION
              + " digits or "
              + MAX_SCALE
              + " decimal places");
    }
    return value;
  }
}
