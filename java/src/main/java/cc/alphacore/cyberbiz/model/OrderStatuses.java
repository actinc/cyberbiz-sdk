package cc.alphacore.cyberbiz.model;

/**
 * The four state machines of an order.
 *
 * @param orderStatus "open", "closed" or "cancelled"
 * @param financialStatus "paid", "pending", "cod", "failed", "abandoned", "refunded",
 *     "no_refunded", "pending_refund", "processing", "remitted", "pending_partial_refund",
 *     "partial_refunded", "refunding" or "refund_failed"
 * @param fulfillmentStatus "unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived",
 *     "received", "returned", "expired", "problem" or "no_need"
 * @param returnStatus "no_need", "request_return", "returning", "checking", "returned", "in_hub",
 *     "problem", "processing", "in_origin_cvs", "refused" or "partial_return"
 */
public record OrderStatuses(
    String orderStatus, String financialStatus, String fulfillmentStatus, String returnStatus) {}
