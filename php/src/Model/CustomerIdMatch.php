<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** One hit of the email / mobile lookup (GET /v1/customers/get_customer_id). */
final class CustomerIdMatch
{
    public function __construct(
        public readonly int $customerId,
        public readonly string $customerEmail,
        public readonly string $customerMobile,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('customer_id'),
            $f->stringOr('customer_email'),
            $f->stringOr('customer_mobile'),
        );
    }
}
