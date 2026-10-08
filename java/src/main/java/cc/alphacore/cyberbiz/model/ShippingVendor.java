package cc.alphacore.cyberbiz.model;

/**
 * The carrier assigned to an order.
 *
 * @param type carrier code, e.g. "custom", "ezcat"
 * @param name carrier name
 */
public record ShippingVendor(String type, String name) {}
