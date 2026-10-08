package cc.alphacore.cyberbiz.model;

/**
 * The POS terminal and salesperson of a POS order; the ids are null on web orders.
 *
 * @param posUserId salesperson id
 * @param posUserEmail salesperson email
 * @param posShopId POS shop id
 * @param posInfo description
 * @param posId POS terminal id
 * @param posName POS terminal name
 */
public record PosInfo(
    Long posUserId,
    String posUserEmail,
    Long posShopId,
    String posInfo,
    Long posId,
    String posName) {}
