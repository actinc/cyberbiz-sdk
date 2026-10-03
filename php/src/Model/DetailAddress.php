<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** An address split into its components. */
final class DetailAddress
{
    public function __construct(
        public readonly string $zip,
        public readonly string $country,
        public readonly string $province,
        public readonly string $city,
        public readonly string $district,
        public readonly string $address1,
        public readonly string $address2,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('zip'),
            $f->stringOr('country'),
            $f->stringOr('province'),
            $f->stringOr('city'),
            $f->stringOr('district'),
            $f->stringOr('address1'),
            $f->stringOr('address2'),
        );
    }
}
