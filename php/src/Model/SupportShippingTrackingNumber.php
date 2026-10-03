<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The booked carrier and tracking number. */
final class SupportShippingTrackingNumber
{
    public function __construct(
        public readonly string $trackingCompany,
        public readonly string $trackingNumber,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('tracking_company'),
            $f->stringOr('tracking_number'),
        );
    }
}
