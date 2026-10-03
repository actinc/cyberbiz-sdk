<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** The price breakdown of an order. */
final class Prices
{
    public function __construct(
        public readonly Money $totalLineItemsPrice,
        public readonly Money $shippingRatePrice,
        public readonly ?OrderDiscounts $discounts,
        public readonly Money $totalPrice,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->moneyOr('total_line_items_price'),
            $f->moneyOr('shipping_rate_price'),
            Nested::object($f, 'discounts', OrderDiscounts::fromFields(...)),
            $f->moneyOr('total_price'),
        );
    }
}
