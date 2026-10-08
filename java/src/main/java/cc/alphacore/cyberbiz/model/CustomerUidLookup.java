package cc.alphacore.cyberbiz.model;

/**
 * The reply of GET /v1/customers/{id}/uid_providers/{provider_type}.
 *
 * @param customerId the customer id
 * @param uid the external UID
 * @param message platform status text, e.g. "查詢成功"
 */
public record CustomerUidLookup(long customerId, String uid, String message) {}
