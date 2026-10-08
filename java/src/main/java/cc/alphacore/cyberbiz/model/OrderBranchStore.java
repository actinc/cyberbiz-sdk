package cc.alphacore.cyberbiz.model;

/**
 * The branch store attached to an order for pickup or express delivery.
 *
 * @param storeNo store number
 * @param name store name
 * @param phone store phone
 * @param county county
 * @param district district
 * @param address street address
 * @param zip postal code
 * @param openingHours opening hours
 * @param lat latitude, or null
 * @param lng longitude, or null
 * @param enabled whether the store is enabled, or null
 * @param sourceType "BranchStore" or "PosShop"
 * @param sourceId the id of the source store, or null
 */
public record OrderBranchStore(
    String storeNo,
    String name,
    String phone,
    String county,
    String district,
    String address,
    String zip,
    String openingHours,
    Double lat,
    Double lng,
    Boolean enabled,
    String sourceType,
    Long sourceId) {}
