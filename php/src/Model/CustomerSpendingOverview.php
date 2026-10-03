<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** A customer's paid and valid orders in a date range. */
final class CustomerSpendingOverview
{
    public function __construct(
        public readonly Money $paidAndValidTotalSpent,
        public readonly int $paidAndValidOrdersCount,
        public readonly Money $paidAndValidAverageSpent,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->moneyOr('paid_and_valid_total_spent'),
            $f->intOr('paid_and_valid_orders_count'),
            $f->moneyOr('paid_and_valid_average_spent'),
        );
    }
}
