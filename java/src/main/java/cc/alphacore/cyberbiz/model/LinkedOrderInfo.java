package cc.alphacore.cyberbiz.model;

/**
 * The affiliate or marketplace that referred an order; source is "非導購訂單" when there was none.
 *
 * @param source the referrer
 * @param shopdotcomRid Shopdotcom RID
 * @param shopdotcomClickId Shopdotcom click id
 * @param lineShoppingEcid LINE Shopping ECID
 * @param lineShoppingAffiliate LINE Shopping affiliate
 * @param ichannelGid iChannel GID
 */
public record LinkedOrderInfo(
    String source,
    String shopdotcomRid,
    String shopdotcomClickId,
    String lineShoppingEcid,
    String lineShoppingAffiliate,
    String ichannelGid) {}
