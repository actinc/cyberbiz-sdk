package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// ProductsService exposes products, variants, options, photos, descriptions,
// tags, shipping bindings, and the v2 product endpoints.
type ProductsService struct {
	client *Client
}

// ProductListOptions are the parameters of List and All.
type ProductListOptions struct {
	ListOptions
}

// ProductSearchOptions are the parameters of Search. At least one of Q and
// Vendor must be set. The endpoint pages with Limit and Offset and sends no
// pagination headers.
type ProductSearchOptions struct {
	Q                 string               `url:"q,omitempty"`
	Limit             int                  `url:"limit,omitempty"`
	Offset            int                  `url:"offset,omitempty"`
	FilterPublished   *bool                `url:"filter_published,omitempty"` // defaults to true on the platform
	OrderBy           ProductSearchOrderBy `url:"order_by,omitempty"`
	Vendor            string               `url:"vendor,omitempty"`
	FilterBranchStore *bool                `url:"filter_branch_store,omitempty"`
}

// ProductCollectionSearchOptions are the parameters of SearchCollection.
type ProductCollectionSearchOptions struct {
	CollectionHandle string               `url:"collection_handle"` // required
	Limit            int                  `url:"limit,omitempty"`
	Offset           int                  `url:"offset,omitempty"`
	FilterPublished  *bool                `url:"filter_published,omitempty"`
	OrderBy          ProductSearchOrderBy `url:"order_by,omitempty"`
}

// ProductRelatedCollectionInput links a product to a collection in an update.
type ProductRelatedCollectionInput struct {
	CollectionID   int64  `json:"collection_id"`
	CollectionType string `json:"collection_type"` // "SmartCollection" or "CustomCollection"
}

// ProductCreateRequest is the body of Create. Title, Handle, Published and
// Price are required.
type ProductCreateRequest struct {
	Title                   string            `json:"title"`
	Handle                  string            `json:"handle"` // storefront URL slug
	Published               bool              `json:"published"`
	Price                   Money             `json:"price"`
	EnglishTitle            *string           `json:"english_title,omitzero"`
	SellFrom                Time              `json:"sell_from,omitzero"`
	SellTo                  Time              `json:"sell_to,omitzero"`
	ChannelName             *string           `json:"channel_name,omitzero"`
	ProductType             *string           `json:"product_type,omitzero"`
	ProductTypeCode         *string           `json:"product_type_code,omitzero"`
	Slogan                  *string           `json:"slogan,omitzero"`
	Brief                   *string           `json:"brief,omitzero"`
	BriefText               *string           `json:"brief_text,omitzero"`
	BriefIncludesHTML       *bool             `json:"brief_includes_html,omitzero"`
	BodyHTML                *string           `json:"body_html,omitzero"`
	Vendor                  *string           `json:"vendor,omitzero"`
	TagsText                *string           `json:"tags_text,omitzero"` // comma-separated tags
	SpecialCollectionID     *int64            `json:"special_collection_id,omitzero"`
	CustomCollectionIDs     *string           `json:"custom_collection_ids,omitzero"` // comma-separated ids
	TaxTypeID               TaxType           `json:"tax_type_id,omitzero"`
	BranchStoreID           *int64            `json:"branch_store_id,omitzero"`
	TemperatureTypes        []TemperatureType `json:"temperature_types,omitzero"`
	Searchable              *bool             `json:"searchable,omitzero"`
	GoogleProductCategoryID *int64            `json:"google_product_category_id,omitzero"`
	RequiredCustomerTags    *string           `json:"required_customer_tags,omitzero"` // comma-separated, custom feature
	// SKU is the SKU of the first variant the platform creates with the
	// product. Required for shops with the POS feature (422 otherwise).
	SKU *string `json:"sku,omitzero"`
}

// ProductUpdateRequest is the body of Update. Every field is optional; only
// the fields set are changed. Pointer fields send an explicit zero value.
type ProductUpdateRequest struct {
	Title                   *string                         `json:"title,omitzero"`
	EnglishTitle            *string                         `json:"english_title,omitzero"`
	Handle                  *string                         `json:"handle,omitzero"`
	Published               *bool                           `json:"published,omitzero"`
	SellFrom                Time                            `json:"sell_from,omitzero"`
	SellTo                  Time                            `json:"sell_to,omitzero"`
	ChannelName             *string                         `json:"channel_name,omitzero"`
	ProductType             *string                         `json:"product_type,omitzero"`
	ProductTypeCode         *string                         `json:"product_type_code,omitzero"`
	Slogan                  *string                         `json:"slogan,omitzero"`
	Brief                   *string                         `json:"brief,omitzero"`
	BriefText               *string                         `json:"brief_text,omitzero"`
	BriefIncludesHTML       *bool                           `json:"brief_includes_html,omitzero"`
	BodyHTML                *string                         `json:"body_html,omitzero"`
	Vendor                  *string                         `json:"vendor,omitzero"`
	TagsText                *string                         `json:"tags_text,omitzero"`
	Price                   *Money                          `json:"price,omitzero"`
	SpecialCollectionID     *int64                          `json:"special_collection_id,omitzero"`
	CustomCollectionIDs     *string                         `json:"custom_collection_ids,omitzero"`
	TaxTypeID               TaxType                         `json:"tax_type_id,omitzero"`
	RelatedCollections      []ProductRelatedCollectionInput `json:"related_collections,omitzero"`
	TemperatureTypes        []TemperatureType               `json:"temperature_types,omitzero"`
	Searchable              *bool                           `json:"searchable,omitzero"`
	GoogleProductCategoryID *int64                          `json:"google_product_category_id,omitzero"`
	RequiredCustomerTags    *string                         `json:"required_customer_tags,omitzero"`
}

