<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Service;

use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Exception\TransportException;
use Actinc\Cyberbiz\Model\AppSettings;
use Actinc\Cyberbiz\Model\ShopInfo;
use Actinc\Cyberbiz\Request;

/** The shop that owns the token, and this app's settings on it. */
final class ShopService
{
    public function __construct(private readonly Endpoint $endpoint) {}

    /**
     * The shop profile (GET /shop).
     *
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function info(): ShopInfo
    {
        return $this->endpoint->object(new Request('GET', 'shop'), ShopInfo::fromFields(...), 'shop_info');
    }

    /**
     * The installed app's record, settings included (GET /settings).
     *
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function settings(): AppSettings
    {
        return $this->endpoint->object(new Request('GET', 'settings'), AppSettings::fromFields(...), 'shop_add_on');
    }

    /**
     * Writes setting fields declared in the app manifest (PUT /settings).
     *
     * @param array<string, mixed> $values field name => data
     *
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function updateSettings(array $values): AppSettings
    {
        $settings = [];
        foreach ($values as $field => $data) {
            $settings[] = ['field' => (string) $field, 'data' => $data];
        }
        $request = new Request('PUT', 'settings', body: ['settings' => $settings]);

        return $this->endpoint->object($request, AppSettings::fromFields(...), 'shop_add_on');
    }
}
