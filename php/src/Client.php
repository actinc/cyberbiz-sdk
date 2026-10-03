<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

/**
 * Client is the entry point to the CYBERBIZ API for one Shop.
 *
 * It holds the configuration every request needs. Sending requests, rate
 * limiting and retries are added on top of it.
 */
final class Client
{
    /** The CYBERBIZ API host shared by every shop. */
    public const DEFAULT_BASE_URI = 'https://app-store-api.cyberbiz.io/';

    private readonly string $token;

    private readonly string $baseUri;

    /**
     * @param string $token   the Shop's API token (a Bearer JWT)
     * @param string $baseUri the API host; override it only in tests
     *
     * @throws \InvalidArgumentException when the token is empty or the base
     *                                   URI is not an absolute http(s) URL
     */
    public function __construct(string $token, string $baseUri = self::DEFAULT_BASE_URI)
    {
        if (trim($token) === '') {
            throw new \InvalidArgumentException('cyberbiz: token must not be empty');
        }
        $scheme = parse_url($baseUri, PHP_URL_SCHEME);
        $host = parse_url($baseUri, PHP_URL_HOST);
        if (!\in_array($scheme, ['http', 'https'], true) || !\is_string($host) || $host === '') {
            throw new \InvalidArgumentException(\sprintf('cyberbiz: base URI must be an absolute http(s) URL, got "%s"', $baseUri));
        }

        $this->token = $token;
        $this->baseUri = rtrim($baseUri, '/') . '/';
    }

    /** The API host requests are sent to, always ending in "/". */
    public function baseUri(): string
    {
        return $this->baseUri;
    }

    /**
     * The Authorization header value for this Shop's requests.
     *
     * @internal used by the request pipeline; never log it
     */
    public function authorization(): string
    {
        return 'Bearer ' . $this->token;
    }

    /**
     * Keeps the token out of var_dump() and print_r() output.
     *
     * @return array{baseUri: string, token: string}
     */
    public function __debugInfo(): array
    {
        return ['baseUri' => $this->baseUri, 'token' => '***'];
    }
}
