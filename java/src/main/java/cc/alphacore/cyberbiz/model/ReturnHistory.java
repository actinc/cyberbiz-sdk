package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.time.OffsetDateTime;

/**
 * One refund made against an order.
 *
 * @param body short description
 * @param price the refunded amount
 * @param refundedAt when it was refunded
 */
public record ReturnHistory(String body, Money price, OffsetDateTime refundedAt) {}
