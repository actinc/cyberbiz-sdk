<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** A response body, or one field in it, does not have the expected shape. */
final class DecodeException extends \UnexpectedValueException implements CyberbizException {}
