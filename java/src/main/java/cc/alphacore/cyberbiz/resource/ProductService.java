package cc.alphacore.cyberbiz.resource;

import cc.alphacore.cyberbiz.CyberbizClient;
import cc.alphacore.cyberbiz.Request;
import cc.alphacore.cyberbiz.exception.ApiException;
import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.exception.TransportException;
import cc.alphacore.cyberbiz.model.Product;
import cc.alphacore.cyberbiz.model.ProductDescriptionSettingName;
import cc.alphacore.cyberbiz.model.ProductOption;
import cc.alphacore.cyberbiz.model.ProductTag;
import cc.alphacore.cyberbiz.model.ProductVariant;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.pagination.PageIterable;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Objects;

/**
 * Products, their variants, options, tags and shipping bindings. Get one from {@link
 * CyberbizClient#products()}.
 *
 * <p>Request bodies and query parameters are maps shaped like the API's JSON (see {@code
 * docs/api/en/cyberbiz-openapi-v1.yaml}), as in the PHP SDK; bodies are encoded with the SDK's JSON
 * configuration, so amounts are sent as exact decimals.
 *
 * <pre>{@code
 * Product tea = client.products().create(Map.of(
 *     "title", "Green Tea", "handle", "green-tea", "published", true,
 *     "price", new BigDecimal("120.50")));
 * }</pre>
 *
 * Every method throws {@link ApiException} for an error response, {@link TransportException} when
 * no response arrived and {@link DecodeException} when the body does not fit the model.
 */
public final class ProductService {
  private static final String PRODUCTS = "/v1/products";
  private static final String VARIANTS = "product_variants";
  private static final String OPTIONS = "product_options";

  private final Calls calls;

  /**
   * Creates the service; prefer {@link CyberbizClient#products()}.
   *
   * @param client the client that sends the requests
   */
  public ProductService(CyberbizClient client) {
    this.calls = new Calls(Objects.requireNonNull(client, "client"));
  }

  /**
   * Returns the first page of products ({@code GET /v1/products}).
   *
   * @return the page
   */
  public Page<Product> list() {
    return list(Map.of());
  }

  /**
   * Returns one page of products ({@code GET /v1/products}).
   *
   * @param query e.g. {@code page} and {@code per_page}
   * @return the page
   */
  public Page<Product> list(Map<String, ?> query) {
    return calls.client().list(get(PRODUCTS, query), Product.class);
  }

  /**
   * Walks every product, page by page ({@code GET /v1/products}).
   *
   * @return the products, fetched lazily
   */
  public PageIterable<Product> all() {
    return all(Map.of());
  }

  /**
   * Walks every product, page by page ({@code GET /v1/products}).
   *
   * @param query e.g. the first {@code page} and {@code per_page}
   * @return the products, fetched lazily
   */
  public PageIterable<Product> all(Map<String, ?> query) {
    return calls.client().all(get(PRODUCTS, query), Product.class);
  }

  /**
   * Returns one product ({@code GET /v1/products/{id}}).
   *
   * @param id the product id
   * @return the product, with {@code id} set
   * @throws NotFoundException also when the API answers with a null body
   */
  public Product get(long id) {
    return calls.one(Request.of("GET", path(id)), Product.class, id);
  }

  /**
   * Creates a product ({@code POST /v1/products}); title, handle, published and price are required.
   *
   * @param product the fields
   * @return the new product
   */
  public Product create(Map<String, ?> product) {
    return calls.object(Params.body(Request.of("POST", PRODUCTS), product), Product.class);
  }

  /**
   * Changes a product ({@code PUT /v1/products/{id}}).
   *
   * @param id the product id
   * @param changes the fields to change
   * @return the product, with {@code id} set
   */
  public Product update(long id, Map<String, ?> changes) {
    return calls.withId(Params.body(Request.of("PUT", path(id)), changes), Product.class, id);
  }

