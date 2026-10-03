<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One set of components of a combo product line item. */
final class RelatedItems
{
    /**
     * @param int $quantity number of combo sets
     * @param list<RelatedItem> $items
     */
    public function __construct(
        public readonly int $quantity,
        public readonly array $items,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('quantity'),
            $f->list('items', RelatedItem::fromFields(...)),
        );
    }
}
