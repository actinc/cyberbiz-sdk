package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;

/**
 * When each lifecycle event of an order happened; null means it has not.
 *
 * @param requestReturnAt return requested
 * @param returnAt returned
 * @param refundAt refunded
 * @param closedAt closed
 * @param cancelledAt cancelled
 * @param expiredAt CVS pickup deadline passed
 * @param confirmedAt confirmed
 */
public record OrderTimings(
    OffsetDateTime requestReturnAt,
    OffsetDateTime returnAt,
    OffsetDateTime refundAt,
    OffsetDateTime closedAt,
    OffsetDateTime cancelledAt,
    OffsetDateTime expiredAt,
    OffsetDateTime confirmedAt) {}