// ProductPosShopBatchCreateRequest is the body of CreateForPosShops. Title
// and PosShopIDs are required; at most 20 POS shops per call.
type ProductPosShopBatchCreateRequest struct {
	Title               string  `json:"title"`
	PosShopIDs          string  `json:"pos_shop_ids"` // comma-separated POS shop ids
	EnglishTitle        *string `json:"english_title,omitzero"`
	SellFrom            Time    `json:"sell_from,omitzero"`
	SellTo              Time    `json:"sell_to,omitzero"`
	ProductType         *string `json:"product_type,omitzero"`
	ProductTypeCode     *string `json:"product_type_code,omitzero"`
	Slogan              *string `json:"slogan,omitzero"`
	Brief               *string `json:"brief,omitzero"`
	BodyHTML            *string `json:"body_html,omitzero"`
	Vendor              *string `json:"vendor,omitzero"`
	TagsText            *string `json:"tags_text,omitzero"`
	Price               *Money  `json:"price,omitzero"`
	SpecialCollectionID *int64  `json:"special_collection_id,omitzero"`
	CustomCollectionIDs *string `json:"custom_collection_ids,omitzero"`
	// SKU is the SKU of the first variant of each created product. The
	// endpoint is POS-only and the platform requires it (422 otherwise).
	SKU *string `json:"sku,omitzero"`
}

// ProductSEOMetaTagsUpdateRequest is the body of UpdateSEOMetaTags.
type ProductSEOMetaTagsUpdateRequest struct {
	Title       *string `json:"title,omitzero"`
	Description *string `json:"description,omitzero"`
	Keywords    *string `json:"keywords,omitzero"`
}

// List returns one page of products (GET /v1/products).
func (s *ProductsService) List(ctx context.Context, opts *ProductListOptions) (*Page[Product], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Product](ctx, s.client, "v1/products", q)
}

// All walks every page of products (GET /v1/products).
func (s *ProductsService) All(ctx context.Context, opts *ProductListOptions) iter.Seq2[Product, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Product, error) bool) { yield(Product{}, err) }
	}
	return listAll[Product](ctx, s.client, "v1/products", q)
}

