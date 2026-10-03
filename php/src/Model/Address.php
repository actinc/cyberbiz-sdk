<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/**
 * A postal address with contact details: an order's billing address or a
 * customer's default address. name is only set on the former, company only
 * on the latter.
 */
final class Address
{
    /**
     * @param string $address single-line full address
     */
    public function __construct(
        public readonly string $name,
        public readonly string $company,
        public readonly string $countryCallingCode,
        public readonly string $phone,
        public readonly string $address,
        public readonly ?DetailAddress $detailAddress,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('name'),
            $f->stringOr('company'),
            $f->stringOr('country_calling_code'),
            $f->stringOr('phone'),
            $f->stringOr('address'),
            Nested::object($f, 'detail_address', DetailAddress::fromFields(...)),
        );
    }
}
