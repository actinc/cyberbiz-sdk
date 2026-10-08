package cc.alphacore.cyberbiz.model;

/**
 * A postal address with contact details: an order's billing address or a customer's default
 * address. name is only set on the former, company only on the latter.
 *
 * @param name contact name
 * @param company company
 * @param countryCallingCode e.g. "+886"
 * @param phone phone
 * @param address single-line full address
 * @param detailAddress the address split into components, or null
 */
public record Address(
    String name,
    String company,
    String countryCallingCode,
    String phone,
    String address,
    DetailAddress detailAddress) {}
