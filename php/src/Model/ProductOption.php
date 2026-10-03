<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A product option such as size or colour. */
final class ProductOption
{
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly int $position,
        public readonly string $types,
    ) {}

    /** @param int|null $id overrides the body's id (detail responses may omit it) */
    public static function fromFields(Fields $f, ?int $id = null): self
    {
        return new self(
            $id ?? $f->intOr('id'),
            $f->stringOr('name'),
            $f->intOr('position'),
            $f->stringOr('types'),
        );
    }
}
