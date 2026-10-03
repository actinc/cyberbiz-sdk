<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The profile of the shop that owns the token (GET /shop). */
final class ShopInfo
{
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly string $primaryDomain,
        public readonly string $email,
        public readonly string $currency,
        public readonly string $language,
        public readonly string $ogImageUrl,
        public readonly string $smsPrefix,
        public readonly string $merchantLocation,
        public readonly string $marketLocation,
        public readonly ?ShopLine $shopLine,
        public readonly ?ShopLineChatBot $shopLineChatBot,
    ) {}

    public static function fromFields(Fields $f): self
    {
        $line = $f->objectOrNull('shop_line');
        $bot = $f->objectOrNull('shop_line_chat_bot');

        return new self(
            $f->intOr('id'),
            $f->stringOr('name'),
            $f->stringOr('primary_domain'),
            $f->stringOr('email'),
            $f->stringOr('currency'),
            $f->stringOr('language'),
            $f->stringOr('og_image_url'),
            $f->stringOr('sms_prefix'),
            $f->stringOr('merchant_location'),
            $f->stringOr('market_location'),
            $line === null ? null : ShopLine::fromFields($line),
            $bot === null ? null : ShopLineChatBot::fromFields($bot),
        );
    }
}
