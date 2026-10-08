package cc.alphacore.cyberbiz.model;

/**
 * One hit of the name lookup (GET /v1/customers/get_customer_id_by_name).
 *
 * @param customerId the customer id
 * @param customerName the customer's name
 */
public record CustomerNameMatch(long customerId, String customerName) {}
