<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The member account that placed an order. */
final class Buyer
{
    public function __construct(
        public readonly string $email,
        public readonly string $mobile,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('email'),
            $f->stringOr('mobile'),
        );
    }
}
