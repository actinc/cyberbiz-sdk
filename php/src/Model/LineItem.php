<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One product variant on an order, fulfillment or return, or in a customer's recent purchases. */
final class LineItem
{
    /**
     * @param string $qc vendor's own item code
     * @param Money $price unit price
     * @param Money $cost unit cost
     * @param string $itemType "normal" or "no"
     * @param string $returnStatus "no_need", "request_return", "returning", "checking", "returned", "in_hub", "problem", "processing", "in_origin_cvs", "refused" or "partial_return"
     * @param list<LineItemDiscount> $discounts
     * @param string $taxTypeId "inclusive_tax", "zero_tax" or "exclusive_tax"
     * @param Money $bonusRedemptionPrice bonus points redeemed for this item; zero when not a bonus mall redemption
     * @param list<RelatedItems> $relatedItems components of a combo product
     * @param string $photo thumbnail path
     * @param list<CustomField> $customFields
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
        public readonly string $itemType,
        public readonly string $returnStatus,
        public readonly string $discountName,
        public readonly array $discounts,
        public readonly Money $totalPriceBeforeDiscounts,
        public readonly Money $totalDiscount,
        public readonly Money $totalPriceAfterDiscounts,
        public readonly string $taxTypeId,
        public readonly Money $bonusRedemptionPrice,
        public readonly array $relatedItems,
        public readonly ?\DateTimeImmutable $createdAt,
        public readonly string $channel,
        public readonly float $weight,
        public readonly string $photo,
        public readonly array $customFields,
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
            $f->stringOr('item_type'),
            $f->stringOr('return_status'),
            $f->stringOr('discount_name'),
            $f->list('discounts', LineItemDiscount::fromFields(...)),
            $f->moneyOr('total_price_before_discounts'),
            $f->moneyOr('total_discount'),
            $f->moneyOr('total_price_after_discounts'),
            $f->stringOr('tax_type_id'),
            $f->moneyOr('bonus_redemption_price'),
            $f->list('related_items', RelatedItems::fromFields(...)),
            $f->time('created_at'),
            $f->stringOr('channel'),
            $f->float('weight'),
            $f->stringOr('photo'),
            $f->list('custom_fields', CustomField::fromFields(...)),
        );
    }
}
