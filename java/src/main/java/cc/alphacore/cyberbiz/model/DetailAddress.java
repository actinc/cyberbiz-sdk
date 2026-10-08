package cc.alphacore.cyberbiz.model;

/**
 * An address split into its components.
 *
 * @param zip postal code
 * @param country country
 * @param province province
 * @param city city
 * @param district district
 * @param address1 first address line
 * @param address2 second address line
 */
public record DetailAddress(
    String zip,
    String country,
    String province,
    String city,
    String district,
    String address1,
    String address2) {}
