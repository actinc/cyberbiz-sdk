package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.util.List;

/**
 * One component variant inside a combo product.
 *
 * @param id the component id
 * @param productId the product id
 * @param productVariantId the variant id
 * @param title product title
 * @param variantTitle variant title
 * @param sku SKU
 * @param qc vendor's own item code
 * @param vendor vendor
 * @param price unit price
 * @param cost unit cost, or null
 * @param quantity units per combo set
 * @param comboProductPriceDifference enterprise-only difference between the component price and its
 *     share of the combo price
 * @param comboProductPriceDiffDetails the difference per unit
 */
public record RelatedItem(
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
    Money comboProductPriceDifference,
    List<Money> comboProductPriceDiffDetails) {

  /** Copies the lists; a missing list becomes empty. */
  public RelatedItem {
    comboProductPriceDiffDetails = Lists.copy(comboProductPriceDiffDetails);
  }
}
