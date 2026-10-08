package cc.alphacore.cyberbiz.model;

/**
 * The special collection a product belongs to.
 *
 * @param id the collection id
 * @param title the collection title
 * @param handle the collection handle used in URLs
 * @param published whether the collection is published, or null
 * @param position the order among the collections, or null
 */
public record ProductSpecialCollectionRef(
    long id, String title, String handle, Boolean published, Integer position) {}
