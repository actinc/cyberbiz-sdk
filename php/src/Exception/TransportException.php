<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** The request failed before a response arrived (DNS, TLS, timeout, ...). */
final class TransportException extends \RuntimeException implements CyberbizException {}
