<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The carrier assigned to an order. */
final class ShippingVendor
{
    /**
     * @param string $type carrier code, e.g. "custom", "ezcat"
     */
    public function __construct(
        public readonly string $type,
        public readonly string $name,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('type'),
            $f->stringOr('name'),
        );
    }
}
