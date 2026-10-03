<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One refund made against an order. */
final class ReturnHistory
{
    /**
     * @param string $body short description
     */
    public function __construct(
        public readonly string $body,
        public readonly Money $price,
        public readonly ?\DateTimeImmutable $refundedAt,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('body'),
            $f->moneyOr('price'),
            $f->time('refunded_at'),
        );
    }
}
