<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/** Maps a Shop Domain (X-Cyberbiz-Domain) to that Shop's App Secret. */
interface SecretResolver
{
    /** The App Secret, or null when the Shop is unknown. */
    public function secretFor(string $shopDomain): ?string;
}
