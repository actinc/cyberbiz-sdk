<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The shop's LINE login integration. */
final class ShopLine
{
    public function __construct(
        public readonly bool $loginEnable,
        public readonly string $liffId,
        public readonly bool $liffEnable,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->boolOr('login_enable'),
            $f->stringOr('liff_id'),
            $f->boolOr('liff_enable'),
        );
    }
}
