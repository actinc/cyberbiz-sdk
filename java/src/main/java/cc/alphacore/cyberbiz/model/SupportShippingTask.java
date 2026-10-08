package cc.alphacore.cyberbiz.model;

/**
 * The outcome for one order of a batch booking.
 *
 * @param orderId the order id, or null
 * @param fulfillmentStatus "unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived",
 *     "received", "returned", "expired", "problem" or "no_need"
 * @param trackingNumbers the booked carrier and number, or null
 * @param message error or warning
 * @param status the booking status
 * @param lineItems comma-separated line item ids
 */
public record SupportShippingTask(
    Long orderId,
    String fulfillmentStatus,
    SupportShippingTrackingNumber trackingNumbers,
    String message,
    String status,
    String lineItems) {}
