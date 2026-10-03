<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One redemption code of a split e-ticket. */
final class EticketTransaction
{
    public function __construct(
        public readonly string $ticketNumber,
        public readonly ?\DateTimeImmutable $usedAt,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('ticket_number'),
            $f->time('used_at'),
        );
    }
}
