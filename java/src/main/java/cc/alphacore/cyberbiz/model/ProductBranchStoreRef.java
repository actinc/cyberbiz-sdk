package cc.alphacore.cyberbiz.model;

/**
 * The branch store a product belongs to.
 *
 * @param id the branch store id
 * @param name the branch store name
 * @param storeNo the store number
 * @param enabled whether the branch store is enabled, or null
 */
public record ProductBranchStoreRef(long id, String name, String storeNo, Boolean enabled) {}
