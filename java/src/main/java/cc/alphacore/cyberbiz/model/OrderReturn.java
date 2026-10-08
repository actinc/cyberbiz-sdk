package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;
import java.util.List;

/**
 * One return shipment of an order (GET /v1/orders/{id}/returns).
 *
 * @param id the return id, or null
 * @param createdAt when it was created
 * @param trackingNumber tracking number
 * @param trackingCompany a carrier code, as in Fulfillment
 * @param lineItems the returned line items
 * @param returnAddress return address
 * @param returnSuda5 return address zip code
 * @param returnReason reason
 * @param returnInfo extra information
 */
public record OrderReturn(
    Long id,
    OffsetDateTime createdAt,
    String trackingNumber,
    String trackingCompany,
    List<LineItem> lineItems,
    String returnAddress,
    String returnSuda5,
    String returnReason,
    String returnInfo) {

  /** Copies the lists; a missing list becomes empty. */
  public OrderReturn {
    lineItems = Lists.copy(lineItems);
  }
}
