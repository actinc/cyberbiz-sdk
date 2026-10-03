<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** Every discount applied to an order. */
final class OrderDiscounts
{
    /**
     * @param Money $specialCollectionDiscount campaign discount
     * @param ShopDiscount|null $shopDiscount shop-wide campaign
     * @param CouponDiscount|null $couponDiscount single-coupon form, deprecated by CYBERBIZ in favour of couponDiscounts
     * @param list<CouponDiscount> $couponDiscounts
     * @param Money $priceDiscount manual discount by staff
     */
    public function __construct(
        public readonly Money $specialCollectionDiscount,
        public readonly Money $vipDiscount,
        public readonly ?ShopDiscount $shopDiscount,
        public readonly ?CouponDiscount $couponDiscount,
        public readonly array $couponDiscounts,
        public readonly Money $bonusConsumed,
        public readonly Money $vipShippingDiscount,
        public readonly Money $couponShippingDiscount,
        public readonly Money $priceDiscount,
        public readonly Money $thirdPartyDiscount,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->moneyOr('special_collection_discount'),
            $f->moneyOr('vip_discount'),
            Nested::object($f, 'shop_discount', ShopDiscount::fromFields(...)),
            Nested::object($f, 'coupon_discount', CouponDiscount::fromFields(...)),
            $f->list('coupon_discounts', CouponDiscount::fromFields(...)),
            $f->moneyOr('bonus_consumed'),
            $f->moneyOr('vip_shipping_discount'),
            $f->moneyOr('coupon_shipping_discount'),
            $f->moneyOr('price_discount'),
            $f->moneyOr('third_party_discount'),
        );
    }
}
