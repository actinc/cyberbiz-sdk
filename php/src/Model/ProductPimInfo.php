<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A product's link to the PIM (product information management) system. */
final class ProductPimInfo
{
    public function __construct(
        public readonly int $id,
        public readonly int $productId,
        public readonly int $productVariantId,
        public readonly int $pimProductId,
        public readonly int $pimVariantId,
        public readonly int $channel,
        public readonly string $channelShopName,
        public readonly bool $isConnected,
        public readonly bool $isSource,
        public readonly int $shopId,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->intOr('product_id'),
            $f->intOr('product_variant_id'),
            $f->intOr('pim_product_id'),
            $f->intOr('pim_variant_id'),
            $f->intOr('channel'),
            $f->stringOr('channel_shop_name'),
            $f->boolOr('is_connected'),
            $f->boolOr('is_source'),
            $f->intOr('shop_id'),
        );
    }
}
