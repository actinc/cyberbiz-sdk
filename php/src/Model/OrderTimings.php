<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** When each lifecycle event of an order happened; null means it has not. */
final class OrderTimings
{
    /**
     * @param \DateTimeImmutable|null $expiredAt CVS pickup deadline passed
     */
    public function __construct(
        public readonly ?\DateTimeImmutable $requestReturnAt,
        public readonly ?\DateTimeImmutable $returnAt,
        public readonly ?\DateTimeImmutable $refundAt,
        public readonly ?\DateTimeImmutable $closedAt,
        public readonly ?\DateTimeImmutable $cancelledAt,
        public readonly ?\DateTimeImmutable $expiredAt,
        public readonly ?\DateTimeImmutable $confirmedAt,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->time('request_return_at'),
            $f->time('return_at'),
            $f->time('refund_at'),
            $f->time('closed_at'),
            $f->time('cancelled_at'),
            $f->time('expired_at'),
            $f->time('confirmed_at'),
        );
    }
}
