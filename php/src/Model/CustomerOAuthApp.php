<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The credential pair of a registered customer-login OAuth application (POST /v2/customer_oauth). */
final class CustomerOAuthApp
{
    public function __construct(
        public readonly string $clientId,
        public readonly string $clientSecret,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('client_id'),
            $f->stringOr('client_secret'),
        );
    }
}
