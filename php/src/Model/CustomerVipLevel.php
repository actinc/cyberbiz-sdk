<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** A VIP level as embedded in CustomerVipInfo. */
final class CustomerVipLevel
{
    /**
     * @param Money $upgradeConditionTotalSpent single-order spend that upgrades to this level
     * @param Money $upgradeConditionTotalSpentInValidityDays spend within the validity window that upgrades to this level
     */
    public function __construct(
        public readonly int $id,
        public readonly int $position,
        public readonly string $name,
        public readonly int $validityDays,
        public readonly int $upgradeValidityDays,
        public readonly Money $upgradeConditionTotalSpent,
        public readonly Money $upgradeConditionTotalSpentInValidityDays,
        public readonly Money $renewalConditionTotalSpent,
        public readonly Money $renewalConditionTotalSpentInValidityDays,
        public readonly bool $bonusPointEnabled,
        public readonly Money $bonusPointThreshold,
        public readonly Money $bonusPointValue,
        public readonly int $bonusPointExpiryDays,
        public readonly bool $birthGiftEnabled,
        public readonly string $birthGiftName,
        public readonly bool $upgradeGiftEnabled,
        public readonly bool $orderDiscountEnabled,
        public readonly bool $freeShippingEnabled,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->intOr('position'),
            $f->stringOr('name'),
            $f->intOr('validity_days'),
            $f->intOr('upgrade_validity_days'),
            $f->moneyOr('upgrade_condition_total_spent'),
            $f->moneyOr('upgrade_condition_total_spent_in_validity_days'),
            $f->moneyOr('renewal_condition_total_spent'),
            $f->moneyOr('renewal_condition_total_spent_in_validity_days'),
            $f->boolOr('bonus_point_enabled'),
            $f->moneyOr('bonus_point_threshold'),
            $f->moneyOr('bonus_point_value'),
            $f->intOr('bonus_point_expiry_days'),
            $f->boolOr('birth_gift_enabled'),
            $f->stringOr('birth_gift_name'),
            $f->boolOr('upgrade_gift_enabled'),
            $f->boolOr('order_discount_enabled'),
            $f->boolOr('free_shipping_enabled'),
        );
    }
}
