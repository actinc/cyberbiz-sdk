package cc.alphacore.cyberbiz.model;

import java.util.List;

/**
 * The outcome for one order of a batch booking.
 *
 * @param orderId the order id, or null
 * @param fulfillmentStatus "unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived",
 *     "received", "returned", "expired", "problem" or "no_need"
 * @param trackingNumbers the order's fulfillments with carrier and number; empty when none
 * @param message error or warning
 * @param status the booking status
 * @param lineItems comma-separated line item ids
 */
public record SupportShippingTask(
    Long orderId,
    String fulfillmentStatus,
    List<SupportShippingTrackingNumber> trackingNumbers,
    String message,
    String status,
    String lineItems) {

  /** Copies the lists; a missing list becomes empty. */
  public SupportShippingTask {
    trackingNumbers = Lists.copy(trackingNumbers);
  }
}
