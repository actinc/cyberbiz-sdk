package cyberbiz

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// VariantDiscountCollectionCreateRequest is the body of
// CreateVariantDiscount. Only Title is required; set the field that matches
// VariantDiscountCollectionType (Amount, Percentage or DiscountAmount).
type VariantDiscountCollectionCreateRequest struct {
	Title                         string              `json:"title"`
	Published                     *bool               `json:"published,omitzero"`
	StartDate                     Time                `json:"start_date,omitzero"`
	EndDate                       Time                `json:"end_date,omitzero"`
	VariantDiscountCollectionType VariantDiscountType `json:"variant_discount_collection_type,omitzero"`
	Amount                        *Money              `json:"amount,omitzero"`          // fixed unit price
	Percentage                    *float64            `json:"percentage,omitzero"`      // percentage off each unit
	DiscountAmount                *Money              `json:"discount_amount,omitzero"` // fixed amount off
	VariantIDs                    []int64             `json:"variant_ids,omitzero"`
}

// VariantDiscountCollectionUpdateRequest is the body of
// UpdateVariantDiscount; every field is optional and unset fields are left
// unchanged. Variants are managed with the *VariantDiscountVariants methods.
type VariantDiscountCollectionUpdateRequest struct {
	Title                         *string             `json:"title,omitzero"`
	Published                     *bool               `json:"published,omitzero"`
	StartDate                     Time                `json:"start_date,omitzero"`
	EndDate                       Time                `json:"end_date,omitzero"`
	VariantDiscountCollectionType VariantDiscountType `json:"variant_discount_collection_type,omitzero"`
	Amount                        *Money              `json:"amount,omitzero"`
	Percentage                    *float64            `json:"percentage,omitzero"`
	DiscountAmount                *Money              `json:"discount_amount,omitzero"`
}

// ListVariantDiscount returns one page of variant-discount collections
// (GET /v1/variant_discount_collections).
func (s *CollectionsService) ListVariantDiscount(ctx context.Context, opts *CollectionListOptions) (*Page[VariantDiscountCollection], error) {
	return collectionsList[VariantDiscountCollection](ctx, s.client, "v1/variant_discount_collections", opts)
}

// AllVariantDiscount walks every page of variant-discount collections
// (GET /v1/variant_discount_collections).
func (s *CollectionsService) AllVariantDiscount(ctx context.Context, opts *CollectionListOptions) iter.Seq2[VariantDiscountCollection, error] {
	return collectionsAll[VariantDiscountCollection](ctx, s.client, "v1/variant_discount_collections", opts)
}

// collectionsVariantDiscountDetail tolerates the documented detail shape, which the
// swagger and Postman both give as a one-element array, as well as a plain
// object.
type collectionsVariantDiscountDetail struct {
	VariantDiscountCollection
}

// UnmarshalJSONFrom decodes either a bare collection object or a one-element
// array wrapping it, leaving the wrapped collection in d.
func (d *collectionsVariantDiscountDetail) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() != '[' {
		return json.UnmarshalDecode(dec, &d.VariantDiscountCollection)
	}
	var items []VariantDiscountCollection
	if err := json.UnmarshalDecode(dec, &items); err != nil {
		return err
	}
	if len(items) > 0 {
		d.VariantDiscountCollection = items[0]
	}
	return nil
}

// GetVariantDiscount returns one variant-discount collection. The docs
// describe the response as a one-element array; both that and a bare object
// are accepted (GET /v1/variant_discount_collections/{id}).
func (s *CollectionsService) GetVariantDiscount(ctx context.Context, id int64) (*VariantDiscountCollection, *Response, error) {
	var out collectionsVariantDiscountDetail
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/variant_discount_collections/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out.VariantDiscountCollection, resp, nil
}

// CreateVariantDiscount creates a variant-discount collection
// (POST /v1/variant_discount_collections).
func (s *CollectionsService) CreateVariantDiscount(ctx context.Context, req *VariantDiscountCollectionCreateRequest) (*VariantDiscountCollection, *Response, error) {
	return collectionsWrite[VariantDiscountCollection](ctx, s.client, http.MethodPost, "v1/variant_discount_collections", req)
}

// UpdateVariantDiscount changes a variant-discount collection
// (PUT /v1/variant_discount_collections/{id}).
func (s *CollectionsService) UpdateVariantDiscount(ctx context.Context, id int64, req *VariantDiscountCollectionUpdateRequest) (*VariantDiscountCollection, *Response, error) {
	return collectionsWrite[VariantDiscountCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/variant_discount_collections/%d", id), req)
}

// DeleteVariantDiscount removes a variant-discount collection
// (DELETE /v1/variant_discount_collections/{id}).
func (s *CollectionsService) DeleteVariantDiscount(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/variant_discount_collections/%d", id), nil)
}

// AddVariantDiscountVariants adds product variants to a variant-discount
// collection (POST /v1/variant_discount_collections/{id}/variants).
func (s *CollectionsService) AddVariantDiscountVariants(ctx context.Context, id int64, variantIDs []int64) (*VariantDiscountCollection, *Response, error) {
	return collectionsWrite[VariantDiscountCollection](ctx, s.client, http.MethodPost,
		fmt.Sprintf("v1/variant_discount_collections/%d/variants", id), collectionsVariantIDsBody{VariantIDs: variantIDs})
}

// ReplaceVariantDiscountVariants replaces the whole variant list of a
// variant-discount collection (PUT /v1/variant_discount_collections/{id}/variants).
func (s *CollectionsService) ReplaceVariantDiscountVariants(ctx context.Context, id int64, variantIDs []int64) (*VariantDiscountCollection, *Response, error) {
	return collectionsWrite[VariantDiscountCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/variant_discount_collections/%d/variants", id), collectionsVariantIDsBody{VariantIDs: variantIDs})
}

// RemoveVariantDiscountVariants removes product variants from a
// variant-discount collection; the ids travel in the query string as
// repeated variant_ids[] parameters
// (DELETE /v1/variant_discount_collections/{id}/variants).
func (s *CollectionsService) RemoveVariantDiscountVariants(ctx context.Context, id int64, variantIDs []int64) (*VariantDiscountCollection, *Response, error) {
	q := url.Values{}
	for _, v := range variantIDs {
		q.Add("variant_ids[]", strconv.FormatInt(v, 10))
	}
	return collectionsDeleteWithQuery[VariantDiscountCollection](ctx, s.client,
		fmt.Sprintf("v1/variant_discount_collections/%d/variants", id), q)
}
