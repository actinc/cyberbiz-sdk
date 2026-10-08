<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/**
 * The receiver's App Secrets are configured wrongly; the request itself may be fine. Respond 500
 * so CYBERBIZ retries the delivery once the configuration is fixed. Deliberately not a
 * WebhookException, so code that answers every WebhookException with a 4xx does not catch it.
 */
abstract class WebhookConfigurationException extends \RuntimeException implements CyberbizException {}
