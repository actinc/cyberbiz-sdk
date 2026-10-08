package cc.alphacore.cyberbiz.model;

import com.google.gson.JsonElement;
import java.time.OffsetDateTime;

/**
 * The installed App's record on the Shop, settings included ({@code GET /settings}, the {@code
 * shop_add_on} object). {@link #toString()} masks the token.
 *
 * @param id the installation id
 * @param settings the setting values, shaped by the App's manifest; a copy, never null ({@code
 *     JsonNull} when absent)
 * @param vendorType the vendor type
 * @param token the installation token; keep it secret
 * @param startAt when the subscription started, or null
 * @param endAt when the subscription ends, or null
 * @param addOnVersion the installed App version, or null
 */
public record AppSettings(
    long id,
    JsonElement settings,
    String vendorType,
    String token,
    OffsetDateTime startAt,
    OffsetDateTime endAt,
    AddOnVersion addOnVersion) {

  /** Copies the free-form settings. */
  public AppSettings {
    settings = Copies.json(settings);
  }

  /** Returns a copy of the settings, so the record stays immutable. */
  @Override
  public JsonElement settings() {
    return settings.deepCopy();
  }

  /** Describes the record with the token masked, so it is safe to log. */
  @Override
  public String toString() {
    return "AppSettings[id="
        + id
        + ", settings="
        + settings
        + ", vendorType="
        + vendorType
        + ", token=***, startAt="
        + startAt
        + ", endAt="
        + endAt
        + ", addOnVersion="
        + addOnVersion
        + "]";
  }
}
