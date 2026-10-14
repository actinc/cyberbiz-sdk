<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Service;

use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Money;
use Actinc\Cyberbiz\Service\CustomersService;
use Actinc\Cyberbiz\Tests\Fake\FakeHttpClient;
use Actinc\Cyberbiz\Time;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class CustomersServiceTest extends TestCase
{
    public function testListDecodesTheGoldenPage(): void
    {
        $h = Harness::golden('v1/GET_v1_customers.json');

        $page = $h->client->customers()->list(['per_page' => 2, 'updated_at_start_time' => new \DateTimeImmutable('2026-01-01 00:00:00', new \DateTimeZone(Time::ZONE))]);

        self::assertSame('GET v1/customers?per_page=2&updated_at_start_time=2026-01-01%2000%3A00%3A00', $h->line());
        self::assertSame(23296, $page->pagination->total);
        $customer = $page->items[0];
        self::assertSame(23300813, $customer->id);
        self::assertSame('enabled', $customer->status);
        self::assertTrue($customer->bonusRemain->equals(Money::of('10012')));
        self::assertSame('2023-10-03 12:04:02', $customer->createdAt?->format('Y-m-d H:i:s'));
        self::assertSame(Time::ZONE, $customer->createdAt->getTimezone()->getName());
        self::assertNull($customer->birthday);
        self::assertSame('line', $customer->uidProviders[0]->providerType);
        self::assertNull($customer->vipInfo);
    }

    public function testGetDecodesTheGoldenCustomerAndSetsTheId(): void
    {
        $h = Harness::golden('v1/GET_v1_customers_{id}.json');

        $customer = $h->client->customers()->get(23300813);

        self::assertSame('GET v1/customers/23300813', $h->line());
        self::assertSame(23300813, $customer->id);
        self::assertSame('2023-11-30 22:30:58', $customer->confirmedAt?->format('Y-m-d H:i:s'));
        self::assertTrue($customer->otherAccumulatedConsumption->equals(Money::zero()));
        self::assertNotNull($customer->address?->detailAddress);
        self::assertSame([], $customer->tags);
    }

    public function testGetTurnsANullBodyIntoNotFound(): void
    {
        $this->expectException(NotFoundException::class);

        Harness::json('null')->client->customers()->get(5);
    }

    public function testLookupGoldenFiles(): void
    {
        $h = Harness::golden(
            'v1/GET_v1_customers_get_customer_id.json',
            'v1/GET_v1_customers_get_customer_id_by_name.json',
            'v1/GET_v1_customers_default_gender_options.json',
            'v1/GET_v1_customers_tags.json',
        );
        $customers = $h->client->customers();

        self::assertSame([], $customers->lookupIds(['customer_emails' => ['a@example.com', 'b@example.com']]));
        $byName = $customers->lookupIdsByName('Syn', 10);
        self::assertSame(['', ''], $customers->defaultGenderOptions());
        $tags = $customers->listTags(['per_page' => 2]);

        self::assertSame('GET v1/customers/get_customer_id?customer_emails=a%40example.com%2Cb%40example.com', $h->line());
        self::assertSame('GET v1/customers/get_customer_id_by_name?customer_name=Syn&limit=10', $h->line(1));
        self::assertSame(40401725, $byName[0]->customerId);
        self::assertSame('GET v1/customers/default_gender_options', $h->line(2));
        self::assertSame('REDACTED', $tags->items[0]->name);
        self::assertFalse($tags->pagination->hasNext());
    }

    public function testDetailGoldenFiles(): void
    {
        $h = Harness::golden(
            'v1/GET_v1_customers_{id}_account_activation_url.json',
            'v1/GET_v1_customers_{id}_uid_providers_line.json',
            'v1/GET_v1_customers_{id}_vip_info.json',
            'v1/GET_v1_customers_{id}_spending_overview.json',
            'v1/GET_v1_customers_{id}_customer_cart_items.json',
            'v1/GET_v1_customers_{id}_message_posts.json',
        );
        $customers = $h->client->customers();

        self::assertStringStartsWith('http://example.cyberbiz.co/account/customer/activate', $customers->accountActivationUrl(1));
        $uid = $customers->getUidProvider(1, 'line');
        $vip = $customers->vipInfo(1);
        $spending = $customers->spendingOverview(1, ['start_date' => new \DateTimeImmutable('2026-01-01'), 'end_date' => '2026-06-30']);

        self::assertSame(23300813, $uid->customerId);
        self::assertSame('查詢成功', $uid->message);
        self::assertSame(23300813, $vip->customerId);
        self::assertNull($vip->currentGroup);
        self::assertNull($vip->extraInfo?->startAt);
        self::assertTrue($vip->extraInfo?->differenceOfTotalSpentForUpgrade->equals(Money::zero()));
        self::assertSame('GET v1/customers/1/spending_overview?start_date=2026-01-01&end_date=2026-06-30', $h->line(3));
        self::assertTrue($spending->paidAndValidTotalSpent->equals(Money::of('1398')));
        self::assertSame(3, $spending->paidAndValidOrdersCount);
        self::assertTrue($spending->paidAndValidAverageSpent->equals(Money::of('466')));
        self::assertSame([], $customers->cartItems(1));
        self::assertSame([], $customers->listMessagePosts(1)->items);
    }

    public function testOrderGoldenFiles(): void
    {
        $h = Harness::golden('v1/GET_v1_customers_{id}_orders.json', 'v1/GET_v1_customers_{id}_recent_purchases.json');
        $customers = $h->client->customers();

        $orders = $customers->listOrders(42, ['per_page' => 2]);
        $recent = $customers->recentPurchases(42, ['start_date' => '2026-01-01', 'end_date' => '2026-03-31', 'max_products' => 5]);

        self::assertSame('GET v1/customers/42/orders?per_page=2', $h->line());
        self::assertSame(31, $orders->pagination->total);
        self::assertSame(49492440, $orders->items[0]->id);
        self::assertSame(1068, $orders->items[0]->orderNumber);
        self::assertTrue($orders->items[0]->subtotalPrice->equals(Money::of('200')));
        self::assertTrue($orders->items[0]->lineItems[0]->discounts[0]->discount->equals(Money::of('866')));
        self::assertSame('bundle_discount', $orders->items[0]->lineItems[0]->discounts[0]->code);
        self::assertSame('GET v1/customers/42/recent_purchases?start_date=2026-01-01&end_date=2026-03-31&max_products=5', $h->line(1));
        self::assertCount(3, $recent);
        self::assertSame(109940716, $recent[0]->id);
        self::assertTrue($recent[0]->price->equals(Money::of('500')));
        self::assertSame('inclusive_tax', $recent[0]->taxTypeId);
        self::assertSame('2026-03-10 11:34:31', $recent[0]->createdAt?->format('Y-m-d H:i:s'));
    }

    public function testV2GoldenFiles(): void
    {
        $h = Harness::golden('v2/GET_v2_customers.json', 'v2/GET_v2_customers_include.json', 'v2/GET_v2_customers_by_uid_provider.json');
        $customers = $h->client->customers();

        $page = $customers->listV2(['per_page' => 2]);
        $selected = $customers->listV2(['ids' => [23300813, 24077804], 'include_params' => ['tags', 'vip_info']]);

        self::assertSame(23300813, $page->items[0]->id);
        self::assertSame(2, $page->pagination->nextPage);
        self::assertSame('GET v2/customers?ids=23300813%2C24077804&include_params=tags%2Cvip_info', $h->line(1));
        self::assertSame(24077804, $selected->items[1]->id);
        try {
            $customers->getByUidProvider('line', 'U-synthetic');
            self::fail('expected NotFoundException');
        } catch (NotFoundException) {
            self::assertSame('GET v2/customers/by_uid_provider?uid=U-synthetic&provider_type=line&provider=line', $h->line(2));
        }
    }

    public function testDecodesSyntheticVipAndMessages(): void
    {
        $h = Harness::json(
            '{"customer_id":1,"current_group":{"id":2,"name":"Gold","customer_tags":["vip"]},"current_level":{"id":3,"bonus_point_enabled":true,"upgrade_condition_total_spent":5000.0},"next_level":null,"extra_info":{"start_at":"2026-01-01 00:00:00","difference_of_total_spent_for_upgrade":1200}}',
            '[{"id":4,"title":"Where is my parcel","status":"replied","comments":[{"id":5,"role":"admin","content":"Shipped","created_at":"2026-02-01 10:00:00"}],"created_at":"2026-02-01 09:00:00"}]',
            '[{"id":6,"product_id":7,"price":120.0,"sku":"SKU-000001","photo_urls":["https://example.com/a.jpg"]}]',
        );
        $customers = $h->client->customers();

        $vip = $customers->vipInfo(1);
        self::assertSame(['vip'], $vip->currentGroup?->customerTags);
        self::assertTrue($vip->currentLevel?->upgradeConditionTotalSpent->equals(Money::of('5000')));
        self::assertNull($vip->nextLevel);
        self::assertTrue($vip->extraInfo?->differenceOfTotalSpentForUpgrade->equals(Money::of('1200')));
        self::assertSame('Shipped', $customers->listMessagePosts(1)->items[0]->comments[0]->content);
        self::assertSame(7, $customers->cartItems(1)[0]->productId);
    }

    public function testGetByUidProviderSendsProviderTypeAndProvider(): void
    {
        $h = Harness::json('{"id":1,"uid_providers":[{"provider_type":"line","uid":"U1"}]}');

        $customer = $h->client->customers()->getByUidProvider('line', 'U1');

        self::assertSame(1, $customer->id);
        self::assertSame('GET v2/customers/by_uid_provider?uid=U1&provider_type=line&provider=line', $h->line());
    }

    public function testCreateOAuthAppSendsTheAppAndReturnsItsCredentials(): void
    {
        $h = new Harness(new FakeHttpClient(null, FakeHttpClient::json(201, '{"client_id":"cid","client_secret":"sec"}')));

        $app = $h->client->customers()->createOAuthApp([
            'app_name' => '會員登入',
            'scopes_description' => '讀取會員資料',
            'description' => 'd',
            'redirect_uri' => 'https://example.com/cb',
            'display_shop_name' => false,
        ]);

        self::assertSame('cid', $app->clientId);
        self::assertSame('sec', $app->clientSecret);
        self::assertSame('POST v2/customer_oauth', $h->line());
        self::assertSame('{"app_name":"會員登入","scopes_description":"讀取會員資料","description":"d","redirect_uri":"https://example.com/cb","display_shop_name":false}', (string) $h->request()->getBody());
    }

    public function testCreateOAuthAppTurnsSuccessFalseIntoAnApiException(): void
    {
        $h = Harness::json('{"success":false,"errors":["redirect uri invalid"]}');

        try {
            $h->client->customers()->createOAuthApp(['app_name' => 'a', 'scopes_description' => 's', 'description' => 'd', 'redirect_uri' => 'x']);
            self::fail('expected ApiException');
        } catch (ApiException $e) {
            self::assertSame(200, $e->statusCode);
            self::assertSame('POST', $e->method);
            self::assertSame('/v2/customer_oauth', $e->path);
            self::assertSame(['redirect uri invalid'], $e->messages);
        }
    }

    public function testVerificationTimesAreSentAsRfc3339(): void
    {
        $h = Harness::json('{"id":1}', '{"id":1}');
        $at = new \DateTimeImmutable('2026-01-01 19:04:05', new \DateTimeZone('UTC'));
        $changes = ['mobile_sms_confirmed_at' => $at, 'confirmed_at' => null];

        $h->client->customers()->create(['name' => 'n', 'confirmed_at' => $at]);
        $h->client->customers()->update(1, $changes);

        self::assertSame('{"name":"n","confirmed_at":"2026-01-02T03:04:05+08:00"}', (string) $h->request()->getBody());
        self::assertSame('{"mobile_sms_confirmed_at":"2026-01-02T03:04:05+08:00","confirmed_at":null}', (string) $h->request(1)->getBody());
        self::assertSame($at, $changes['mobile_sms_confirmed_at'], 'the caller\'s array is untouched');
    }

    public function testMessagePostsDecodeOrderAndAdminObjects(): void
    {
        $h = Harness::json('[{"id":1,"order":{"id":2,"type":"Order","name":"#1001"},"comments":[{"id":3,"role":"admin","admin":{"id":4,"name":"Ops","email":"ops@example.com"}},{"id":5,"role":"customer","admin":null}]},{"id":6,"order":null}]');

        $posts = $h->client->customers()->listMessagePosts(1)->items;

        self::assertSame(2, $posts[0]->order?->id);
        self::assertSame('Order', $posts[0]->order->type);
        self::assertSame('#1001', $posts[0]->order->name);
        self::assertSame(4, $posts[0]->comments[0]->admin?->id);
        self::assertSame('Ops', $posts[0]->comments[0]->admin->name);
        self::assertSame('ops@example.com', $posts[0]->comments[0]->admin->email);
        self::assertNull($posts[0]->comments[1]->admin);
        self::assertNull($posts[1]->order);
    }

    public function testUpdateKeepsTheIdAndSendsExplicitNulls(): void
    {
        $h = Harness::json('{"name":"Synthetic"}');

        $customer = $h->client->customers()->update(8, ['confirmed_at' => null, 'other_accumulated_consumption' => Money::of('0')]);

        self::assertSame(8, $customer->id);
        self::assertSame('PUT v1/customers/8', $h->line());
        self::assertSame('{"confirmed_at":null,"other_accumulated_consumption":0.00}', (string) $h->request()->getBody());
    }

    /**
     * Every remaining method: method, path, query and body.
     *
     * @param \Closure(CustomersService): mixed $call
     */
    #[DataProvider('requestShapes')]
    public function testRequestShape(\Closure $call, string $reply, string $line, mixed $body): void
    {
        $h = Harness::json($reply);

        $call($h->client->customers());

        self::assertSame($line, $h->line());
        self::assertSame($body, (string) $h->request()->getBody() === '' ? null : $h->body());
    }

    /** @return iterable<string, array{\Closure(CustomersService): mixed, string, string, mixed}> */
    public static function requestShapes(): iterable
    {
        yield 'all' => [static fn(CustomersService $s) => iterator_to_array($s->all()), '[]', 'GET v1/customers?page=1&per_page=50', null];
        yield 'create' => [static fn(CustomersService $s) => $s->create(['name' => 'Synthetic', 'password' => 'not-a-real-secret', 'enable_cvs_pickup' => false, 'accepts_marketing' => false]), '{"id":1}', 'POST v1/customers', ['name' => 'Synthetic', 'password' => 'not-a-real-secret', 'enable_cvs_pickup' => false, 'accepts_marketing' => false]];
        yield 'lookup by mobile' => [static fn(CustomersService $s) => $s->lookupIds(['customer_mobiles' => ['0900000000']]), '[]', 'GET v1/customers/get_customer_id?customer_mobiles=0900000000', null];
        yield 'lookup by name without limit' => [static fn(CustomersService $s) => $s->lookupIdsByName('A B'), '[]', 'GET v1/customers/get_customer_id_by_name?customer_name=A%20B', null];
        yield 'consume bonus' => [static fn(CustomersService $s) => $s->consumeBonusPoints(3, ['consume_all' => true]), '', 'POST v1/customers/3/consume_bonus_points', ['consume_all' => true]];
        yield 'register code' => [static fn(CustomersService $s) => $s->updateRegisterCode(3, ['secret_key' => 'k', 'register_code' => 'R1', 'accepts_marketing' => false]), '', 'POST v1/customers/3/update_register_code_and_accepts_marketing', ['secret_key' => 'k', 'register_code' => 'R1', 'accepts_marketing' => false]];
        yield 'set uid' => [static fn(CustomersService $s) => $s->setUidProvider(3, 'line_at', 'U1'), '', 'PUT v1/customers/3/uid_providers/line_at', ['uid' => 'U1']];
        yield 'all orders' => [static fn(CustomersService $s) => iterator_to_array($s->allOrders(3, ['per_page' => 10])), '[]', 'GET v1/customers/3/orders?page=1&per_page=10', null];
        yield 'message posts page' => [static fn(CustomersService $s) => $s->listMessagePosts(3, ['page' => 2]), '[]', 'GET v1/customers/3/message_posts?page=2', null];
        yield 'all v2' => [static fn(CustomersService $s) => iterator_to_array($s->allV2(['include_params' => ['tags']])), '[]', 'GET v2/customers?page=1&include_params=tags&per_page=50', null];
        yield 'escapes provider type' => [static fn(CustomersService $s) => $s->getUidProvider(3, 'a/b'), '{}', 'GET v1/customers/3/uid_providers/a%2Fb', null];
    }
}
