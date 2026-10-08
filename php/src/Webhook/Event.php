<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Json;

/** One authenticated webhook. */
final class Event
{
    /**
     * @param string                $type            the X-Cyberbiz-Event value, e.g. "orders/paid"
     * @param string                $shopDomain      the Shop Domain that identifies the Shop
     * @param string                $customDomain    the merchant's storefront hostname; informational only
     * @param string                $domainSignature empty when the header was absent
     * @param array<string, string> $headers         lower-cased header names
     * @param string                $body            the raw body, byte for byte
     * @param string                $appId           the App whose secret verified the body, as named by a
     *                                               CredentialResolver; empty with a SecretResolver
     */
    public function __construct(
        public readonly string $type,
        public readonly string $shopDomain,
        public readonly string $customDomain,
        public readonly string $signature,
        public readonly string $domainSignature,
        public readonly array $headers,
        public readonly string $body,
        public readonly string $appId = '',
    ) {}

    /** The documented event, or null for an event this SDK does not know yet. */
    public function eventType(): ?EventType
    {
        return EventType::tryFrom($this->type);
    }

    /**
     * The decoded body, decimals kept exact (see Json::decode).
     *
     * @throws DecodeException
     */
    public function decode(): mixed
    {
        return Json::decode($this->body);
    }

    /**
     * The body as typed fields, for building payload models.
     *
     * @throws DecodeException when the body is not a JSON object
     */
    public function fields(): Fields
    {
        return Fields::of($this->decode());
    }
}
