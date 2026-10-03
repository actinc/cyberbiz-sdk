<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** A shop-wide campaign discount on an order. */
final class ShopDiscount
{
    public function __construct(
        public readonly string $name,
        public readonly Money $amount,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('name'),
            $f->moneyOr('amount'),
        );
    }
}
