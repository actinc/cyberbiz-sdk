<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The branch store attached to an order for pickup or express delivery. */
final class OrderBranchStore
{
    /**
     * @param string $sourceType "BranchStore" or "PosShop"
     */
    public function __construct(
        public readonly string $storeNo,
        public readonly string $name,
        public readonly string $phone,
        public readonly string $county,
        public readonly string $district,
        public readonly string $address,
        public readonly string $zip,
        public readonly string $openingHours,
        public readonly float $lat,
        public readonly float $lng,
        public readonly bool $enabled,
        public readonly string $sourceType,
        public readonly int $sourceId,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('store_no'),
            $f->stringOr('name'),
            $f->stringOr('phone'),
            $f->stringOr('county'),
            $f->stringOr('district'),
            $f->stringOr('address'),
            $f->stringOr('zip'),
            $f->stringOr('opening_hours'),
            $f->float('lat'),
            $f->float('lng'),
            $f->boolOr('enabled'),
            $f->stringOr('source_type'),
            $f->intOr('source_id'),
        );
    }
}
