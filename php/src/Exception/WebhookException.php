<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** Every webhook rejection extends this; answer it with a 4xx. */
abstract class WebhookException extends \RuntimeException implements CyberbizException {}
