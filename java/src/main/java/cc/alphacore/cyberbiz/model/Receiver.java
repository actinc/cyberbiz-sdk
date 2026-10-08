package cc.alphacore.cyberbiz.model;

/**
 * The shipping recipient of an order.
 *
 * @param name recipient name
 * @param countryCallingCode e.g. "+886"
 * @param phone phone
 * @param address single-line full address
 * @param detailAddress the address split into components, or null
 * @param cvsStoreId convenience store number for pickup
 * @param allpayLogisticsId ECPay logistics id
 */
public record Receiver(
    String name,
    String countryCallingCode,
    String phone,
    String address,
    DetailAddress detailAddress,
    String cvsStoreId,
    String allpayLogisticsId) {}
