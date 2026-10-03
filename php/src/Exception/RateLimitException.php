<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** A 429 that survived every retry. */
final class RateLimitException extends ApiException {}
