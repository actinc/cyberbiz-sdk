<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** A product description section the shop has configured. */
final class ProductDescriptionSettingName
{
    public function __construct(
        public readonly string $settingName,
        public readonly string $title,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('setting_name'),
            $f->stringOr('title'),
        );
    }
}
