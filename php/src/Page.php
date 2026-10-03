<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

/**
 * One page of a list response.
 *
 * @template T
 */
final class Page
{
    /** @param list<T> $items */
    public function __construct(
        public readonly array $items,
        public readonly Pagination $pagination,
        public readonly Response $response,
    ) {}
}
