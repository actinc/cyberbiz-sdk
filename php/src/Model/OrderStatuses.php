<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The four state machines of an order. */
final class OrderStatuses
{
    /**
     * @param string $orderStatus "open", "closed" or "cancelled"
     * @param string $financialStatus "paid", "pending", "cod", "failed", "abandoned", "refunded", "no_refunded", "pending_refund", "processing", "remitted", "pending_partial_refund", "partial_refunded", "refunding" or "refund_failed"
     * @param string $fulfillmentStatus "unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived", "received", "returned", "expired", "problem" or "no_need"
     * @param string $returnStatus "no_need", "request_return", "returning", "checking", "returned", "in_hub", "problem", "processing", "in_origin_cvs", "refused" or "partial_return"
     */
    public function __construct(
        public readonly string $orderStatus,
        public readonly string $financialStatus,
        public readonly string $fulfillmentStatus,
        public readonly string $returnStatus,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('order_status'),
            $f->stringOr('financial_status'),
            $f->stringOr('fulfillment_status'),
            $f->stringOr('return_status'),
        );
    }
}
