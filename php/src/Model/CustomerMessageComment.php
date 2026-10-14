<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One message in a customer service thread. */
final class CustomerMessageComment
{
    /**
     * @param string $role who replied
     * @param CustomerMessageAdmin|null $admin the replying admin when role is admin, else null
     */
    public function __construct(
        public readonly int $id,
        public readonly string $role,
        public readonly ?CustomerMessageAdmin $admin,
        public readonly string $content,
        public readonly ?\DateTimeImmutable $createdAt,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('role'),
            Nested::object($f, 'admin', CustomerMessageAdmin::fromFields(...)),
            $f->stringOr('content'),
            $f->time('created_at'),
        );
    }
}
