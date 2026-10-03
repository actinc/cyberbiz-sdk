<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One payment recorded against an order. */
final class OrderTransaction
{
    /**
     * @param string $kindName e.g. 已收款
     * @param string $paidTypeName e.g. 手動
     */
    public function __construct(
        public readonly int $id,
        public readonly Money $amount,
        public readonly string $kindName,
        public readonly string $paidTypeName,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->moneyOr('amount'),
            $f->stringOr('kind_name'),
            $f->stringOr('paid_type_name'),
        );
    }
}