  /**
   * Removes a product ({@code DELETE /v1/products/{id}}).
   *
   * @param id the product id
   */
  public void delete(long id) {
    calls.call(Request.of("DELETE", path(id)));
  }

  /**
   * Creates the same product in several POS shops ({@code POST /v1/products/pos_shop_batch}).
   *
   * @param batch the fields, e.g. {@code pos_shop_ids} and the product
   * @return the created products
   */
  public List<Product> createForPosShops(Map<String, ?> batch) {
    Request request = Params.body(Request.of("POST", PRODUCTS + "/pos_shop_batch"), batch);
    return calls.list(request, Product.class);
  }

  /**
   * Finds products by keyword or vendor ({@code GET /v1/products/search}). The endpoint pages with
   * {@code limit} and {@code offset} and sends no pagination headers.
   *
   * @param query e.g. {@code q}, {@code vendor}, {@code limit}, {@code offset}, {@code
   *     filter_published}, {@code order_by}, {@code filter_branch_store}
   * @return the matching products
   */
  public List<Product> search(Map<String, ?> query) {
    return calls.list(get(PRODUCTS + "/search", query), Product.class);
  }

  /**
   * Returns the products of a collection ({@code GET /v1/products/search/collection}).
   *
   * @param query {@code collection_handle}, and e.g. {@code limit}, {@code offset}, {@code
   *     filter_published}, {@code order_by}
   * @return the products
   */
  public List<Product> searchCollection(Map<String, ?> query) {
    return calls.list(get(PRODUCTS + "/search/collection", query), Product.class);
  }

  /**
   * Changes a product's SEO fields ({@code PUT /v1/products/{id}/seo_meta_tags}).
   *
   * @param id the product id
   * @param tags e.g. {@code title}, {@code description}, {@code keywords}
   * @return the product, with {@code id} set
   */
  public Product updateSeoMetaTags(long id, Map<String, ?> tags) {
    Request request = Params.body(Request.of("PUT", path(id, "seo_meta_tags")), tags);
    return calls.withId(request, Product.class, id);
  }

  /**
   * Returns a product's tags ({@code GET /v1/products/{id}/product_tags}).
   *
   * @param productId the product id
   * @return the tags
   */
  public List<ProductTag> listTags(long productId) {
    return calls.list(Request.of("GET", path(productId, "product_tags")), ProductTag.class);
  }

  /**
   * Attaches tags and returns the resulting list ({@code PUT /v1/products/{id}/product_tags/add}).
   *
   * @param productId the product id
   * @param tags the tag names, sent comma-separated
   * @return the product's tags after the change
   */
  public List<ProductTag> addTags(long productId, List<String> tags) {
    return putTags(path(productId, "product_tags/add"), tags);
  }

  /**
   * Detaches tags and returns the resulting list ({@code PUT
   * /v1/products/{id}/product_tags/remove}).
   *
   * @param productId the product id
   * @param tags the tag names, sent comma-separated
   * @return the product's tags after the change
   */
  public List<ProductTag> removeTags(long productId, List<String> tags) {
    return putTags(path(productId, "product_tags/remove"), tags);
  }

  /**
   * Returns every shipping method a product can be bound to ({@code GET
   * /v1/products/bind_shippings}).
   *
   * @return the shipping method names
   */
  public List<String> listBindableShippings() {
    return shippingNames(Request.of("GET", PRODUCTS + "/bind_shippings"));
  }

  /**
   * Returns the shipping methods bound to a product ({@code GET /v1/products/{id}/bind_shippings}).
   *
   * @param productId the product id
   * @return the shipping method names
   */
  public List<String> getBindShippings(long productId) {
    return shippingNames(Request.of("GET", path(productId, "bind_shippings")));
  }

  /**
   * Replaces the shipping methods bound to a product ({@code POST
   * /v1/products/{id}/bind_shippings}).
   *
   * @param productId the product id
   * @param shippingNames the shipping method names
   * @return the bound shipping method names after the change
   */
  public List<String> bindShippings(long productId, List<String> shippingNames) {
    Objects.requireNonNull(shippingNames, "shippingNames");
    Request request =
        Params.body(
            Request.of("POST", path(productId, "bind_shippings")),
            Map.of("shipping_names", shippingNames));
    return shippingNames(request);
  }

