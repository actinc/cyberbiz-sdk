package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;
import java.util.List;

/**
 * One shipment of an order.
 *
 * @param id the fulfillment id
 * @param trackingCompany carrier code such as "ezcat", "e_can", "postserv", "hct", "SEVEN",
 *     "FAMILY", "HILIFE", "SEVEN_C2C", "FAMILY_C2C", "SF", "cyberbiz_express" or "other"
 * @param trackingNumber tracking number
 * @param fulfilledAt when it shipped
 * @param receivedAt when it was received; null until then
 * @param status "unshipped", "preparing", "cancel", "fulfilled", "partial", "arrived", "received",
 *     "returned", "expired", "problem" or "no_need"
 * @param lineItems the shipped line items
 * @param trackingUrl Uber Direct and Pandago only
 * @param cvsShippingType "seven", "seven_c2c", "family", "family_c2c", "family_cold",
 *     "family_cold_c2c", "hilife", "hilife_cold", "hilife_c2c", "ezcat_cvs", "ezcat_cvs_cold" or
 *     "ezcat_cvs_refrigerate"; only set by the v2 CVS shipping endpoint
 */
public record Fulfillment(
    long id,
    String trackingCompany,
    String trackingNumber,
    OffsetDateTime fulfilledAt,
    OffsetDateTime receivedAt,
    String status,
    List<LineItem> lineItems,
    String trackingUrl,
    String cvsShippingType) {

  /** Copies the lists; a missing list becomes empty. */
  public Fulfillment {
    lineItems = Lists.copy(lineItems);
  }
}
