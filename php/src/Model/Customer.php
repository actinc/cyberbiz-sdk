<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/**
 * A shop member (GET /v1/customers, /v1/customers/{id}, /v2/customers, and
 * embedded in an order). The v1 detail response omits id;
 * CustomersService::get() fills it in. vipInfo is only present on v2 lists
 * requested with include_params=vip_info.
 */
final class Customer
{
    /**
     * @param string $status "pending", "validate", "enabled", "disabled", "invited", "declined" or "warning"
     * @param string $countryCallingCode e.g. "+886"
     * @param list<Tag> $tags
     * @param string $gender free text; shops define their own options
     * @param \DateTimeImmutable|null $birthday a date (midnight, Asia/Taipei)
     * @param Money $otherAccumulatedConsumption spend accumulated through other channels
     * @param list<CustomField> $customFields
     * @param \DateTimeImmutable|null $confirmedAt email verified at; null when never verified
     * @param \DateTimeImmutable|null $mobileSmsConfirmedAt mobile verified at; null when never verified
     * @param Money $bonusRemain remaining bonus points
     * @param list<CustomerUidProvider> $uidProviders
     */
    public function __construct(
        public readonly int $id,
        public readonly string $name,
        public readonly string $status,
        public readonly string $email,
        public readonly string $countryCallingCode,
        public readonly string $mobile,
        public readonly bool $enableCvsPickup,
        public readonly bool $enableCvsCod,
        public readonly bool $enableHomeDeliveryCod,
        public readonly bool $acceptsMarketing,
        public readonly bool $acceptsEmailNotification,
        public readonly array $tags,
        public readonly ?Address $address,
        public readonly string $gender,
        public readonly ?\DateTimeImmutable $birthday,
        public readonly Money $otherAccumulatedConsumption,
        public readonly ?\DateTimeImmutable $otherAccumulatedConsumptionExpiredAt,
        public readonly string $note,
        public readonly array $customFields,
        public readonly ?\DateTimeImmutable $createdAt,
        public readonly ?\DateTimeImmutable $updatedAt,
        public readonly ?\DateTimeImmutable $confirmedAt,
        public readonly ?\DateTimeImmutable $mobileSmsConfirmedAt,
        public readonly Money $bonusRemain,
        public readonly array $uidProviders,
        public readonly ?CustomerVipInfo $vipInfo,
    ) {}

    /** @param int|null $id overrides the body's id (detail responses may omit it) */
    public static function fromFields(Fields $f, ?int $id = null): self
    {
        return new self(
            $id ?? $f->intOr('id'),
            $f->stringOr('name'),
            $f->stringOr('status'),
            $f->stringOr('email'),
            $f->stringOr('country_calling_code'),
            $f->stringOr('mobile'),
            $f->boolOr('enable_cvs_pickup'),
            $f->boolOr('enable_cvs_cod'),
            $f->boolOr('enable_home_delivery_cod'),
            $f->boolOr('accepts_marketing'),
            $f->boolOr('accepts_email_notification'),
            $f->list('tags', Tag::fromFields(...)),
            Nested::object($f, 'address', Address::fromFields(...)),
            $f->stringOr('gender'),
            $f->time('birthday'),
            $f->moneyOr('other_accumulated_consumption'),
            $f->time('other_accumulated_consumption_expired_at'),
            $f->stringOr('note'),
            $f->list('custom_fields', CustomField::fromFields(...)),
            $f->time('created_at'),
            $f->time('updated_at'),
            $f->time('confirmed_at'),
            $f->time('mobile_sms_confirmed_at'),
            $f->moneyOr('bonus_remain'),
            $f->list('uid_providers', CustomerUidProvider::fromFields(...)),
            Nested::object($f, 'vip_info', CustomerVipInfo::fromFields(...)),
        );
    }
}
