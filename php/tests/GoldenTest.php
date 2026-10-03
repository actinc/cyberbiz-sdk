<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Pagination;
use PHPUnit\Framework\TestCase;

/** Self-tests of the Golden File harness the resource models will use. */
final class GoldenTest extends TestCase
{
    public function testReadsASharedGoldenFile(): void
    {
        [$id, $currency] = Golden::object('app/GET_shop.json', static function (Fields $f): array {
            $shop = $f->object('shop_info');

            return [$shop->int('id'), $shop->string('currency')];
        });

        self::assertSame(26721, $id);
        self::assertSame('TWD', $currency);
    }

    public function testReadsTheRecordedPaginationHeaders(): void
    {
        $pagination = Pagination::fromResponse(Golden::response('v1/GET_v1_products.json'));

        self::assertSame(1, $pagination->page);
        self::assertSame(2, $pagination->perPage);
        self::assertSame(195, $pagination->total);
        self::assertSame(98, $pagination->totalPages);
        self::assertSame(2, $pagination->nextPage);
        self::assertTrue($pagination->hasNext());
    }

    public function testAMismatchNamesTheFileAndTheField(): void
    {
        try {
            Golden::object(__DIR__ . '/fixtures/wrong-type.json', static fn(Fields $f): array => $f->list('line_items', static fn(Fields $i): mixed => $i->money('price')));
            self::fail('expected DecodeException');
        } catch (DecodeException $e) {
            self::assertSame('wrong-type.json: $.line_items[1].price: expected money, got array', $e->getMessage());
        }
    }
}
