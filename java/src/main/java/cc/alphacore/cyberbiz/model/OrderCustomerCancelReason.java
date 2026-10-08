package cc.alphacore.cyberbiz.model;

/**
 * The reason a customer gave when cancelling an order.
 *
 * @param source where the order was cancelled
 * @param reasonId the reason id, or null
 * @param reasonDetail the reason text
 */
public record OrderCustomerCancelReason(String source, Long reasonId, String reasonDetail) {}
