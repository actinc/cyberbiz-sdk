package cc.alphacore.cyberbiz.model;

/**
 * A fulfillment created by a batch booking; trackingNumber is empty until the carrier assigns one.
 *
 * @param id the fulfillment id, or null
 * @param orderId the order id, or null
 * @param trackingNumber tracking number
 * @param trackingCompany carrier code
 */
public record ShippingFulfillment(
    Long id, Long orderId, String trackingNumber, String trackingCompany) {}
