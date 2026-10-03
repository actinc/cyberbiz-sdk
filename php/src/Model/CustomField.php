<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A shop-defined field value on a customer or line item. id is 0 where the response omits it. */
final class CustomField
{
    /**
     * @param string $name field key
     * @param string $label display name
     */
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly string $label,
        public readonly string $value,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('name'),
            $f->stringOr('label'),
            $f->stringOr('value'),
        );
    }
}
