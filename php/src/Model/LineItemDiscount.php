<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One discount applied to a line item. */
final class LineItemDiscount
{
    /**
     * @param int $position which unit of the line the discount hit
     * @param int $id discount type id
     * @param string $code e.g. "bundle_discount"
     */
    public function __construct(
        public readonly int $position,
        public readonly int $id,
        public readonly string $code,
        public readonly string $name,
        public readonly Money $discount,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('position'),
            $f->intOr('id'),
            $f->stringOr('code'),
            $f->stringOr('name'),
            $f->moneyOr('discount'),
        );
    }
}
