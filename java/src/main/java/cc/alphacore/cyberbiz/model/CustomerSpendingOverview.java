package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * A customer's paid and valid orders in a date range.
 *
 * @param paidAndValidTotalSpent total spent
 * @param paidAndValidOrdersCount number of orders
 * @param paidAndValidAverageSpent average per order
 */
public record CustomerSpendingOverview(
    Money paidAndValidTotalSpent, int paidAndValidOrdersCount, Money paidAndValidAverageSpent) {}
