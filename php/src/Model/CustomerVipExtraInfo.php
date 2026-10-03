<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/**
 * The validity window of the current VIP level and how far the customer is
 * from renewing or upgrading; differences are zero when there is no such
 * condition.
 */
final class CustomerVipExtraInfo
{
    public function __construct(
        public readonly ?\DateTimeImmutable $startAt,
        public readonly ?\DateTimeImmutable $endAt,
        public readonly Money $differenceOfTotalSpentInValidityDaysForRenewal,
        public readonly Money $differenceOfTotalSpentForRenewal,
        public readonly Money $differenceOfTotalSpentInValidityDaysForUpgrade,
        public readonly Money $differenceOfTotalSpentForUpgrade,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->time('start_at'),
            $f->time('end_at'),
            $f->moneyOr('difference_of_total_spent_in_validity_days_for_renewal'),
            $f->moneyOr('difference_of_total_spent_for_renewal'),
            $f->moneyOr('difference_of_total_spent_in_validity_days_for_upgrade'),
            $f->moneyOr('difference_of_total_spent_for_upgrade'),
        );
    }
}
