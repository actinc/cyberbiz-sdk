package cc.alphacore.cyberbiz.model;

/**
 * A product photo.
 *
 * @param id the photo id
 * @param url the image URL
 * @param position the order among the product's photos, or null
 */
public record ProductPhoto(long id, String url, Integer position) {}
