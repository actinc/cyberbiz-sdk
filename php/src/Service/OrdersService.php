<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Service;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Exception\TransportException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Json;
use Actinc\Cyberbiz\Model\Fulfillment;
use Actinc\Cyberbiz\Model\FulfillmentInfo;
use Actinc\Cyberbiz\Model\Order;
use Actinc\Cyberbiz\Model\OrderEticket;
use Actinc\Cyberbiz\Model\OrderNumberId;
use Actinc\Cyberbiz\Model\OrderReturn;
use Actinc\Cyberbiz\Model\OrderTransaction;
use Actinc\Cyberbiz\Model\SupportShippingBatchResult;
use Actinc\Cyberbiz\Model\SupportShippingsResult;
use Actinc\Cyberbiz\Page;
use Actinc\Cyberbiz\Request;
use Actinc\Cyberbiz\Response;

/**
 * Orders, their fulfillments, payments, returns and e-tickets, plus the v2
 * shipping-label endpoints.
 *
 * Request bodies and query options are associative arrays shaped like the
 * API's JSON (see docs/api/en/cyberbiz-openapi-v1.yaml and -v2.yaml). In
 * queries a list is sent comma-separated and a DateTimeInterface as a
 * CYBERBIZ timestamp; in bodies Money is sent exactly. The write endpoints
 * that answer with an order return null when the platform replies with an
 * empty or null body.
 *
 * @phpstan-type OrderListQuery array{page?: int, per_page?: int, offset?: int, start_time?: \DateTimeInterface|string, end_time?: \DateTimeInterface|string, updated_at_start_time?: \DateTimeInterface|string, updated_at_end_time?: \DateTimeInterface|string, closed_at_start_time?: \DateTimeInterface|string, closed_at_end_time?: \DateTimeInterface|string, refund_at_start_time?: \DateTimeInterface|string, refund_at_end_time?: \DateTimeInterface|string, cancelled_at_start_time?: \DateTimeInterface|string, cancelled_at_end_time?: \DateTimeInterface|string, statuses?: list<string>, financial_statuses?: list<string>, fulfillment_statuses?: list<string>, return_statuses?: list<string>, tags?: list<string>, excluded_tags?: list<string>, data_source?: string, vendor?: string}
 * @phpstan-type PageQuery array{page?: int, per_page?: int, offset?: int}
 *
 * @throws ApiException|TransportException|DecodeException|\JsonException from every method
 */
final class OrdersService
{
    public function __construct(
        private readonly Client $client,
        private readonly Endpoint $endpoint,
    ) {}

    /**
     * One page of orders (GET /v1/orders). Time ranges are inclusive;
     * data_source is "ec" or "pos".
     *
     * @param OrderListQuery $query
     *
     * @return Page<Order>
     */
    public function list(array $query = []): Page
    {
        return $this->client->page(new Request('GET', 'v1/orders', Options::query($query)), Endpoint::mapper(Order::fromFields(...)));
    }

    /**
     * Every order, page by page (GET /v1/orders).
     *
     * @param OrderListQuery $query
     *
     * @return \Generator<int, Order>
     */
    public function all(array $query = []): \Generator
    {
        return $this->client->each(new Request('GET', 'v1/orders', Options::query($query)), Endpoint::mapper(Order::fromFields(...)));
    }

    /**
     * One order (GET /v1/orders/{id}).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function get(int $id): Order
    {
        return $this->endpoint->one(new Request('GET', self::path($id)), Order::fromFields(...));
    }

    /**
     * Maps shop-facing order numbers to API ids (GET /v1/orders/get_order_id).
     *
     * @param list<int> $orderNumbers
     *
     * @return list<OrderNumberId>
     */
    public function lookupIds(array $orderNumbers): array
    {
        $request = new Request('GET', 'v1/orders/get_order_id', ['order_numbers' => Options::join($orderNumbers)]);

        return $this->endpoint->list($request, OrderNumberId::fromFields(...));
    }

    /**
     * Changes an order's note and delivery preferences (PUT /v1/orders/{id}).
     * delivery_date ("YYYY-MM-DD") and delivery_time (0-3) are custom
     * features CYBERBIZ must enable.
     *
     * @param array{note?: string, delivery_date?: string, delivery_time?: int} $changes
     */
    public function update(int $id, array $changes): ?Order
    {
        return $this->order(new Request('PUT', self::path($id), body: $changes));
    }

