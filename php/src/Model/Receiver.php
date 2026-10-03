<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The shipping recipient of an order. */
final class Receiver
{
    /**
     * @param string $cvsStoreId convenience store number for pickup
     * @param string $allpayLogisticsId ECPay logistics id
     */
    public function __construct(
        public readonly string $name,
        public readonly string $countryCallingCode,
        public readonly string $phone,
        public readonly string $address,
        public readonly ?DetailAddress $detailAddress,
        public readonly string $cvsStoreId,
        public readonly string $allpayLogisticsId,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('name'),
            $f->stringOr('country_calling_code'),
            $f->stringOr('phone'),
            $f->stringOr('address'),
            Nested::object($f, 'detail_address', DetailAddress::fromFields(...)),
            $f->stringOr('cvs_store_id'),
            $f->stringOr('allpay_logistics_id'),
        );
    }
}
