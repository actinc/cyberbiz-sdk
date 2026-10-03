<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** A required X-Cyberbiz-* header is absent (respond 400). */
final class MissingHeaderException extends WebhookException {}
