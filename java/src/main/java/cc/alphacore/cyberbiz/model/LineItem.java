package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.time.OffsetDateTime;
import java.util.List;

/**
 * One product variant on an order, fulfillment or return, or in a customer's recent purchases.
 *
 * @param id the line item id
 * @param productId the product id
 * @param productVariantId the variant id
 * @param title product title
 * @param variantTitle variant title
 * @param sku SKU
 * @param qc vendor's own item code
 * @param vendor vendor
 * @param price unit price
 * @param cost unit cost, or null
 * @param quantity units ordered
 * @param itemType "normal" or "no"
 * @param returnStatus "no_need", "request_return", "returning", "checking", "returned", "in_hub",
 *     "problem", "processing", "in_origin_cvs", "refused" or "partial_return"
 * @param discountName discount name
 * @param discounts discounts applied to this line
 * @param totalPriceBeforeDiscounts line total before discounts
 * @param totalDiscount total discount on this line
 * @param totalPriceAfterDiscounts line total after discounts
 * @param taxTypeId "inclusive_tax", "zero_tax" or "exclusive_tax"
 * @param bonusRedemptionPrice bonus points redeemed for this item; null when not a bonus mall
 *     redemption
 * @param relatedItems components of a combo product
 * @param createdAt when the line was created
 * @param channel sales channel
 * @param weight unit weight
 * @param photo thumbnail path
 * @param customFields shop-defined fields
 */
public record LineItem(
    long id,
    long productId,
    long productVariantId,
    String title,
    String variantTitle,
    String sku,
    String qc,
    String vendor,
    Money price,
    Money cost,
    int quantity,
    String itemType,
    String returnStatus,
    String discountName,
    List<LineItemDiscount> discounts,
    Money totalPriceBeforeDiscounts,
    Money totalDiscount,
    Money totalPriceAfterDiscounts,
    String taxTypeId,
    Money bonusRedemptionPrice,
    List<RelatedItems> relatedItems,
    OffsetDateTime createdAt,
    String channel,
    double weight,
    String photo,
    List<CustomField> customFields) {

  /** Copies the lists; a missing list becomes empty. */
  public LineItem {
    discounts = Lists.copy(discounts);
    relatedItems = Lists.copy(relatedItems);
    customFields = Lists.copy(customFields);
  }
}
