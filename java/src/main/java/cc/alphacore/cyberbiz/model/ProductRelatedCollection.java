package cc.alphacore.cyberbiz.model;

/**
 * A collection shown as related to a product.
 *
 * @param id the link id
 * @param relatableCollectionId the related collection's id, or null
 * @param relatableCollectionType the related collection's type
 */
public record ProductRelatedCollection(
    long id, Long relatableCollectionId, String relatableCollectionType) {}
