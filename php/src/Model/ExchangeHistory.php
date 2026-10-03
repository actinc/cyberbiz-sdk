<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** One POS exchange performed on an order. */
final class ExchangeHistory
{
    /**
     * @param list<OrderPaymentInfo> $multiplePaymentInfos
     * @param list<ExchangeLineItem> $lineItems
     */
    public function __construct(
        public readonly ?\DateTimeImmutable $createdAt,
        public readonly Money $price,
        public readonly Money $orderPriceBefore,
        public readonly Money $orderPriceAfter,
        public readonly int $posShopId,
        public readonly int $posId,
        public readonly string $paymentName,
        public readonly string $paymentMethod,
        public readonly array $multiplePaymentInfos,
        public readonly array $lineItems,
        public readonly ?OrderEinvoice $einvoice,
        public readonly string $paperInvoiceNo,
        public readonly string $paperCompanyNo,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->time('created_at'),
            $f->moneyOr('price'),
            $f->moneyOr('order_price_before'),
            $f->moneyOr('order_price_after'),
            $f->intOr('pos_shop_id'),
            $f->intOr('pos_id'),
            $f->stringOr('payment_name'),
            $f->stringOr('payment_method'),
            $f->list('multiple_payment_infos', OrderPaymentInfo::fromFields(...)),
            $f->list('line_items', ExchangeLineItem::fromFields(...)),
            Nested::object($f, 'einvoice', OrderEinvoice::fromFields(...)),
            $f->stringOr('paper_invoice_no'),
            $f->stringOr('paper_company_no'),
        );
    }
}
