<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The UTM attribution captured at checkout (custom feature). */
final class UtmTracking
{
    public function __construct(
        public readonly string $utmSource,
        public readonly string $utmMedium,
        public readonly string $utmCampaign,
        public readonly string $utmContent,
        public readonly string $utmTerm,
        public readonly ?\DateTimeImmutable $utmClickTime,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('utm_source'),
            $f->stringOr('utm_medium'),
            $f->stringOr('utm_campaign'),
            $f->stringOr('utm_content'),
            $f->stringOr('utm_term'),
            $f->time('utm_click_time'),
        );
    }
}
