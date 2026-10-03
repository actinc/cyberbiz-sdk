<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The outcome of a v2 batch home-delivery booking (POST /v2/orders/fulfillments/support_shippings). */
final class SupportShippingsResult
{
    /**
     * @param list<FailedShippingOrder> $failedOrders
     * @param list<ShippingFulfillment> $fulfillments
     */
    public function __construct(
        public readonly array $failedOrders,
        public readonly array $fulfillments,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->list('failed_orders', FailedShippingOrder::fromFields(...)),
            $f->list('fulfillments', ShippingFulfillment::fromFields(...)),
        );
    }
}
