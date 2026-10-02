package cyberbiz

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// MaxPerPage is the largest page size the platform accepts.
const MaxPerPage = 50

// ListOptions are the pagination parameters shared by every list endpoint.
// Page is 1-based; PerPage defaults to 50 and cannot exceed [MaxPerPage];
// Offset skips that many records before paging starts. The first record on a
// page is offset + (page-1)*per_page + 1.
type ListOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
	Offset  int `url:"offset,omitempty"`
}

// Pagination is what CYBERBIZ reports in the X-Page, X-Per-Page, X-Offset,
// X-Total, X-Total-Pages, X-Next-Page and X-Prev-Page response headers.
// NextPage and PrevPage are zero when there is no such page.
type Pagination struct {
	Page       int
	PerPage    int
	Offset     int
	Total      int
	TotalPages int
	NextPage   int
	PrevPage   int
}

// HasNext reports whether another page follows this one.
func (p Pagination) HasNext() bool { return p.NextPage > 0 }

// Page is one page of a list response.
type Page[T any] struct {
	Items      []T
	Pagination Pagination
	Response   *Response
}

func parsePagination(h http.Header) Pagination {
	get := func(name string) int {
		n, _ := strconv.Atoi(h.Get(name))
		return n
	}
	return Pagination{
		Page:       get("X-Page"),
		PerPage:    get("X-Per-Page"),
		Offset:     get("X-Offset"),
		Total:      get("X-Total"),
		TotalPages: get("X-Total-Pages"),
		NextPage:   get("X-Next-Page"),
		PrevPage:   get("X-Prev-Page"),
	}
}

// pageFetcher fetches page number page.
type pageFetcher[T any] func(ctx context.Context, page int) (*Page[T], error)

// allPages walks every page starting from startPage and yields each item.
// It stops at the first error, which is yielded with a zero item.
func allPages[T any](ctx context.Context, startPage int, fetch pageFetcher[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		page := max(startPage, 1)
		for {
			result, err := fetch(ctx, page)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range result.Items {
				if !yield(item, nil) {
					return
				}
			}
			if !result.Pagination.HasNext() || len(result.Items) == 0 {
				return
			}
			page = result.Pagination.NextPage
		}
	}
}

// withPage returns a copy of q with the page parameter replaced.
func withPage(q url.Values, page int) url.Values {
	out := q.Clone()
	if out == nil {
		out = url.Values{}
	}
	out.Set("page", strconv.Itoa(page))
	if out.Get("per_page") == "" {
		out.Set("per_page", strconv.Itoa(MaxPerPage))
	}
	return out
}
