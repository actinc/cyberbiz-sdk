<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The reason a customer gave when cancelling an order. */
final class OrderCustomerCancelReason
{
    public function __construct(
        public readonly string $source,
        public readonly int $reasonId,
        public readonly string $reasonDetail,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('source'),
            $f->intOr('reason_id'),
            $f->stringOr('reason_detail'),
        );
    }
}
