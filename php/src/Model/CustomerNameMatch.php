<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One hit of the name lookup (GET /v1/customers/get_customer_id_by_name). */
final class CustomerNameMatch
{
    public function __construct(
        public readonly int $customerId,
        public readonly string $customerName,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('customer_id'),
            $f->stringOr('customer_name'),
        );
    }
}
