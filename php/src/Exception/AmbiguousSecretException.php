<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** The body Signature verified under more than one App's secret: two Apps share a secret (respond 500). */
final class AmbiguousSecretException extends WebhookConfigurationException {}
