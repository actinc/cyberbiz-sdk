package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.time.OffsetDateTime;
import java.util.List;

/**
 * One shop order (GET /v1/orders, GET /v1/orders/{id}, GET /v1/customers/{id}/orders). Only the
 * list responses carry {@code token}.
 *
 * @param id the order id
 * @param subtotalPrice line items only, before shipping
 * @param createdAt when the order was placed
 * @param updatedAt when the order last changed
 * @param orderNumber the shop-facing sequential number (the swagger says string; the platform sends
 *     an integer)
 * @param orderName e.g. "#1101"
 * @param customer the member who placed the order, or null
 * @param buyer the account that placed the order, or null
 * @param receiver the shipping recipient, or null
 * @param billingAddress the billing address, or null
 * @param lineItems the ordered variants
 * @param shippingType the shipping method code
 * @param shippingName the shipping method name
 * @param shippingVendor the carrier, or null
 * @param logisticsId ECPay logistics order id
 * @param deliveryDate requested delivery day (midnight, Asia/Taipei), custom feature; null when not
 *     set
 * @param deliveryTime requested delivery slot 0-3, custom feature
 * @param delegate staff email when the order was placed on the customer's behalf
 * @param fulfillments the shipments
 * @param paymentName the payment method name
 * @param paymentMethod the payment method code
 * @param paymentUrl the payment page
 * @param multiplePaymentInfos POS split payments
 * @param prices the price breakdown, or null
 * @param card4no last four digits of the card
 * @param transactionNumber the payment provider's transaction number
 * @param merchantTradeNo the merchant trade number sent to the payment provider
 * @param einvoice the electronic invoice, or null
 * @param paperInvoiceNo paper invoice number
 * @param paperCompanyNo tax id on the paper invoice
 * @param prepaymentPaperInvoiceNo paper invoice number of a prepayment
 * @param prepaymentPaperCompanyNo tax id on the prepayment paper invoice
 * @param statuses the four state machines, or null
 * @param timings lifecycle timestamps, or null
 * @param returnHistories refunds made against the order
 * @param note the customer's note
 * @param branchStore pickup store, or null
 * @param referralCode referral code
 * @param checkoutReferralCode referral code entered at checkout
 * @param checkoutReferralUserName name of the referrer entered at checkout
 * @param registerReferralCode referral code the customer registered with
 * @param totalBonusRedemptionPrice bonus points spent in the bonus mall, in the shop currency
 * @param posInfo POS terminal and salesperson, or null
 * @param exchangeHistories POS exchanges
 * @param linkedOrderInfo the referring affiliate, or null
 * @param tags the order's tags
 * @param expressDeliveryBranchStore the store shipping an express delivery, or null
 * @param shippingStatus the carrier's shipping status
 * @param extraInfo extra information
 * @param fromDevice e.g. 桌機, 手機
 * @param customerCancelReasonDetail the customer's cancel reason, or null
 * @param serialNumbers campaign serial numbers
 * @param orderWeight total weight, or null
 * @param warehouseTypeId warehouse type, or null
 * @param utmTracking UTM attribution (custom feature), or null
 * @param token the order's public token, list endpoints only
 */
public record Order(
    long id,
    Money subtotalPrice,
    OffsetDateTime createdAt,
    OffsetDateTime updatedAt,
    long orderNumber,
    String orderName,
    Customer customer,
    Buyer buyer,
    Receiver receiver,
    Address billingAddress,
    List<LineItem> lineItems,
    String shippingType,
    String shippingName,
    ShippingVendor shippingVendor,
    String logisticsId,
    OffsetDateTime deliveryDate,
    int deliveryTime,
    String delegate,
    List<Fulfillment> fulfillments,
    String paymentName,
    String paymentMethod,
    String paymentUrl,
    List<OrderPaymentInfo> multiplePaymentInfos,
    Prices prices,
    String card4no,
    String transactionNumber,
    String merchantTradeNo,
    OrderEinvoice einvoice,
    String paperInvoiceNo,
    String paperCompanyNo,
    String prepaymentPaperInvoiceNo,
    String prepaymentPaperCompanyNo,
    OrderStatuses statuses,
    OrderTimings timings,
    List<ReturnHistory> returnHistories,
    String note,
    OrderBranchStore branchStore,
    String referralCode,
    String checkoutReferralCode,
    String checkoutReferralUserName,
    String registerReferralCode,
    Money totalBonusRedemptionPrice,
    PosInfo posInfo,
    List<ExchangeHistory> exchangeHistories,
    LinkedOrderInfo linkedOrderInfo,
    List<Tag> tags,
    OrderBranchStore expressDeliveryBranchStore,
    String shippingStatus,
    String extraInfo,
    String fromDevice,
    OrderCustomerCancelReason customerCancelReasonDetail,
    List<String> serialNumbers,
    Double orderWeight,
    Integer warehouseTypeId,
    UtmTracking utmTracking,
    String token) {

  /** Copies the lists; a missing list becomes empty. */
  public Order {
    lineItems = Lists.copy(lineItems);
    fulfillments = Lists.copy(fulfillments);
    multiplePaymentInfos = Lists.copy(multiplePaymentInfos);
    returnHistories = Lists.copy(returnHistories);
    exchangeHistories = Lists.copy(exchangeHistories);
    tags = Lists.copy(tags);
    serialNumbers = Lists.copy(serialNumbers);
  }
}
