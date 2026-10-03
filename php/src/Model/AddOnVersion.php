<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The app version installed on the shop. */
final class AddOnVersion
{
    public function __construct(
        public readonly string $webhookUrl,
        public readonly ?AppManifest $manifest,
        public readonly bool $embedded,
        public readonly string $status,
        public readonly string $scopes,
    ) {}

    public static function fromFields(Fields $f): self
    {
        $manifest = $f->objectOrNull('manifest');

        return new self(
            $f->stringOr('webhook_url'),
            $manifest === null ? null : AppManifest::fromFields($manifest),
            $f->boolOr('embedded'),
            $f->stringOr('status'),
            $f->stringOr('scopes'),
        );
    }
}
