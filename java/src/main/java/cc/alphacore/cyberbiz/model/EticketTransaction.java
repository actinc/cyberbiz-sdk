package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;

/**
 * One redemption code of a split e-ticket.
 *
 * @param ticketNumber redemption code
 * @param usedAt when it was redeemed; null until then
 */
public record EticketTransaction(String ticketNumber, OffsetDateTime usedAt) {}
