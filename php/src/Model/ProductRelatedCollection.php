<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A collection related to a product. */
final class ProductRelatedCollection
{
    public function __construct(
        public readonly int $id,
        public readonly int $relatableCollectionId,
        public readonly string $relatableCollectionType,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->intOr('relatable_collection_id'),
            $f->stringOr('relatable_collection_type'),
        );
    }
}
