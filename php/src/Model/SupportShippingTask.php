<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The outcome for one order of a batch booking. */
final class SupportShippingTask
{
    /**
     * @param string $fulfillmentStatus "unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived", "received", "returned", "expired", "problem" or "no_need"
     * @param string $message error or warning
     * @param list<SupportShippingTrackingNumber> $trackingNumbers the order's fulfillments with carrier and number
     * @param string $lineItems comma-separated line item ids
     */
    public function __construct(
        public readonly int $orderId,
        public readonly string $fulfillmentStatus,
        public readonly array $trackingNumbers,
        public readonly string $message,
        public readonly string $status,
        public readonly string $lineItems,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('order_id'),
            $f->stringOr('fulfillment_status'),
            $f->list('tracking_numbers', SupportShippingTrackingNumber::fromFields(...)),
            $f->stringOr('message'),
            $f->stringOr('status'),
            $f->stringOr('line_items'),
        );
    }
}
