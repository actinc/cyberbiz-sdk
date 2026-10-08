package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * One payment of a split (multi-method) POS payment.
 *
 * @param name the payment method
 * @param amount the amount paid
 */
public record OrderPaymentInfo(String name, Money amount) {}
