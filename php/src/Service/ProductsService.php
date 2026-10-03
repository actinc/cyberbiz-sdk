<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Service;

use Actinc\Cyberbiz\Client;
use Actinc\Cyberbiz\Exception\ApiException;
use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Exception\NotFoundException;
use Actinc\Cyberbiz\Exception\TransportException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Model\Product;
use Actinc\Cyberbiz\Model\ProductDescriptionSettingName;
use Actinc\Cyberbiz\Model\ProductOption;
use Actinc\Cyberbiz\Model\ProductTag;
use Actinc\Cyberbiz\Model\ProductVariant;
use Actinc\Cyberbiz\Page;
use Actinc\Cyberbiz\Request;

/**
 * Products, their variants, options, tags and shipping bindings.
 *
 * Request bodies are associative arrays shaped like the API's JSON (see
 * docs/api/en/cyberbiz-openapi-v1.yaml); Money and DateTimeInterface values
 * are encoded exactly.
 *
 * @throws ApiException|TransportException|DecodeException|\JsonException from every method
 */
final class ProductsService
{
    public function __construct(
        private readonly Client $client,
        private readonly Endpoint $endpoint,
    ) {}

    /**
     * One page of products (GET /v1/products).
     *
     * @param array{page?: int, per_page?: int} $query
     *
     * @return Page<Product>
     */
    public function list(array $query = []): Page
    {
        return $this->client->page(new Request('GET', 'v1/products', $query), Endpoint::mapper(self::product()));
    }

    /**
     * Every product, page by page (GET /v1/products).
     *
     * @param array{page?: int, per_page?: int} $query
     *
     * @return \Generator<int, Product>
     */
    public function all(array $query = []): \Generator
    {
        return $this->client->each(new Request('GET', 'v1/products', $query), Endpoint::mapper(self::product()));
    }

    /**
     * One product (GET /v1/products/{id}).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function get(int $id): Product
    {
        return $this->endpoint->one(new Request('GET', self::path($id)), self::product($id));
    }

    /**
     * Creates a product (POST /v1/products). title, handle, published and
     * price are required.
     *
     * @param array<string, mixed> $product
     */
    public function create(array $product): Product
    {
        return $this->endpoint->object(new Request('POST', 'v1/products', body: $product), self::product());
    }

    /**
     * Changes a product (PUT /v1/products/{id}).
     *
     * @param array<string, mixed> $changes
     */
    public function update(int $id, array $changes): Product
    {
        return $this->endpoint->object(new Request('PUT', self::path($id), body: $changes), self::product($id));
    }

    /** Removes a product (DELETE /v1/products/{id}). */
    public function delete(int $id): void
    {
        $this->endpoint->call(new Request('DELETE', self::path($id)));
    }

    /**
     * Creates the same product in several POS shops (POST /v1/products/pos_shop_batch).
     *
     * @param array<string, mixed> $batch
     *
     * @return list<Product>
     */
    public function createForPosShops(array $batch): array
    {
        return $this->endpoint->list(new Request('POST', 'v1/products/pos_shop_batch', body: $batch), self::product());
    }

    /**
     * Finds products by keyword or vendor (GET /v1/products/search). The
     * endpoint pages with limit and offset and sends no pagination headers.
     *
     * @param array{q?: string, vendor?: string, limit?: int, offset?: int, filter_published?: bool, order_by?: string, filter_branch_store?: bool} $query
     *
     * @return list<Product>
     */
    public function search(array $query): array
    {
        return $this->endpoint->list(new Request('GET', 'v1/products/search', $query), self::product());
    }

    /**
     * The products of a collection (GET /v1/products/search/collection).
     *
     * @param array{collection_handle: string, limit?: int, offset?: int, filter_published?: bool, order_by?: string} $query
     *
     * @return list<Product>
     */
    public function searchCollection(array $query): array
    {
        return $this->endpoint->list(new Request('GET', 'v1/products/search/collection', $query), self::product());
    }

