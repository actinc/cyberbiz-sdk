<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The sales channel of a product. */
final class ProductChannel
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