    /**
     * Replaces an order's tags (PUT /v1/orders/{id}/tags).
     *
     * @param list<string> $tags
     */
    public function updateTags(int $id, array $tags): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'tags'), body: ['tags' => $tags]));
    }

    /** Sets the logistics message shown for an order (PUT /v1/orders/{id}/update_warehouse_note). */
    public function updateWarehouseNote(int $id, string $note): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'update_warehouse_note'), body: ['warehouse_note' => $note]));
    }

    /**
     * Opens or closes an order (PUT /v1/orders/{id}/update_status).
     *
     * @param string $status "open" or "closed"
     */
    public function updateStatus(int $id, string $status): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'update_status'), body: ['status' => $status]));
    }

    /** Asks the platform to refresh an order's payment state (PUT /v1/orders/{id}/update_financial_status). */
    public function updateFinancialStatus(int $id): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'update_financial_status')));
    }

    /** Sets the fulfillment status back to unshipped (PUT /v1/orders/{id}/unshipped). */
    public function markUnshipped(int $id): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'unshipped')));
    }

    /** Sets the fulfillment status to preparing (PUT /v1/orders/{id}/preparing). */
    public function markPreparing(int $id): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'preparing')));
    }

    /**
     * Cancels an order (PUT /v1/orders/{id}/cancelled). cancel_reason is
     * required: "customer", "duplicate", "not_pay", "fraud", "inventory",
     * "forgot", "not_shipped_yet", "pos_sp_return", "card_paid_fail" or
     * "other". cancel_reason_detail is the customer-side reason:
     * "wait_too_long", "want_to_use_other_discount", "modify_shipment_location",
     * "have_concern_about_product", "price_too_high", "operation_mistake" or "other".
     *
     * @param array<string, mixed> $cancel
     */
    public function cancel(int $id, array $cancel): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'cancelled'), body: $cancel));
    }

    /**
     * Moves an order's return status by hand (PUT /v1/orders/{id}/manual_return).
     *
     * @param string $operation "manual_returning", "manual_check_goods", "manual_return_refuse" or "manual_return_done"
     */
    public function manualReturn(int $id, string $operation): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'manual_return'), body: ['operation' => $operation]));
    }

    /** Switches an order to merchant-arranged shipping (PUT /v1/orders/{id}/change_to_custom_shipping). */
    public function changeToCustomShipping(int $id): ?Order
    {
        return $this->order(new Request('PUT', self::path($id, 'change_to_custom_shipping')));
    }

    /**
     * One page of an order's payments (GET /v1/orders/{id}/transactions).
     *
     * @param PageQuery $query
     *
     * @return Page<OrderTransaction>
     */
    public function listTransactions(int $id, array $query = []): Page
    {
        return $this->client->page(new Request('GET', self::path($id, 'transactions'), $query), Endpoint::mapper(OrderTransaction::fromFields(...)));
    }

    /**
     * Records a manual payment against an order (POST /v1/orders/{id}/transactions).
     *
     * @param array{kind: string, paid_type: string} $transaction kind "capture", paid_type "manual"
     */
    public function createTransaction(int $id, array $transaction): OrderTransaction
    {
        return $this->endpoint->object(new Request('POST', self::path($id, 'transactions'), body: $transaction), OrderTransaction::fromFields(...));
    }

    /**
     * An order's return shipments. Not paginated; the swagger documents an
     * object but the platform sends an array (GET /v1/orders/{id}/returns).
     *
     * @return list<OrderReturn>
     */
    public function listReturns(int $id): array
    {
        return $this->endpoint->list(new Request('GET', self::path($id, 'returns')), OrderReturn::fromFields(...));
    }

    /**
     * One page of e-tickets; the shop needs the e-ticket feature or the
     * platform answers 403 (GET /v1/order_etickets).
     *
     * @param array{page?: int, per_page?: int, offset?: int, search_column?: string, q?: string} $query search_column: "phone", "ticket_number", "title", "name" or "order_name"
     *
     * @return Page<OrderEticket>
     */
    public function listEtickets(array $query = []): Page
    {
        return $this->client->page(new Request('GET', 'v1/order_etickets', $query), Endpoint::mapper(OrderEticket::fromFields(...)));
    }

    /**
     * Redeems units of an e-ticket at a branch store (POST /v1/order_etickets/submit_redeem).
     *
     * @param array{ticket_number: string, redeem_quantity: int, user_id: int, branch_store_id: int} $redeem
     */
    public function redeemEticket(array $redeem): OrderEticket
    {
        return $this->endpoint->object(new Request('POST', 'v1/order_etickets/submit_redeem', body: $redeem), OrderEticket::fromFields(...));
    }

    /**
     * One page of an order's fulfillments (GET /v1/orders/{id}/fulfillments).
     *
     * @param PageQuery $query
     *
     * @return Page<Fulfillment>
     */
    public function listFulfillments(int $orderId, array $query = []): Page
    {
        return $this->client->page(new Request('GET', self::path($orderId, 'fulfillments'), $query), Endpoint::mapper(Fulfillment::fromFields(...)));
    }

    /**
     * One fulfillment (GET /v1/orders/{id}/fulfillments/{fulfillment_id}).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function getFulfillment(int $orderId, int $fulfillmentId): Fulfillment
    {
        return $this->endpoint->one(new Request('GET', self::path($orderId, 'fulfillments/' . $fulfillmentId)), Fulfillment::fromFields(...));
    }

    /**
     * Ships line items with a merchant-arranged carrier
     * (POST /v1/orders/{id}/fulfillments/custom_shipping). line_item_ids may
     * be a list; it is sent comma-separated.
     *
     * @param array{line_item_ids: list<int>|string, tracking_number: string, tracking_company: string, notify_customer: bool} $shipping
     */
    public function createCustomShipping(int $orderId, array $shipping): ?Fulfillment
    {
        return $this->fulfillment(self::path($orderId, 'fulfillments/custom_shipping'), Options::lineItemIds($shipping));
    }

    /**
     * Books a home-delivery shipment through CYBERBIZ. The platform may answer
     * 202 while the tracking number is being issued; the result is then null
     * and the fulfillment can be read later with listFulfillments()
     * (POST /v1/orders/{id}/fulfillments/support_shipping). source: "ezcat",
     * "pelican", "sf" or "hct"; temperature: "normal" or "cold";
     * fridge_or_frozen: "none", "fridge" or "frozen".
     *
     * @param array<string, mixed> $shipping line_item_ids may be a list
     */
    public function createSupportShipping(int $orderId, array $shipping): ?Fulfillment
    {
        return $this->fulfillment(self::path($orderId, 'fulfillments/support_shipping'), Options::lineItemIds($shipping));
    }

    /**
     * Books a convenience-store shipment through the v1 endpoint; prefer
     * createCvsShipping(), whose reply carries the label type
     * (POST /v1/orders/{id}/fulfillments/cvs_shipping).
     *
     * @param array{measurement?: string, size?: int} $shipping measurement "S60" or "S105"
     */
    public function createCvsShippingV1(int $orderId, array $shipping = []): ?Fulfillment
    {
        return $this->fulfillment(self::path($orderId, 'fulfillments/cvs_shipping'), $shipping);
    }

    /**
     * Ships part of a Hi-Life order (POST /v1/orders/{id}/fulfillments/partial_cvs_shipping).
     *
     * @param array{items: list<array{line_item_id: int, quantity: int}>, charge?: \Actinc\Cyberbiz\Money} $shipping
     */
    public function createPartialCvsShipping(int $orderId, array $shipping): ?Fulfillment
    {
        return $this->fulfillment(self::path($orderId, 'fulfillments/partial_cvs_shipping'), $shipping);
    }

    /** Closes partial shipping on a Hi-Life order (POST /v1/orders/{id}/fulfillments/partial_cvs_shipping/conclude). */
    public function concludePartialCvsShipping(int $orderId): ?Order
    {
        return $this->order(new Request('POST', self::path($orderId, 'fulfillments/partial_cvs_shipping/conclude')));
    }

    /**
     * Ships line items by express delivery from a branch store
     * (POST /v1/orders/{id}/fulfillments/express_delivery_shipping).
     *
     * @param list<int> $lineItemIds
     */
    public function createExpressDeliveryShipping(int $orderId, array $lineItemIds): ?Fulfillment
    {
        return $this->fulfillment(self::path($orderId, 'fulfillments/express_delivery_shipping'), ['line_item_ids' => Options::join($lineItemIds)]);
    }

    /** Marks a CVS order as arrived at the store (POST /v1/orders/{id}/fulfillments/arrived). */
    public function markArrived(int $orderId): ?Order
    {
        return $this->order(new Request('POST', self::path($orderId, 'fulfillments/arrived')));
    }

    /** Marks an order as received by the customer (POST /v1/orders/{id}/fulfillments/received). */
    public function markReceived(int $orderId): ?Order
    {
        return $this->order(new Request('POST', self::path($orderId, 'fulfillments/received')));
    }

    /** Marks a CVS order as not collected in time (POST /v1/orders/{id}/fulfillments/expired). */
    public function markExpired(int $orderId): ?Order
    {
        return $this->order(new Request('POST', self::path($orderId, 'fulfillments/expired')));
    }

    /**
     * The CVS label data of the given fulfillments (POST /v1/orders/fulfillment_infos).
     *
     * @param list<int> $fulfillmentIds
     *
     * @return list<FulfillmentInfo>
     */
    public function getFulfillmentInfos(array $fulfillmentIds): array
    {
        $request = new Request('POST', 'v1/orders/fulfillment_infos', body: ['fulfillment_ids' => $fulfillmentIds]);

        return $this->endpoint->list($request, FulfillmentInfo::fromFields(...));
    }

    /**
     * Books home-delivery shipments for several orders through the v1
     * endpoint (POST /v1/orders/fulfillments/support_shipping). Each entry's
     * line_item_ids may be a list.
     *
     * @param array{support_shippings: list<array<string, mixed>>, email: string} $batch
     */
    public function createSupportShippingBatch(array $batch): SupportShippingBatchResult
    {
        $batch['support_shippings'] = array_map(Options::lineItemIds(...), $batch['support_shippings']);
        $request = new Request('POST', 'v1/orders/fulfillments/support_shipping', body: $batch);

        return $this->endpoint->object($request, SupportShippingBatchResult::fromFields(...));
    }

    /**
     * Books a convenience-store shipment; only a reply with a non-empty
     * cvsShippingType can be printed with printCvsShippingLabels()
     * (POST /v2/orders/{id}/cvs_shipping).
     *
     * @param array{measurement?: string, size?: int} $shipping
     */
    public function createCvsShipping(int $orderId, array $shipping = []): ?Fulfillment
    {
        return $this->fulfillment('v2/orders/' . $orderId . '/cvs_shipping', $shipping);
    }

    /**
     * A zip archive of CVS shipping labels, in the returned Response's body
     * (POST /v2/orders/cvs_shipping_labels).
     *
     * @param array{shipping_type: string, fulfillment_ids: list<int>} $labels
     */
    public function printCvsShippingLabels(array $labels): Response
    {
        return $this->endpoint->call(new Request('POST', 'v2/orders/cvs_shipping_labels', body: $labels));
    }

    /**
     * Books home-delivery labels for several orders
     * (POST /v2/orders/fulfillments/support_shippings).
     *
     * @param array<string, mixed> $shippings shipping_type, size, temperature, ..., shipping_orders: list<array{order_id: int, line_item_ids: list<int>}>
     */
    public function createSupportShippings(array $shippings): SupportShippingsResult
    {
        $request = new Request('POST', 'v2/orders/fulfillments/support_shippings', body: $shippings);

        return $this->endpoint->object($request, SupportShippingsResult::fromFields(...));
    }

    /**
     * A zip archive of home-delivery labels, in the returned Response's body
     * (POST /v2/orders/fulfillments/support_shipping_labels).
     *
     * @param array{shipping_type: string, print_type?: string, order_ids: list<int>} $labels print_type "normal" or "thermal", HCT only
     */
    public function printSupportShippingLabels(array $labels): Response
    {
        return $this->endpoint->call(new Request('POST', 'v2/orders/fulfillments/support_shipping_labels', body: $labels));
    }

    /** Sends a write whose reply is the updated order, or nothing. */
    private function order(Request $request): ?Order
    {
        $fields = self::optional($this->endpoint->call($request));

        return $fields === null ? null : Order::fromFields($fields);
    }

    /**
     * Posts a shipment; a 202, empty or null reply yields null.
     *
     * @param array<string, mixed> $body
     */
    private function fulfillment(string $path, array $body): ?Fulfillment
    {
        $response = $this->endpoint->call(new Request('POST', $path, body: $body));
        $fields = $response->statusCode === 202 ? null : self::optional($response);

        return $fields === null ? null : Fulfillment::fromFields($fields);
    }

    private static function optional(Response $response): ?Fields
    {
        return $response->isNull() || trim($response->body) === '' ? null : Fields::of(Json::decode($response->body));
    }

    private static function path(int $orderId, string $suffix = ''): string
    {
        return 'v1/orders/' . $orderId . ($suffix === '' ? '' : '/' . $suffix);
    }
}
