package cc.alphacore.cyberbiz.model;

import java.util.List;

/**
 * An electronic ticket sold on an order.
 *
 * @param title ticket title
 * @param ticketNumber redemption code
 * @param availableQuantity units left, or null
 * @param usedQuantity units redeemed, or null
 * @param enabled whether the ticket can be redeemed, or null
 * @param orderId the order id, or null
 * @param productId the product id, or null
 * @param separate split into one code per unit, or null
 * @param transactions the per-unit codes
 */
public record OrderEticket(
    String title,
    String ticketNumber,
    Integer availableQuantity,
    Integer usedQuantity,
    Boolean enabled,
    Long orderId,
    Long productId,
    Boolean separate,
    List<EticketTransaction> transactions) {

  /** Copies the lists; a missing list becomes empty. */
  public OrderEticket {
    transactions = Lists.copy(transactions);
  }
}
