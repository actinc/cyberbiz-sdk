<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A product's SEO meta tags. */
final class ProductSeoMetaTags
{
    public function __construct(
        public readonly string $title,
        public readonly string $description,
        public readonly string $keywords,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('title'),
            $f->stringOr('description'),
            $f->stringOr('keywords'),
        );
    }
}
