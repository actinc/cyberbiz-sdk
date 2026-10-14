package cc.alphacore.cyberbiz.model;

/**
 * The shop admin who wrote a reply in a customer service thread.
 *
 * @param id the admin id, or null
 * @param name the admin's name
 * @param email the admin's email
 */
public record CustomerMessageAdmin(Long id, String name, String email) {}
