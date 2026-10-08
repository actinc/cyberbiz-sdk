package cc.alphacore.cyberbiz.model;

/**
 * The App version installed on a Shop.
 *
 * @param webhookUrl the URL webhooks are sent to
 * @param manifest the App manifest, or null
 * @param embedded whether the App is embedded in the Shop admin
 * @param status the installation status, e.g. {@code init}
 * @param scopes the granted scopes, space-separated
 */
public record AddOnVersion(
    String webhookUrl, AppManifest manifest, boolean embedded, String status, String scopes) {}