  /**
   * Returns the description sections the Shop supports ({@code GET
   * /v1/products/get_product_description_setting_names}).
   *
   * @return the sections
   */
  public List<ProductDescriptionSettingName> listDescriptionSettingNames() {
    Request request = Request.of("GET", PRODUCTS + "/get_product_description_setting_names");
    return calls.list(request, ProductDescriptionSettingName.class);
  }

  /**
   * Returns a product's variants ({@code GET /v1/products/{id}/product_variants}).
   *
   * @param productId the product id
   * @return the variants
   */
  public List<ProductVariant> listVariants(long productId) {
    Request request = Request.of("GET", path(productId, VARIANTS));
    return calls.list(request, ProductVariant.class);
  }

  /**
   * Returns one variant ({@code GET /v1/products/{id}/product_variants/{variant_id}}).
   *
   * @param productId the product id
   * @param variantId the variant id
   * @return the variant, with {@code id} set
   * @throws NotFoundException also when the API answers with a null body
   */
  public ProductVariant getVariant(long productId, long variantId) {
    Request request = Request.of("GET", path(productId, VARIANTS + "/" + variantId));
    return calls.one(request, ProductVariant.class, variantId);
  }

  /**
   * Adds a variant ({@code POST /v1/products/{id}/product_variants}).
   *
   * @param productId the product id
   * @param variant the fields, e.g. {@code price}, {@code sku}, {@code option1}
   * @return the new variant
   */
  public ProductVariant createVariant(long productId, Map<String, ?> variant) {
    Request request = Params.body(Request.of("POST", path(productId, VARIANTS)), variant);
    return calls.object(request, ProductVariant.class);
  }

  /**
   * Changes a variant ({@code PUT /v1/products/{id}/product_variants/{variant_id}}).
   *
   * @param productId the product id
   * @param variantId the variant id
   * @param changes the fields to change
   * @return the variant, with {@code id} set
   */
  public ProductVariant updateVariant(long productId, long variantId, Map<String, ?> changes) {
    Request request =
        Params.body(Request.of("PUT", path(productId, VARIANTS + "/" + variantId)), changes);
    return calls.withId(request, ProductVariant.class, variantId);
  }

  /**
   * Removes a variant ({@code DELETE /v1/products/{id}/product_variants/{variant_id}}).
   *
   * @param productId the product id
   * @param variantId the variant id
   */
  public void deleteVariant(long productId, long variantId) {
    calls.call(Request.of("DELETE", path(productId, VARIANTS + "/" + variantId)));
  }

  /**
   * Returns the first page of the variants with a SKU ({@code GET
   * /v1/products/sku/{sku}/product_variants}).
   *
   * @param sku the SKU; percent-encoded as one path segment, so "/" cannot add a separator; must
   *     not be empty, "." or "..", nor longer than 256 characters (IllegalArgumentException)
   * @return the page
   */
  public Page<ProductVariant> listVariantsBySku(String sku) {
    return listVariantsBySku(sku, Map.of());
  }

  /**
   * Returns one page of the variants with a SKU ({@code GET
   * /v1/products/sku/{sku}/product_variants}).
   *
   * @param sku the SKU; percent-encoded as one path segment, so "/" cannot add a separator; must
   *     not be empty, "." or "..", nor longer than 256 characters (IllegalArgumentException)
   * @param query e.g. {@code page} and {@code per_page}
   * @return the page
   */
  public Page<ProductVariant> listVariantsBySku(String sku, Map<String, ?> query) {
    return calls.client().list(get(skuPath(sku), query), ProductVariant.class);
  }

