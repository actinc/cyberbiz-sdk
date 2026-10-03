<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests\Service;

use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Model\ProductVariant;
use Actinc\Cyberbiz\Money;
use Actinc\Cyberbiz\Service\ProductsService;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class ProductsServiceTest extends TestCase
{
    public function testListDecodesTheGoldenPage(): void
    {
        $h = Harness::golden('v1/GET_v1_products.json');

        $page = $h->client->products()->list(['page' => 1, 'per_page' => 2]);

        self::assertSame('GET v1/products?page=1&per_page=2', $h->line());
        self::assertCount(2, $page->items);
        self::assertSame(69458894, $page->items[0]->id);
        self::assertTrue($page->items[0]->price->equals(Money::of('200')));
        self::assertSame(['常溫'], $page->items[0]->temperatureTypes);
        self::assertSame('2026-07-10 20:22:05', $page->items[0]->createdAt?->format('Y-m-d H:i:s'));
        self::assertSame(195, $page->pagination->total);
    }

    public function testAllFollowsTheNextPage(): void
    {
        $h = Harness::json('[{"id":1},{"id":2}]', '[{"id":3}]');

        $ids = array_map(static fn($p): int => $p->id, iterator_to_array($h->client->products()->all(['per_page' => 2]), false));

        self::assertSame([1, 2], $ids, 'no X-Next-Page header, so it stops after one page');
        self::assertSame('GET v1/products?page=1&per_page=2', $h->line());
    }

    public function testGetDecodesTheGoldenProductAndSetsTheId(): void
    {
        $h = Harness::golden('v1/GET_v1_products_{id}.json');

        $product = $h->client->products()->get(69458894);

        self::assertSame('GET v1/products/69458894', $h->line());
        self::assertSame(69458894, $product->id);
        self::assertSame('Hi Rocket Tea CYB 覆蓋 LY 測試', $product->title);
        self::assertNotSame([], $product->productVariants);
        self::assertSame(69458894, $product->productVariants[0]->productId);
    }

    public function testGetTurnsANullBodyIntoNotFound(): void
    {
        $h = Harness::json('null');

        try {
            $h->client->products()->get(5);
            self::fail('expected NotFoundException');
        } catch (NotFoundException $e) {
            self::assertSame(404, $e->statusCode);
            self::assertSame(['resource is null'], $e->messages);
        }
    }

    public function testSearchDecodesTheGoldenFile(): void
    {
        $h = Harness::golden('v1/GET_v1_products_search_query.json');

        $products = $h->client->products()->search(['q' => 'API', 'limit' => 2, 'filter_published' => false]);

        self::assertSame('GET v1/products/search?q=API&limit=2&filter_published=false', $h->line());
        self::assertCount(2, $products);
        self::assertSame(69221930, $products[0]->id);
    }

    public function testSearchCollectionDecodesTheGoldenFile(): void
    {
        $h = Harness::golden('v1/GET_v1_products_search_collection.json');

        $products = $h->client->products()->searchCollection(['collection_handle' => 'frontpage']);

        self::assertSame('GET v1/products/search/collection?collection_handle=frontpage', $h->line());
        self::assertSame(63639028, $products[0]->id);
    }

    public function testVariantGoldenFiles(): void
    {
        $h = Harness::golden(
            'v1/GET_v1_products_{id}_product_variants.json',
            'v1/GET_v1_products_{id}_product_variants_{id}.json',
            'v1/GET_v1_products_sku_{id}_product_variants.json',
        );
        $products = $h->client->products();

        $list = $products->listVariants(69458894);
        $one = $products->getVariant(69458894, 84683248);
        $bySku = $products->listVariantsBySku('SKU-5ac62f', ['per_page' => 2]);

        self::assertSame(84683248, $list[0]->id);
        self::assertSame('SKU-5ac62f', $list[0]->sku);
        self::assertSame(84683248, $one->id);
        self::assertSame(69458894, $one->productId);
        self::assertTrue($one->price->equals(Money::of('200')));
        self::assertSame(80027916, $bySku->items[0]->id);
        self::assertSame('GET v1/products/sku/SKU-5ac62f/product_variants?per_page=2', $h->line(2));
    }

    public function testOptionGoldenFiles(): void
    {
        $h = Harness::golden('v1/GET_v1_products_{id}_product_options.json', 'v1/GET_v1_products_{id}_product_options_{id}.json');
        $products = $h->client->products();

        self::assertSame([], $products->listOptions(1));
        $option = $products->getOption(1, 42);

        self::assertSame(42, $option->id);
        self::assertSame(1, $option->position);
        self::assertSame('頸枕,眼罩', $option->types);
    }

    public function testTagShippingAndDescriptionGoldenFiles(): void
    {
        $h = Harness::golden(
            'v1/GET_v1_products_{id}_product_tags.json',
            'v1/GET_v1_products_bind_shippings.json',
            'v1/GET_v1_products_{id}_bind_shippings.json',
            'v1/GET_v1_products_get_product_description_setting_names.json',
        );
        $products = $h->client->products();

        self::assertSame([], $products->listTags(1));
        self::assertSame(['黑貓宅急便', '門市取貨', '便利袋', '門市取貨（預設）'], $products->listBindableShippings());
        self::assertSame(['門市取貨（預設）', '黑貓宅急便', '門市取貨', '便利袋'], $products->getBindShippings(1));
        self::assertSame('GET v1/products/1/bind_shippings', $h->line(2));
        self::assertSame([], $products->listDescriptionSettingNames());
    }

    public function testEscapesTheSkuInThePath(): void
    {
        $h = Harness::json('[]');

        $h->client->products()->listVariantsBySku('A/B C');

        self::assertSame('GET v1/products/sku/A%2FB%20C/product_variants', $h->line());
    }

    public function testCreateSendsMoneyExactly(): void
    {
        $h = Harness::json('{"id":9,"title":"Tea","price":"12.50"}');

        $product = $h->client->products()->create(['title' => 'Tea', 'handle' => 'tea', 'published' => true, 'price' => Money::of('12.5')]);

        self::assertSame('POST v1/products', $h->line());
        self::assertSame('{"title":"Tea","handle":"tea","published":true,"price":12.50}', (string) $h->request()->getBody());
        self::assertSame(9, $product->id);
    }

    /**
     * Every write and the remaining reads: method, path and body.
     *
     * @param \Closure(ProductsService): mixed $call
     */
    #[DataProvider('requestShapes')]
    public function testRequestShape(\Closure $call, string $reply, string $line, mixed $body): void
    {
        $h = Harness::json($reply);

        $call($h->client->products());

        self::assertSame($line, $h->line());
        self::assertSame($body, (string) $h->request()->getBody() === '' ? null : $h->body());
    }

    /** @return iterable<string, array{\Closure(ProductsService): mixed, string, string, mixed}> */
    public static function requestShapes(): iterable
    {
        yield 'update' => [static fn(ProductsService $s) => $s->update(3, ['title' => 'New']), '{}', 'PUT v1/products/3', ['title' => 'New']];
        yield 'delete' => [static fn(ProductsService $s) => $s->delete(3), '', 'DELETE v1/products/3', null];
        yield 'pos shop batch' => [static fn(ProductsService $s) => $s->createForPosShops(['pos_shop_ids' => [1, 2]]), '[]', 'POST v1/products/pos_shop_batch', ['pos_shop_ids' => [1, 2]]];
        yield 'seo meta tags' => [static fn(ProductsService $s) => $s->updateSeoMetaTags(3, ['title' => 'T']), '{}', 'PUT v1/products/3/seo_meta_tags', ['title' => 'T']];
        yield 'add tags' => [static fn(ProductsService $s) => $s->addTags(3, ['a', 'b']), '[]', 'PUT v1/products/3/product_tags/add', ['tags' => 'a,b']];
        yield 'remove tags' => [static fn(ProductsService $s) => $s->removeTags(3, ['a']), '[]', 'PUT v1/products/3/product_tags/remove', ['tags' => 'a']];
        yield 'bind shippings' => [static fn(ProductsService $s) => $s->bindShippings(3, ['宅配']), '{"shipping_names":["宅配"]}', 'POST v1/products/3/bind_shippings', ['shipping_names' => ['宅配']]];
        yield 'create variant' => [static fn(ProductsService $s) => $s->createVariant(3, ['sku' => 'S1']), '{}', 'POST v1/products/3/product_variants', ['sku' => 'S1']];
        yield 'update variant' => [static fn(ProductsService $s) => $s->updateVariant(3, 4, ['sku' => 'S2']), '{}', 'PUT v1/products/3/product_variants/4', ['sku' => 'S2']];
        yield 'delete variant' => [static fn(ProductsService $s) => $s->deleteVariant(3, 4), '', 'DELETE v1/products/3/product_variants/4', null];
        yield 'create option' => [static fn(ProductsService $s) => $s->createOption(3, ['name' => 'Size']), '{}', 'POST v1/products/3/product_options', ['name' => 'Size']];
        yield 'update option' => [static fn(ProductsService $s) => $s->updateOption(3, 5, ['name' => 'Color']), '{}', 'PUT v1/products/3/product_options/5', ['name' => 'Color']];
        yield 'delete option' => [static fn(ProductsService $s) => $s->deleteOption(3, 5), '', 'DELETE v1/products/3/product_options/5', null];
        yield 'all variants by sku' => [static fn(ProductsService $s) => iterator_to_array($s->allVariantsBySku('S1')), '[]', 'GET v1/products/sku/S1/product_variants?page=1&per_page=50', null];
    }

    public function testUpdateKeepsTheRequestedIds(): void
    {
        $h = Harness::json('{}', '{}', '{}', '{}');
        $products = $h->client->products();

        self::assertSame(3, $products->update(3, [])->id);
        self::assertSame(3, $products->updateSeoMetaTags(3, [])->id);
        self::assertInstanceOf(ProductVariant::class, $variant = $products->updateVariant(3, 4, []));
        self::assertSame(4, $variant->id);
        self::assertSame(5, $products->updateOption(3, 5, [])->id);
    }
}
