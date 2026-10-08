package cc.alphacore.cyberbiz.model;

/**
 * An order a batch booking could not ship.
 *
 * @param orderId the order id, or null
 * @param message why
 */
public record FailedShippingOrder(Long orderId, String message) {}
