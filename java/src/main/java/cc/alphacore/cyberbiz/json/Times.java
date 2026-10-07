package cc.alphacore.cyberbiz.json;

import cc.alphacore.cyberbiz.exception.DecodeException;
import java.time.DateTimeException;
import java.time.LocalDateTime;
import java.time.OffsetDateTime;
import java.time.ZoneId;
import java.time.ZoneOffset;
import java.time.chrono.IsoChronology;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeFormatterBuilder;
import java.time.format.ResolverStyle;
import java.time.temporal.ChronoField;
import java.time.temporal.TemporalAccessor;
import java.util.Locale;

/**
 * CYBERBIZ timestamps. The platform sends {@code "2026-07-10 20:22:05"} with no zone, always
 * meaning Asia/Taipei (ADR-0004), and occasionally ISO 8601 with an offset. Every parsed value is
 * returned in Asia/Taipei (offset +08:00); the zone is deliberately not configurable.
 */
public final class Times {

  /** The zone of every CYBERBIZ timestamp. */
  public static final ZoneId ZONE = ZoneId.of("Asia/Taipei");

  /** The platform layout, {@code 2026-07-10 20:22:05}, used when sending a time. */
  public static final DateTimeFormatter FORMAT =
      DateTimeFormatter.ofPattern("uuuu-MM-dd HH:mm:ss", Locale.ROOT);

  /**
   * Date, then optionally " HH:mm[:ss[.fraction]]" and an offset ("Z", "+08:00", "+0800", "+08"),
   * with or without a space before it. A "T" separator is turned into a space before parsing.
   */
  private static final DateTimeFormatter PARSER =
      new DateTimeFormatterBuilder()
          .append(DateTimeFormatter.ISO_LOCAL_DATE)
          .optionalStart()
          .appendLiteral(' ')
          .appendValue(ChronoField.HOUR_OF_DAY, 2)
          .appendLiteral(':')
          .appendValue(ChronoField.MINUTE_OF_HOUR, 2)
          .optionalStart()
          .appendLiteral(':')
          .appendValue(ChronoField.SECOND_OF_MINUTE, 2)
          .optionalStart()
          .appendFraction(ChronoField.NANO_OF_SECOND, 0, 9, true)
          .optionalEnd()
          .optionalEnd()
          .optionalStart()
          .appendLiteral(' ')
          .optionalEnd()
          .optionalStart()
          .appendOffset("+HH:MM", "Z")
          .optionalEnd()
          .optionalStart()
          .appendOffset("+HHmm", "Z")
          .optionalEnd()
          .optionalEnd()
          .parseDefaulting(ChronoField.HOUR_OF_DAY, 0)
          .parseDefaulting(ChronoField.MINUTE_OF_HOUR, 0)
          .toFormatter(Locale.ROOT)
          .withChronology(IsoChronology.INSTANCE)
          .withResolverStyle(ResolverStyle.STRICT);

  private Times() {}

  /**
   * Parses a CYBERBIZ timestamp. A value without an offset is read in Asia/Taipei; one with an
   * offset is converted to Asia/Taipei. A date alone means midnight in Taipei.
   *
   * @param text the timestamp, e.g. {@code "2026-07-10 20:22:05"} or {@code "2026-07-10T12:22:05Z"}
   * @return the time at offset +08:00, or null for null or blank text
   * @throws DecodeException when the text is in no accepted format
   */
  public static OffsetDateTime parse(String text) {
    if (text == null || text.isBlank()) {
      return null;
    }
    try {
      return parseStrict(text);
    } catch (DateTimeException e) {
      throw new DecodeException(
          "cyberbiz: cannot parse \"" + text.strip() + "\" as a CYBERBIZ timestamp", e);
    }
  }

  /**
   * Formats a time in the platform layout, converted to Asia/Taipei.
   *
   * @param time the time
   * @return e.g. {@code "2026-07-10 20:22:05"}
   */
  public static String format(OffsetDateTime time) {
    return FORMAT.format(time.atZoneSameInstant(ZONE));
  }

  /** Parses non-blank text; throws {@link DateTimeException} when no format matches. */
  static OffsetDateTime parseStrict(String text) {
    String value = text.strip();
    if (value.length() > 10 && value.charAt(10) == 'T') {
      value = value.substring(0, 10) + ' ' + value.substring(11);
    }
    TemporalAccessor parsed = PARSER.parse(value);
    LocalDateTime local = LocalDateTime.from(parsed);
    if (parsed.isSupported(ChronoField.OFFSET_SECONDS)) {
      return local.atOffset(ZoneOffset.from(parsed)).atZoneSameInstant(ZONE).toOffsetDateTime();
    }
    return local.atZone(ZONE).toOffsetDateTime();
  }
}
