package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.time.OffsetDateTime;

/**
 * The validity window of the current VIP level and how far the customer is from renewing or
 * upgrading; null when there is no such condition.
 *
 * @param startAt start of the validity window
 * @param endAt end of the validity window
 * @param differenceOfTotalSpentInValidityDaysForRenewal spend within the window still needed to
 *     renew
 * @param differenceOfTotalSpentForRenewal single-order spend still needed to renew
 * @param differenceOfTotalSpentInValidityDaysForUpgrade spend within the window still needed to
 *     upgrade
 * @param differenceOfTotalSpentForUpgrade single-order spend still needed to upgrade
 */
public record CustomerVipExtraInfo(
    OffsetDateTime startAt,
    OffsetDateTime endAt,
    Money differenceOfTotalSpentInValidityDaysForRenewal,
    Money differenceOfTotalSpentForRenewal,
    Money differenceOfTotalSpentInValidityDaysForUpgrade,
    Money differenceOfTotalSpentForUpgrade) {}
