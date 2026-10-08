package cc.alphacore.cyberbiz.model;

/**
 * The booked carrier and tracking number.
 *
 * @param trackingCompany carrier code
 * @param trackingNumber tracking number
 */
public record SupportShippingTrackingNumber(String trackingCompany, String trackingNumber) {}
