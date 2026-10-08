package cc.alphacore.cyberbiz.model;

/**
 * How a variant is linked to the product information management (PIM) system.
 *
 * @param id the link id
 * @param productId the product id, or null
 * @param productVariantId the variant id, or null
 * @param pimProductId the PIM product id, or null
 * @param pimVariantId the PIM variant id, or null
 * @param channel the PIM channel, or null
 * @param channelShopName the shop name on that channel
 * @param isConnected whether the link is active, or null
 * @param isSource whether this Shop is the source of the data, or null
 * @param shopId the Shop id, or null
 */
public record ProductPimInfo(
    long id,
    Long productId,
    Long productVariantId,
    Long pimProductId,
    Long pimVariantId,
    Integer channel,
    String channelShopName,
    Boolean isConnected,
    Boolean isSource,
    Long shopId) {}
