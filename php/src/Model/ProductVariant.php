<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One purchasable variant of a product (size, colour, ...). */
final class ProductVariant
{
    /**
     * @param list<string>         $photoUrls
     * @param list<ProductPimInfo> $pimInfos
     */
    public function __construct(
        public readonly int $id,
        public readonly int $productId,
        public readonly string $name,
        public readonly int $position,
        public readonly Money $price,
        public readonly Money $cost,
        public readonly Money $compareAtPrice,
        public readonly float $meas,
        public readonly Money $maxUsableBonus,
        public readonly float $weight,
        public readonly string $option1,
        public readonly string $option2,
        public readonly string $option3,
        public readonly bool $inventoryManagement,
        public readonly int $inventoryQuantity,
        public readonly int $sold,
        public readonly int $safetyInventoryQuantity,
        public readonly string $inventoryPolicy,
        public readonly string $sku,
        public readonly string $qc,
        public readonly bool $requiresShipping,
        public readonly ?\DateTimeImmutable $createdAt,
        public readonly ?\DateTimeImmutable $updatedAt,
        public readonly bool $honeycombSync,
        public readonly string $vendor,
        public readonly array $photoUrls,
        public readonly array $pimInfos,
    ) {}

    /** @param int|null $id overrides the body's id (detail responses may omit it) */
    public static function fromFields(Fields $f, ?int $id = null): self
    {
        return new self(
            $id ?? $f->intOr('id'),
            $f->intOr('product_id'),
            $f->stringOr('name'),
            $f->intOr('position'),
            $f->moneyOr('price'),
            $f->moneyOr('cost'),
            $f->moneyOr('compare_at_price'),
            $f->float('meas'),
            $f->moneyOr('max_usable_bonus'),
            $f->float('weight'),
            $f->stringOr('option1'),
            $f->stringOr('option2'),
            $f->stringOr('option3'),
            $f->boolOr('inventory_management'),
            $f->intOr('inventory_quantity'),
            $f->intOr('sold'),
            $f->intOr('safety_inventory_quantity'),
            $f->stringOr('inventory_policy'),
            $f->stringOr('sku'),
            $f->stringOr('qc'),
            $f->boolOr('requires_shipping'),
            $f->time('created_at'),
            $f->time('updated_at'),
            $f->boolOr('honeycomb_sync'),
            $f->stringOr('vendor'),
            $f->strings('photo_urls'),
            $f->list('pim_infos', ProductPimInfo::fromFields(...)),
        );
    }
}
