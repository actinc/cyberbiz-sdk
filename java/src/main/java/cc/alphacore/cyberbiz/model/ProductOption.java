package cc.alphacore.cyberbiz.model;

/**
 * A product option such as size or colour; its values are the variants' {@code option1..3}.
 *
 * @param id the option id; set from the request for a single-option read or update
 * @param name the option name
 * @param position the order among the product's options, from 1
 * @param types the option values, comma-separated, e.g. {@code S,M,L}
 */
public record ProductOption(long id, String name, int position, String types) {}
