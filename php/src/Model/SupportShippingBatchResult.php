<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;

/** The outcome of a v1 batch home-delivery booking (POST /v1/orders/fulfillments/support_shipping). */
final class SupportShippingBatchResult
{
    /**
     * @param list<SupportShippingTask> $results
     */
    public function __construct(
        public readonly string $requestId,
        public readonly array $results,
    ) {}

    public static function fromFields(Fields $f): self
    {
        return new self(
            $f->stringOr('request_id'),
            $f->list('results', SupportShippingTask::fromFields(...)),
        );
    }
}
