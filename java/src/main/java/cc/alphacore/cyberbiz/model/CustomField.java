package cc.alphacore.cyberbiz.model;

/**
 * A shop-defined field value on a customer or line item; id is null where the response omits it.
 *
 * @param id the field id, or null
 * @param name field key
 * @param label display name
 * @param value the value
 */
public record CustomField(Long id, String name, String label, String value) {}
