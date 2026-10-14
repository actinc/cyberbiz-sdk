<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Service;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Exception\TransportException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Model\Customer;
use Actinc\Cyberbiz\Model\CustomerIdMatch;
use Actinc\Cyberbiz\Model\CustomerMessagePost;
use Actinc\Cyberbiz\Model\CustomerNameMatch;
use Actinc\Cyberbiz\Model\CustomerOAuthApp;
use Actinc\Cyberbiz\Model\CustomerSpendingOverview;
use Actinc\Cyberbiz\Model\CustomerUidLookup;
use Actinc\Cyberbiz\Model\CustomerVipInfo;
use Actinc\Cyberbiz\Model\LineItem;
use Actinc\Cyberbiz\Model\Order;
use Actinc\Cyberbiz\Model\ProductVariant;
use Actinc\Cyberbiz\Model\Tag;
use Actinc\Cyberbiz\Page;
use Actinc\Cyberbiz\Request;
use Actinc\Cyberbiz\Time;

/**
 * Customers, their orders, cart, service threads, VIP state and external
 * login identities, plus the v2 customer lookups.
 *
 * Request bodies and query options are associative arrays shaped like the
 * API's JSON (see docs/api/en/cyberbiz-openapi-v1.yaml and -v2.yaml). In
 * queries a list is sent comma-separated and a DateTimeInterface as a
 * CYBERBIZ timestamp (a date for start_date / end_date).
 *
 * @phpstan-type PageQuery array{page?: int, per_page?: int, offset?: int}
 * @phpstan-type ListV2Query array{page?: int, per_page?: int, offset?: int, ids?: list<int>, include_params?: list<string>}
 * @phpstan-type DateValue \DateTimeInterface|string
 *
 * @throws ApiException|TransportException|DecodeException|\JsonException from every method
 */
final class CustomersService
{
    /** The verification timestamps the platform parses as ISO 8601 (Ruby Time.iso8601). */
    private const RFC3339_FIELDS = ['confirmed_at', 'mobile_sms_confirmed_at'];

    public function __construct(
        private readonly Client $client,
        private readonly Endpoint $endpoint,
    ) {}

    /**
     * One page of customers (GET /v1/customers).
     *
     * @param array{page?: int, per_page?: int, offset?: int, updated_at_start_time?: \DateTimeInterface|string, updated_at_end_time?: \DateTimeInterface|string} $query
     *
     * @return Page<Customer>
     */
    public function list(array $query = []): Page
    {
        return $this->client->page(new Request('GET', 'v1/customers', Options::query($query)), Endpoint::mapper(self::customer()));
    }

    /**
     * Every customer, page by page (GET /v1/customers).
     *
     * @param array{page?: int, per_page?: int, offset?: int, updated_at_start_time?: \DateTimeInterface|string, updated_at_end_time?: \DateTimeInterface|string} $query
     *
     * @return \Generator<int, Customer>
     */
    public function all(array $query = []): \Generator
    {
        return $this->client->each(new Request('GET', 'v1/customers', Options::query($query)), Endpoint::mapper(self::customer()));
    }

    /**
     * One customer; the response omits the id, which is filled in
     * (GET /v1/customers/{id}).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function get(int $id): Customer
    {
        return $this->endpoint->one(new Request('GET', self::path($id)), self::customer($id));
    }

    /**
     * Creates a customer (POST /v1/customers). birthday and
     * other_accumulated_consumption_expired_at are "YYYY-MM-DD"; tags_text
     * is comma-separated; status is one of the Customer::$status values.
     * A DateTimeInterface under confirmed_at or mobile_sms_confirmed_at is
     * sent as RFC 3339 in Asia/Taipei ("2026-01-02T03:04:05+08:00"): the
     * platform rejects the space-separated API layout there.
     *
     * @param array<string, mixed> $customer
     */
    public function create(array $customer): Customer
    {
        return $this->endpoint->object(new Request('POST', 'v1/customers', body: self::verificationTimes($customer)), self::customer());
    }

    /**
     * Changes a customer (PUT /v1/customers/{id}). Sending confirmed_at or
     * mobile_sms_confirmed_at as null clears that verification; a
     * DateTimeInterface there is sent as RFC 3339, as in create().
     *
     * @param array<string, mixed> $changes
     */
    public function update(int $id, array $changes): Customer
    {
        return $this->endpoint->object(new Request('PUT', self::path($id), body: self::verificationTimes($changes)), self::customer($id));
    }

    /**
     * Finds customer ids by email or mobile; no match is an empty list
     * (GET /v1/customers/get_customer_id).
     *
     * @param array{customer_emails?: list<string>, customer_mobiles?: list<string>} $query
     *
     * @return list<CustomerIdMatch>
     */
    public function lookupIds(array $query): array
    {
        return $this->endpoint->list(new Request('GET', 'v1/customers/get_customer_id', Options::query($query)), CustomerIdMatch::fromFields(...));
    }

