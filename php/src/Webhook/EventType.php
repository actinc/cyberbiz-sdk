<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/** Every documented webhook event, in documentation order. */
enum EventType: string
{
    case CustomersCreate = 'customers/create';
    case CustomersUpdate = 'customers/update';
    case UidProvidersCreate = 'uid_providers/create';
    case UidProvidersUpdate = 'uid_providers/update';
    case BonusPointsCreate = 'bonus_points/create';
    case BonusPointsUpdate = 'bonus_points/update';
    case BonusPointsDestroy = 'bonus_points/destroy';
    case CommentBonusCreate = 'comment_bonus/create';
    case OrdersCreate = 'orders/create';
    case OrdersPaid = 'orders/paid';
    case OrdersPreparing = 'orders/preparing';
    case OrdersFulfilled = 'orders/fulfilled';
    case OrdersReceived = 'orders/received';
    case OrdersArrived = 'orders/arrived';
    case OrdersExpired = 'orders/expired';
    case OrdersCancelled = 'orders/cancelled';
    case OrdersReturned = 'orders/returned';
    case OrdersPartialReturn = 'orders/partial_return';
    case OrdersRefunded = 'orders/refunded';
    case OrdersPartialRefunded = 'orders/partial_refunded';
    case OrdersClosed = 'orders/closed';
    case OrdersRequestReturn = 'orders/request_return';
    case OrdersOpened = 'orders/opened';
    case ExpressDeliveryOrdersUpdate = 'express_delivery_orders/update';
    case ProductsCreate = 'products/create';
    case ProductsUpdate = 'products/update';
    case ProductsDelete = 'products/delete';
    case VariantsCreate = 'variants/create';
    case VariantsUpdate = 'variants/update';
    case VariantsDelete = 'variants/delete';
    case CouponsCreate = 'coupons/create';
    case CouponsUpdate = 'coupons/update';
    case CouponsDestroy = 'coupons/destroy';
    case CustomerVipLevelUpdate = 'customer_vip_level/update';
    case AppsUninstall = 'apps/uninstall';
    // Added from the source-derived webhook collection; the last four are sent by
    // Merchant webhooks only and have no resource prefix.
    case OrdersUpdate = 'orders/update';
    case ProductsUpdateOptions = 'products/update_options';
    case ProductsUpdatePhotos = 'products/update_photos';
    case VariantsOutOfStock = 'variants/out_of_stock';
    case CartAdd = 'cart/add';
    case CustomerTagsUpdate = 'customer_tags/update';
    case AffiliateVendorOrdersCreate = 'affiliate_vendor_orders/create';
    case AffiliateVendorOrdersCancelled = 'affiliate_vendor_orders/cancelled';
    case AffiliateVendorOrdersClosed = 'affiliate_vendor_orders/closed';
    case CollectionsCreate = 'collections/create';
    case CollectionsUpdate = 'collections/update';
    case CollectionsDelete = 'collections/delete';
    case MenusCreate = 'menus/create';
    case MenusUpdate = 'menus/update';
    case MenusDestroy = 'menus/destroy';
    case DomainsCreate = 'domains/create';
    case DomainsUpdate = 'domains/update';
    case DomainsDestroy = 'domains/destroy';
    case BulkOperationsFinish = 'bulk_operations/finish';
    case LineChatBotUpdate = 'line_chat_bot/update';
    case FbeAdAccountsReclaim = 'fbe_ad_accounts/reclaim';
    case PosWalletTransactionsTopup = 'pos_wallet_transactions/topup';
    case PosWalletTransactionsRefund = 'pos_wallet_transactions/refund';
    case StockInvoicesCreated = 'stock_invoices/created';
    case StockInvoicesDone = 'stock_invoices/done';
    case StockInvoicesCancelled = 'stock_invoices/cancelled';
    case StockReceiptsCreated = 'stock_receipts/created';
    case StockReceiptsDone = 'stock_receipts/done';
    case StockReceiptsCancelled = 'stock_receipts/cancelled';
    case StockRequisitionsCreated = 'stock_requisitions/created';
    case StockRequisitionsApproved = 'stock_requisitions/approved';
    case StockRequisitionsCancelled = 'stock_requisitions/cancelled';
    case StoredOrdersRetrieved = 'stored_orders/retrieved';
    case StoredOrdersRefunded = 'stored_orders/refunded';
    case SendEdmFinish = 'send_edm_finish';
    case SendSmsFinish = 'send_sms_finish';
    case SendLineOaMessageFinish = 'send_line_oa_message_finish';
    case GetCustomerIdsFinish = 'get_customer_ids_finish';

    /** "orders" for "orders/paid". */
    public function resource(): string
    {
        return strstr($this->value, '/', true) ?: $this->value;
    }

    /** "paid" for "orders/paid". */
    public function action(): string
    {
        return substr((string) strrchr($this->value, '/'), 1);
    }
}
