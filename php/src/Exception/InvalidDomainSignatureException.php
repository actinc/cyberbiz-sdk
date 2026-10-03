<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** X-Cyberbiz-Domain-Hmac-Sha256 does not match the Shop Domain (respond 401). */
final class InvalidDomainSignatureException extends WebhookException {}
