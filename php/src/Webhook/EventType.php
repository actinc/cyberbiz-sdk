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
