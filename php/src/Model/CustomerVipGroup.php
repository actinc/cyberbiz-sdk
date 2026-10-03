<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The VIP group a customer belongs to, as embedded in CustomerVipInfo. */
final class CustomerVipGroup
{
    /**
     * @param list<string> $customerTags
     */
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly int $position,
        public readonly string $descriptionUrl,
        public readonly array $customerTags,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->stringOr('name'),
            $f->intOr('position'),
            $f->stringOr('description_url'),
            $f->strings('customer_tags'),
        );
    }
}
