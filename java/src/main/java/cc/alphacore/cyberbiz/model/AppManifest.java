package cc.alphacore.cyberbiz.model;

import com.google.gson.JsonElement;
import java.util.List;

/**
 * An App's manifest.
 *
 * @param name the App name
 * @param version the App version
 * @param scopes the requested scopes, space-separated
 * @param manifestVersion the manifest format version
 * @param type the App type, e.g. {@code other}
 * @param webhookEvents the Events the App subscribes to, e.g. {@code orders/paid}; unmodifiable
 * @param settingFields the declared setting fields as free-form JSON; a copy, never null ({@code
 *     JsonNull} when absent)
 */
public record AppManifest(
    String name,
    String version,
    String scopes,
    int manifestVersion,
    String type,
    List<String> webhookEvents,
    JsonElement settingFields) {

  /** Copies the list and the free-form JSON. */
  public AppManifest {
    webhookEvents = Lists.copy(webhookEvents);
    settingFields = Copies.json(settingFields);
  }

  /** Returns a copy of the setting fields, so the record stays immutable. */
  @Override
  public JsonElement settingFields() {
    return settingFields.deepCopy();
  }
}
