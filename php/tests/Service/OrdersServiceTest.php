<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Service;

use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Money;
use Actinc\Cyberbiz\Service\OrdersService;
use Actinc\Cyberbiz\Tests\Fake\FakeHttpClient;
use Actinc\Cyberbiz\Time;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class OrdersServiceTest extends TestCase
{
    public function testListDecodesTheGoldenPage(): void
    {
        $h = Harness::golden('v1/GET_v1_orders.json');

        $page = $h->client->orders()->list(['page' => 1, 'per_page' => 2]);

        self::assertSame('GET v1/orders?page=1&per_page=2', $h->line());
        self::assertSame(92, $page->pagination->total);
        $order = $page->items[1];
        self::assertSame(56944000, $order->id);
        self::assertSame(1102, $order->orderNumber);
        self::assertTrue($order->subtotalPrice->equals(Money::of('9999')));
        self::assertSame('2026-09-07 12:39:33', $order->createdAt?->format('Y-m-d H:i:s'));
        self::assertSame(Time::ZONE, $order->createdAt->getTimezone()->getName());
        self::assertSame('pending', $order->statuses?->financialStatus);
        self::assertSame(42058619, $order->customer?->id);
        self::assertSame('inclusive_tax', $order->lineItems[0]->taxTypeId);
        self::assertNotSame('', $order->token);
    }

    public function testGetDecodesTheGoldenOrder(): void
    {
        $h = Harness::golden('v1/GET_v1_orders_{id}.json');

        $order = $h->client->orders()->get(56943817);

        self::assertSame('GET v1/orders/56943817', $h->line());
        self::assertSame(56943817, $order->id);
        self::assertSame(1101, $order->orderNumber);
        self::assertSame('#1101', $order->orderName);
        self::assertSame('1990-01-01', $order->customer?->birthday?->format('Y-m-d'));
        self::assertTrue($order->customer->bonusRemain->equals(Money::of('5000')));
        self::assertNull($order->customer->otherAccumulatedConsumptionExpiredAt);
        self::assertSame('custom', $order->shippingVendor?->type);
        self::assertNull($order->deliveryDate);
        self::assertNull($order->einvoice);
        self::assertNull($order->timings?->closedAt);
        self::assertSame('2026-09-07 12:35:05', $order->timings?->confirmedAt?->format('Y-m-d H:i:s'));
        self::assertTrue($order->prices?->discounts?->vipDiscount->equals(Money::zero()));
        self::assertNull($order->prices->discounts->shopDiscount);
        $component = $order->lineItems[0]->relatedItems[0]->items[0];
        self::assertTrue($component->comboProductPriceDifference->equals(Money::of('9991')));
        self::assertTrue($component->comboProductPriceDiffDetails[0]->equals(Money::of('9991')));
        self::assertSame('非導購訂單', $order->linkedOrderInfo?->source);
        self::assertSame(0, $order->posInfo?->posUserId);
        self::assertSame(0, $order->warehouseTypeId);
        self::assertSame('', $order->token, 'detail responses carry no token');
    }

    public function testGetTurnsANullBodyIntoNotFound(): void
    {
        $this->expectException(NotFoundException::class);

        Harness::json('null')->client->orders()->get(5);
    }

    public function testLookupIdsDecodesTheGoldenFile(): void
    {
        $h = Harness::golden('v1/GET_v1_orders_get_order_id.json');

        $ids = $h->client->orders()->lookupIds([1102, 1101]);

        self::assertSame('GET v1/orders/get_order_id?order_numbers=1102%2C1101', $h->line());
        self::assertSame(1102, $ids[0]->orderNumber);
        self::assertSame(56944000, $ids[0]->orderId);
    }

    public function testFulfillmentGoldenFiles(): void
    {
        $h = Harness::golden('v1/GET_v1_orders_{id}_fulfillments.json', 'v1/GET_v1_orders_{id}_fulfillments_{id}.json');
        $orders = $h->client->orders();

        self::assertSame([], $orders->listFulfillments(7, ['per_page' => 2])->items);
        $fulfillment = $orders->getFulfillment(7, 23064624);

        self::assertSame('GET v1/orders/7/fulfillments?per_page=2', $h->line());
        self::assertSame('GET v1/orders/7/fulfillments/23064624', $h->line(1));
        self::assertSame(23064624, $fulfillment->id);
        self::assertSame('other', $fulfillment->trackingCompany);
        self::assertSame('fulfilled', $fulfillment->status);
        self::assertSame('2026-06-03 21:39:47', $fulfillment->fulfilledAt?->format('Y-m-d H:i:s'));
        self::assertNull($fulfillment->receivedAt);
        self::assertTrue($fulfillment->lineItems[0]->price->equals(Money::of('888')));
        self::assertTrue($fulfillment->lineItems[0]->relatedItems[0]->items[0]->cost->equals(Money::of('100')));
    }

    public function testReturnAndTransactionGoldenFiles(): void
    {
        $h = Harness::golden('v1/GET_v1_orders_{id}_returns.json', 'v1/GET_v1_orders_{id}_transactions.json');
        $orders = $h->client->orders();

        self::assertSame([], $orders->listReturns(7));
        self::assertSame([], $orders->listTransactions(7)->items);
        self::assertSame('GET v1/orders/7/returns', $h->line());
        self::assertSame('GET v1/orders/7/transactions', $h->line(1));
    }

    public function testListEncodesFiltersCommaSeparatedAndTimesInTaipei(): void
    {
        $h = Harness::json('[]');

        $h->client->orders()->list([
            'start_time' => new \DateTimeImmutable('2026-01-01 00:00:00', new \DateTimeZone('UTC')),
            'end_time' => '2026-01-31 23:59:59',
            'statuses' => ['open', 'closed'],
            'tags' => ['vip'],
            'data_source' => 'ec',
        ]);

        self::assertSame(
            'GET v1/orders?start_time=2026-01-01%2008%3A00%3A00&end_time=2026-01-31%2023%3A59%3A59&statuses=open%2Cclosed&tags=vip&data_source=ec',
            $h->line(),
        );
    }

    public function testAllFollowsTheNextPage(): void
    {
        $next = FakeHttpClient::json(200, '[{"id":1}]', ['X-Next-Page' => '2']);
        $h = new Harness(new FakeHttpClient(null, $next, FakeHttpClient::json(200, '[{"id":2}]')));

        $ids = array_map(static fn($o): int => $o->id, iterator_to_array($h->client->orders()->all(['statuses' => ['open']]), false));

        self::assertSame([1, 2], $ids);
        self::assertSame('GET v1/orders?page=2&statuses=open&per_page=50', $h->line(1));
    }

    public function testWritesReturnNullForAnEmptyOrNullReply(): void
    {
        $h = Harness::json('', 'null', '{"id":9,"order_number":1009}');
        $orders = $h->client->orders();

        self::assertNull($orders->markPreparing(9));
        self::assertNull($orders->markArrived(9));
        self::assertSame(1009, $orders->updateStatus(9, 'closed')?->orderNumber);
    }

    public function testShipmentsReturnNullWhileTheCarrierIsPending(): void
    {
        $h = new Harness(new FakeHttpClient(null, FakeHttpClient::json(202, '{"id":1}'), FakeHttpClient::json(200, '{"id":5,"cvs_shipping_type":"seven"}')));
        $orders = $h->client->orders();

        self::assertNull($orders->createSupportShipping(9, ['line_item_ids' => [1, 2], 'source' => 'ezcat']));
        self::assertSame('seven', $orders->createCvsShipping(9)?->cvsShippingType);
        self::assertSame(['line_item_ids' => '1,2', 'source' => 'ezcat'], $h->body());
    }

    public function testCancelSendsMoneyExactly(): void
    {
        $h = Harness::json('{"id":9}');

        $h->client->orders()->cancel(9, ['cancel_reason' => 'other', 'email' => false, 'refund_shopdotcom' => Money::of('10.5')]);

        self::assertSame('PUT v1/orders/9/cancelled', $h->line());
        self::assertSame('{"cancel_reason":"other","email":false,"refund_shopdotcom":10.50}', (string) $h->request()->getBody());
    }

    public function testDecodesSyntheticSubResources(): void
    {
        $h = Harness::json(
            '[{"id":1,"amount":100.0,"kind_name":"已收款","paid_type_name":"手動"}]',
            '[{"id":2,"created_at":"2026-01-02 03:04:05","tracking_company":"ezcat","line_items":[{"id":3,"price":"50"}],"return_suda5":"100"}]',
            '[{"title":"Ticket","ticket_number":"T-1","available_quantity":2,"enabled":true,"separate":true,"transactions":[{"ticket_number":"T-1-1","used_at":"2026-01-02 03:04:05"}]}]',
            '[{"id":4,"id_verification":true,"amount":60,"store_no":"S1"}]',
            '{"request_id":"r-1","results":[{"order_id":5,"tracking_numbers":{"tracking_company":"hct","tracking_number":"N1"},"line_items":"1,2"}]}',
            '{"failed_orders":[{"order_id":6,"message":"no stock"}],"fulfillments":[{"id":7,"order_id":8,"tracking_number":""}]}',
        );
        $orders = $h->client->orders();

        self::assertTrue($orders->listTransactions(1)->items[0]->amount->equals(Money::of('100')));
        self::assertSame('100', $orders->listReturns(1)[0]->returnSuda5);
        self::assertSame('2026-01-02', $orders->listEtickets()->items[0]->transactions[0]->usedAt?->format('Y-m-d'));
        self::assertTrue($orders->getFulfillmentInfos([4])[0]->idVerification);
        self::assertSame('hct', $orders->createSupportShippingBatch(['support_shippings' => [], 'email' => 'ops@example.com'])->results[0]->trackingNumbers?->trackingCompany);
        self::assertSame(6, $orders->createSupportShippings([])->failedOrders[0]->orderId);
    }

    /**
     * Every method not covered above: method, path, query and body.
     *
     * @param \Closure(OrdersService): mixed $call
     */
    #[DataProvider('requestShapes')]
    public function testRequestShape(\Closure $call, string $reply, string $line, mixed $body): void
    {
        $h = Harness::json($reply);

        $call($h->client->orders());

        self::assertSame($line, $h->line());
        self::assertSame($body, (string) $h->request()->getBody() === '' ? null : $h->body());
    }

    /** @return iterable<string, array{\Closure(OrdersService): mixed, string, string, mixed}> */
    public static function requestShapes(): iterable
    {
        yield 'update' => [static fn(OrdersService $s) => $s->update(3, ['note' => 'n', 'delivery_time' => 0]), '{}', 'PUT v1/orders/3', ['note' => 'n', 'delivery_time' => 0]];
        yield 'update tags' => [static fn(OrdersService $s) => $s->updateTags(3, ['a', 'b']), '{}', 'PUT v1/orders/3/tags', ['tags' => ['a', 'b']]];
        yield 'warehouse note' => [static fn(OrdersService $s) => $s->updateWarehouseNote(3, 'fragile'), '{}', 'PUT v1/orders/3/update_warehouse_note', ['warehouse_note' => 'fragile']];
        yield 'update status' => [static fn(OrdersService $s) => $s->updateStatus(3, 'open'), '{}', 'PUT v1/orders/3/update_status', ['status' => 'open']];
        yield 'financial status' => [static fn(OrdersService $s) => $s->updateFinancialStatus(3), '{}', 'PUT v1/orders/3/update_financial_status', null];
        yield 'unshipped' => [static fn(OrdersService $s) => $s->markUnshipped(3), '{}', 'PUT v1/orders/3/unshipped', null];
        yield 'preparing' => [static fn(OrdersService $s) => $s->markPreparing(3), '{}', 'PUT v1/orders/3/preparing', null];
        yield 'manual return' => [static fn(OrdersService $s) => $s->manualReturn(3, 'manual_return_done'), '{}', 'PUT v1/orders/3/manual_return', ['operation' => 'manual_return_done']];
        yield 'custom shipping switch' => [static fn(OrdersService $s) => $s->changeToCustomShipping(3), '{}', 'PUT v1/orders/3/change_to_custom_shipping', null];
        yield 'transactions page' => [static fn(OrdersService $s) => $s->listTransactions(3, ['page' => 2]), '[]', 'GET v1/orders/3/transactions?page=2', null];
        yield 'create transaction' => [static fn(OrdersService $s) => $s->createTransaction(3, ['kind' => 'capture', 'paid_type' => 'manual']), '{}', 'POST v1/orders/3/transactions', ['kind' => 'capture', 'paid_type' => 'manual']];
        yield 'etickets' => [static fn(OrdersService $s) => $s->listEtickets(['search_column' => 'phone', 'q' => '0900']), '[]', 'GET v1/order_etickets?search_column=phone&q=0900', null];
        yield 'redeem eticket' => [static fn(OrdersService $s) => $s->redeemEticket(['ticket_number' => 'T', 'redeem_quantity' => 1, 'user_id' => 2, 'branch_store_id' => 3]), '{}', 'POST v1/order_etickets/submit_redeem', ['ticket_number' => 'T', 'redeem_quantity' => 1, 'user_id' => 2, 'branch_store_id' => 3]];
        yield 'custom shipping' => [static fn(OrdersService $s) => $s->createCustomShipping(3, ['line_item_ids' => [4, 5], 'tracking_number' => 'N', 'tracking_company' => 'ezcat', 'notify_customer' => false]), '{}', 'POST v1/orders/3/fulfillments/custom_shipping', ['line_item_ids' => '4,5', 'tracking_number' => 'N', 'tracking_company' => 'ezcat', 'notify_customer' => false]];
        yield 'support shipping' => [static fn(OrdersService $s) => $s->createSupportShipping(3, ['line_item_ids' => '4', 'size' => 60, 'is_fragile' => false]), '{}', 'POST v1/orders/3/fulfillments/support_shipping', ['line_item_ids' => '4', 'size' => 60, 'is_fragile' => false]];
        yield 'cvs shipping v1' => [static fn(OrdersService $s) => $s->createCvsShippingV1(3, ['measurement' => 'S60']), '{}', 'POST v1/orders/3/fulfillments/cvs_shipping', ['measurement' => 'S60']];
        yield 'partial cvs' => [static fn(OrdersService $s) => $s->createPartialCvsShipping(3, ['items' => [['line_item_id' => 4, 'quantity' => 1]]]), '{}', 'POST v1/orders/3/fulfillments/partial_cvs_shipping', ['items' => [['line_item_id' => 4, 'quantity' => 1]]]];
        yield 'conclude partial cvs' => [static fn(OrdersService $s) => $s->concludePartialCvsShipping(3), '{}', 'POST v1/orders/3/fulfillments/partial_cvs_shipping/conclude', null];
        yield 'express delivery' => [static fn(OrdersService $s) => $s->createExpressDeliveryShipping(3, [4, 5]), '{}', 'POST v1/orders/3/fulfillments/express_delivery_shipping', ['line_item_ids' => '4,5']];
        yield 'arrived' => [static fn(OrdersService $s) => $s->markArrived(3), '{}', 'POST v1/orders/3/fulfillments/arrived', null];
        yield 'received' => [static fn(OrdersService $s) => $s->markReceived(3), '{}', 'POST v1/orders/3/fulfillments/received', null];
        yield 'expired' => [static fn(OrdersService $s) => $s->markExpired(3), '{}', 'POST v1/orders/3/fulfillments/expired', null];
        yield 'fulfillment infos' => [static fn(OrdersService $s) => $s->getFulfillmentInfos([4, 5]), '[]', 'POST v1/orders/fulfillment_infos', ['fulfillment_ids' => [4, 5]]];
        yield 'support shipping batch' => [static fn(OrdersService $s) => $s->createSupportShippingBatch(['support_shippings' => [['order_id' => 3, 'line_item_ids' => [4]]], 'email' => 'ops@example.com']), '{}', 'POST v1/orders/fulfillments/support_shipping', ['support_shippings' => [['order_id' => 3, 'line_item_ids' => '4']], 'email' => 'ops@example.com']];
        yield 'cvs shipping v2' => [static fn(OrdersService $s) => $s->createCvsShipping(3, ['size' => 60]), '{}', 'POST v2/orders/3/cvs_shipping', ['size' => 60]];
        yield 'cvs labels' => [static fn(OrdersService $s) => $s->printCvsShippingLabels(['shipping_type' => 'seven', 'fulfillment_ids' => [6]]), 'PK', 'POST v2/orders/cvs_shipping_labels', ['shipping_type' => 'seven', 'fulfillment_ids' => [6]]];
        yield 'support shippings v2' => [static fn(OrdersService $s) => $s->createSupportShippings(['shipping_type' => 'hct', 'shipping_orders' => [['order_id' => 3, 'line_item_ids' => [4]]]]), '{}', 'POST v2/orders/fulfillments/support_shippings', ['shipping_type' => 'hct', 'shipping_orders' => [['order_id' => 3, 'line_item_ids' => [4]]]]];
        yield 'support labels' => [static fn(OrdersService $s) => $s->printSupportShippingLabels(['shipping_type' => 'hct', 'print_type' => 'thermal', 'order_ids' => [3]]), 'PK', 'POST v2/orders/fulfillments/support_shipping_labels', ['shipping_type' => 'hct', 'print_type' => 'thermal', 'order_ids' => [3]]];
    }

    public function testLabelsReturnTheRawArchive(): void
    {
        $h = Harness::json('PK-zip-bytes');

        self::assertSame('PK-zip-bytes', $h->client->orders()->printCvsShippingLabels(['shipping_type' => 'seven', 'fulfillment_ids' => [1]])->body);
    }

    public function testRejectsNonScalarLineItemIds(): void
    {
        $this->expectException(\InvalidArgumentException::class);

        Harness::json('{}')->client->orders()->createSupportShipping(1, ['line_item_ids' => [[1]]]);
    }
}
