package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * One discount applied to a line item.
 *
 * @param position which unit of the line the discount hit
 * @param id discount type id
 * @param code e.g. "bundle_discount"
 * @param name discount name
 * @param discount the amount
 */
public record LineItemDiscount(int position, long id, String code, String name, Money discount) {}
