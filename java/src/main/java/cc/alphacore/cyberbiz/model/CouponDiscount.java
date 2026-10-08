package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * One coupon applied to an order; id is only present in the coupon_discounts list.
 *
 * @param id the discount id, or null
 * @param name the coupon name
 * @param code the coupon code
 * @param amount the discount
 * @param couponId the coupon id, or null
 */
public record CouponDiscount(Long id, String name, String code, Money amount, Long couponId) {}
