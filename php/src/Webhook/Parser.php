<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

use Actinc\Cyberbiz\Exception\BodyTooLargeException;
use Actinc\Cyberbiz\Exception\InvalidDomainSignatureException;
use Actinc\Cyberbiz\Exception\InvalidSignatureException;
use Actinc\Cyberbiz\Exception\MissingHeaderException;
use Actinc\Cyberbiz\Exception\UnknownShopException;
use Psr\Http\Message\ServerRequestInterface;

/**
 * Authenticates CYBERBIZ App webhooks. Every request carries
 * X-Cyberbiz-Event, X-Cyberbiz-Domain (the Shop Domain) and
 * X-Cyberbiz-Hmac-Sha256 (the body Signature); X-Cyberbiz-Domain-Hmac-Sha256
 * is checked when present.
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

    public function __construct(
        private readonly SecretResolver $secrets,
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
        $secret = $this->secrets->secretFor($shopDomain);
        if ($secret === null || $secret === '') {
            throw new UnknownShopException(\sprintf('webhook: unknown shop "%s"', $shopDomain));
        }
        if (!Signature::verify($body, $signature, $secret)) {
            throw new InvalidSignatureException(\sprintf('webhook: invalid signature for %s from "%s"', $type, $shopDomain));
        }
        $domainSignature = $headers[strtolower(self::HEADER_DOMAIN_SIGNATURE)] ?? '';
        if ($this->checkDomainSignature && $domainSignature !== '' && !Signature::verifyDomain($shopDomain, $domainSignature, $secret)) {
            throw new InvalidDomainSignatureException(\sprintf('webhook: invalid domain signature for "%s"', $shopDomain));
        }
        $customDomain = $headers[strtolower(self::HEADER_SHOP_DOMAIN)] ?? '';

        return new Event($type, $shopDomain, $customDomain, $signature, $domainSignature, $headers, $body);
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
