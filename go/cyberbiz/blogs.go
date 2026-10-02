package cyberbiz

import (
	"context"
	"fmt"
	"io"
	"iter"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// BlogsService exposes blogs, their articles, article tags and SEO tags
// (/v1/blogs). The feature is licensed per shop; an unlicensed shop gets
// 401 "無權使用該 API" from every endpoint.
type BlogsService struct {
	client *Client
}

// BlogCreateRequest is the body of Create.
type BlogCreateRequest struct {
	Title       string          `json:"title"`
	Handle      *string         `json:"handle,omitzero"`
	Commentable BlogCommentable `json:"commentable,omitzero"`
}

// BlogUpdateRequest is the body of Update; every field is optional.
type BlogUpdateRequest struct {
	Title       *string         `json:"title,omitzero"`
	Handle      *string         `json:"handle,omitzero"`
	Commentable BlogCommentable `json:"commentable,omitzero"`
}

// BlogSEOMetaTagsUpdateRequest is the body of UpdateSEOMetaTags.
type BlogSEOMetaTagsUpdateRequest struct {
	Title       *string `json:"title,omitzero"`
	Description *string `json:"description,omitzero"`
	Keywords    *string `json:"keywords,omitzero"`
}

// ArticleListOptions filters ListArticles and AllArticles.
type ArticleListOptions struct {
	ListOptions
	Published    *bool  `url:"published,omitempty"`
	Pinned       *bool  `url:"pinned,omitempty"`
	CreatedAtMin Time   `url:"created_at_min,omitempty"`
	CreatedAtMax Time   `url:"created_at_max,omitempty"`
	Tag          string `url:"tag,omitempty"`
	// Title and Author are substring searches.
	Title  string `url:"title,omitempty"`
	Author string `url:"author,omitempty"`
}

// ArticleCreateRequest is the body of CreateArticle.
type ArticleCreateRequest struct {
	Title          string  `json:"title"`
	BodyHTML       *string `json:"body_html,omitzero"`
	Author         *string `json:"author,omitzero"`
	Handle         *string `json:"handle,omitzero"`
	Published      *bool   `json:"published,omitzero"`
	PublishedAt    Time    `json:"published_at,omitzero"`
	PublishedEndAt Time    `json:"published_end_at,omitzero"`
	Pinned         *bool   `json:"pinned,omitzero"`
	// TagsText is the comma-separated list of tag names.
	TagsText *string `json:"tags_text,omitzero"`
}

// ArticleUpdateRequest is the body of UpdateArticle; every field is optional.
type ArticleUpdateRequest struct {
	Title          *string `json:"title,omitzero"`
	BodyHTML       *string `json:"body_html,omitzero"`
	Author         *string `json:"author,omitzero"`
	Handle         *string `json:"handle,omitzero"`
	Published      *bool   `json:"published,omitzero"`
	PublishedAt    Time    `json:"published_at,omitzero"`
	PublishedEndAt Time    `json:"published_end_at,omitzero"`
	Pinned         *bool   `json:"pinned,omitzero"`
	TagsText       *string `json:"tags_text,omitzero"`
}

// List returns one page of blogs (GET /v1/blogs).
func (s *BlogsService) List(ctx context.Context, opts *ListOptions) (*Page[Blog], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Blog](ctx, s.client, "v1/blogs", q)
}

// All walks every page of blogs (GET /v1/blogs).
func (s *BlogsService) All(ctx context.Context, opts *ListOptions) iter.Seq2[Blog, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Blog, error) bool) { yield(Blog{}, err) }
	}
	return listAll[Blog](ctx, s.client, "v1/blogs", q)
}

