<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One message in a customer service thread. */
final class CustomerMessageComment
{
    /**
     * @param string $role who replied
     * @param string $admin admin details when role is admin
     */
    public function __construct(
        public readonly int $id,
        public readonly string $role,
        public readonly string $admin,
        public readonly string $content,
        public readonly ?\DateTimeImmutable $createdAt,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('role'),
            $f->stringOr('admin'),
            $f->stringOr('content'),
            $f->time('created_at'),
        );
    }
}
