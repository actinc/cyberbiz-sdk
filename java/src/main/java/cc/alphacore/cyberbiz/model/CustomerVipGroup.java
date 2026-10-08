package cc.alphacore.cyberbiz.model;

import java.util.List;

/**
 * The VIP group a customer belongs to, as embedded in CustomerVipInfo.
 *
 * @param id the group id, or null
 * @param name name
 * @param position sort position, or null
 * @param descriptionUrl description page
 * @param customerTags tags given to members
 */
public record CustomerVipGroup(
    Long id, String name, Integer position, String descriptionUrl, List<String> customerTags) {

  /** Copies the lists; a missing list becomes empty. */
  public CustomerVipGroup {
    customerTags = Lists.copy(customerTags);
  }
}
