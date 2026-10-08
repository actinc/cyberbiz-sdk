package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.time.OffsetDateTime;
import java.util.List;

/**
 * One POS exchange performed on an order.
 *
 * @param createdAt when the exchange happened
 * @param price the price difference
 * @param orderPriceBefore order total before
 * @param orderPriceAfter order total after
 * @param posShopId POS shop id, or null
 * @param posId POS terminal id, or null
 * @param paymentName payment method name
 * @param paymentMethod payment method code
 * @param multiplePaymentInfos split payments
 * @param lineItems the exchanged variants
 * @param einvoice the invoice, or null
 * @param paperInvoiceNo paper invoice number
 * @param paperCompanyNo tax id on the paper invoice
 */
public record ExchangeHistory(
    OffsetDateTime createdAt,
    Money price,
    Money orderPriceBefore,
    Money orderPriceAfter,
    Long posShopId,
    Long posId,
    String paymentName,
    String paymentMethod,
    List<OrderPaymentInfo> multiplePaymentInfos,
    List<ExchangeLineItem> lineItems,
    OrderEinvoice einvoice,
    String paperInvoiceNo,
    String paperCompanyNo) {

  /** Copies the lists; a missing list becomes empty. */
  public ExchangeHistory {
    multiplePaymentInfos = Lists.copy(multiplePaymentInfos);
    lineItems = Lists.copy(lineItems);
  }
}
