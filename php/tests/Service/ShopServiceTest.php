<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Service;

use PHPUnit\Framework\TestCase;

final class ShopServiceTest extends TestCase
{
    public function testInfoDecodesTheGoldenFile(): void
    {
        $h = Harness::golden('app/GET_shop.json');

        $shop = $h->client->shop()->info();

        self::assertSame('GET shop', $h->line());
        self::assertSame(26721, $shop->id);
        self::assertSame('TWD', $shop->currency);
        self::assertSame('zh-TW', $shop->language);
        self::assertSame('example.cyberbiz.co', $shop->primaryDomain);
        self::assertNotNull($shop->shopLine);
        self::assertTrue($shop->shopLine->loginEnable);
        self::assertNull($shop->shopLineChatBot);
    }

    public function testSettingsDecodesTheGoldenFile(): void
    {
        $h = Harness::golden('app/GET_settings.json');

        $settings = $h->client->shop()->settings();

        self::assertSame('GET settings', $h->line());
        self::assertSame(29384, $settings->id);
        self::assertNull($settings->startAt);
        self::assertNotNull($settings->addOnVersion);
        self::assertSame('init', $settings->addOnVersion->status);
        self::assertTrue($settings->addOnVersion->embedded);
        self::assertNotNull($settings->addOnVersion->manifest);
        self::assertNotSame([], $settings->addOnVersion->manifest->webhookEvents);
    }

    public function testSettingsKeepTheAppTokenOutOfDebugOutput(): void
    {
        $h = Harness::json('{"shop_add_on":{"id":1,"token":"synthetic-app-token"}}');

        $settings = $h->client->shop()->settings();

        self::assertSame('synthetic-app-token', $settings->token);
        self::assertStringNotContainsString('synthetic-app-token', print_r($settings, true));
    }

    public function testUpdateSettingsSendsFieldDataPairs(): void
    {
        $h = Harness::json('{"shop_add_on":{"id":7,"settings":{"color":"red"}}}');

        $settings = $h->client->shop()->updateSettings(['color' => 'red', 'limit' => 3]);

        self::assertSame('PUT settings', $h->line());
        self::assertSame(['settings' => [['field' => 'color', 'data' => 'red'], ['field' => 'limit', 'data' => 3]]], $h->body());
        self::assertSame(7, $settings->id);
        self::assertSame(['color' => 'red'], $settings->settings);
    }
}