// Get returns one blog (GET /v1/blogs/{id}).
func (s *BlogsService) Get(ctx context.Context, id int64) (*Blog, *Response, error) {
	var out Blog
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/blogs/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Create creates a blog (POST /v1/blogs).
func (s *BlogsService) Create(ctx context.Context, req *BlogCreateRequest) (*Blog, *Response, error) {
	var out Blog
	resp, err := s.client.post(ctx, "v1/blogs", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Update changes a blog (PUT /v1/blogs/{id}).
func (s *BlogsService) Update(ctx context.Context, id int64, req *BlogUpdateRequest) (*Blog, *Response, error) {
	var out Blog
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/blogs/%d", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Delete removes a blog (DELETE /v1/blogs/{id}).
func (s *BlogsService) Delete(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/blogs/%d", id), nil)
}

// GetSEOMetaTags returns a blog's SEO tags (GET /v1/blogs/{id}/seo_meta_tags).
func (s *BlogsService) GetSEOMetaTags(ctx context.Context, blogID int64) (*BlogSEOMetaTags, *Response, error) {
	var out BlogSEOMetaTags
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/blogs/%d/seo_meta_tags", blogID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateSEOMetaTags changes a blog's SEO tags (PUT /v1/blogs/{id}/seo_meta_tags).
func (s *BlogsService) UpdateSEOMetaTags(ctx context.Context, blogID int64, req *BlogSEOMetaTagsUpdateRequest) (*BlogSEOMetaTags, *Response, error) {
	var out BlogSEOMetaTags
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/blogs/%d/seo_meta_tags", blogID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ListArticles returns one page of a blog's articles (GET /v1/blogs/{id}/articles).
func (s *BlogsService) ListArticles(ctx context.Context, blogID int64, opts *ArticleListOptions) (*Page[Article], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Article](ctx, s.client, fmt.Sprintf("v1/blogs/%d/articles", blogID), q)
}

// AllArticles walks every page of a blog's articles (GET /v1/blogs/{id}/articles).
func (s *BlogsService) AllArticles(ctx context.Context, blogID int64, opts *ArticleListOptions) iter.Seq2[Article, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Article, error) bool) { yield(Article{}, err) }
	}
	return listAll[Article](ctx, s.client, fmt.Sprintf("v1/blogs/%d/articles", blogID), q)
}

// GetArticle returns one article (GET /v1/blogs/{id}/articles/{article_id}).
func (s *BlogsService) GetArticle(ctx context.Context, blogID, articleID int64) (*Article, *Response, error) {
	var out Article
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/blogs/%d/articles/%d", blogID, articleID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = articleID
	return &out, resp, nil
}

// CreateArticle adds an article to a blog (POST /v1/blogs/{id}/articles).
func (s *BlogsService) CreateArticle(ctx context.Context, blogID int64, req *ArticleCreateRequest) (*Article, *Response, error) {
	var out Article
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/blogs/%d/articles", blogID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateArticle changes an article (PUT /v1/blogs/{id}/articles/{article_id}).
func (s *BlogsService) UpdateArticle(ctx context.Context, blogID, articleID int64, req *ArticleUpdateRequest) (*Article, *Response, error) {
	var out Article
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/blogs/%d/articles/%d", blogID, articleID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = articleID
	return &out, resp, nil
}

// DeleteArticle removes an article (DELETE /v1/blogs/{id}/articles/{article_id}).
func (s *BlogsService) DeleteArticle(ctx context.Context, blogID, articleID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/blogs/%d/articles/%d", blogID, articleID), nil)
}

// ListArticleTags returns an article's tags; the endpoint is not paginated
// (GET /v1/blogs/{id}/articles/{article_id}/tags).
func (s *BlogsService) ListArticleTags(ctx context.Context, blogID, articleID int64) ([]ArticleTag, *Response, error) {
	var out []ArticleTag
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/blogs/%d/articles/%d/tags", blogID, articleID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

type blogsTagsBody struct {
	Tags string `json:"tags"`
}

// AddArticleTags adds the comma-separated tags to an article and returns the
// resulting tag list (PUT /v1/blogs/{id}/articles/{article_id}/tags/add).
func (s *BlogsService) AddArticleTags(ctx context.Context, blogID, articleID int64, tags string) ([]ArticleTag, *Response, error) {
	return s.changeArticleTags(ctx, blogID, articleID, "add", tags)
}

// RemoveArticleTags removes the comma-separated tags from an article and
// returns the resulting tag list
// (PUT /v1/blogs/{id}/articles/{article_id}/tags/remove).
func (s *BlogsService) RemoveArticleTags(ctx context.Context, blogID, articleID int64, tags string) ([]ArticleTag, *Response, error) {
	return s.changeArticleTags(ctx, blogID, articleID, "remove", tags)
}

func (s *BlogsService) changeArticleTags(ctx context.Context, blogID, articleID int64, action, tags string) ([]ArticleTag, *Response, error) {
	var out []ArticleTag
	path := fmt.Sprintf("v1/blogs/%d/articles/%d/tags/%s", blogID, articleID, action)
	resp, err := s.client.put(ctx, path, blogsTagsBody{Tags: tags}, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// UploadArticleCoverImage sets an article's cover image from an image
// stream, sent as multipart form data
// (POST /v1/blogs/{id}/articles/{article_id}/cover_image).
func (s *BlogsService) UploadArticleCoverImage(ctx context.Context, blogID, articleID int64, filename string, image io.Reader) (*Response, error) {
	path := fmt.Sprintf("v1/blogs/%d/articles/%d/cover_image", blogID, articleID)
	req, err := assetsMultipartRequest(path, "cover_image", filename, image)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// DeleteArticleCoverImage removes an article's cover image
// (DELETE /v1/blogs/{id}/articles/{article_id}/cover_image).
func (s *BlogsService) DeleteArticleCoverImage(ctx context.Context, blogID, articleID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/blogs/%d/articles/%d/cover_image", blogID, articleID), nil)
}
