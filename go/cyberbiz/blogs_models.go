package cyberbiz

// BlogCommentable is a blog's comment policy.
type BlogCommentable string

// Known BlogCommentable values, as returned in a blog's commentable field.
const (
	BlogCommentableNo       BlogCommentable = "no"       // comments disabled
	BlogCommentableModerate BlogCommentable = "moderate" // comments held for approval
	BlogCommentableYes      BlogCommentable = "yes"      // comments published immediately
)

// Blog is a storefront blog (GET /v1/blogs). The blog feature is not
// licensed on the test shop, so no successful Golden File exists; the shape
// follows the swagger. The detail response omits id; Get fills it in.
type Blog struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// Handle is the URL slug, as in /blogs/{handle}.
	Handle        string          `json:"handle"`
	Commentable   BlogCommentable `json:"commentable"`
	Position      int             `json:"position"`
	ArticlesCount int             `json:"articles_count"`
	CreatedAt     Time            `json:"created_at"`
	UpdatedAt     Time            `json:"updated_at"`
}

// BlogSEOMetaTags are a blog's SEO tags (GET /v1/blogs/{id}/seo_meta_tags).
type BlogSEOMetaTags struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
}

// Article is a blog post (GET /v1/blogs/{id}/articles). The detail response
// omits id; GetArticle fills it in.
type Article struct {
	ID     int64  `json:"id"`
	BlogID int64  `json:"blog_id"`
	Title  string `json:"title"`
	// BodyHTML is the article content; the platform strips HTML unless the
	// app is authorised to store it verbatim.
	BodyHTML string `json:"body_html"`
	Author   string `json:"author"`
	// Handle is the URL slug, as in /blogs/{blog}/{handle}.
	Handle    string `json:"handle"`
	Published bool   `json:"published"`
	// PublishedAt and PublishedEndAt bound the publication window.
	PublishedAt    Time         `json:"published_at"`
	PublishedEndAt Time         `json:"published_end_at"`
	Pinned         bool         `json:"pinned"`
	Tags           []ArticleTag `json:"tags"`
	CoverImageURL  string       `json:"cover_image_url"`
	CreatedAt      Time         `json:"created_at"`
	UpdatedAt      Time         `json:"updated_at"`
}

// ArticleTag is a tag on an article.
type ArticleTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Category is the platform's numeric tag category.
	Category int `json:"category"`
}
