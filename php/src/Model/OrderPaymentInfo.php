<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One payment of a split (multi-method) POS payment. */
final class OrderPaymentInfo
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
