package cyberbiz

import (
	"context"
	"iter"
	"net/http"
	"net/url"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CollectionsService exposes every collection kind: custom, smart, special,
// add-buy, limit, variant-discount and VIP (/v1/*_collections), plus the v2
// searchable lists of custom and smart collections (/v2/*_collections).
// Methods are grouped by kind: ListCustom, GetSmart, CreateSpecial, ...
type CollectionsService struct {
	client *Client
}

// CollectionListOptions are the parameters of every v1 collection list.
type CollectionListOptions struct {
	ListOptions
}

// CollectionSearchOptions are the parameters of the v2 collection lists,
// which add a title/handle search.
type CollectionSearchOptions struct {
	ListOptions
	// Q matches against the collection title or handle (substring).
	Q string `url:"q,omitempty"`
}

// collectionsProductIDsBody is the body of every "product_ids" membership endpoint.
type collectionsProductIDsBody struct {
	ProductIDs IDList `json:"product_ids"`
}

// collectionsVariantIDsBody is the body of the variant membership endpoints; the
// platform takes a JSON array here rather than a comma-separated string.
type collectionsVariantIDsBody struct {
	VariantIDs []int64 `json:"variant_ids"`
}

// collectionsList fetches one page of any collection kind.
func collectionsList[T any](ctx context.Context, c *Client, path string, opts any) (*Page[T], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[T](ctx, c, path, q)
}

// collectionsAll walks every page of any collection kind.
func collectionsAll[T any](ctx context.Context, c *Client, path string, opts any) iter.Seq2[T, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(T, error) bool) {
			var zero T
			yield(zero, err)
		}
	}
	return listAll[T](ctx, c, path, q)
}

// collectionsGet fetches one collection and stamps id onto it, because the
// v1 detail entities omit the id.
func collectionsGet[T any](ctx context.Context, c *Client, path string, setID func(*T)) (*T, *Response, error) {
	var out T
	resp, err := c.getOne(ctx, path, nil, &out)
	if err != nil {
		return nil, resp, err
	}
	setID(&out)
	return &out, resp, nil
}

// collectionsWrite sends a create or update and decodes the returned entity.
func collectionsWrite[T any](ctx context.Context, c *Client, method, path string, body any) (*T, *Response, error) {
	var out T
	var resp *Response
	var err error
	switch method {
	case http.MethodPost:
		resp, err = c.post(ctx, path, body, &out)
	default:
		resp, err = c.put(ctx, path, body, &out)
	}
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// collectionsDeleteWithQuery sends a DELETE whose parameters travel in the query
// string, as the membership-removal endpoints require, and decodes the
// returned entity.
func collectionsDeleteWithQuery[T any](ctx context.Context, c *Client, path string, q url.Values) (*T, *Response, error) {
	var out T
	resp, err := c.Do(ctx, &Request{Method: http.MethodDelete, Path: path, Query: q}, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// collectionsProductIDsQuery is the query string of the DELETE .../products endpoints.
func collectionsProductIDsQuery(ids IDList) url.Values {
	return url.Values{"product_ids": {ids.String()}}
}
