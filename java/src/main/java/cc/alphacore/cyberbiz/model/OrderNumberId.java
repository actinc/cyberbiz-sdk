package cc.alphacore.cyberbiz.model;

/**
 * Maps a shop-facing order number to its API id (GET /v1/orders/get_order_id).
 *
 * @param orderNumber the shop-facing number
 * @param orderId the API id
 */
public record OrderNumberId(long orderNumber, long orderId) {}
