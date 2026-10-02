package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// AddBuyCollectionCreateRequest is the body of CreateAddBuy. Title,
// Published, Price and ItemLimit are required.
type AddBuyCollectionCreateRequest struct {
	Title         string        `json:"title"`
	Published     bool          `json:"published"`
	ProductsOrder ProductsOrder `json:"products_order,omitzero"`
	StartDate     Time          `json:"start_date,omitzero"`
	EndDate       Time          `json:"end_date,omitzero"`
	// Price is the minimum order subtotal that unlocks the add-on offer.
	Price Money `json:"price"`
	// ItemLimit is the maximum number of add-on items per order.
	ItemLimit int `json:"item_limit"`
}

// AddBuyCollectionUpdateRequest is the body of UpdateAddBuy; every field is
// optional and unset fields are left unchanged.
type AddBuyCollectionUpdateRequest struct {
	Title         *string       `json:"title,omitzero"`
	Published     *bool         `json:"published,omitzero"`
	ProductsOrder ProductsOrder `json:"products_order,omitzero"`
	StartDate     Time          `json:"start_date,omitzero"`
	EndDate       Time          `json:"end_date,omitzero"`
	Price         *Money        `json:"price,omitzero"`
	ItemLimit     *int          `json:"item_limit,omitzero"`
}

// AddBuyProductRequest is the body of AddAddBuyProduct: the product offered
// and its add-on price.
type AddBuyProductRequest struct {
	ProductID int64 `json:"product_id"`
	Price     Money `json:"price"`
}

// AddBuyProductUpdateRequest is the body of UpdateAddBuyProduct.
type AddBuyProductUpdateRequest struct {
	Price *Money `json:"price,omitzero"`
}

// ListAddBuy returns one page of add-buy collections (GET /v1/add_buy_collections).
func (s *CollectionsService) ListAddBuy(ctx context.Context, opts *CollectionListOptions) (*Page[AddBuyCollection], error) {
	return collectionsList[AddBuyCollection](ctx, s.client, "v1/add_buy_collections", opts)
}

// AllAddBuy walks every page of add-buy collections (GET /v1/add_buy_collections).
func (s *CollectionsService) AllAddBuy(ctx context.Context, opts *CollectionListOptions) iter.Seq2[AddBuyCollection, error] {
	return collectionsAll[AddBuyCollection](ctx, s.client, "v1/add_buy_collections", opts)
}

// GetAddBuy returns one add-buy collection (GET /v1/add_buy_collections/{id}).
func (s *CollectionsService) GetAddBuy(ctx context.Context, id int64) (*AddBuyCollection, *Response, error) {
	return collectionsGet(ctx, s.client, fmt.Sprintf("v1/add_buy_collections/%d", id),
		func(c *AddBuyCollection) { c.ID = id })
}

// CreateAddBuy creates an add-buy collection (POST /v1/add_buy_collections).
func (s *CollectionsService) CreateAddBuy(ctx context.Context, req *AddBuyCollectionCreateRequest) (*AddBuyCollection, *Response, error) {
	return collectionsWrite[AddBuyCollection](ctx, s.client, http.MethodPost, "v1/add_buy_collections", req)
}

// UpdateAddBuy changes an add-buy collection (PUT /v1/add_buy_collections/{id}).
func (s *CollectionsService) UpdateAddBuy(ctx context.Context, id int64, req *AddBuyCollectionUpdateRequest) (*AddBuyCollection, *Response, error) {
	return collectionsWrite[AddBuyCollection](ctx, s.client, http.MethodPut, fmt.Sprintf("v1/add_buy_collections/%d", id), req)
}

// DeleteAddBuy removes an add-buy collection (DELETE /v1/add_buy_collections/{id}).
func (s *CollectionsService) DeleteAddBuy(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/add_buy_collections/%d", id), nil)
}

// AddAddBuyProduct adds one product with its add-on price
// (POST /v1/add_buy_collections/{id}/products).
func (s *CollectionsService) AddAddBuyProduct(ctx context.Context, id int64, req *AddBuyProductRequest) (*AddBuyCollection, *Response, error) {
	return collectionsWrite[AddBuyCollection](ctx, s.client, http.MethodPost,
		fmt.Sprintf("v1/add_buy_collections/%d/products", id), req)
}

// UpdateAddBuyProduct changes the add-on price of one product
// (PUT /v1/add_buy_collections/{id}/products/{product_id}).
func (s *CollectionsService) UpdateAddBuyProduct(ctx context.Context, id, productID int64, req *AddBuyProductUpdateRequest) (*AddBuyCollection, *Response, error) {
	return collectionsWrite[AddBuyCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/add_buy_collections/%d/products/%d", id, productID), req)
}

// RemoveAddBuyProduct removes one product from an add-buy collection
// (DELETE /v1/add_buy_collections/{id}/products/{product_id}).
func (s *CollectionsService) RemoveAddBuyProduct(ctx context.Context, id, productID int64) (*AddBuyCollection, *Response, error) {
	var out AddBuyCollection
	resp, err := s.client.delete(ctx, fmt.Sprintf("v1/add_buy_collections/%d/products/%d", id, productID), &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ReorderAddBuyProducts sets the manual product order of an add-buy
// collection (PUT /v1/add_buy_collections/{id}/products/manual_order).
func (s *CollectionsService) ReorderAddBuyProducts(ctx context.Context, id int64, productIDs IDList) (*AddBuyCollection, *Response, error) {
	return collectionsWrite[AddBuyCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/add_buy_collections/%d/products/manual_order", id), collectionsProductIDsBody{ProductIDs: productIDs})
}
