<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One return shipment of an order (GET /v1/orders/{id}/returns). */
final class OrderReturn
{
    /**
     * @param string $trackingCompany a carrier code, as in Fulfillment
     * @param list<LineItem> $lineItems
     * @param string $returnSuda5 return address zip code
     */
    public function __construct(
        public readonly int $id,
        public readonly ?\DateTimeImmutable $createdAt,
        public readonly string $trackingNumber,
        public readonly string $trackingCompany,
        public readonly array $lineItems,
        public readonly string $returnAddress,
        public readonly string $returnSuda5,
        public readonly string $returnReason,
        public readonly string $returnInfo,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->time('created_at'),
            $f->stringOr('tracking_number'),
            $f->stringOr('tracking_company'),
            $f->list('line_items', LineItem::fromFields(...)),
            $f->stringOr('return_address'),
            $f->stringOr('return_suda5'),
            $f->stringOr('return_reason'),
            $f->stringOr('return_info'),
        );
    }
}
