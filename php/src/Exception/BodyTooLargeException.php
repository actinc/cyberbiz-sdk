<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** The body exceeds the size limit (respond 413). */
final class BodyTooLargeException extends WebhookException {}
