package cc.alphacore.cyberbiz.model;

import java.util.List;

/**
 * One set of components of a combo product line item.
 *
 * @param quantity number of combo sets
 * @param items the components
 */
public record RelatedItems(int quantity, List<RelatedItem> items) {

  /** Copies the lists; a missing list becomes empty. */
  public RelatedItems {
    items = Lists.copy(items);
  }
}
