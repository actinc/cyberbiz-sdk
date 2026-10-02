package cyberbiz

import (
	"bytes"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

const blogsArticleSample = `{"blog_id":3,"title":"開幕","body_html":"<p>hi</p>","author":"小編","handle":"opening","published":true,
	"published_at":"2026-03-01 09:00:00","published_end_at":null,"pinned":false,"tags":[{"id":1,"name":"news","category":0}],
	"cover_image_url":"https://cdn.example/c.jpg","created_at":"2026-02-28 12:00:00","updated_at":"2026-02-28 12:00:00"}`

func TestBlogsGoldenUnlicensedIsUnauthorized(t *testing.T) {
	// Every blog Golden File is the 401 the unlicensed test shop returns;
	// replay each one and check it surfaces as ErrUnauthorized.
	goldens := map[string]func(c *Client) error{
		"errors/GET_v1_blogs.json":                         func(c *Client) error { _, err := c.Blogs.List(testCtx, nil); return err },
		"errors/GET_v1_blogs_{id}.json":                    func(c *Client) error { _, _, err := c.Blogs.Get(testCtx, 1); return err },
		"errors/GET_v1_blogs_{id}_articles.json":           func(c *Client) error { _, err := c.Blogs.ListArticles(testCtx, 1, nil); return err },
		"errors/GET_v1_blogs_{id}_articles_{id}.json":      func(c *Client) error { _, _, err := c.Blogs.GetArticle(testCtx, 1, 2); return err },
		"errors/GET_v1_blogs_{id}_articles_{id}_tags.json": func(c *Client) error { _, _, err := c.Blogs.ListArticleTags(testCtx, 1, 2); return err },
		"errors/GET_v1_blogs_{id}_seo_meta_tags.json":      func(c *Client) error { _, _, err := c.Blogs.GetSEOMetaTags(testCtx, 1); return err },
	}
	for name, call := range goldens {
		body := readGolden(t, name)
		c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(body)
		}, WithMaxRetries(0))
		err := call(c)
		var apiErr *APIError
		if !errors.As(err, &apiErr) || !errors.Is(err, ErrUnauthorized) {
			t.Errorf("%s: got %v", name, err)
			continue
		}
		if len(apiErr.Messages) != 1 || apiErr.Messages[0] != "無權使用該 API" {
			t.Errorf("%s: messages = %q", name, apiErr.Messages)
		}
	}
}

func TestBlogsDecodeSwaggerShapes(t *testing.T) {
	c, _ := New("tok")
	var blog Blog
	body := `{"title":"News","handle":"news","commentable":"moderate","position":1,"articles_count":4,
		"created_at":"2026-01-01 00:00:00","updated_at":null,"id":3}`
	if err := c.decode([]byte(body), &blog); err != nil {
		t.Fatal(err)
	}
	if blog.ID != 3 || blog.Commentable != BlogCommentableModerate || blog.ArticlesCount != 4 || !blog.UpdatedAt.IsZero() {
		t.Errorf("blog = %+v", blog)
	}
	var art Article
	if err := c.decode([]byte(blogsArticleSample), &art); err != nil {
		t.Fatal(err)
	}
	if art.BlogID != 3 || art.BodyHTML != "<p>hi</p>" || !art.Published || art.PublishedAt.String() != "2026-03-01 09:00:00" {
		t.Errorf("article = %+v", art)
	}
	if len(art.Tags) != 1 || art.Tags[0].Name != "news" || !art.PublishedEndAt.IsZero() {
		t.Errorf("article = %+v", art)
	}
	var seo BlogSEOMetaTags
	if err := c.decode([]byte(`{"title":"t","description":null,"keywords":"a,b"}`), &seo); err != nil {
		t.Fatal(err)
	}
	if seo.Title != "t" || seo.Description != "" || seo.Keywords != "a,b" {
		t.Errorf("seo = %+v", seo)
	}
}

