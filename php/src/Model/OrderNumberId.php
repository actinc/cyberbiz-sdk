<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** Maps a shop-facing order number to its API id (GET /v1/orders/get_order_id). */
final class OrderNumberId
{
    public function __construct(
        public readonly int $orderNumber,
        public readonly int $orderId,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('order_number'),
            $f->intOr('order_id'),
        );
    }
}
