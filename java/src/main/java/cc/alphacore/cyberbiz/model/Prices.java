package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * The price breakdown of an order.
 *
 * @param totalLineItemsPrice sum of the line items
 * @param shippingRatePrice shipping fee
 * @param discounts every discount, or null
 * @param totalPrice the amount to pay
 */
public record Prices(
    Money totalLineItemsPrice,
    Money shippingRatePrice,
    OrderDiscounts discounts,
    Money totalPrice) {}
