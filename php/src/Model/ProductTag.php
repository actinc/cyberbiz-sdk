<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A product tag. */
final class ProductTag
{
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly int $category,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('name'),
            $f->intOr('category'),
        );
    }
}
