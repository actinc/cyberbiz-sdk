package cc.alphacore.cyberbiz.model;

import java.util.List;

/**
 * The outcome of a v1 batch home-delivery booking (POST /v1/orders/fulfillments/support_shipping).
 *
 * @param requestId the booking request id
 * @param results one result per order
 */
public record SupportShippingBatchResult(String requestId, List<SupportShippingTask> results) {

  /** Copies the lists; a missing list becomes empty. */
  public SupportShippingBatchResult {
    results = Lists.copy(results);
  }
}
