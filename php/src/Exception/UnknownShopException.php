<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Exception;

/** No App Secret is known for the Shop Domain (respond 401). */
final class UnknownShopException extends WebhookException {}
