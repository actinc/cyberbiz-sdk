package cc.alphacore.cyberbiz.model;

/**
 * One hit of the email / mobile lookup (GET /v1/customers/get_customer_id).
 *
 * @param customerId the customer id, or null
 * @param customerEmail the matched email
 * @param customerMobile the matched mobile
 */
public record CustomerIdMatch(Long customerId, String customerEmail, String customerMobile) {}
