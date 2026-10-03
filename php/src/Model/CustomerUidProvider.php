<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** Links a customer to an external login identity. */
final class CustomerUidProvider
{
    /**
     * @param string $providerType "line", "line_at" or "facebook"
     */
    public function __construct(
        public readonly string $providerType,
        public readonly string $uid,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('provider_type'),
            $f->stringOr('uid'),
        );
    }
}
