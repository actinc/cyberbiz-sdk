<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** An order a batch booking could not ship. */
final class FailedShippingOrder
{
    public function __construct(
        public readonly int $orderId,
        public readonly string $message,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('order_id'),
            $f->stringOr('message'),
        );
    }
}
