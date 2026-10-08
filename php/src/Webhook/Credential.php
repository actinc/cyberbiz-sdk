<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/**
 * One App's webhook secret for a Shop. CYBERBIZ sends no App identifier header, so the App behind
 * a webhook is identified only by the Credential whose secret verifies the body Signature.
 */
final class Credential
{
    /**
     * @param string $appId  your own identifier for the App, reported as Event::$appId
     * @param string $secret the App Secret; an empty secret never verifies
     */
    public function __construct(
        public readonly string $appId,
        #[\SensitiveParameter]
        public readonly string $secret,
    ) {}

    /**
     * Never shows the secret in var_dump() or print_r().
     *
     * @return array{appId: string, secret: string}
     */
    public function __debugInfo(): array
    {
        return ['appId' => $this->appId, 'secret' => '***'];
    }
}
