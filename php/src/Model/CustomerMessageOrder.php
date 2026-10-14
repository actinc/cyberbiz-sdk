<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The order a customer service thread refers to. */
final class CustomerMessageOrder
{
    /**
     * @param string $type the kind of order the thread is linked to
     * @param string $name the order name shown to the customer
     */
    public function __construct(
        public readonly int $id,
        public readonly string $type,
        public readonly string $name,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('type'),
            $f->stringOr('name'),
        );
    }
}
