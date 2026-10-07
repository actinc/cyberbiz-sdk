<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Exception\TransportException;
use Actinc\Cyberbiz\Http\Backoff;
use Actinc\Cyberbiz\Http\Clock;
use Actinc\Cyberbiz\Http\ErrorMapper;
use Actinc\Cyberbiz\Http\RateLimiter;
use Actinc\Cyberbiz\Http\SystemClock;
use Actinc\Cyberbiz\Service\CustomersService;
use Actinc\Cyberbiz\Service\Endpoint;
use Actinc\Cyberbiz\Service\OrdersService;
use Actinc\Cyberbiz\Service\ProductsService;
use Actinc\Cyberbiz\Service\ShopService;
use Http\Discovery\Psr17FactoryDiscovery;
use Http\Discovery\Psr18ClientDiscovery;
use Psr\Http\Client\ClientExceptionInterface;
use Psr\Http\Client\ClientInterface;
use Psr\Http\Message\RequestFactoryInterface;
use Psr\Http\Message\RequestInterface;
use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\StreamFactoryInterface;

/**
 * Client talks to the CYBERBIZ API on behalf of exactly one Shop. It applies
 * the platform's rate limit, retries 429 and gateway errors, and turns error
 * responses into typed exceptions.
 */
final class Client
{
    public const VERSION = '0.1.0';

    /** The CYBERBIZ API host shared by every shop. */
    public const DEFAULT_BASE_URI = 'https://app-store-api.cyberbiz.io/';

    /** The platform limit of requests per second. */
    public const DEFAULT_RATE_LIMIT = 5.0;

    /** Retries after a 429 or a transient 5xx. */
    public const DEFAULT_MAX_RETRIES = 3;

    private const RETRYABLE_STATUS = [429, 502, 503, 504];

    private readonly string $token;

    private readonly string $baseUri;

    private readonly ClientInterface $http;

    private readonly RequestFactoryInterface $requests;

    private readonly StreamFactoryInterface $streams;

    private readonly ?RateLimiter $limiter;

    private readonly Clock $clock;

    private readonly Backoff $backoff;

    /**
     * @param string                       $token      the Shop's API token (a Bearer JWT)
     * @param string                       $baseUri    the API host; override it only in tests or behind a proxy
     * @param ClientInterface|null         $httpClient any PSR-18 client; discovered when null
     * @param float                        $rateLimit  requests per second; 0 disables limiting
     * @param int                          $maxRetries retries after 429/502/503/504 or a network error
     *
     * @throws \InvalidArgumentException for an empty token, a bad base URI or negative limits
     * @throws \LogicException           when no PSR-18 client or PSR-17 factory can be discovered
     */
    public function __construct(
        string $token,
        string $baseUri = self::DEFAULT_BASE_URI,
        ?ClientInterface $httpClient = null,
        ?RequestFactoryInterface $requestFactory = null,
        ?StreamFactoryInterface $streamFactory = null,
        float $rateLimit = self::DEFAULT_RATE_LIMIT,
        private readonly int $maxRetries = self::DEFAULT_MAX_RETRIES,
        ?Backoff $backoff = null,
        ?Clock $clock = null,
        private readonly string $userAgent = 'cyberbiz-sdk-php/' . self::VERSION,
    ) {
        $this->token = self::validToken($token);
        $this->baseUri = self::validBaseUri($baseUri);
        if ($rateLimit < 0 || $maxRetries < 0) {
            throw new \InvalidArgumentException('cyberbiz: rate limit and max retries must not be negative');
        }
        [$this->http, $this->requests, $this->streams] = self::transport($httpClient, $requestFactory, $streamFactory);
        $this->clock = $clock ?? new SystemClock();
        $this->limiter = $rateLimit > 0 ? new RateLimiter($rateLimit, $this->clock) : null;
        $this->backoff = $backoff ?? new Backoff();
    }

    /** The API host requests are sent to, always ending in "/". */
    public function baseUri(): string
    {
        return $this->baseUri;
    }

    /** The shop that owns the token, and this app's settings on it. */
    public function shop(): ShopService
    {
        return new ShopService(new Endpoint($this));
    }

    /** Products, variants, options, tags and shipping bindings. */
    public function products(): ProductsService
    {
        return new ProductsService($this, new Endpoint($this));
    }

    /** Orders, fulfillments, payments, returns, e-tickets and shipping labels. */
    public function orders(): OrdersService
    {
        return new OrdersService($this, new Endpoint($this));
    }

    /** Customers, their orders, cart, VIP state and login identities. */
    public function customers(): CustomersService
    {
        return new CustomersService($this, new Endpoint($this));
    }

    /**
     * Sends a request, retrying 429 for every method and 502/503/504 and
     * network errors for idempotent methods only, and returns the 2xx response.
     *
     * @throws ApiException       for an error response that survived every retry
     * @throws TransportException when the last attempt failed before a response
     * @throws \JsonException     when the request body cannot be encoded
     */
    public function send(Request $request): Response
    {
        $psr = $this->build($request);
        for ($attempt = 0; ; ++$attempt) {
            [$response, $error] = $this->attempt($psr);
            if (!$this->shouldRetry($request, $response, $attempt)) {
                break;
            }
            $this->clock->sleep($this->retryDelay($response, $attempt + 1));
        }
        if ($response === null) {
            throw $error ?? new TransportException('cyberbiz: no response');
        }
        ErrorMapper::check($request, $response);

        return $response;
    }