    /**
     * Finds customers whose name starts with $prefix; $limit caps the hits
     * (up to 50), 0 means the platform default (GET /v1/customers/get_customer_id_by_name).
     *
     * @return list<CustomerNameMatch>
     */
    public function lookupIdsByName(string $prefix, int $limit = 0): array
    {
        $query = ['customer_name' => $prefix, 'limit' => $limit > 0 ? $limit : null];

        return $this->endpoint->list(new Request('GET', 'v1/customers/get_customer_id_by_name', $query), CustomerNameMatch::fromFields(...));
    }

    /**
     * The shop's extra gender options besides male and female
     * (GET /v1/customers/default_gender_options).
     *
     * @return list<string>
     */
    public function defaultGenderOptions(): array
    {
        $request = new Request('GET', 'v1/customers/default_gender_options');

        return $this->endpoint->list($request, static fn(Fields $f): string => $f->stringOr('default_gender_option'));
    }

    /**
     * One page of customer tags (GET /v1/customers/tags).
     *
     * @param PageQuery $query
     *
     * @return Page<Tag>
     */
    public function listTags(array $query = []): Page
    {
        return $this->client->page(new Request('GET', 'v1/customers/tags', $query), Endpoint::mapper(Tag::fromFields(...)));
    }

    /**
     * The customer's account activation link (GET /v1/customers/{id}/account_activation_url).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function accountActivationUrl(int $id): string
    {
        $request = new Request('GET', self::path($id, 'account_activation_url'));

        return $this->endpoint->one($request, static fn(Fields $f): string => $f->stringOr('account_activation_url'));
    }

    /**
     * Deducts bonus points (POST /v1/customers/{id}/consume_bonus_points).
     * consume_all spends every remaining point and amount is then ignored.
     *
     * @param array{amount?: \Actinc\Cyberbiz\Money, consume_all?: bool} $consume
     */
    public function consumeBonusPoints(int $id, array $consume): void
    {
        $this->endpoint->call(new Request('POST', self::path($id, 'consume_bonus_points'), body: $consume));
    }

    /**
     * Records the referrer code and marketing consent of a customer who
     * registered through a third-party login
     * (POST /v1/customers/{id}/update_register_code_and_accepts_marketing).
     *
     * @param array{secret_key: string, register_code: string, accepts_marketing?: bool} $registration
     */
    public function updateRegisterCode(int $id, array $registration): void
    {
        $this->endpoint->call(new Request('POST', self::path($id, 'update_register_code_and_accepts_marketing'), body: $registration));
    }

    /**
     * The customer's external UID for a provider
     * (GET /v1/customers/{id}/uid_providers/{provider_type}).
     *
     * @param string $providerType "line", "line_at" or "facebook"
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function getUidProvider(int $id, string $providerType): CustomerUidLookup
    {
        return $this->endpoint->one(new Request('GET', self::uidPath($id, $providerType)), CustomerUidLookup::fromFields(...));
    }

    /**
     * Creates or replaces the customer's external UID for a provider
     * (PUT /v1/customers/{id}/uid_providers/{provider_type}).
     *
     * @param string $providerType "line", "line_at" or "facebook"
     */
    public function setUidProvider(int $id, string $providerType, string $uid): void
    {
        $this->endpoint->call(new Request('PUT', self::uidPath($id, $providerType), body: ['uid' => $uid]));
    }

    /**
     * The customer's VIP membership state (GET /v1/customers/{id}/vip_info).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function vipInfo(int $id): CustomerVipInfo
    {
        return $this->endpoint->one(new Request('GET', self::path($id, 'vip_info')), CustomerVipInfo::fromFields(...));
    }

    /**
     * The customer's paid and valid orders between two dates, both required
     * (GET /v1/customers/{id}/spending_overview).
     *
     * @param array{start_date: DateValue, end_date: DateValue} $query
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function spendingOverview(int $id, array $query): CustomerSpendingOverview
    {
        $request = new Request('GET', self::path($id, 'spending_overview'), Options::query($query, Options::DATE));

        return $this->endpoint->one($request, CustomerSpendingOverview::fromFields(...));
    }

    /**
     * The variants waiting in the customer's web-shop cart
     * (GET /v1/customers/{id}/customer_cart_items).
     *
     * @return list<ProductVariant>
     */
    public function cartItems(int $id): array
    {
        $request = new Request('GET', self::path($id, 'customer_cart_items'));

        return $this->endpoint->list($request, static fn(Fields $f): ProductVariant => ProductVariant::fromFields($f));
    }

    /**
     * One page of the customer's service threads (GET /v1/customers/{id}/message_posts).
     *
     * @param PageQuery $query
     *
     * @return Page<CustomerMessagePost>
     */
    public function listMessagePosts(int $id, array $query = []): Page
    {
        return $this->client->page(new Request('GET', self::path($id, 'message_posts'), $query), Endpoint::mapper(CustomerMessagePost::fromFields(...)));
    }

