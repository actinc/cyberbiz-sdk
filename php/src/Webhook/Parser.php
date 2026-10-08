<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

use Actinc\Cyberbiz\Exception\AmbiguousSecretException;
use Actinc\Cyberbiz\Exception\BodyTooLargeException;
use Actinc\Cyberbiz\Exception\InvalidDomainSignatureException;
use Actinc\Cyberbiz\Exception\InvalidSignatureException;
use Actinc\Cyberbiz\Exception\MissingHeaderException;
use Actinc\Cyberbiz\Exception\TooManyCredentialsException;
use Actinc\Cyberbiz\Exception\UnknownShopException;
use Psr\Http\Message\ServerRequestInterface;

/**
 * Authenticates CYBERBIZ App webhooks. Every request carries
 * X-Cyberbiz-Event, X-Cyberbiz-Domain (the Shop Domain) and
 * X-Cyberbiz-Hmac-Sha256 (the body Signature); X-Cyberbiz-Domain-Hmac-Sha256
 * is checked when present.
 *
 * With a CredentialResolver, every candidate App Secret of the Shop is tried in constant time;
 * exactly one must verify the body, and Event::$appId names its App. Two matches throw
 * AmbiguousSecretException, more than MAX_CREDENTIALS candidates TooManyCredentialsException
 * (both configuration errors: respond 500). The Domain Signature is verified with the matched
 * secret.
 */
final class Parser
{
    public const HEADER_EVENT = 'X-Cyberbiz-Event';
    public const HEADER_DOMAIN = 'X-Cyberbiz-Domain';
    public const HEADER_SHOP_DOMAIN = 'X-Cyberbiz-Shop-Domain';
    public const HEADER_SIGNATURE = 'X-Cyberbiz-Hmac-Sha256';
    public const HEADER_DOMAIN_SIGNATURE = 'X-Cyberbiz-Domain-Hmac-Sha256';

    /** The largest body accepted, as in the Go SDK (2 MiB). */
    public const MAX_BODY_BYTES = 2 << 20;

    /** The most candidate App Secrets accepted for one Shop; each costs one HMAC over the body. */
    public const MAX_CREDENTIALS = 16;

    public function __construct(
        private readonly SecretResolver|CredentialResolver $secrets,
        private readonly int $maxBodyBytes = self::MAX_BODY_BYTES,
        private readonly bool $checkDomainSignature = true,
    ) {}

    /** Authenticates a PSR-7 request. The body stream is rewound afterwards when seekable. */
    public function parse(ServerRequestInterface $request): Event
    {
        $stream = $request->getBody();
        $body = (string) $stream;
        if ($stream->isSeekable()) {
            $stream->rewind();
        }
        $headers = [];
        foreach ($request->getHeaders() as $name => $values) {
            $headers[(string) $name] = implode(', ', $values);
        }

        return $this->parseRaw($body, $headers);
    }

    /**
     * Authenticates a raw body and its headers, for frameworks without PSR-7.
     * Header names may be in any case, or in $_SERVER form (HTTP_X_CYBERBIZ_EVENT),
     * so `parseRaw(file_get_contents('php://input'), $_SERVER)` works.
     *
     * @param array<array-key, mixed> $headers
     *
     * @throws MissingHeaderException|BodyTooLargeException|UnknownShopException|InvalidSignatureException|InvalidDomainSignatureException
     * @throws AmbiguousSecretException|TooManyCredentialsException
     */
    public function parseRaw(string $body, array $headers): Event
    {
        $headers = self::normalise($headers);
        $type = self::required($headers, self::HEADER_EVENT);
        $shopDomain = self::required($headers, self::HEADER_DOMAIN);
        $signature = self::required($headers, self::HEADER_SIGNATURE);
        if (\strlen($body) > $this->maxBodyBytes) {
            throw new BodyTooLargeException(\sprintf('webhook: body exceeds %d bytes', $this->maxBodyBytes));
        }
        $credential = self::match($body, $signature, $this->credentials($shopDomain), $type, $shopDomain);
        $domainSignature = $headers[strtolower(self::HEADER_DOMAIN_SIGNATURE)] ?? '';
        if ($this->checkDomainSignature && $domainSignature !== '' && !Signature::verifyDomain($shopDomain, $domainSignature, $credential->secret)) {
            throw new InvalidDomainSignatureException(\sprintf('webhook: invalid domain signature for "%s"', $shopDomain));
        }
        $customDomain = $headers[strtolower(self::HEADER_SHOP_DOMAIN)] ?? '';

        return new Event($type, $shopDomain, $customDomain, $signature, $domainSignature, $headers, $body, $credential->appId);
    }

    /**
     * The candidates with a non-empty secret; a SecretResolver yields one with an empty App ID.
     *
     * @return non-empty-list<Credential>
     */
    private function credentials(string $shopDomain): array
    {
        if ($this->secrets instanceof CredentialResolver) {
            $candidates = $this->secrets->credentialsFor($shopDomain);
            if (\count($candidates) > self::MAX_CREDENTIALS) {
                throw new TooManyCredentialsException(\sprintf('webhook: %d app credentials for "%s", at most %d', \count($candidates), $shopDomain, self::MAX_CREDENTIALS));
            }
        } else {
            $candidates = [new Credential('', $this->secrets->secretFor($shopDomain) ?? '')];
        }
        $usable = array_values(array_filter($candidates, static fn(Credential $c): bool => $c->secret !== ''));
        if ($usable === []) {
            throw new UnknownShopException(\sprintf('webhook: unknown shop "%s"', $shopDomain));
        }

        return $usable;
    }

    /**
     * The one candidate whose secret verifies the body. Every candidate is checked, each in
     * constant time, so the time taken does not depend on which one matched.
     *
     * @param non-empty-list<Credential> $candidates
     */
    private static function match(string $body, string $signature, array $candidates, string $type, string $shopDomain): Credential
    {
        $matched = null;
        $matches = 0;
        foreach ($candidates as $candidate) {
            if (Signature::verify($body, $signature, $candidate->secret)) {
                $matched = $candidate;
                ++$matches;
            }
        }
        if ($matched === null) {
            throw new InvalidSignatureException(\sprintf('webhook: invalid signature for %s from "%s"', $type, $shopDomain));
        }
        if ($matches > 1) {
            throw new AmbiguousSecretException(\sprintf('webhook: %d apps of "%s" share the secret that signed this body', $matches, $shopDomain));
        }

        return $matched;
    }

    /**
     * @param array<array-key, mixed> $headers
     *
     * @return array<string, string> lower-cased dash-separated names
     */
    private static function normalise(array $headers): array
    {
        $out = [];
        foreach ($headers as $name => $value) {
            $name = strtolower((string) $name);
            if (str_starts_with($name, 'http_')) {
                $name = str_replace('_', '-', substr($name, 5));
            }
            $value = \is_array($value) ? implode(', ', array_filter($value, is_string(...))) : $value;
            if (\is_string($value)) {
                $out[$name] = trim($value);
            }
        }

        return $out;
    }

    /** @param array<string, string> $headers */
    private static function required(array $headers, string $name): string
    {
        $value = $headers[strtolower($name)] ?? '';
        if ($value === '') {
            throw new MissingHeaderException('webhook: missing header ' . $name);
        }

        return $value;
    }
}
