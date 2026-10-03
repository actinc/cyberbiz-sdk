<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** An electronic ticket sold on an order. */
final class OrderEticket
{
    /**
     * @param string $ticketNumber redemption code
     * @param bool $separate split into one code per unit
     * @param list<EticketTransaction> $transactions
     */
    public function __construct(
        public readonly string $title,
        public readonly string $ticketNumber,
        public readonly int $availableQuantity,
        public readonly int $usedQuantity,
        public readonly bool $enabled,
        public readonly int $orderId,
        public readonly int $productId,
        public readonly bool $separate,
        public readonly array $transactions,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('title'),
            $f->stringOr('ticket_number'),
            $f->intOr('available_quantity'),
            $f->intOr('used_quantity'),
            $f->boolOr('enabled'),
            $f->intOr('order_id'),
            $f->intOr('product_id'),
            $f->boolOr('separate'),
            $f->list('transactions', EticketTransaction::fromFields(...)),
        );
    }
}
