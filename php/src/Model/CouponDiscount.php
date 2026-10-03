<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One coupon applied to an order; id is only present in the coupon_discounts list. */
final class CouponDiscount
{
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly string $code,
        public readonly Money $amount,
        public readonly int $couponId,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('name'),
            $f->stringOr('code'),
            $f->moneyOr('amount'),
            $f->intOr('coupon_id'),
        );
    }
}
