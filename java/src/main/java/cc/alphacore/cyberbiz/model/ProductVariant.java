package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.time.OffsetDateTime;
import java.util.List;

/**
 * A purchasable variant of a product: one combination of option values with its own price, SKU and
 * stock.
 *
 * @param id the variant id; set from the request for a single-variant read or update
 * @param productId the product the variant belongs to
 * @param name the variant name
 * @param position the order among the product's variants, from 1
 * @param price the selling price
 * @param cost the cost, or null
 * @param compareAtPrice the original price shown struck through, or null
 * @param meas the volume used for shipping
 * @param maxUsableBonus the most bonus points a customer can redeem on it
 * @param weight the weight used for shipping
 * @param option1 the value of the first option, or null
 * @param option2 the value of the second option, or null
 * @param option3 the value of the third option, or null
 * @param inventoryManagement whether CYBERBIZ tracks the stock
 * @param inventoryQuantity the quantity in stock
 * @param sold the quantity sold
 * @param safetyInventoryQuantity the safety stock level, or null
 * @param inventoryPolicy what happens when the stock runs out, e.g. {@code deny}
 * @param sku the SKU, or null
 * @param qc the quality-control code, or null
 * @param requiresShipping whether the variant is shipped
 * @param createdAt when the variant was created, in Asia/Taipei
 * @param updatedAt when the variant last changed, in Asia/Taipei
 * @param honeycombSync whether the variant syncs to Honeycomb
 * @param vendor the vendor, or null
 * @param photoUrls the variant's photo URLs; unmodifiable
 * @param pimInfos the variant's PIM links; unmodifiable
 */
public record ProductVariant(
    long id,
    long productId,
    String name,
    int position,
    Money price,
    Money cost,
    Money compareAtPrice,
    double meas,
    Money maxUsableBonus,
    double weight,
    String option1,
    String option2,
    String option3,
    boolean inventoryManagement,
    int inventoryQuantity,
    int sold,
    Integer safetyInventoryQuantity,
    String inventoryPolicy,
    String sku,
    String qc,
    boolean requiresShipping,
    OffsetDateTime createdAt,
    OffsetDateTime updatedAt,
    boolean honeycombSync,
    String vendor,
    List<String> photoUrls,
    List<ProductPimInfo> pimInfos) {

  /** Copies the lists; an absent list becomes empty. */
  public ProductVariant {
    photoUrls = Lists.copy(photoUrls);
    pimInfos = Lists.copy(pimInfos);
  }
}
