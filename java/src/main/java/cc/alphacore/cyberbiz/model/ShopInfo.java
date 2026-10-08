package cc.alphacore.cyberbiz.model;

/**
 * The Shop that owns the API token ({@code GET /shop}, the {@code shop_info} object).
 *
 * @param id the Shop id
 * @param name the Shop name
 * @param primaryDomain the primary domain, e.g. {@code example.cyberbiz.co}
 * @param email the contact e-mail
 * @param currency the currency code, e.g. {@code TWD}
 * @param language the storefront language, e.g. {@code zh-TW}
 * @param ogImageUrl the Open Graph image URL, often protocol-relative
 * @param smsPrefix the prefix put before SMS messages
 * @param merchantLocation where the merchant is located
 * @param marketLocation the market the Shop sells to
 * @param shopLine the LINE login settings, or null
 * @param shopLineChatBot the LINE chat bot, or null when none is connected
 */
public record ShopInfo(
    long id,
    String name,
    String primaryDomain,
    String email,
    String currency,
    String language,
    String ogImageUrl,
    String smsPrefix,
    String merchantLocation,
    String marketLocation,
    ShopLine shopLine,
    ShopLineChatBot shopLineChatBot) {}
