<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/**
 * One shop order (GET /v1/orders, GET /v1/orders/{id}, GET
 * /v1/customers/{id}/orders). The list responses also carry token.
 */
final class Order
{
    /**
     * @param Money $subtotalPrice line items only, before shipping
     * @param int $orderNumber the shop-facing sequential number (the swagger says string; the platform sends an integer)
     * @param string $orderName e.g. "#1101"
     * @param list<LineItem> $lineItems
     * @param string $logisticsId ECPay logistics order id
     * @param \DateTimeImmutable|null $deliveryDate requested delivery day (a date), custom feature
     * @param int $deliveryTime requested delivery slot 0-3, custom feature
     * @param string $delegate staff email when the order was placed on the customer's behalf
     * @param list<Fulfillment> $fulfillments
     * @param list<OrderPaymentInfo> $multiplePaymentInfos POS split payments
     * @param string $card4no last four digits of the card
     * @param list<ReturnHistory> $returnHistories
     * @param OrderBranchStore|null $branchStore pickup store
     * @param Money $totalBonusRedemptionPrice bonus points spent in the bonus mall, in the shop currency
     * @param list<ExchangeHistory> $exchangeHistories
     * @param list<Tag> $tags
     * @param string $fromDevice e.g. 桌機, 手機
     * @param list<string> $serialNumbers campaign serial numbers
     * @param string $token the order's public token, list endpoints only
     */
    public function __construct(
        public readonly int $id,
        public readonly Money $subtotalPrice,
        public readonly ?\DateTimeImmutable $createdAt,
        public readonly ?\DateTimeImmutable $updatedAt,
        public readonly int $orderNumber,
        public readonly string $orderName,
        public readonly ?Customer $customer,
        public readonly ?Buyer $buyer,
        public readonly ?Receiver $receiver,
        public readonly ?Address $billingAddress,
        public readonly array $lineItems,
        public readonly string $shippingType,
        public readonly string $shippingName,
        public readonly ?ShippingVendor $shippingVendor,
        public readonly string $logisticsId,
        public readonly ?\DateTimeImmutable $deliveryDate,
        public readonly int $deliveryTime,
        public readonly string $delegate,
        public readonly array $fulfillments,
        public readonly string $paymentName,
        public readonly string $paymentMethod,
        public readonly string $paymentUrl,
        public readonly array $multiplePaymentInfos,
        public readonly ?Prices $prices,
        public readonly string $card4no,
        public readonly string $transactionNumber,
        public readonly string $merchantTradeNo,
        public readonly ?OrderEinvoice $einvoice,
        public readonly string $paperInvoiceNo,
        public readonly string $paperCompanyNo,
        public readonly string $prepaymentPaperInvoiceNo,
        public readonly string $prepaymentPaperCompanyNo,
        public readonly ?OrderStatuses $statuses,
        public readonly ?OrderTimings $timings,
        public readonly array $returnHistories,
        public readonly string $note,
        public readonly ?OrderBranchStore $branchStore,
        public readonly string $referralCode,
        public readonly string $checkoutReferralCode,
        public readonly string $checkoutReferralUserName,
        public readonly string $registerReferralCode,
        public readonly Money $totalBonusRedemptionPrice,
        public readonly ?PosInfo $posInfo,
        public readonly array $exchangeHistories,
        public readonly ?LinkedOrderInfo $linkedOrderInfo,
        public readonly array $tags,
        public readonly ?OrderBranchStore $expressDeliveryBranchStore,
        public readonly string $shippingStatus,
        public readonly string $extraInfo,
        public readonly string $fromDevice,
        public readonly ?OrderCustomerCancelReason $customerCancelReasonDetail,
        public readonly array $serialNumbers,
        public readonly float $orderWeight,
        public readonly int $warehouseTypeId,
        public readonly ?UtmTracking $utmTracking,
        public readonly string $token,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->intOr('id'),
            $f->moneyOr('subtotal_price'),
            $f->time('created_at'),
            $f->time('updated_at'),
            $f->intOr('order_number'),
            $f->stringOr('order_name'),
            Nested::object($f, 'customer', Customer::fromFields(...)),
            Nested::object($f, 'buyer', Buyer::fromFields(...)),
            Nested::object($f, 'receiver', Receiver::fromFields(...)),
            Nested::object($f, 'billing_address', Address::fromFields(...)),
            $f->list('line_items', LineItem::fromFields(...)),
            $f->stringOr('shipping_type'),
            $f->stringOr('shipping_name'),
            Nested::object($f, 'shipping_vendor', ShippingVendor::fromFields(...)),
            $f->stringOr('logistics_id'),
            $f->time('delivery_date'),
            $f->intOr('delivery_time'),
            $f->stringOr('delegate'),
            $f->list('fulfillments', Fulfillment::fromFields(...)),
            $f->stringOr('payment_name'),
            $f->stringOr('payment_method'),
            $f->stringOr('payment_url'),
            $f->list('multiple_payment_infos', OrderPaymentInfo::fromFields(...)),
            Nested::object($f, 'prices', Prices::fromFields(...)),
            $f->stringOr('card4no'),
            $f->stringOr('transaction_number'),
            $f->stringOr('merchant_trade_no'),
            Nested::object($f, 'einvoice', OrderEinvoice::fromFields(...)),
            $f->stringOr('paper_invoice_no'),
            $f->stringOr('paper_company_no'),
            $f->stringOr('prepayment_paper_invoice_no'),
            $f->stringOr('prepayment_paper_company_no'),
            Nested::object($f, 'statuses', OrderStatuses::fromFields(...)),
            Nested::object($f, 'timings', OrderTimings::fromFields(...)),
            $f->list('return_histories', ReturnHistory::fromFields(...)),
            $f->stringOr('note'),
            Nested::object($f, 'branch_store', OrderBranchStore::fromFields(...)),
            $f->stringOr('referral_code'),
            $f->stringOr('checkout_referral_code'),
            $f->stringOr('checkout_referral_user_name'),
            $f->stringOr('register_referral_code'),
            $f->moneyOr('total_bonus_redemption_price'),
            Nested::object($f, 'pos_info', PosInfo::fromFields(...)),
            $f->list('exchange_histories', ExchangeHistory::fromFields(...)),
            Nested::object($f, 'linked_order_info', LinkedOrderInfo::fromFields(...)),
            $f->list('tags', Tag::fromFields(...)),
            Nested::object($f, 'express_delivery_branch_store', OrderBranchStore::fromFields(...)),
            $f->stringOr('shipping_status'),
            $f->stringOr('extra_info'),
            $f->stringOr('from_device'),
            Nested::object($f, 'customer_cancel_reason_detail', OrderCustomerCancelReason::fromFields(...)),
            $f->strings('serial_numbers'),
            $f->float('order_weight'),
            $f->intOr('warehouse_type_id'),
            Nested::object($f, 'utm_tracking', UtmTracking::fromFields(...)),
            $f->stringOr('token'),
        );
    }
}
