<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

/**
 * What CYBERBIZ reports in the X-Page, X-Per-Page, X-Offset, X-Total,
 * X-Total-Pages, X-Next-Page and X-Prev-Page headers. nextPage and prevPage
 * are 0 when there is no such page.
 */
final class Pagination
{
    /** The largest page size the platform accepts. */
    public const MAX_PER_PAGE = 50;

    public function __construct(
        public readonly int $page = 0,
        public readonly int $perPage = 0,
        public readonly int $offset = 0,
        public readonly int $total = 0,
        public readonly int $totalPages = 0,
        public readonly int $nextPage = 0,
        public readonly int $prevPage = 0,
    ) {}

    public static function fromResponse(Response $response): self
    {
        $int = static fn(string $name): int => (int) $response->header($name);

        return new self(
            $int('X-Page'),
            $int('X-Per-Page'),
            $int('X-Offset'),
            $int('X-Total'),
            $int('X-Total-Pages'),
            $int('X-Next-Page'),
            $int('X-Prev-Page'),
        );
    }

    public function hasNext(): bool
    {
        return $this->nextPage > 0;
    }
}
