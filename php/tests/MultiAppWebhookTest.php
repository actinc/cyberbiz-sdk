<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Exception\AmbiguousSecretException;
use Actinc\Cyberbiz\Exception\InvalidDomainSignatureException;
use Actinc\Cyberbiz\Exception\InvalidSignatureException;
use Actinc\Cyberbiz\Exception\TooManyCredentialsException;
use Actinc\Cyberbiz\Exception\UnknownShopException;
use Actinc\Cyberbiz\Webhook\AppSecrets;
use Actinc\Cyberbiz\Webhook\Credential;
use Actinc\Cyberbiz\Webhook\CredentialResolver;
use Actinc\Cyberbiz\Webhook\Parser;
use Actinc\Cyberbiz\Webhook\Signature;
use Actinc\Cyberbiz\Webhook\StaticSecret;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

/**
 * Several Apps on one Shop (CBSDK-41): CYBERBIZ sends no App identifier, so the App is the one
 * whose secret verifies the body Signature.
 */
final class MultiAppWebhookTest extends TestCase
{
    private const SHOP = 'shop-a.cyberbiz.co';
    private const SECRET_A = 'app-a-test-secret';
    private const SECRET_B = 'app-b-test-secret';
    private const BODY = '{"id":1001,"name":"#1001"}';

    private static function twoApps(): Parser
    {
        return new Parser(new AppSecrets([self::SHOP => ['app-a' => self::SECRET_A, 'app-b' => self::SECRET_B]]));
    }

    /** @return array<string, string> a delivery for SHOP signed with $secret */
    private static function headers(string $secret, ?string $domainSignature = null): array
    {
        $headers = [
            'X-Cyberbiz-Event' => 'orders/paid',
            'X-Cyberbiz-Domain' => self::SHOP,
            'X-Cyberbiz-Hmac-Sha256' => Signature::sign(self::BODY, $secret),
        ];
        if ($domainSignature !== null) {
            $headers['X-Cyberbiz-Domain-Hmac-Sha256'] = $domainSignature;
        }

        return $headers;
    }

    /** @param list<Credential> $credentials */
    private static function resolver(array $credentials): CredentialResolver
    {
        return new class ($credentials) implements CredentialResolver {
            /** @param list<Credential> $credentials */
            public function __construct(private readonly array $credentials) {}

            public function credentialsFor(string $shopDomain): array
            {
                return $this->credentials;
            }
        };
    }

    /** @return list<Credential> */
    private static function numbered(int $count): array
    {
        $out = [];
        for ($i = 0; $i < $count; ++$i) {
            $out[] = new Credential('app-' . $i, 'secret-' . $i);
        }

        return $out;
    }

    /** AC1 */
    public function testIdentifiesTheAppWhoseSecretVerified(): void
    {
        self::assertSame('app-b', self::twoApps()->parseRaw(self::BODY, self::headers(self::SECRET_B))->appId);
        self::assertSame('app-a', self::twoApps()->parseRaw(self::BODY, self::headers(self::SECRET_A))->appId);
    }

    /** AC2 */
    public function testRejectsWhenNoCandidateVerifies(): void
    {
        $this->expectException(InvalidSignatureException::class);
        self::twoApps()->parseRaw(self::BODY, self::headers('app-c-test-secret'));
    }

    /** AC3 */
    public function testRejectsTwoAppsSharingASecret(): void
    {
        $parser = new Parser(new AppSecrets([self::SHOP => ['app-a' => self::SECRET_A, 'app-b' => self::SECRET_A]]));
        $this->expectException(AmbiguousSecretException::class);
        $parser->parseRaw(self::BODY, self::headers(self::SECRET_A));
    }

    /** AC5 */
    public function testRejectsMoreCandidatesThanTheCap(): void
    {
        $credentials = self::numbered(Parser::MAX_CREDENTIALS + 1);
        $this->expectException(TooManyCredentialsException::class);
        (new Parser(self::resolver($credentials)))->parseRaw(self::BODY, self::headers('secret-0'));
    }

    public function testAcceptsExactlyTheCap(): void
    {
        $parser = new Parser(self::resolver(self::numbered(Parser::MAX_CREDENTIALS)));

        self::assertSame('app-15', $parser->parseRaw(self::BODY, self::headers('secret-15'))->appId);
    }

    public function testVerifiesTheDomainSignatureWithTheMatchedSecret(): void
    {
        $ok = self::headers(self::SECRET_B, Signature::signDomain(self::SHOP, self::SECRET_B));
        self::assertSame('app-b', self::twoApps()->parseRaw(self::BODY, $ok)->appId);

        $this->expectException(InvalidDomainSignatureException::class);
        self::twoApps()->parseRaw(self::BODY, self::headers(self::SECRET_B, Signature::signDomain(self::SHOP, self::SECRET_A)));
    }

    public function testIgnoresEmptySecrets(): void
    {
        $parser = new Parser(self::resolver([new Credential('app-a', ''), new Credential('app-b', self::SECRET_B)]));

        self::assertSame('app-b', $parser->parseRaw(self::BODY, self::headers(self::SECRET_B))->appId);
    }

    /** @return iterable<string, array{CredentialResolver}> */
    public static function emptyResolvers(): iterable
    {
        yield 'no credentials' => [self::resolver([])];
        yield 'only empty secrets' => [self::resolver([new Credential('app-a', ''), new Credential('app-b', '')])];
        yield 'unknown shop' => [new AppSecrets(['shop-b.cyberbiz.co' => ['app-a' => self::SECRET_A]])];
    }

    #[DataProvider('emptyResolvers')]
    public function testFailsClosedWithoutSecrets(CredentialResolver $resolver): void
    {
        $this->expectException(UnknownShopException::class);
        (new Parser($resolver))->parseRaw(self::BODY, self::headers(self::SECRET_A));
    }

    public function testASecretResolverLeavesTheAppIdEmpty(): void
    {
        $event = (new Parser(new StaticSecret(self::SECRET_A)))->parseRaw(self::BODY, self::headers(self::SECRET_A));

        self::assertSame('', $event->appId);
    }

    public function testAppSecretsMatchesDomainsIgnoringCase(): void
    {
        $parser = new Parser(new AppSecrets(['Shop-A.cyberbiz.co' => ['app-b' => self::SECRET_B]]));

        self::assertSame('app-b', $parser->parseRaw(self::BODY, self::headers(self::SECRET_B))->appId);
    }

    public function testNeverDumpsTheSecrets(): void
    {
        $dump = print_r(new Credential('app-a', self::SECRET_A), true)
            . print_r(new AppSecrets([self::SHOP => ['app-a' => self::SECRET_A]]), true);

        self::assertStringNotContainsString(self::SECRET_A, $dump);
        self::assertStringContainsString('app-a', $dump);
    }
}
