<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One component variant inside a combo product. */
final class RelatedItem
{
    /**
     * @param Money $comboProductPriceDifference enterprise-only difference between the component price and its share of the combo price
     * @param list<Money> $comboProductPriceDiffDetails
     */
    public function __construct(
        public readonly int $id,
        public readonly int $productId,
        public readonly int $productVariantId,
        public readonly string $title,
        public readonly string $variantTitle,
        public readonly string $sku,
        public readonly string $qc,
        public readonly string $vendor,
        public readonly Money $price,
        public readonly Money $cost,
        public readonly int $quantity,
        public readonly Money $comboProductPriceDifference,
        public readonly array $comboProductPriceDiffDetails,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->intOr('product_id'),
            $f->intOr('product_variant_id'),
            $f->stringOr('title'),
            $f->stringOr('variant_title'),
            $f->stringOr('sku'),
            $f->stringOr('qc'),
            $f->stringOr('vendor'),
            $f->moneyOr('price'),
            $f->moneyOr('cost'),
            $f->intOr('quantity'),
            $f->moneyOr('combo_product_price_difference'),
            Nested::moneys($f, 'combo_product_price_diff_details'),
        );
    }
}
