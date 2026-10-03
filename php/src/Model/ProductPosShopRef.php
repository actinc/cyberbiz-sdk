<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The POS shop a product belongs to. */
final class ProductPosShopRef
{
    public function __construct(
        public readonly int $id,
        public readonly string $name,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('name'),
        );
    }
}
