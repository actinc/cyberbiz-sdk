<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/**
 * A customer's VIP membership state (GET /v1/customers/{id}/vip_info); group
 * and levels are null outside a VIP programme.
 */
final class CustomerVipInfo
{
    public function __construct(
        public readonly int $customerId,
        public readonly ?CustomerVipGroup $currentGroup,
        public readonly ?CustomerVipLevel $currentLevel,
        public readonly ?CustomerVipLevel $nextLevel,
        public readonly ?CustomerVipExtraInfo $extraInfo,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('customer_id'),
            Nested::object($f, 'current_group', CustomerVipGroup::fromFields(...)),
            Nested::object($f, 'current_level', CustomerVipLevel::fromFields(...)),
            Nested::object($f, 'next_level', CustomerVipLevel::fromFields(...)),
            Nested::object($f, 'extra_info', CustomerVipExtraInfo::fromFields(...)),
        );
    }
}
