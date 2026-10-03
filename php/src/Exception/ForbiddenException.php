<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** A 403: a missing token scope or plugin, or an action the resource's state forbids. */
final class ForbiddenException extends ApiException {}
