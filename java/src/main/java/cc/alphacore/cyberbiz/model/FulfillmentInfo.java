package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * The label data of one CVS fulfillment (POST /v1/orders/fulfillment_infos).
 *
 * @param id the fulfillment id, or null
 * @param idVerification pickup requires an id check, or null
 * @param receiver recipient name
 * @param receiverPhone recipient phone
 * @param amount amount to collect
 * @param trackingNumber tracking number
 * @param storeName store name
 * @param storeNo store number
 * @param barcode barcode
 * @param shopName shop name
 * @param orderNo order number
 * @param shopPhone shop phone
 * @param shopUrl shop URL
 * @param supplierName Hi-Life only
 * @param imageUrl FamilyMart only
 * @param pdf base64 PDF, Hi-Life only
 * @param html 7-11 C2C only
 * @param pdfJson Hi-Life cold chain only
 */
public record FulfillmentInfo(
    Long id,
    Boolean idVerification,
    String receiver,
    String receiverPhone,
    Money amount,
    String trackingNumber,
    String storeName,
    String storeNo,
    String barcode,
    String shopName,
    String orderNo,
    String shopPhone,
    String shopUrl,
    String supplierName,
    String imageUrl,
    String pdf,
    String html,
    String pdfJson) {}
