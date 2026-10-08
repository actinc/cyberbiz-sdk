<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/** A fixed Shop Domain => App ID => App Secret map (several Apps per Shop). Domains match ignoring case. */
final class AppSecrets implements CredentialResolver
{
    /** @var array<string, list<Credential>> */
    private readonly array $credentials;

    /**
     * @param array<string, array<string, string>> $secrets keyed by Shop Domain, then by App ID, e.g.
     *                                                      ['shop-a.cyberbiz.co' => ['app-a' => '...', 'app-b' => '...']]
     */
    public function __construct(#[\SensitiveParameter] array $secrets)
    {
        $credentials = [];
        foreach ($secrets as $shopDomain => $apps) {
            $key = strtolower((string) $shopDomain);
            foreach ($apps as $appId => $secret) {
                $credentials[$key][] = new Credential((string) $appId, $secret);
            }
        }
        $this->credentials = $credentials;
    }

    public function credentialsFor(string $shopDomain): array
    {
        return $this->credentials[strtolower($shopDomain)] ?? [];
    }

    /**
     * Lists the Shop Domains and App IDs, never the secrets.
     *
     * @return array<string, list<string>>
     */
    public function __debugInfo(): array
    {
        return array_map(
            static fn(array $creds): array => array_map(static fn(Credential $c): string => $c->appId, $creds),
            $this->credentials,
        );
    }
}
