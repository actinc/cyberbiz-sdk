<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** A 401: the token is invalid, or the shop has not licensed the feature. */
final class AuthenticationException extends ApiException {}
