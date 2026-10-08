package cc.alphacore.cyberbiz.model;

/**
 * A product's SEO fields; each is null when the Shop has not set it.
 *
 * @param title the page title
 * @param description the meta description
 * @param keywords the meta keywords
 */
public record ProductSeoMetaTags(String title, String description, String keywords) {}