// Get returns one product (GET /v1/products/{id}).
func (s *ProductsService) Get(ctx context.Context, id int64) (*Product, *Response, error) {
	var out Product
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/products/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Create creates a product (POST /v1/products).
func (s *ProductsService) Create(ctx context.Context, req *ProductCreateRequest) (*Product, *Response, error) {
	var out Product
	resp, err := s.client.post(ctx, "v1/products", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Update changes a product (PUT /v1/products/{id}).
//
// The platform fills tax_type_id and temperature_types with defaults
// (inclusive_tax, 常溫) when a request leaves them out, which would silently
// reset a zero-rated, tax-exempt, refrigerated or frozen product. When
// TaxTypeID or TemperatureTypes is unset, Update therefore first reads the
// product (GET /v1/products/{id}) and sends its current values. Set both to
// skip that extra request. req itself is not modified.
func (s *ProductsService) Update(ctx context.Context, id int64, req *ProductUpdateRequest) (*Product, *Response, error) {
	body, resp, err := s.withCurrentTaxAndTemperature(ctx, id, req)
	if err != nil {
		return nil, resp, err
	}
	var out Product
	resp, err = s.client.put(ctx, fmt.Sprintf("v1/products/%d", id), body, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// withCurrentTaxAndTemperature returns a copy of req whose TaxTypeID and
// TemperatureTypes are filled from the stored product when unset.
func (s *ProductsService) withCurrentTaxAndTemperature(ctx context.Context, id int64, req *ProductUpdateRequest) (*ProductUpdateRequest, *Response, error) {
	body := ProductUpdateRequest{}
	if req != nil {
		body = *req
	}
	if body.TaxTypeID != "" && len(body.TemperatureTypes) > 0 {
		return &body, nil, nil
	}
	current, resp, err := s.Get(ctx, id)
	if err != nil {
		return nil, resp, fmt.Errorf("cyberbiz: read product %d before update: %w", id, err)
	}
	if body.TaxTypeID == "" {
		body.TaxTypeID = current.TaxTypeID
	}
	if len(body.TemperatureTypes) == 0 {
		body.TemperatureTypes = slices.Clone(current.TemperatureTypes)
	}
	return &body, resp, nil
}

// Delete removes a product (DELETE /v1/products/{id}).
func (s *ProductsService) Delete(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/products/%d", id), nil)
}

// CreateForPosShops creates the same product in several POS shops at once
// (POST /v1/products/pos_shop_batch).
func (s *ProductsService) CreateForPosShops(ctx context.Context, req *ProductPosShopBatchCreateRequest) ([]Product, *Response, error) {
	var out []Product
	resp, err := s.client.post(ctx, "v1/products/pos_shop_batch", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// Search finds products by keyword or vendor (GET /v1/products/search). The
// endpoint pages with Limit and Offset and sends no pagination headers.
func (s *ProductsService) Search(ctx context.Context, opts *ProductSearchOptions) ([]Product, *Response, error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, nil, err
	}
	var out []Product
	resp, err := s.client.get(ctx, "v1/products/search", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// SearchCollection lists the products of a collection by handle
// (GET /v1/products/search/collection).
func (s *ProductsService) SearchCollection(ctx context.Context, opts *ProductCollectionSearchOptions) ([]Product, *Response, error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, nil, err
	}
	var out []Product
	resp, err := s.client.get(ctx, "v1/products/search/collection", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// UpdateSEOMetaTags changes a product's SEO fields and returns the product
// (PUT /v1/products/{id}/seo_meta_tags).
func (s *ProductsService) UpdateSEOMetaTags(ctx context.Context, id int64, req *ProductSEOMetaTagsUpdateRequest) (*Product, *Response, error) {
	var out Product
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/products/%d/seo_meta_tags", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// ListTags returns a product's tags (GET /v1/products/{id}/product_tags).
func (s *ProductsService) ListTags(ctx context.Context, productID int64) ([]ProductTag, *Response, error) {
	var out []ProductTag
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/products/%d/product_tags", productID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// productsTagsBody is the body of the product_tags/add and /remove endpoints, which
// take the tags as one comma-separated string.
type productsTagsBody struct {
	Tags string `json:"tags"`
}

// AddTags attaches tags to a product and returns the resulting tag list
// (PUT /v1/products/{id}/product_tags/add).
func (s *ProductsService) AddTags(ctx context.Context, productID int64, tags []string) ([]ProductTag, *Response, error) {
	return s.putTags(ctx, fmt.Sprintf("v1/products/%d/product_tags/add", productID), tags)
}

// RemoveTags detaches tags from a product and returns the resulting tag
// list (PUT /v1/products/{id}/product_tags/remove).
func (s *ProductsService) RemoveTags(ctx context.Context, productID int64, tags []string) ([]ProductTag, *Response, error) {
	return s.putTags(ctx, fmt.Sprintf("v1/products/%d/product_tags/remove", productID), tags)
}

func (s *ProductsService) putTags(ctx context.Context, path string, tags []string) ([]ProductTag, *Response, error) {
	var out []ProductTag
	resp, err := s.client.put(ctx, path, productsTagsBody{Tags: strings.Join(tags, ",")}, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// ListBindableShippings returns the names of every shipping method a product
// can be bound to (GET /v1/products/bind_shippings).
func (s *ProductsService) ListBindableShippings(ctx context.Context) ([]string, *Response, error) {
	var out ProductShippingNames
	resp, err := s.client.get(ctx, "v1/products/bind_shippings", nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out.ShippingNames, resp, nil
}

// GetBindShippings returns the shipping methods bound to a product
// (GET /v1/products/{id}/bind_shippings).
func (s *ProductsService) GetBindShippings(ctx context.Context, productID int64) ([]string, *Response, error) {
	var out ProductShippingNames
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/products/%d/bind_shippings", productID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out.ShippingNames, resp, nil
}

// BindShippings replaces the shipping methods bound to a product
// (POST /v1/products/{id}/bind_shippings).
func (s *ProductsService) BindShippings(ctx context.Context, productID int64, shippingNames []string) ([]string, *Response, error) {
	var out ProductShippingNames
	body := ProductShippingNames{ShippingNames: shippingNames}
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/products/%d/bind_shippings", productID), body, &out)
	if err != nil {
		return nil, resp, err
	}
	return out.ShippingNames, resp, nil
}

// ListDescriptionSettingNames returns the description sections the shop
// supports (GET /v1/products/get_product_description_setting_names).
func (s *ProductsService) ListDescriptionSettingNames(ctx context.Context) ([]ProductDescriptionSettingName, *Response, error) {
	var out []ProductDescriptionSettingName
	resp, err := s.client.get(ctx, "v1/products/get_product_description_setting_names", nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}
