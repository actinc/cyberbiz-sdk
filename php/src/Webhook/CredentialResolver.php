<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/**
 * Returns every App Secret configured for a Shop Domain, for one receiver that serves several
 * Apps installed on the same Shop. Pass it to Parser in place of a SecretResolver; Event::$appId
 * then names the App whose secret verified the body.
 */
interface CredentialResolver
{
    /**
     * The candidates for a Shop: at most Parser::MAX_CREDENTIALS. Return an empty list for an
     * unknown Shop; empty secrets are ignored.
     *
     * @return list<Credential>
     */
    public function credentialsFor(string $shopDomain): array;
}
