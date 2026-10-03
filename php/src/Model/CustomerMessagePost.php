<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One customer service thread (GET /v1/customers/{id}/message_posts). */
final class CustomerMessagePost
{
    /**
     * @param string $order the order the thread is about, as text
     * @param string $status "reply_yet" or "replied"
     * @param list<CustomerMessageComment> $comments
     */
    public function __construct(
        public readonly int $id,
        public readonly string $title,
        public readonly string $category,
        public readonly string $order,
        public readonly string $status,
        public readonly array $comments,
        public readonly ?\DateTimeImmutable $createdAt,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('title'),
            $f->stringOr('category'),
            $f->stringOr('order'),
            $f->stringOr('status'),
            $f->list('comments', CustomerMessageComment::fromFields(...)),
            $f->time('created_at'),
        );
    }
}
