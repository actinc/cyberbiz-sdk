<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The manifest the app was registered with. */
final class AppManifest
{
    /**
     * @param list<string> $webhookEvents
     * @param mixed        $settingFields shape defined by the app
     */
    public function __construct(
        public readonly string $name,
        public readonly string $version,
        public readonly string $scopes,
        public readonly int $manifestVersion,
        public readonly string $type,
        public readonly array $webhookEvents,
        public readonly mixed $settingFields,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('name'),
            $f->stringOr('version'),
            $f->stringOr('scopes'),
            $f->intOr('manifest_version'),
            $f->stringOr('type'),
            $f->strings('webhook_events'),
            $f->raw('setting_fields'),
        );
    }
}
