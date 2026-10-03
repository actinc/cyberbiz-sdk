<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The POS terminal and salesperson of a POS order. */
final class PosInfo
{
    public function __construct(
        public readonly int $posUserId,
        public readonly string $posUserEmail,
        public readonly int $posShopId,
        public readonly string $posInfo,
        public readonly int $posId,
        public readonly string $posName,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('pos_user_id'),
            $f->stringOr('pos_user_email'),
            $f->intOr('pos_shop_id'),
            $f->stringOr('pos_info'),
            $f->intOr('pos_id'),
            $f->stringOr('pos_name'),
        );
    }
}
