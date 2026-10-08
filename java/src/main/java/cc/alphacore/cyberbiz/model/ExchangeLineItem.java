package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;

/**
 * One variant exchanged in an ExchangeHistory.
 *
 * @param productVariantId the variant id, or null
 * @param name the name
 * @param sku SKU
 * @param qc vendor's own item code
 * @param price unit price
 * @param quantity units, or null
 */
public record ExchangeLineItem(
    Long productVariantId, String name, String sku, String qc, Money price, Integer quantity) {}
