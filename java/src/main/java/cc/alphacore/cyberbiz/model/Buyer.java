package cc.alphacore.cyberbiz.model;

/**
 * The member account that placed an order.
 *
 * @param email email
 * @param mobile mobile number
 */
public record Buyer(String email, String mobile) {}
