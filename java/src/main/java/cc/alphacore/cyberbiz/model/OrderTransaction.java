package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * One payment recorded against an order.
 *
 * @param id the transaction id, or null
 * @param amount the amount
 * @param kindName e.g. 已收款
 * @param paidTypeName e.g. 手動
 */
public record OrderTransaction(Long id, Money amount, String kindName, String paidTypeName) {}
