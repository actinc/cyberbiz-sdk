package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * A shop-wide campaign discount on an order.
 *
 * @param name the campaign name
 * @param amount the discount
 */
public record ShopDiscount(String name, Money amount) {}
