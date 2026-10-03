<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** The label data of one CVS fulfillment (POST /v1/orders/fulfillment_infos). */
final class FulfillmentInfo
{
    /**
     * @param bool $idVerification pickup requires an id check
     * @param string $supplierName Hi-Life only
     * @param string $imageUrl FamilyMart only
     * @param string $pdf base64 PDF, Hi-Life only
     * @param string $html 7-11 C2C only
     * @param string $pdfJson Hi-Life cold chain only
     */
    public function __construct(
        public readonly int $id,
        public readonly bool $idVerification,
        public readonly string $receiver,
        public readonly string $receiverPhone,
        public readonly Money $amount,
        public readonly string $trackingNumber,
        public readonly string $storeName,
        public readonly string $storeNo,
        public readonly string $barcode,
        public readonly string $shopName,
        public readonly string $orderNo,
        public readonly string $shopPhone,
        public readonly string $shopUrl,
        public readonly string $supplierName,
        public readonly string $imageUrl,
        public readonly string $pdf,
        public readonly string $html,
        public readonly string $pdfJson,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->boolOr('id_verification'),
            $f->stringOr('receiver'),
            $f->stringOr('receiver_phone'),
            $f->moneyOr('amount'),
            $f->stringOr('tracking_number'),
            $f->stringOr('store_name'),
            $f->stringOr('store_no'),
            $f->stringOr('barcode'),
            $f->stringOr('shop_name'),
            $f->stringOr('order_no'),
            $f->stringOr('shop_phone'),
            $f->stringOr('shop_url'),
            $f->stringOr('supplier_name'),
            $f->stringOr('image_url'),
            $f->stringOr('pdf'),
            $f->stringOr('html'),
            $f->stringOr('pdf_json'),
        );
    }
}
