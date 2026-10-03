<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The shop's LINE chat bot, when one is linked. */
final class ShopLineChatBot
{
    public function __construct(
        public readonly string $botId,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('bot_id'),
        );
    }
}
