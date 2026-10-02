package cyberbiz

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// ContentService exposes the v2 storefront content endpoints: custom pages
// (/v2/pages), menus (/v2/menus), categories (/v2/categories), product feeds
// (/v2/product_feeds) and shop e-mails (/v2/shop_emails).
type ContentService struct {
	client *Client
}

// PageCreateRequest is the body of CreatePage.
type PageCreateRequest struct {
	Title string `json:"title"`
}

// PageUpdateRequest is the body of UpdatePage; every field is optional. To
// change the page HTML set both SectionID and SectionHTML; only the custom
// HTML section created by CreatePage can be changed.
type PageUpdateRequest struct {
	Title  *string    `json:"title,omitzero"`
	Handle *string    `json:"handle,omitzero"`
	Status PageStatus `json:"status,omitzero"`
	// SectionID is the custom HTML section id from CustomPage.HTMLSections.
	SectionID *string `json:"section_id,omitzero"`
	// SectionHTML is the new raw HTML; the SDK applies the platform's
	// required encoding (HTML-escaped JSON string) on the wire.
	SectionHTML *string `json:"-"`
}

// contentPageUpdateBody is PageUpdateRequest as sent on the wire.
type contentPageUpdateBody struct {
	PageUpdateRequest
	SectionContent string `json:"section_content,omitzero"`
}

// contentEncodeSectionHTML produces the section_content wire form: the HTML
// as a JSON string with <, > and & escaped as \u003c, \u003e and \u0026.
func contentEncodeSectionHTML(html string) (string, error) {
	b, err := json.Marshal(html, jsontext.EscapeForHTML(true))
	if err != nil {
		return "", fmt.Errorf("cyberbiz: encoding section content: %w", err)
	}
	return string(b), nil
}

// CreatePage creates a custom page with one HTML section (POST /v2/pages).
func (s *ContentService) CreatePage(ctx context.Context, req *PageCreateRequest) (*CustomPage, *Response, error) {
	var out CustomPage
	resp, err := s.client.post(ctx, "v2/pages", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdatePage changes a custom page created by CreatePage
// (PUT /v2/pages/{id}).
func (s *ContentService) UpdatePage(ctx context.Context, id int64, req *PageUpdateRequest) (*CustomPage, *Response, error) {
	body := contentPageUpdateBody{}
	if req != nil {
		body.PageUpdateRequest = *req
		if req.SectionHTML != nil {
			encoded, err := contentEncodeSectionHTML(*req.SectionHTML)
			if err != nil {
				return nil, nil, err
			}
			body.SectionContent = encoded
		}
	}
	var out CustomPage
	resp, err := s.client.put(ctx, fmt.Sprintf("v2/pages/%d", id), body, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ListMenus returns one page of menus without their items (GET /v2/menus).
func (s *ContentService) ListMenus(ctx context.Context, opts *ListOptions) (*Page[Menu], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Menu](ctx, s.client, "v2/menus", q)
}

// GetMenu returns one menu with its nested items (GET /v2/menus/{id}).
func (s *ContentService) GetMenu(ctx context.Context, id int64) (*Menu, *Response, error) {
	var out Menu
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v2/menus/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CategoryListOptions filters ListCategories and AllCategories.
type CategoryListOptions struct {
	ListOptions
	// Q matches category titles.
	Q string `url:"q,omitempty"`
}

// ListCategories returns one page of categories (GET /v2/categories).
func (s *ContentService) ListCategories(ctx context.Context, opts *CategoryListOptions) (*Page[Category], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Category](ctx, s.client, "v2/categories", q)
}

// AllCategories walks every page of categories (GET /v2/categories).
func (s *ContentService) AllCategories(ctx context.Context, opts *CategoryListOptions) iter.Seq2[Category, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Category, error) bool) { yield(Category{}, err) }
	}
	return listAll[Category](ctx, s.client, "v2/categories", q)
}

type contentProductFeedsEnvelope struct {
	ProductFeeds []ProductFeed `json:"product_feeds"`
}

// ProductFeeds returns the shop's product feed URLs (GET /v2/product_feeds).
func (s *ContentService) ProductFeeds(ctx context.Context) ([]ProductFeed, *Response, error) {
	var env contentProductFeedsEnvelope
	resp, err := s.client.get(ctx, "v2/product_feeds", nil, &env)
	if err != nil {
		return nil, resp, err
	}
	return env.ProductFeeds, resp, nil
}

// ShopEmails returns the shop's contact and notification addresses
// (GET /v2/shop_emails/shop_emails).
func (s *ContentService) ShopEmails(ctx context.Context) (*ShopEmails, *Response, error) {
	var out ShopEmails
	resp, err := s.client.get(ctx, "v2/shop_emails/shop_emails", nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
