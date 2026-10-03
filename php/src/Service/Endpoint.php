<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Service;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Exception\TransportException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Json;
use Actinc\Cyberbiz\Request;
use Actinc\Cyberbiz\Response;

/**
 * Shared request-and-decode steps of the resource services.
 *
 * @internal
 */
final class Endpoint
{
    public function __construct(private readonly Client $client) {}

    /**
     * Sends a request whose response is one JSON object and builds a model
     * from it, optionally unwrapping an envelope key first.
     *
     * @template T
     *
     * @param callable(Fields): T $build
     *
     * @return T
     *
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function object(Request $request, callable $build, ?string $envelope = null): mixed
    {
        $fields = Fields::of(self::decode($this->client->send($request), $request));

        return $build($envelope === null ? $fields : $fields->object($envelope));
    }

    /**
     * Like object() for a GET by id: the API answers some missing resources
     * with 200 and a null body, which becomes a NotFoundException.
     *
     * @template T
     *
     * @param callable(Fields): T $build
     *
     * @return T
     *
     * @throws NotFoundException when the body is null
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function one(Request $request, callable $build): mixed
    {
        $response = $this->client->send($request);
        if ($response->isNull()) {
            throw new NotFoundException(404, $request->method, $request->path, $response->requestId(), ['resource is null'], $response->body);
        }

        return $build(Fields::of(self::decode($response, $request)));
    }

    /**
     * Sends a request whose response is a JSON array (unpaginated) and builds
     * a model from each item.
     *
     * @template T
     *
     * @param callable(Fields): T $build
     *
     * @return list<T>
     *
     * @throws ApiException|TransportException|DecodeException|\JsonException
     */
    public function list(Request $request, callable $build): array
    {
        return $this->client->page($request, self::mapper($build))->items;
    }

    /** Sends a request and ignores its body. */
    public function call(Request $request): Response
    {
        return $this->client->send($request);
    }

    /**
     * Turns a model builder into a mapper over decoded list items, keeping
     * the item index in decoding errors.
     *
     * @template T
     *
     * @param callable(Fields): T $build
     *
     * @return \Closure(mixed): T
     */
    public static function mapper(callable $build): \Closure
    {
        $index = 0;

        return static function (mixed $item) use ($build, &$index): mixed {
            return $build(Fields::of($item, \sprintf('$[%d]', $index++)));
        };
    }

    private static function decode(Response $response, Request $request): mixed
    {
        if (trim($response->body) === '') {
            throw new DecodeException(\sprintf('cyberbiz: %s %s: empty response body', strtoupper($request->method), $request->path));
        }

        return Json::decode($response->body);
    }
}
