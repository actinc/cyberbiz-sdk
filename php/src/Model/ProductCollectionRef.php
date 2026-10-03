<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A custom collection a product belongs to. */
final class ProductCollectionRef
{
    public function __construct(
        public readonly int $id,
        public readonly string $title,
        public readonly string $handle,
        public readonly bool $published,
        public readonly string $bodyHtml,
        public readonly string $productsOrderName,
        public readonly int $position,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('title'),
            $f->stringOr('handle'),
            $f->boolOr('published'),
            $f->stringOr('body_html'),
            $f->stringOr('products_order_name'),
            $f->intOr('position'),
        );
    }
}
