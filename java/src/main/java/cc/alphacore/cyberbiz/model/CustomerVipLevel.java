package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * A VIP level as embedded in CustomerVipInfo.
 *
 * @param id the level id, or null
 * @param position rank, or null
 * @param name name
 * @param validityDays validity in days, or null
 * @param upgradeValidityDays upgrade window in days, or null
 * @param upgradeConditionTotalSpent single-order spend that upgrades to this level
 * @param upgradeConditionTotalSpentInValidityDays spend within the validity window that upgrades to
 *     this level
 * @param renewalConditionTotalSpent single-order spend that renews this level
 * @param renewalConditionTotalSpentInValidityDays spend within the validity window that renews this
 *     level
 * @param bonusPointEnabled whether purchases earn bonus points, or null
 * @param bonusPointThreshold spend per bonus award
 * @param bonusPointValue bonus points per award
 * @param bonusPointExpiryDays bonus point lifetime in days, or null
 * @param birthGiftEnabled whether there is a birthday gift, or null
 * @param birthGiftName birthday gift
 * @param upgradeGiftEnabled whether there is an upgrade gift, or null
 * @param orderDiscountEnabled whether the level gives an order discount, or null
 * @param freeShippingEnabled whether the level ships free, or null
 */
public record CustomerVipLevel(
    Long id,
    Integer position,
    String name,
    Integer validityDays,
    Integer upgradeValidityDays,
    Money upgradeConditionTotalSpent,
    Money upgradeConditionTotalSpentInValidityDays,
    Money renewalConditionTotalSpent,
    Money renewalConditionTotalSpentInValidityDays,
    Boolean bonusPointEnabled,
    Money bonusPointThreshold,
    Money bonusPointValue,
    Integer bonusPointExpiryDays,
    Boolean birthGiftEnabled,
    String birthGiftName,
    Boolean upgradeGiftEnabled,
    Boolean orderDiscountEnabled,
    Boolean freeShippingEnabled) {}
