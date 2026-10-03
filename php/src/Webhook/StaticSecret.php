<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/** One App Secret for every Shop (single-shop integrations). */
final class StaticSecret implements SecretResolver
{
    public function __construct(private readonly string $secret) {}

    public function secretFor(string $shopDomain): ?string
    {
        return $this->secret === '' ? null : $this->secret;
    }
}
