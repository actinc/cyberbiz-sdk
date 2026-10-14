package cc.alphacore.cyberbiz.webhook;

import java.util.Optional;

/** Every documented webhook Event, in documentation order (docs/api/en/webhooks.md). */
public enum EventType {
  /** A customer registered. */
  CUSTOMERS_CREATE("customers/create"),
  /** A customer's profile changed. */
  CUSTOMERS_UPDATE("customers/update"),
  /** A social-login uid was linked to a customer. */
  UID_PROVIDERS_CREATE("uid_providers/create"),
  /** A social-login uid of a customer changed. */
  UID_PROVIDERS_UPDATE("uid_providers/update"),
  /** Bonus points were granted. */
  BONUS_POINTS_CREATE("bonus_points/create"),
  /** Bonus points were used. */
  BONUS_POINTS_UPDATE("bonus_points/update"),
  /** A bonus point record was deleted. */
  BONUS_POINTS_DESTROY("bonus_points/destroy"),
  /** Bonus points were granted for an approved product review. */
  COMMENT_BONUS_CREATE("comment_bonus/create"),
  /** An order was placed. */
  ORDERS_CREATE("orders/create"),
  /** An order was paid. */
  ORDERS_PAID("orders/paid"),
  /** An order is being prepared for shipment. */
  ORDERS_PREPARING("orders/preparing"),
  /** An order was shipped. */
  ORDERS_FULFILLED("orders/fulfilled"),
  /** An order was received by the customer. */
  ORDERS_RECEIVED("orders/received"),
  /** An order arrived at the pickup store. */
  ORDERS_ARRIVED("orders/arrived"),
  /** An order was not picked up in time. */
  ORDERS_EXPIRED("orders/expired"),
  /** An order was cancelled. */
  ORDERS_CANCELLED("orders/cancelled"),
  /** An order was returned. */
  ORDERS_RETURNED("orders/returned"),
  /** Part of an order was returned. */
  ORDERS_PARTIAL_RETURN("orders/partial_return"),
  /** An order was refunded. */
  ORDERS_REFUNDED("orders/refunded"),
  /** Part of an order was refunded. */
  ORDERS_PARTIAL_REFUNDED("orders/partial_refunded"),
  /** An order was closed. */
  ORDERS_CLOSED("orders/closed"),
  /** The customer requested a return. */
  ORDERS_REQUEST_RETURN("orders/request_return"),
  /** An order was (re)opened. */
  ORDERS_OPENED("orders/opened"),
  /** An express-delivery order changed. */
  EXPRESS_DELIVERY_ORDERS_UPDATE("express_delivery_orders/update"),
  /** A product was created. */
  PRODUCTS_CREATE("products/create"),
  /** A product changed. */
  PRODUCTS_UPDATE("products/update"),
  /** A product was deleted. */
  PRODUCTS_DELETE("products/delete"),
  /** A product variant was created. */
  VARIANTS_CREATE("variants/create"),
  /** A product variant changed. */
  VARIANTS_UPDATE("variants/update"),
  /** A product variant was deleted. */
  VARIANTS_DELETE("variants/delete"),
  /** A coupon was created. */
  COUPONS_CREATE("coupons/create"),
  /** A coupon was used. */
  COUPONS_UPDATE("coupons/update"),
  /** A coupon was deleted. */
  COUPONS_DESTROY("coupons/destroy"),
  /** A customer's VIP level changed. */
  CUSTOMER_VIP_LEVEL_UPDATE("customer_vip_level/update"),
  /** The app was uninstalled from the shop. */
  APPS_UNINSTALL("apps/uninstall"),
  // Added from the source-derived webhook collection; the last four are sent by Merchant
  // webhooks only and have no resource prefix.
  ORDERS_UPDATE("orders/update"),
  PRODUCTS_UPDATE_OPTIONS("products/update_options"),
  PRODUCTS_UPDATE_PHOTOS("products/update_photos"),
  VARIANTS_OUT_OF_STOCK("variants/out_of_stock"),
  CART_ADD("cart/add"),
  CUSTOMER_TAGS_UPDATE("customer_tags/update"),
  AFFILIATE_VENDOR_ORDERS_CREATE("affiliate_vendor_orders/create"),
  AFFILIATE_VENDOR_ORDERS_CANCELLED("affiliate_vendor_orders/cancelled"),
  AFFILIATE_VENDOR_ORDERS_CLOSED("affiliate_vendor_orders/closed"),
  COLLECTIONS_CREATE("collections/create"),
  COLLECTIONS_UPDATE("collections/update"),
  COLLECTIONS_DELETE("collections/delete"),
  MENUS_CREATE("menus/create"),
  MENUS_UPDATE("menus/update"),
  MENUS_DESTROY("menus/destroy"),
  DOMAINS_CREATE("domains/create"),
  DOMAINS_UPDATE("domains/update"),
  DOMAINS_DESTROY("domains/destroy"),
  BULK_OPERATIONS_FINISH("bulk_operations/finish"),
  LINE_CHAT_BOT_UPDATE("line_chat_bot/update"),
  FBE_AD_ACCOUNTS_RECLAIM("fbe_ad_accounts/reclaim"),
  POS_WALLET_TRANSACTIONS_TOPUP("pos_wallet_transactions/topup"),
  POS_WALLET_TRANSACTIONS_REFUND("pos_wallet_transactions/refund"),
  STOCK_INVOICES_CREATED("stock_invoices/created"),
  STOCK_INVOICES_DONE("stock_invoices/done"),
  STOCK_INVOICES_CANCELLED("stock_invoices/cancelled"),
  STOCK_RECEIPTS_CREATED("stock_receipts/created"),
  STOCK_RECEIPTS_DONE("stock_receipts/done"),
  STOCK_RECEIPTS_CANCELLED("stock_receipts/cancelled"),
  STOCK_REQUISITIONS_CREATED("stock_requisitions/created"),
  STOCK_REQUISITIONS_APPROVED("stock_requisitions/approved"),
  STOCK_REQUISITIONS_CANCELLED("stock_requisitions/cancelled"),
  STORED_ORDERS_RETRIEVED("stored_orders/retrieved"),
  STORED_ORDERS_REFUNDED("stored_orders/refunded"),
  SEND_EDM_FINISH("send_edm_finish"),
  SEND_SMS_FINISH("send_sms_finish"),
  SEND_LINE_OA_MESSAGE_FINISH("send_line_oa_message_finish"),
  GET_CUSTOMER_IDS_FINISH("get_customer_ids_finish");

  /** The X-Cyberbiz-Event value. */
  private final String value;

  EventType(String value) {
    this.value = value;
  }

  /**
   * Returns the documented Event for an X-Cyberbiz-Event value.
   *
   * @param value e.g. "orders/paid"
   * @return the Event, or empty for one this SDK does not know yet
   */
  public static Optional<EventType> of(String value) {
    for (EventType type : values()) {
      if (type.value.equals(value)) {
        return Optional.of(type);
      }
    }
    return Optional.empty();
  }

  /** Returns the X-Cyberbiz-Event value, e.g. "orders/paid". */
  public String value() {
    return value;
  }

  /**
   * Returns the resource, "orders" for "orders/paid"; the whole value for an Event without a slash,
   * such as "send_sms_finish".
   */
  public String resource() {
    int slash = value.indexOf('/');
    return slash < 0 ? value : value.substring(0, slash);
  }

  /** Returns the action, "paid" for "orders/paid"; empty for an Event without a slash. */
  public String action() {
    int slash = value.lastIndexOf('/');
    return slash < 0 ? "" : value.substring(slash + 1);
  }
}
