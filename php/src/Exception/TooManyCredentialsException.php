<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** The CredentialResolver returned more than Parser::MAX_CREDENTIALS candidates for one Shop (respond 500). */
final class TooManyCredentialsException extends WebhookConfigurationException {}
