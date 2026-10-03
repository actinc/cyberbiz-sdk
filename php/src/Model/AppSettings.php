<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/**
 * The installed app's record for this shop (GET /settings). $token is the
 * app's own credential: never log it or store it in plain text.
 */
final class AppSettings
{
    /** @param mixed $settings shape defined by the app manifest */
    public function __construct(
        public readonly int $id,
        public readonly mixed $settings,
        public readonly string $vendorType,
        public readonly string $token,
        public readonly ?\DateTimeImmutable $startAt,
        public readonly ?\DateTimeImmutable $endAt,
        public readonly ?AddOnVersion $addOnVersion,
    ) {}

    public static function fromFields(Fields $f): self
    {
        $version = $f->objectOrNull('add_on_version');

        return new self(
            $f->intOr('id'),
            $f->raw('settings'),
            $f->stringOr('vendor_type'),
            $f->stringOr('token'),
            $f->time('start_at'),
            $f->time('end_at'),
            $version === null ? null : AddOnVersion::fromFields($version),
        );
    }

    /**
     * Keeps the token out of var_dump() and print_r() output.
     *
     * @return array<string, mixed>
     */
    public function __debugInfo(): array
    {
        return [
            'id' => $this->id,
            'settings' => $this->settings,
            'vendorType' => $this->vendorType,
            'token' => '***',
            'startAt' => $this->startAt,
            'endAt' => $this->endAt,
            'addOnVersion' => $this->addOnVersion,
        ];
    }
}
