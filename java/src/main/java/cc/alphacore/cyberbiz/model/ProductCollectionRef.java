package cc.alphacore.cyberbiz.model;

/**
 * A custom collection that contains a product.
 *
 * @param id the collection id
 * @param title the collection title
 * @param handle the collection handle used in URLs
 * @param published whether the collection is published
 * @param bodyHtml the collection description as HTML, or null
 * @param productsOrderName how the collection orders its products
 * @param position the order among the collections
 */
public record ProductCollectionRef(
    long id,
    String title,
    String handle,
    boolean published,
    String bodyHtml,
    String productsOrderName,
    int position) {}
