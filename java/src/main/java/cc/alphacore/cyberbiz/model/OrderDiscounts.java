package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.util.List;

/**
 * Every discount applied to an order; an amount is null when the response omits it.
 *
 * @param specialCollectionDiscount campaign discount
 * @param vipDiscount VIP discount
 * @param shopDiscount shop-wide campaign, or null
 * @param couponDiscount single-coupon form, deprecated by CYBERBIZ in favour of couponDiscounts; or
 *     null
 * @param couponDiscounts the applied coupons
 * @param bonusConsumed bonus points spent
 * @param vipShippingDiscount VIP shipping discount
 * @param couponShippingDiscount coupon shipping discount
 * @param priceDiscount manual discount by staff
 * @param thirdPartyDiscount discount by a third party
 */
public record OrderDiscounts(
    Money specialCollectionDiscount,
    Money vipDiscount,
    ShopDiscount shopDiscount,
    CouponDiscount couponDiscount,
    List<CouponDiscount> couponDiscounts,
    Money bonusConsumed,
    Money vipShippingDiscount,
    Money couponShippingDiscount,
    Money priceDiscount,
    Money thirdPartyDiscount) {

  /** Copies the lists; a missing list becomes empty. */
  public OrderDiscounts {
    couponDiscounts = Lists.copy(couponDiscounts);
  }
}
