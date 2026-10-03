<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A fulfillment created by a batch booking; trackingNumber is empty until the carrier assigns one. */
final class ShippingFulfillment
{
    public function __construct(
        public readonly int $id,
        public readonly int $orderId,
        public readonly string $trackingNumber,
        public readonly string $trackingCompany,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->intOr('order_id'),
            $f->stringOr('tracking_number'),
            $f->stringOr('tracking_company'),
        );
    }
}
