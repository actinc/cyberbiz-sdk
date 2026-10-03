<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The electronic invoice attached to an order or exchange. */
final class OrderEinvoice
{
    /**
     * @param string $title invoice title (buyer name)
     * @param string $companyNo buyer's tax id
     * @param string $invoiceStatus "issue", "issue_invalid", "allowance" or "partial_allowance"
     * @param \DateTimeImmutable|null $invalidAt voided or allowance time; null while valid
     * @param string $invoiceType "default", "company", "phone_barcode", "nature_person", "donate" or "paper_invoice"
     * @param string $loveCode donation code
     * @param string $naturePerson citizen digital certificate
     */
    public function __construct(
        public readonly string $title,
        public readonly int $orderId,
        public readonly string $companyNo,
        public readonly string $invoiceNo,
        public readonly string $invoiceStatus,
        public readonly ?\DateTimeImmutable $invoiceAt,
        public readonly ?\DateTimeImmutable $invalidAt,
        public readonly string $randomNum,
        public readonly string $invoiceType,
        public readonly string $loveCode,
        public readonly string $phoneBarcode,
        public readonly string $naturePerson,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('title'),
            $f->intOr('order_id'),
            $f->stringOr('company_no'),
            $f->stringOr('invoice_no'),
            $f->stringOr('invoice_status'),
            $f->time('invoice_at'),
            $f->time('invalid_at'),
            $f->stringOr('random_num'),
            $f->stringOr('invoice_type'),
            $f->stringOr('love_code'),
            $f->stringOr('phone_barcode'),
            $f->stringOr('nature_person'),
        );
    }
}
