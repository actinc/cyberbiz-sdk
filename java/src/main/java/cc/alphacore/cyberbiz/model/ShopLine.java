package cc.alphacore.cyberbiz.model;

/**
 * A Shop's LINE login and LIFF settings.
 *
 * @param loginEnable whether customers can log in with LINE
 * @param liffId the LIFF app id, or null
 * @param liffEnable whether the LIFF app is enabled
 */
public record ShopLine(boolean loginEnable, String liffId, boolean liffEnable) {}
