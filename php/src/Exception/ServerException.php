<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** A 5xx that survived every retry. */
final class ServerException extends ApiException {}
