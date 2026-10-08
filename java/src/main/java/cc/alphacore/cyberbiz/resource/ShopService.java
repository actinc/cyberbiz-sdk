package cc.alphacore.cyberbiz.resource;

import cc.alphacore.cyberbiz.CyberbizClient;
import cc.alphacore.cyberbiz.Request;
import cc.alphacore.cyberbiz.exception.ApiException;
import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.exception.TransportException;
import cc.alphacore.cyberbiz.model.AppSettings;
import cc.alphacore.cyberbiz.model.ShopInfo;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * The Shop that owns the token, and this App's settings on it. Get one from {@link
 * CyberbizClient#shop()}. Every method throws {@link ApiException} for an error response, {@link
 * TransportException} when no response arrived and {@link DecodeException} when the body does not
 * fit the model.
 */
public final class ShopService {
  private final Calls calls;

  /**
   * Creates the service; prefer {@link CyberbizClient#shop()}.
   *
   * @param client the client that sends the requests
   */
  public ShopService(CyberbizClient client) {
    this.calls = new Calls(Objects.requireNonNull(client, "client"));
  }

  /**
   * Returns the Shop profile ({@code GET /shop}).
   *
   * @return the Shop
   */
  public ShopInfo info() {
    return calls.object(Request.of("GET", "/shop"), ShopInfo.class, "shop_info");
  }

  /**
   * Returns the installed App's record, settings included ({@code GET /settings}).
   *
   * @return the App's settings
   */
  public AppSettings settings() {
    return calls.object(Request.of("GET", "/settings"), AppSettings.class, "shop_add_on");
  }

  /**
   * Writes setting fields declared in the App manifest ({@code PUT /settings}). Each entry is sent
   * as {@code {"field": name, "data": value}}, in the map's order.
   *
   * @param values field name to value; values are encoded with the SDK's JSON configuration
   * @return the App's settings after the change
   */
  public AppSettings updateSettings(Map<String, ?> values) {
    Objects.requireNonNull(values, "values");
    List<Map<String, Object>> settings = new ArrayList<>();
    values.forEach(
        (field, data) -> {
          Map<String, Object> entry = new LinkedHashMap<>();
          entry.put("field", field);
          entry.put("data", data);
          settings.add(entry);
        });
    Request request = Params.body(Request.of("PUT", "/settings"), Map.of("settings", settings));
    return calls.object(request, AppSettings.class, "shop_add_on");
  }
}
