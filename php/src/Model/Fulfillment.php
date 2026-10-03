<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One shipment of an order. */
final class Fulfillment
{
    /**
     * @param string $trackingCompany carrier code such as "ezcat", "e_can", "postserv", "hct", "SEVEN", "FAMILY", "HILIFE", "SEVEN_C2C", "FAMILY_C2C", "SF", "cyberbiz_express" or "other" (see the Go SDK's TrackingCompany for the full list)
     * @param string $status "unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived", "received", "returned", "expired", "problem" or "no_need"
     * @param list<LineItem> $lineItems
     * @param string $trackingUrl Uber Direct and Pandago only
     * @param string $cvsShippingType "seven", "seven_c2c", "family", "family_c2c", "family_cold", "family_cold_c2c", "hilife", "hilife_cold", "hilife_c2c", "ezcat_cvs", "ezcat_cvs_cold" or "ezcat_cvs_refrigerate"; only set by the v2 CVS shipping endpoint
     */
    public function __construct(
        public readonly int $id,
        public readonly string $trackingCompany,
        public readonly string $trackingNumber,
        public readonly ?\DateTimeImmutable $fulfilledAt,
        public readonly ?\DateTimeImmutable $receivedAt,
        public readonly string $status,
        public readonly array $lineItems,
        public readonly string $trackingUrl,
        public readonly string $cvsShippingType,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('tracking_company'),
            $f->stringOr('tracking_number'),
            $f->time('fulfilled_at'),
            $f->time('received_at'),
            $f->stringOr('status'),
            $f->list('line_items', LineItem::fromFields(...)),
            $f->stringOr('tracking_url'),
            $f->stringOr('cvs_shipping_type'),
        );
    }
}
