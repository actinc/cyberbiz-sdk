<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The affiliate or marketplace that referred an order; source is "非導購訂單" when there was none. */
final class LinkedOrderInfo
{
    public function __construct(
        public readonly string $source,
        public readonly string $shopdotcomRid,
        public readonly string $shopdotcomClickId,
        public readonly string $lineShoppingEcid,
        public readonly string $lineShoppingAffiliate,
        public readonly string $ichannelGid,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('source'),
            $f->stringOr('shopdotcom_rid'),
            $f->stringOr('shopdotcom_click_id'),
            $f->stringOr('line_shopping_ecid'),
            $f->stringOr('line_shopping_affiliate'),
            $f->stringOr('ichannel_gid'),
        );
    }
}
