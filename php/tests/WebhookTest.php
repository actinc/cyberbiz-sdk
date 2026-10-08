<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Exception\BodyTooLargeException;
use Actinc\Cyberbiz\Exception\InvalidDomainSignatureException;
use Actinc\Cyberbiz\Exception\InvalidSignatureException;
use Actinc\Cyberbiz\Exception\MissingHeaderException;
use Actinc\Cyberbiz\Exception\UnknownShopException;
use Actinc\Cyberbiz\Webhook\EventType;
use Actinc\Cyberbiz\Webhook\Parser;
use Actinc\Cyberbiz\Webhook\SecretMap;
use Actinc\Cyberbiz\Webhook\Signature;
use Actinc\Cyberbiz\Webhook\StaticSecret;
use Nyholm\Psr7\ServerRequest;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class WebhookTest extends TestCase
{
    /** @return iterable<string, array{array{event: string, headers: array<string, string>, base64: string, body: string}}> */
    public static function samples(): iterable
    {
        foreach (WebhookSamples::all() as $event => $sample) {
            yield $event => [$sample];
        }
    }

    /** @return array{event: string, headers: array<string, string>, base64: string, body: string} */
    private static function sample(string $event = 'orders/paid'): array
    {
        return WebhookSamples::all()[$event];
    }

    private static function parser(): Parser
    {
        return new Parser(new StaticSecret(WebhookSamples::SECRET));
    }

    public function testCoversEveryDocumentedEvent(): void
    {
        self::assertSame(array_map(static fn(EventType $t): string => $t->value, EventType::cases()), array_keys(WebhookSamples::all()));
    }

    /** @param array{event: string, headers: array<string, string>, base64: string, body: string} $sample */
    #[DataProvider('samples')]
    public function testAuthenticatesTheHexSampleAndKnowsItsEvent(array $sample): void
    {
        $event = self::parser()->parseRaw($sample['body'], $sample['headers']);

        self::assertSame($sample['event'], $event->type);
        self::assertSame($sample['event'], $event->eventType()?->value);
        self::assertSame('example.cyberbiz.co', $event->shopDomain);
        self::assertSame('www.example-shop.com', $event->customDomain);
        self::assertIsArray($event->decode());
    }

    /** @param array{event: string, headers: array<string, string>, base64: string, body: string} $sample */
    #[DataProvider('samples')]
    public function testAuthenticatesTheBase64FormToo(array $sample): void
    {
        $headers = ['X-Cyberbiz-Hmac-Sha256' => $sample['base64']] + $sample['headers'];

        self::assertSame($sample['event'], self::parser()->parseRaw($sample['body'], $headers)->type);
    }

    public function testAcceptsUppercaseHex(): void
    {
        $s = self::sample();
        $headers = ['X-Cyberbiz-Hmac-Sha256' => strtoupper($s['headers']['X-Cyberbiz-Hmac-Sha256'])] + $s['headers'];

        self::assertSame('orders/paid', self::parser()->parseRaw($s['body'], $headers)->type);
    }

    public function testRejectsATamperedBody(): void
    {
        $s = self::sample();

        $this->expectException(InvalidSignatureException::class);
        self::parser()->parseRaw(str_replace('"id": 1', '"id": 2', $s['body']), $s['headers']);
    }

    public function testRejectsAWrongDomainSignatureSeparately(): void
    {
        $s = self::sample();
        $headers = ['X-Cyberbiz-Domain-Hmac-Sha256' => Signature::signDomain('other.cyberbiz.co', WebhookSamples::SECRET)] + $s['headers'];

        try {
            self::parser()->parseRaw($s['body'], $headers);
            self::fail('expected InvalidDomainSignatureException');
        } catch (InvalidDomainSignatureException) {
            // A distinct class from InvalidSignatureException: the body itself verified.
        }
        $lenient = new Parser(new StaticSecret(WebhookSamples::SECRET), checkDomainSignature: false);
        self::assertSame('orders/paid', $lenient->parseRaw($s['body'], $headers)->type);
    }

    public function testResolvesTheSecretPerShop(): void
    {
        $s = self::sample();
        $parser = new Parser(new SecretMap(['Example.cyberbiz.co' => WebhookSamples::SECRET, 'other.cyberbiz.co' => 'other-secret']));
        self::assertSame('orders/paid', $parser->parseRaw($s['body'], $s['headers'])->type);

        $wrong = new Parser(new SecretMap(['example.cyberbiz.co' => 'other-secret']));
        try {
            $wrong->parseRaw($s['body'], $s['headers']);
            self::fail('expected InvalidSignatureException');
        } catch (InvalidSignatureException) {
        }

        $this->expectException(UnknownShopException::class);
        (new Parser(new SecretMap(['other.cyberbiz.co' => 'other-secret'])))->parseRaw($s['body'], $s['headers']);
    }

    /**
     * Cross-shop replay (CBSDK-39): every shop has its own App Secret, so shop A's delivery resent
     * with shop B's X-Cyberbiz-Domain fails the body Signature under B's secret, whatever Domain
     * Signature it carries. The Domain Signature stays optional (checked only when present).
     *
     * @return iterable<string, array{?string}>
     */
    public static function crossShopDomainSignatures(): iterable
    {
        yield 'no domain signature' => [null];
        yield "shop B's domain signature" => [Signature::signDomain('shop-b.cyberbiz.co', 'shop-b-test-secret')];
        yield "shop A's domain signature" => [Signature::signDomain('shop-a.cyberbiz.co', 'shop-a-test-secret')];
    }

    #[DataProvider('crossShopDomainSignatures')]
    public function testRejectsACrossShopReplay(?string $domainSignature): void
    {
        $this->expectException(InvalidSignatureException::class);
        self::crossShopParser()->parseRaw(...self::shopADelivery('shop-b.cyberbiz.co', $domainSignature));
    }

    public function testAcceptsShopADeliveryUnderPerShopSecrets(): void
    {
        $delivery = self::shopADelivery('shop-a.cyberbiz.co', Signature::signDomain('shop-a.cyberbiz.co', 'shop-a-test-secret'));

        self::assertSame('shop-a.cyberbiz.co', self::crossShopParser()->parseRaw(...$delivery)->shopDomain);
    }

    private static function crossShopParser(): Parser
    {
        return new Parser(new SecretMap(['shop-a.cyberbiz.co' => 'shop-a-test-secret', 'shop-b.cyberbiz.co' => 'shop-b-test-secret']));
    }

    /** @return array{string, array<string, string>} shop A's body and body Signature, labelled with $domain */
    private static function shopADelivery(string $domain, ?string $domainSignature): array
    {
        $body = '{"id":1001,"name":"#1001"}';
        $headers = [
            'X-Cyberbiz-Event' => 'orders/paid',
            'X-Cyberbiz-Domain' => $domain,
            'X-Cyberbiz-Hmac-Sha256' => Signature::sign($body, 'shop-a-test-secret'),
        ];
        if ($domainSignature !== null) {
            $headers['X-Cyberbiz-Domain-Hmac-Sha256'] = $domainSignature;
        }

        return [$body, $headers];
    }

    public function testRequiresTheCyberbizHeaders(): void
    {
        $s = self::sample();
        unset($s['headers']['X-Cyberbiz-Event']);

        $this->expectExceptionObject(new MissingHeaderException('webhook: missing header X-Cyberbiz-Event'));
        self::parser()->parseRaw($s['body'], $s['headers']);
    }

    public function testLimitsTheBodySize(): void
    {
        $s = self::sample();

        $this->expectException(BodyTooLargeException::class);
        (new Parser(new StaticSecret(WebhookSamples::SECRET), maxBodyBytes: 10))->parseRaw($s['body'], $s['headers']);
    }

    public function testAcceptsServerStyleHeaders(): void
    {
        $s = self::sample();
        $server = ['REQUEST_METHOD' => 'POST'];
        foreach ($s['headers'] as $name => $value) {
            $server['HTTP_' . strtoupper(str_replace('-', '_', $name))] = $value;
        }

        self::assertSame('orders/paid', self::parser()->parseRaw($s['body'], $server)->type);
    }

    public function testParsesAPsr7RequestAndRewindsTheBody(): void
    {
        $s = self::sample('products/update');
        $request = new ServerRequest('POST', '/webhooks/cyberbiz', $s['headers'], $s['body']);

        $event = self::parser()->parse($request);

        self::assertSame(EventType::ProductsUpdate, $event->eventType());
        self::assertSame($s['body'], $request->getBody()->getContents());
    }

    public function testSigningMatchesTheDocumentedSamples(): void
    {
        $s = self::sample('apps/uninstall');

        self::assertSame($s['headers']['X-Cyberbiz-Hmac-Sha256'], Signature::sign($s['body'], WebhookSamples::SECRET));
        self::assertSame($s['headers']['X-Cyberbiz-Domain-Hmac-Sha256'], Signature::signDomain('example.cyberbiz.co', WebhookSamples::SECRET));
        self::assertFalse(Signature::verify($s['body'], Signature::sign($s['body'], ''), ''));
        self::assertFalse(Signature::verify($s['body'], '', WebhookSamples::SECRET));
    }

    public function testSplitsEventTypes(): void
    {
        self::assertSame('orders', EventType::OrdersPartialRefunded->resource());
        self::assertSame('partial_refunded', EventType::OrdersPartialRefunded->action());
    }

    public function testKeepsUnknownEventsButReportsNoEventType(): void
    {
        $body = '{"id":1}';
        $headers = [
            'X-Cyberbiz-Event' => 'carts/create',
            'X-Cyberbiz-Domain' => 'example.cyberbiz.co',
            'X-Cyberbiz-Hmac-Sha256' => Signature::sign($body, WebhookSamples::SECRET),
        ];

        $event = self::parser()->parseRaw($body, $headers);

        self::assertSame('carts/create', $event->type);
        self::assertNull($event->eventType());
        self::assertSame(1, $event->fields()->int('id'));
    }
}
