package cc.alphacore.cyberbiz.model;

/**
 * Links a customer to an external login identity.
 *
 * @param providerType "line", "line_at" or "facebook"
 * @param uid the external UID
 */
public record CustomerUidProvider(String providerType, String uid) {}
