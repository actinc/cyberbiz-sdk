package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;

/**
 * The electronic invoice attached to an order or exchange.
 *
 * @param title invoice title (buyer name)
 * @param orderId the order id, or null
 * @param companyNo buyer's tax id
 * @param invoiceNo invoice number
 * @param invoiceStatus "issue", "issue_invalid", "allowance" or "partial_allowance"
 * @param invoiceAt issued at
 * @param invalidAt voided or allowance time; null while valid
 * @param randomNum the invoice's random number
 * @param invoiceType "default", "company", "phone_barcode", "nature_person", "donate" or
 *     "paper_invoice"
 * @param loveCode donation code
 * @param phoneBarcode mobile barcode carrier
 * @param naturePerson citizen digital certificate
 */
public record OrderEinvoice(
    String title,
    Long orderId,
    String companyNo,
    String invoiceNo,
    String invoiceStatus,
    OffsetDateTime invoiceAt,
    OffsetDateTime invalidAt,
    String randomNum,
    String invoiceType,
    String loveCode,
    String phoneBarcode,
    String naturePerson) {}