    /**
     * One page of the customer's orders, as full orders (GET /v1/customers/{id}/orders).
     *
     * @param PageQuery $query
     *
     * @return Page<Order>
     */
    public function listOrders(int $id, array $query = []): Page
    {
        return $this->client->page(new Request('GET', self::path($id, 'orders'), $query), Endpoint::mapper(Order::fromFields(...)));
    }

    /**
     * Every order of the customer, page by page (GET /v1/customers/{id}/orders).
     *
     * @param PageQuery $query
     *
     * @return \Generator<int, Order>
     */
    public function allOrders(int $id, array $query = []): \Generator
    {
        return $this->client->each(new Request('GET', self::path($id, 'orders'), $query), Endpoint::mapper(Order::fromFields(...)));
    }

    /**
     * The line items of the customer's recent valid orders in a date range;
     * every option is required (GET /v1/customers/{id}/recent_purchases).
     *
     * @param array{start_date: DateValue, end_date: DateValue, max_products: int} $query
     *
     * @return list<LineItem>
     */
    public function recentPurchases(int $id, array $query): array
    {
        $request = new Request('GET', self::path($id, 'recent_purchases'), Options::query($query, Options::DATE));

        return $this->endpoint->list($request, LineItem::fromFields(...));
    }

    /**
     * One page of customers through the v2 endpoint, which can select by id
     * (at most 100; paging is then ignored) and include "uid_providers",
     * "tags" or "vip_info" (GET /v2/customers).
     *
     * @param ListV2Query $query
     *
     * @return Page<Customer>
     */
    public function listV2(array $query = []): Page
    {
        return $this->client->page(new Request('GET', 'v2/customers', Options::query($query)), Endpoint::mapper(self::customer()));
    }

    /**
     * Every customer through the v2 endpoint, page by page (GET /v2/customers).
     *
     * @param ListV2Query $query
     *
     * @return \Generator<int, Customer>
     */
    public function allV2(array $query = []): \Generator
    {
        return $this->client->each(new Request('GET', 'v2/customers', Options::query($query)), Endpoint::mapper(self::customer()));
    }

    /**
     * The customer linked to an external identity. The platform declares and
     * filters on "provider_type"; "provider" is sent too for compatibility
     * with deployments the SDK was first tested against
     * (GET /v2/customers/by_uid_provider).
     *
     * @param string $providerType "line", "line_at" or "facebook"
     *
     * @throws NotFoundException also for an unknown uid (the API answers 200 null)
     */
    public function getByUidProvider(string $providerType, string $uid): Customer
    {
        $request = new Request('GET', 'v2/customers/by_uid_provider', ['uid' => $uid, 'provider_type' => $providerType, 'provider' => $providerType]);

        return $this->endpoint->one($request, self::customer());
    }

    /**
     * Registers a customer-login OAuth application for the shop and returns
     * its client id and secret; it needs the customer_oauth scope.
     * scopes_description is shown to customers on the consent page;
     * display_shop_name shows the shop name in its title
     * (POST /v2/customer_oauth).
     *
     * @param array{app_name: string, scopes_description: string, description: string, redirect_uri: string, display_shop_name?: bool} $app
     *
     * @throws ApiException also for a {"success":false,"errors":[...]} reply
     */
    public function createOAuthApp(array $app): CustomerOAuthApp
    {
        return $this->endpoint->objectOrFailure(new Request('POST', 'v2/customer_oauth', body: $app), CustomerOAuthApp::fromFields(...), 'client_id');
    }

    private static function path(int $customerId, string $suffix = ''): string
    {
        return 'v1/customers/' . $customerId . ($suffix === '' ? '' : '/' . $suffix);
    }

    private static function uidPath(int $customerId, string $providerType): string
    {
        return self::path($customerId, 'uid_providers/' . rawurlencode($providerType));
    }

    /**
     * Encodes the DateTimeInterface verification timestamps of a body as
     * RFC 3339; null (clears) and strings pass through.
     *
     * @param array<string, mixed> $body
     *
     * @return array<string, mixed>
     */
    private static function verificationTimes(array $body): array
    {
        foreach (self::RFC3339_FIELDS as $key) {
            $value = $body[$key] ?? null;
            if ($value instanceof \DateTimeInterface) {
                $body[$key] = \DateTimeImmutable::createFromInterface($value)->setTimezone(new \DateTimeZone(Time::ZONE))->format(\DATE_RFC3339);
            }
        }

        return $body;
    }

    /** @return \Closure(Fields): Customer */
    private static function customer(?int $id = null): \Closure
    {
        return static fn(Fields $f): Customer => Customer::fromFields($f, $id);
    }
}
