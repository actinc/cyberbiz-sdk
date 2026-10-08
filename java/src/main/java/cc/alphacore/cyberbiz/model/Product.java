package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import com.google.gson.JsonElement;
import java.time.OffsetDateTime;
import java.util.List;

/**
 * A product with its variants, options, tags, photos and collections. Lists are unmodifiable and
 * empty when the API leaves them out.
 *
 * @param id the product id; set from the request for a single-product read or update, whose
 *     response omits it
 * @param title the product title
 * @param handle the handle used in the product URL, or null
 * @param englishTitle the English title, or null
 * @param productUrl the storefront URL, often protocol-relative
 * @param published whether the product is published
 * @param sellFrom when sales start, or null
 * @param sellTo when sales end, or null
 * @param productType the product type, or null
 * @param productTypeCode the product type code, or null
 * @param slogan the slogan
 * @param brief the short description, possibly HTML
 * @param briefText the short description as plain text, or null
 * @param briefIncludesHtml whether {@code brief} contains HTML
 * @param bodyHtml the full description as HTML
 * @param vendor the vendor, or null
 * @param price the price shown for the product
 * @param sellWeight the weight used for shipping
 * @param taxTypeId the tax type
 * @param customCollections the custom collections containing the product
 * @param specialCollection the special collection, or null
 * @param tags the product's tags
 * @param productVariants the variants
 * @param productOptions the options
 * @param posShop the POS shop, or null
 * @param photos the photos
 * @param photoUrls the photo URLs
 * @param channel the sales channel, or null
 * @param relatedCollections the related collections
 * @param branchStore the branch store, or null
 * @param createdAt when the product was created, in Asia/Taipei
 * @param updatedAt when the product last changed, in Asia/Taipei
 * @param temperatureTypes the shipping temperature types, e.g. {@code normal}
 * @param searchable whether the product appears in storefront search
 * @param googleProductCategoryId the Google product category, or null
 * @param productCustomFields the Shop-defined custom fields as free-form JSON; a copy, never null
 *     ({@code JsonNull} when absent)
 * @param seoMetaTags the SEO fields, or null
 * @param requiredCustomerTags customer tags required to buy the product
 */
public record Product(
    long id,
    String title,
    String handle,
    String englishTitle,
    String productUrl,
    boolean published,
    OffsetDateTime sellFrom,
    OffsetDateTime sellTo,
    String productType,
    String productTypeCode,
    String slogan,
    String brief,
    String briefText,
    boolean briefIncludesHtml,
    String bodyHtml,
    String vendor,
    Money price,
    double sellWeight,
    String taxTypeId,
    List<ProductCollectionRef> customCollections,
    ProductSpecialCollectionRef specialCollection,
    List<ProductTag> tags,
    List<ProductVariant> productVariants,
    List<ProductOption> productOptions,
    ProductPosShopRef posShop,
    List<ProductPhoto> photos,
    List<String> photoUrls,
    ProductChannel channel,
    List<ProductRelatedCollection> relatedCollections,
    ProductBranchStoreRef branchStore,
    OffsetDateTime createdAt,
    OffsetDateTime updatedAt,
    List<String> temperatureTypes,
    boolean searchable,
    Long googleProductCategoryId,
    JsonElement productCustomFields,
    ProductSeoMetaTags seoMetaTags,
    List<String> requiredCustomerTags) {

  /** Copies the lists and the free-form JSON; an absent list becomes empty. */
  public Product {
    customCollections = Lists.copy(customCollections);
    tags = Lists.copy(tags);
    productVariants = Lists.copy(productVariants);
    productOptions = Lists.copy(productOptions);
    photos = Lists.copy(photos);
    photoUrls = Lists.copy(photoUrls);
    relatedCollections = Lists.copy(relatedCollections);
    temperatureTypes = Lists.copy(temperatureTypes);
    productCustomFields = Copies.json(productCustomFields);
    requiredCustomerTags = Lists.copy(requiredCustomerTags);
  }

  /** Returns a copy of the custom fields, so the record stays immutable. */
  @Override
  public JsonElement productCustomFields() {
    return productCustomFields.deepCopy();
  }
}
