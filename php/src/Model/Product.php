<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Model;

use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Money;

/** A product. Some fields appear only in some responses, as noted. */
final class Product
{
    /**
     * @param list<ProductCollectionRef>     $customCollections
     * @param list<ProductTag>               $tags
     * @param list<ProductVariant>           $productVariants
     * @param list<ProductOption>            $productOptions
     * @param list<ProductPhoto>             $photos             detail response only
     * @param list<string>                   $photoUrls          list and search responses only
     * @param list<ProductRelatedCollection> $relatedCollections
     * @param string                         $taxTypeId          "inclusive_tax", "zero_tax" or "exclusive_tax"
     * @param list<string>                   $temperatureTypes   "常溫", "冷藏" or "冷凍"
     * @param mixed                          $productCustomFields
     * @param list<string>                   $requiredCustomerTags
     */
    public function __construct(
        public readonly int $id,
        public readonly string $title,
        public readonly string $handle,
        public readonly string $englishTitle,
        public readonly string $productUrl,
        public readonly bool $published,
        public readonly ?\DateTimeImmutable $sellFrom,
        public readonly ?\DateTimeImmutable $sellTo,
        public readonly string $productType,
        public readonly string $productTypeCode,
        public readonly string $slogan,
        public readonly string $brief,
        public readonly string $briefText,
        public readonly bool $briefIncludesHtml,
        public readonly string $bodyHtml,
        public readonly string $vendor,
        public readonly Money $price,
        public readonly float $sellWeight,
        public readonly string $taxTypeId,
        public readonly array $customCollections,
        public readonly ?ProductSpecialCollectionRef $specialCollection,
        public readonly array $tags,
        public readonly array $productVariants,
        public readonly array $productOptions,
        public readonly ?ProductPosShopRef $posShop,
        public readonly array $photos,
        public readonly array $photoUrls,
        public readonly ?ProductChannel $channel,
        public readonly array $relatedCollections,
        public readonly ?ProductBranchStoreRef $branchStore,
        public readonly ?\DateTimeImmutable $createdAt,
        public readonly ?\DateTimeImmutable $updatedAt,
        public readonly array $temperatureTypes,
        public readonly bool $searchable,
        public readonly int $googleProductCategoryId,
        public readonly mixed $productCustomFields,
        public readonly ?ProductSeoMetaTags $seoMetaTags,
        public readonly array $requiredCustomerTags,
    ) {}

    /** @param int|null $id overrides the body's id (detail responses may omit it) */
    public static function fromFields(Fields $f, ?int $id = null): self
    {
        return new self(
            $id ?? $f->intOr('id'),
            $f->stringOr('title'),
            $f->stringOr('handle'),
            $f->stringOr('english_title'),
            $f->stringOr('product_url'),
            $f->boolOr('published'),
            $f->time('sell_from'),
            $f->time('sell_to'),
            $f->stringOr('product_type'),
            $f->stringOr('product_type_code'),
            $f->stringOr('slogan'),
            $f->stringOr('brief'),
            $f->stringOr('brief_text'),
            $f->boolOr('brief_includes_html'),
            $f->stringOr('body_html'),
            $f->stringOr('vendor'),
            $f->moneyOr('price'),
            $f->float('sell_weight'),
            $f->stringOr('tax_type_id'),
            $f->list('custom_collections', ProductCollectionRef::fromFields(...)),
            self::optional($f, 'special_collection', ProductSpecialCollectionRef::fromFields(...)),
            $f->list('tags', ProductTag::fromFields(...)),
            $f->list('product_variants', static fn(Fields $v): ProductVariant => ProductVariant::fromFields($v)),
            $f->list('product_options', ProductOption::fromFields(...)),
            self::optional($f, 'pos_shop', ProductPosShopRef::fromFields(...)),
            $f->list('photos', ProductPhoto::fromFields(...)),
            $f->strings('photo_urls'),
            self::optional($f, 'channel', ProductChannel::fromFields(...)),
            $f->list('related_collections', ProductRelatedCollection::fromFields(...)),
            self::optional($f, 'branch_store', ProductBranchStoreRef::fromFields(...)),
            $f->time('created_at'),
            $f->time('updated_at'),
            $f->strings('temperature_types'),
            $f->boolOr('searchable'),
            $f->intOr('google_product_category_id'),
            $f->raw('product_custom_fields'),
            self::optional($f, 'seo_meta_tags', ProductSeoMetaTags::fromFields(...)),
            $f->strings('required_customer_tags'),
        );
    }

    /**
     * @template T
     *
     * @param callable(Fields): T $build
     *
     * @return T|null
     */
    private static function optional(Fields $f, string $key, callable $build): mixed
    {
        $object = $f->objectOrNull($key);

        return $object === null ? null : $build($object);
    }
}
