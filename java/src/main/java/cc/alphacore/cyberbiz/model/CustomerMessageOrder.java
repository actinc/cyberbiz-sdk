package cc.alphacore.cyberbiz.model;

/**
 * The order a customer service thread refers to.
 *
 * @param id the order id, or null
 * @param type the kind of order the thread is linked to
 * @param name the order name shown to the customer, e.g. "#1001"
 */
public record CustomerMessageOrder(Long id, String type, String name) {}
