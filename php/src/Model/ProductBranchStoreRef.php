<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The branch store a product belongs to. */
final class ProductBranchStoreRef
{
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly string $storeNo,
        public readonly bool $enabled,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('name'),
            $f->stringOr('store_no'),
            $f->boolOr('enabled'),
        );
    }
}
