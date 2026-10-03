<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A product photo. */
final class ProductPhoto
{
    public function __construct(
        public readonly int $id,
        public readonly string $url,
        public readonly int $position,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('url'),
            $f->intOr('position'),
        );
    }
}