    /**
     * Fetches one page of a list endpoint: the decoded JSON array, mapped
     * item by item, plus the pagination headers.
     *
     * @template T
     *
     * @param (callable(mixed): T)|null $map converts each decoded item; identity when null
     *
     * @return Page<($map is null ? mixed : T)>
     *
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function page(Request $request, ?callable $map = null): Page
    {
        $response = $this->send($request);
        $decoded = $response->isNull() || trim($response->body) === '' ? [] : Json::decode($response->body);
        if (!\is_array($decoded) || !array_is_list($decoded)) {
            throw new DecodeException(\sprintf('cyberbiz: %s %s: expected a JSON array', strtoupper($request->method), $request->path));
        }
        $items = $map === null ? $decoded : array_map($map, $decoded);

        return new Page($items, Pagination::fromResponse($response), $response);
    }

    /**
     * Walks every page of a list endpoint, starting at the request's "page"
     * (default 1, per_page default 50), and yields each item. It follows
     * X-Next-Page and stops at the last page or an empty page, so it sends
     * exactly one request per page.
     *
     * @template T
     *
     * @param (callable(mixed): T)|null $map
     *
     * @return \Generator<int, ($map is null ? mixed : T)>
     *
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function each(Request $request, ?callable $map = null): \Generator
    {
        $start = $request->query['page'] ?? 1;
        $page = max(1, \is_numeric($start) ? (int) $start : 1);
        while (true) {
            $result = $this->page(self::withPage($request, $page), $map);
            yield from $result->items;
            if (!$result->pagination->hasNext() || $result->items === []) {
                return;
            }
            $page = $result->pagination->nextPage;
        }
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

    /** @return array{0: ?Response, 1: ?TransportException} */
    private function attempt(RequestInterface $psr): array
    {
        $this->limiter?->wait();
        try {
            $reply = $this->http->sendRequest($psr);
        } catch (ClientExceptionInterface $e) {
            return [null, new TransportException('cyberbiz: ' . $e->getMessage(), 0, $e)];
        }

        return [new Response($reply->getStatusCode(), self::headers($reply), (string) $reply->getBody()), null];
    }

    /** @return array<string, list<string>> */
    private static function headers(ResponseInterface $reply): array
    {
        $out = [];
        foreach ($reply->getHeaders() as $name => $values) {
            $out[(string) $name] = array_values($values);
        }

        return $out;
    }

    private function shouldRetry(Request $request, ?Response $response, int $attempt): bool
    {
        if ($attempt >= $this->maxRetries) {
            return false;
        }
        if ($response === null) {
            return $request->isIdempotent();
        }

        // A 502/503/504 may arrive after the server already acted on the request, so only a 429
        // (rejected before processing) is safe to repeat for POST and PATCH.
        return $request->isIdempotent()
            ? \in_array($response->statusCode, self::RETRYABLE_STATUS, true)
            : $response->statusCode === 429;
    }

    private function retryDelay(?Response $response, int $attempt): float
    {
        $retryAfter = $response === null ? null : Backoff::retryAfter($response->header('Retry-After'), time());

        return $retryAfter ?? $this->backoff->delay($attempt);
    }

    /** @throws \JsonException */
    private function build(Request $request): RequestInterface
    {
        $psr = $this->requests->createRequest(strtoupper($request->method), $this->uri($request))
            ->withHeader('Authorization', 'Bearer ' . $this->token)
            ->withHeader('Accept', 'application/json')
            ->withHeader('User-Agent', $this->userAgent);
        if ($request->body !== null) {
            $json = \is_string($request->body) ? $request->body : Json::encode($request->body);
            $psr = $psr->withHeader('Content-Type', 'application/json')->withBody($this->streams->createStream($json));
        }
        foreach ($request->headers as $name => $value) {
            $psr = $psr->withHeader($name, $value);
        }

        return $psr;
    }

    private function uri(Request $request): string
    {
        $uri = $this->baseUri . ltrim($request->path, '/');
        $query = Query::encode($request->query);

        return $query === '' ? $uri : $uri . (str_contains($uri, '?') ? '&' : '?') . $query;
    }

    private static function withPage(Request $request, int $page): Request
    {
        $query = ['page' => $page] + $request->query;
        $query['page'] = $page;
        $query['per_page'] ??= Pagination::MAX_PER_PAGE;

        return new Request($request->method, $request->path, $query, $request->body, $request->headers);
    }

    private static function validToken(string $token): string
    {
        if (trim($token) === '') {
            throw new \InvalidArgumentException('cyberbiz: token must not be empty');
        }

        return $token;
    }

    private static function validBaseUri(string $baseUri): string
    {
        $scheme = parse_url($baseUri, \PHP_URL_SCHEME);
        $host = parse_url($baseUri, \PHP_URL_HOST);
        if (!\in_array($scheme, ['http', 'https'], true) || !\is_string($host) || $host === '') {
            throw new \InvalidArgumentException(\sprintf('cyberbiz: base URI must be an absolute http(s) URL, got "%s"', $baseUri));
        }

        return rtrim($baseUri, '/') . '/';
    }

    /** @return array{0: ClientInterface, 1: RequestFactoryInterface, 2: StreamFactoryInterface} */
    private static function transport(?ClientInterface $http, ?RequestFactoryInterface $requests, ?StreamFactoryInterface $streams): array
    {
        try {
            return [
                $http ?? Psr18ClientDiscovery::find(),
                $requests ?? Psr17FactoryDiscovery::findRequestFactory(),
                $streams ?? Psr17FactoryDiscovery::findStreamFactory(),
            ];
        } catch (\Http\Discovery\Exception $e) {
            throw new \LogicException('cyberbiz: install a PSR-18 HTTP client and PSR-17 factories (e.g. guzzlehttp/guzzle), or pass them to Client: ' . $e->getMessage(), 0, $e);
        }
    }
}