  /**
   * Walks every variant with a SKU, page by page ({@code GET
   * /v1/products/sku/{sku}/product_variants}).
   *
   * @param sku the SKU; percent-encoded as one path segment, so "/" cannot add a separator; must
   *     not be empty, "." or "..", nor longer than 256 characters (IllegalArgumentException)
   * @return the variants, fetched lazily
   */
  public PageIterable<ProductVariant> allVariantsBySku(String sku) {
    return allVariantsBySku(sku, Map.of());
  }

  /**
   * Walks every variant with a SKU, page by page ({@code GET
   * /v1/products/sku/{sku}/product_variants}).
   *
   * @param sku the SKU; percent-encoded as one path segment, so "/" cannot add a separator; must
   *     not be empty, "." or "..", nor longer than 256 characters (IllegalArgumentException)
   * @param query e.g. the first {@code page} and {@code per_page}
   * @return the variants, fetched lazily
   */
  public PageIterable<ProductVariant> allVariantsBySku(String sku, Map<String, ?> query) {
    return calls.client().all(get(skuPath(sku), query), ProductVariant.class);
  }

  /**
   * Returns a product's options ({@code GET /v1/products/{id}/product_options}).
   *
   * @param productId the product id
   * @return the options
   */
  public List<ProductOption> listOptions(long productId) {
    return calls.list(Request.of("GET", path(productId, OPTIONS)), ProductOption.class);
  }

  /**
   * Returns one option ({@code GET /v1/products/{id}/product_options/{option_id}}).
   *
   * @param productId the product id
   * @param optionId the option id
   * @return the option, with {@code id} set
   * @throws NotFoundException also when the API answers with a null body
   */
  public ProductOption getOption(long productId, long optionId) {
    Request request = Request.of("GET", path(productId, OPTIONS + "/" + optionId));
    return calls.one(request, ProductOption.class, optionId);
  }

  /**
   * Adds an option ({@code POST /v1/products/{id}/product_options}).
   *
   * @param productId the product id
   * @param option the fields, e.g. {@code name} and {@code types}
   * @return the new option
   */
  public ProductOption createOption(long productId, Map<String, ?> option) {
    Request request = Params.body(Request.of("POST", path(productId, OPTIONS)), option);
    return calls.object(request, ProductOption.class);
  }

  /**
   * Changes an option ({@code PUT /v1/products/{id}/product_options/{option_id}}).
   *
   * @param productId the product id
   * @param optionId the option id
   * @param changes the fields to change
   * @return the option, with {@code id} set
   */
  public ProductOption updateOption(long productId, long optionId, Map<String, ?> changes) {
    Request request =
        Params.body(Request.of("PUT", path(productId, OPTIONS + "/" + optionId)), changes);
    return calls.withId(request, ProductOption.class, optionId);
  }

  /**
   * Removes an option ({@code DELETE /v1/products/{id}/product_options/{option_id}}).
   *
   * @param productId the product id
   * @param optionId the option id
   */
  public void deleteOption(long productId, long optionId) {
    calls.call(Request.of("DELETE", path(productId, OPTIONS + "/" + optionId)));
  }

  private List<ProductTag> putTags(String path, List<String> tags) {
    Request request = Params.body(Request.of("PUT", path), Map.of("tags", String.join(",", tags)));
    return calls.list(request, ProductTag.class);
  }

  private List<String> shippingNames(Request request) {
    List<String> names = calls.object(request, ShippingNames.class).shippingNames();
    return names == null ? List.of() : Collections.unmodifiableList(new ArrayList<>(names));
  }

  private static Request get(String path, Map<String, ?> query) {
    return Params.query(Request.of("GET", path), query);
  }

  private static String path(long productId) {
    return PRODUCTS + "/" + productId;
  }

  private static String path(long productId, String suffix) {
    return path(productId) + "/" + suffix;
  }

  private static String skuPath(String sku) {
    return PRODUCTS + "/sku/" + Params.segment(sku) + "/" + VARIANTS;
  }

  /** The body of the bind_shippings endpoints. */
  private record ShippingNames(List<String> shippingNames) {}
}
