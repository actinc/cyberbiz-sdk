package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/url"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// ProductVariantCreateRequest is the body of CreateVariant. Position,
// InventoryManagement, InventoryQuantity, InventoryPolicy and
// RequiresShipping are required.
type ProductVariantCreateRequest struct {
	Position                int             `json:"position"`
	InventoryManagement     bool            `json:"inventory_management"`
	InventoryQuantity       int             `json:"inventory_quantity"` // ignored when the shop uses POS
	InventoryPolicy         InventoryPolicy `json:"inventory_policy"`
	RequiresShipping        bool            `json:"requires_shipping"`
	Price                   *Money          `json:"price,omitzero"`
	Cost                    *Money          `json:"cost,omitzero"`
	CompareAtPrice          *Money          `json:"compare_at_price,omitzero"`
	Meas                    *float64        `json:"meas,omitzero"`
	MaxUsableBonus          *Money          `json:"max_usable_bonus,omitzero"`
	Weight                  *float64        `json:"weight,omitzero"` // kilograms
	Option1                 *string         `json:"option1,omitzero"`
	Option2                 *string         `json:"option2,omitzero"`
	Option3                 *string         `json:"option3,omitzero"`
	Sold                    *int            `json:"sold,omitzero"`
	SafetyInventoryQuantity *int            `json:"safety_inventory_quantity,omitzero"`
	SKU                     *string         `json:"sku,omitzero"`
	QC                      *string         `json:"qc,omitzero"`
	BonusCollectionTitles   *string         `json:"bonus_collection_titles,omitzero"` // comma-separated
	HoneycombSync           *bool           `json:"honeycomb_sync,omitzero"`
}

// ProductVariantUpdateRequest is the body of UpdateVariant. Every field is
// optional; only the fields set are changed.
type ProductVariantUpdateRequest struct {
	Position                *int            `json:"position,omitzero"`
	Price                   *Money          `json:"price,omitzero"`
	Cost                    *Money          `json:"cost,omitzero"`
	CompareAtPrice          *Money          `json:"compare_at_price,omitzero"`
	Meas                    *float64        `json:"meas,omitzero"`
	MaxUsableBonus          *Money          `json:"max_usable_bonus,omitzero"`
	Weight                  *float64        `json:"weight,omitzero"`
	Option1                 *string         `json:"option1,omitzero"`
	Option2                 *string         `json:"option2,omitzero"`
	Option3                 *string         `json:"option3,omitzero"`
	InventoryManagement     *bool           `json:"inventory_management,omitzero"`
	InventoryQuantity       *int            `json:"inventory_quantity,omitzero"`
	Sold                    *int            `json:"sold,omitzero"`
	SafetyInventoryQuantity *int            `json:"safety_inventory_quantity,omitzero"`
	InventoryPolicy         InventoryPolicy `json:"inventory_policy,omitzero"`
	SKU                     *string         `json:"sku,omitzero"`
	QC                      *string         `json:"qc,omitzero"`
	RequiresShipping        *bool           `json:"requires_shipping,omitzero"`
	BonusCollectionTitles   *string         `json:"bonus_collection_titles,omitzero"`
	HoneycombSync           *bool           `json:"honeycomb_sync,omitzero"`
}

// ProductOptionCreateRequest is the body of CreateOption. Both fields are
// required.
type ProductOptionCreateRequest struct {
	Name  string `json:"name"`
	Types string `json:"types"` // comma-separated option values
}

// ProductOptionUpdateRequest is the body of UpdateOption.
type ProductOptionUpdateRequest struct {
	Name  *string `json:"name,omitzero"`
	Types *string `json:"types,omitzero"`
}

// ListVariants returns every variant of a product; the endpoint is not
// paginated (GET /v1/products/{id}/product_variants).
func (s *ProductsService) ListVariants(ctx context.Context, productID int64) ([]ProductVariant, *Response, error) {
	var out []ProductVariant
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/products/%d/product_variants", productID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// GetVariant returns one variant (GET /v1/products/{id}/product_variants/{id}).
func (s *ProductsService) GetVariant(ctx context.Context, productID, variantID int64) (*ProductVariant, *Response, error) {
	var out ProductVariant
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/products/%d/product_variants/%d", productID, variantID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = variantID
	return &out, resp, nil
}

// CreateVariant adds a variant to a product (POST /v1/products/{id}/product_variants).
func (s *ProductsService) CreateVariant(ctx context.Context, productID int64, req *ProductVariantCreateRequest) (*ProductVariant, *Response, error) {
	var out ProductVariant
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/products/%d/product_variants", productID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateVariant changes a variant (PUT /v1/products/{id}/product_variants/{id}).
func (s *ProductsService) UpdateVariant(ctx context.Context, productID, variantID int64, req *ProductVariantUpdateRequest) (*ProductVariant, *Response, error) {
	var out ProductVariant
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/products/%d/product_variants/%d", productID, variantID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = variantID
	return &out, resp, nil
}

// DeleteVariant removes a variant (DELETE /v1/products/{id}/product_variants/{id}).
func (s *ProductsService) DeleteVariant(ctx context.Context, productID, variantID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/products/%d/product_variants/%d", productID, variantID), nil)
}

// ListVariantsBySKU returns one page of the variants that carry sku across
// all products (GET /v1/products/sku/{sku}/product_variants).
func (s *ProductsService) ListVariantsBySKU(ctx context.Context, sku string, opts *ListOptions) (*Page[ProductVariant], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[ProductVariant](ctx, s.client, productsVariantsBySKUPath(sku), q)
}

// AllVariantsBySKU walks every page of the variants that carry sku
// (GET /v1/products/sku/{sku}/product_variants).
func (s *ProductsService) AllVariantsBySKU(ctx context.Context, sku string, opts *ListOptions) iter.Seq2[ProductVariant, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(ProductVariant, error) bool) { yield(ProductVariant{}, err) }
	}
	return listAll[ProductVariant](ctx, s.client, productsVariantsBySKUPath(sku), q)
}

func productsVariantsBySKUPath(sku string) string {
	return "v1/products/sku/" + url.PathEscape(sku) + "/product_variants"
}

// ListOptions returns the option axes of a product
// (GET /v1/products/{id}/product_options).
func (s *ProductsService) ListOptions(ctx context.Context, productID int64) ([]ProductOption, *Response, error) {
	var out []ProductOption
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/products/%d/product_options", productID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// GetOption returns one option axis (GET /v1/products/{id}/product_options/{id}).
func (s *ProductsService) GetOption(ctx context.Context, productID, optionID int64) (*ProductOption, *Response, error) {
	var out ProductOption
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/products/%d/product_options/%d", productID, optionID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = optionID
	return &out, resp, nil
}

// CreateOption adds an option axis to a product (POST /v1/products/{id}/product_options).
func (s *ProductsService) CreateOption(ctx context.Context, productID int64, req *ProductOptionCreateRequest) (*ProductOption, *Response, error) {
	var out ProductOption
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/products/%d/product_options", productID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateOption changes an option axis (PUT /v1/products/{id}/product_options/{id}).
func (s *ProductsService) UpdateOption(ctx context.Context, productID, optionID int64, req *ProductOptionUpdateRequest) (*ProductOption, *Response, error) {
	var out ProductOption
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/products/%d/product_options/%d", productID, optionID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = optionID
	return &out, resp, nil
}

// DeleteOption removes an option axis (DELETE /v1/products/{id}/product_options/{id}).
func (s *ProductsService) DeleteOption(ctx context.Context, productID, optionID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/products/%d/product_options/%d", productID, optionID), nil)
}