    /**
     * Changes a product's SEO fields (PUT /v1/products/{id}/seo_meta_tags).
     *
     * @param array<string, mixed> $tags
     */
    public function updateSeoMetaTags(int $id, array $tags): Product
    {
        $request = new Request('PUT', self::path($id, 'seo_meta_tags'), body: $tags);

        return $this->endpoint->object($request, self::product($id));
    }

    /**
     * A product's tags (GET /v1/products/{id}/product_tags).
     *
     * @return list<ProductTag>
     */
    public function listTags(int $productId): array
    {
        return $this->endpoint->list(new Request('GET', self::path($productId, 'product_tags')), ProductTag::fromFields(...));
    }

    /**
     * Attaches tags and returns the resulting list (PUT .../product_tags/add).
     *
     * @param list<string> $tags
     *
     * @return list<ProductTag>
     */
    public function addTags(int $productId, array $tags): array
    {
        return $this->putTags(self::path($productId, 'product_tags/add'), $tags);
    }

    /**
     * Detaches tags and returns the resulting list (PUT .../product_tags/remove).
     *
     * @param list<string> $tags
     *
     * @return list<ProductTag>
     */
    public function removeTags(int $productId, array $tags): array
    {
        return $this->putTags(self::path($productId, 'product_tags/remove'), $tags);
    }

    /**
     * Every shipping method a product can be bound to (GET /v1/products/bind_shippings).
     *
     * @return list<string>
     */
    public function listBindableShippings(): array
    {
        return $this->shippingNames(new Request('GET', 'v1/products/bind_shippings'));
    }

    /**
     * The shipping methods bound to a product (GET /v1/products/{id}/bind_shippings).
     *
     * @return list<string>
     */
    public function getBindShippings(int $productId): array
    {
        return $this->shippingNames(new Request('GET', self::path($productId, 'bind_shippings')));
    }

    /**
     * Replaces the shipping methods bound to a product (POST /v1/products/{id}/bind_shippings).
     *
     * @param list<string> $shippingNames
     *
     * @return list<string>
     */
    public function bindShippings(int $productId, array $shippingNames): array
    {
        $body = ['shipping_names' => $shippingNames];

        return $this->shippingNames(new Request('POST', self::path($productId, 'bind_shippings'), body: $body));
    }

    /**
     * The description sections the shop supports
     * (GET /v1/products/get_product_description_setting_names).
     *
     * @return list<ProductDescriptionSettingName>
     */
    public function listDescriptionSettingNames(): array
    {
        $request = new Request('GET', 'v1/products/get_product_description_setting_names');

        return $this->endpoint->list($request, ProductDescriptionSettingName::fromFields(...));
    }

    /**
     * A product's variants (GET /v1/products/{id}/product_variants).
     *
     * @return list<ProductVariant>
     */
    public function listVariants(int $productId): array
    {
        return $this->endpoint->list(new Request('GET', self::path($productId, 'product_variants')), self::variant());
    }

    /**
     * One variant (GET /v1/products/{id}/product_variants/{variant_id}).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function getVariant(int $productId, int $variantId): ProductVariant
    {
        $request = new Request('GET', self::path($productId, 'product_variants/' . $variantId));

        return $this->endpoint->one($request, self::variant($variantId));
    }

    /**
     * Adds a variant (POST /v1/products/{id}/product_variants).
     *
     * @param array<string, mixed> $variant
     */
    public function createVariant(int $productId, array $variant): ProductVariant
    {
        $request = new Request('POST', self::path($productId, 'product_variants'), body: $variant);

        return $this->endpoint->object($request, self::variant());
    }

    /**
     * Changes a variant (PUT /v1/products/{id}/product_variants/{variant_id}).
     *
     * @param array<string, mixed> $changes
     */
    public function updateVariant(int $productId, int $variantId, array $changes): ProductVariant
    {
        $request = new Request('PUT', self::path($productId, 'product_variants/' . $variantId), body: $changes);

        return $this->endpoint->object($request, self::variant($variantId));
    }

