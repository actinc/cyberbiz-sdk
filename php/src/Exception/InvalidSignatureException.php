<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** X-Cyberbiz-Hmac-Sha256 does not match the body (respond 401). */
final class InvalidSignatureException extends WebhookException {}
