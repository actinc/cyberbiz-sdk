package cc.alphacore.cyberbiz.model;

import java.util.List;

/**
 * The outcome of a v2 batch home-delivery booking (POST /v2/orders/fulfillments/support_shippings).
 *
 * @param failedOrders orders that could not be booked
 * @param fulfillments the created fulfillments
 */
public record SupportShippingsResult(
    List<FailedShippingOrder> failedOrders, List<ShippingFulfillment> fulfillments) {

  /** Copies the lists; a missing list becomes empty. */
  public SupportShippingsResult {
    failedOrders = Lists.copy(failedOrders);
    fulfillments = Lists.copy(fulfillments);
  }
}
