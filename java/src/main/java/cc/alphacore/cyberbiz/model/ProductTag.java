package cc.alphacore.cyberbiz.model;

/**
 * A tag attached to a product.
 *
 * @param id the tag id
 * @param name the tag text
 * @param category the tag category, or null
 */
public record ProductTag(long id, String name, Integer category) {}