    /** Removes a variant (DELETE /v1/products/{id}/product_variants/{variant_id}). */
    public function deleteVariant(int $productId, int $variantId): void
    {
        $this->endpoint->call(new Request('DELETE', self::path($productId, 'product_variants/' . $variantId)));
    }

    /**
     * One page of the variants with a SKU (GET /v1/products/sku/{sku}/product_variants).
     *
     * @param array{page?: int, per_page?: int} $query
     *
     * @return Page<ProductVariant>
     */
    public function listVariantsBySku(string $sku, array $query = []): Page
    {
        return $this->client->page(new Request('GET', self::skuPath($sku), $query), Endpoint::mapper(self::variant()));
    }

    /**
     * Every variant with a SKU, page by page.
     *
     * @param array{page?: int, per_page?: int} $query
     *
     * @return \Generator<int, ProductVariant>
     */
    public function allVariantsBySku(string $sku, array $query = []): \Generator
    {
        return $this->client->each(new Request('GET', self::skuPath($sku), $query), Endpoint::mapper(self::variant()));
    }

    /**
     * A product's options (GET /v1/products/{id}/product_options).
     *
     * @return list<ProductOption>
     */
    public function listOptions(int $productId): array
    {
        return $this->endpoint->list(new Request('GET', self::path($productId, 'product_options')), self::option());
    }

    /**
     * One option (GET /v1/products/{id}/product_options/{option_id}).
     *
     * @throws NotFoundException also when the API answers with a null body
     */
    public function getOption(int $productId, int $optionId): ProductOption
    {
        $request = new Request('GET', self::path($productId, 'product_options/' . $optionId));

        return $this->endpoint->one($request, self::option($optionId));
    }

    /**
     * Adds an option (POST /v1/products/{id}/product_options).
     *
     * @param array<string, mixed> $option
     */
    public function createOption(int $productId, array $option): ProductOption
    {
        $request = new Request('POST', self::path($productId, 'product_options'), body: $option);

        return $this->endpoint->object($request, self::option());
    }

    /**
     * Changes an option (PUT /v1/products/{id}/product_options/{option_id}).
     *
     * @param array<string, mixed> $changes
     */
    public function updateOption(int $productId, int $optionId, array $changes): ProductOption
    {
        $request = new Request('PUT', self::path($productId, 'product_options/' . $optionId), body: $changes);

        return $this->endpoint->object($request, self::option($optionId));
    }

    /** Removes an option (DELETE /v1/products/{id}/product_options/{option_id}). */
    public function deleteOption(int $productId, int $optionId): void
    {
        $this->endpoint->call(new Request('DELETE', self::path($productId, 'product_options/' . $optionId)));
    }

    /**
     * @param list<string> $tags
     *
     * @return list<ProductTag>
     */
    private function putTags(string $path, array $tags): array
    {
        $request = new Request('PUT', $path, body: ['tags' => implode(',', $tags)]);

        return $this->endpoint->list($request, ProductTag::fromFields(...));
    }

    /** @return list<string> */
    private function shippingNames(Request $request): array
    {
        return $this->endpoint->object($request, static fn(Fields $f): array => $f->strings('shipping_names'));
    }

    private static function path(int $productId, string $suffix = ''): string
    {
        return 'v1/products/' . $productId . ($suffix === '' ? '' : '/' . $suffix);
    }

    private static function skuPath(string $sku): string
    {
        return 'v1/products/sku/' . rawurlencode($sku) . '/product_variants';
    }

    /** @return \Closure(Fields): Product */
    private static function product(?int $id = null): \Closure
    {
        return static fn(Fields $f): Product => Product::fromFields($f, $id);
    }

    /** @return \Closure(Fields): ProductVariant */
    private static function variant(?int $id = null): \Closure
    {
        return static fn(Fields $f): ProductVariant => ProductVariant::fromFields($f, $id);
    }

    /** @return \Closure(Fields): ProductOption */
    private static function option(?int $id = null): \Closure
    {
        return static fn(Fields $f): ProductOption => ProductOption::fromFields($f, $id);
    }
}
