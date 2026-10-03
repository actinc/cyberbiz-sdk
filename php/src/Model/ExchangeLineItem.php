<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One variant exchanged in an ExchangeHistory. */
final class ExchangeLineItem
{
    public function __construct(
        public readonly int $productVariantId,
        public readonly string $name,
        public readonly string $sku,
        public readonly string $qc,
        public readonly Money $price,
        public readonly int $quantity,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('product_variant_id'),
            $f->stringOr('name'),
            $f->stringOr('sku'),
            $f->stringOr('qc'),
            $f->moneyOr('price'),
            $f->intOr('quantity'),
        );
    }
}
