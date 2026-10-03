<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The reply of GET /v1/customers/{id}/uid_providers/{provider_type}. */
final class CustomerUidLookup
{
    /**
     * @param string $message platform status text, e.g. "查詢成功"
     */
    public function __construct(
        public readonly int $customerId,
        public readonly string $uid,
        public readonly string $message,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('customer_id'),
            $f->stringOr('uid'),
            $f->stringOr('message'),
        );
    }
}