func TestBlogsRequests(t *testing.T) {
	lc, lcall := discountsSpyClient(t, 200, `[]`)
	if _, err := lc.Blogs.List(testCtx, &ListOptions{Page: 2}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "GET", "/v1/blogs")
	if lcall.Query.Get("page") != "2" {
		t.Errorf("query = %v", lcall.Query)
	}

	c, call := discountsSpyClient(t, 200, `{"title":"News","handle":"news"}`)
	blog, _, err := c.Blogs.Get(testCtx, 3)
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/blogs/3")
	if blog.ID != 3 {
		t.Errorf("id = %d", blog.ID)
	}

	handle := "news"
	if _, _, err := c.Blogs.Create(testCtx, &BlogCreateRequest{Title: "News", Handle: &handle, Commentable: BlogCommentableNo}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/blogs")
	discountsAssertJSONBody(t, call, `{"title":"News","handle":"news","commentable":"no"}`)

	empty := ""
	if _, _, err := c.Blogs.Update(testCtx, 3, &BlogUpdateRequest{Handle: &empty}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/blogs/3")
	discountsAssertJSONBody(t, call, `{"handle":""}`)

	if _, err := c.Blogs.Delete(testCtx, 3); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "DELETE", "/v1/blogs/3")

	if _, _, err := c.Blogs.GetSEOMetaTags(testCtx, 3); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/blogs/3/seo_meta_tags")

	kw := "a,b"
	if _, _, err := c.Blogs.UpdateSEOMetaTags(testCtx, 3, &BlogSEOMetaTagsUpdateRequest{Keywords: &kw}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/blogs/3/seo_meta_tags")
	discountsAssertJSONBody(t, call, `{"keywords":"a,b"}`)
}

func TestBlogsArticleRequests(t *testing.T) {
	lc, lcall := discountsSpyClient(t, 200, `[]`)
	published := true
	pinned := false
	since, _ := ParseTime("2026-01-01 00:00:00")
	if _, err := lc.Blogs.ListArticles(testCtx, 3, &ArticleListOptions{Published: &published, Pinned: &pinned, CreatedAtMin: since, Tag: "news", Title: "開", Author: "小編"}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "GET", "/v1/blogs/3/articles")
	q := lcall.Query
	if q.Get("published") != "true" || q.Get("pinned") != "false" || q.Get("created_at_min") != "2026-01-01 00:00:00" ||
		q.Get("tag") != "news" || q.Get("title") != "開" || q.Get("author") != "小編" || q.Has("created_at_max") {
		t.Errorf("query = %v", q)
	}

	tags, _, err := lc.Blogs.ListArticleTags(testCtx, 3, 7)
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "GET", "/v1/blogs/3/articles/7/tags")
	if len(tags) != 0 {
		t.Errorf("tags = %+v", tags)
	}

	if _, _, err := lc.Blogs.AddArticleTags(testCtx, 3, 7, "a,b"); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "PUT", "/v1/blogs/3/articles/7/tags/add")
	discountsAssertJSONBody(t, lcall, `{"tags":"a,b"}`)

	if _, _, err := lc.Blogs.RemoveArticleTags(testCtx, 3, 7, "a"); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "PUT", "/v1/blogs/3/articles/7/tags/remove")
	discountsAssertJSONBody(t, lcall, `{"tags":"a"}`)

	c, call := discountsSpyClient(t, 201, blogsArticleSample)
	at, _ := ParseTime("2026-03-01 09:00:00")
	html := "<p>hi</p>"
	art, _, err := c.Blogs.CreateArticle(testCtx, 3, &ArticleCreateRequest{Title: "開幕", BodyHTML: &html, Published: &published, PublishedAt: at, Pinned: &pinned})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/blogs/3/articles")
	discountsAssertJSONBody(t, call, `{"title":"開幕","body_html":"<p>hi</p>","published":true,"published_at":"2026-03-01 09:00:00","pinned":false}`)
	if art.Handle != "opening" {
		t.Errorf("article = %+v", art)
	}

	got, _, err := c.Blogs.GetArticle(testCtx, 3, 7)
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/blogs/3/articles/7")
	if got.ID != 7 {
		t.Errorf("id = %d", got.ID)
	}

	tagsText := ""
	if _, _, err := c.Blogs.UpdateArticle(testCtx, 3, 7, &ArticleUpdateRequest{TagsText: &tagsText}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v1/blogs/3/articles/7")
	discountsAssertJSONBody(t, call, `{"tags_text":""}`)

	if _, err := c.Blogs.DeleteArticle(testCtx, 3, 7); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "DELETE", "/v1/blogs/3/articles/7")

	if _, err := c.Blogs.UploadArticleCoverImage(testCtx, 3, 7, "cover.jpg", strings.NewReader("JPG")); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/blogs/3/articles/7/cover_image")
	_, params, err := mime.ParseMediaType(call.ContentType)
	if err != nil {
		t.Fatal(err)
	}
	part, err := multipart.NewReader(bytes.NewReader(call.Body), params["boundary"]).NextPart()
	if err != nil || part.FormName() != "cover_image" || part.FileName() != "cover.jpg" {
		t.Errorf("part = %v %v", part, err)
	}

	if _, err := c.Blogs.DeleteArticleCoverImage(testCtx, 3, 7); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "DELETE", "/v1/blogs/3/articles/7/cover_image")
}
