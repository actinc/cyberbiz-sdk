<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/** A fixed Shop Domain => App Secret map (multi-shop integrations). */
final class SecretMap implements SecretResolver
{
    /** @var array<string, string> */
    private readonly array $secrets;

    /** @param array<string, string> $secrets keyed by Shop Domain, e.g. "example.cyberbiz.co" */
    public function __construct(array $secrets)
    {
        $this->secrets = array_change_key_case($secrets, \CASE_LOWER);
    }

    public function secretFor(string $shopDomain): ?string
    {
        $secret = $this->secrets[strtolower($shopDomain)] ?? null;

        return $secret === '' ? null : $secret;
    }
}
