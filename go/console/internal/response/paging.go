package response

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Page is the list envelope shared by the outbound and inbound routes.
type Page[T any] struct {
	Items   []T   `json:"items"`
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
}

const (
	defaultPerPage = 50
	maxPerPage     = 200
)

// Paging reads page and per_page with defaults and bounds.
func Paging(c *gin.Context) (page, perPage int) {
	page, _ = strconv.Atoi(c.Query("page"))
	perPage, _ = strconv.Atoi(c.Query("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}

// TimeParam parses an RFC 3339 or YYYY-MM-DD query parameter; a blank
// value yields nil.
func TimeParam(c *gin.Context, name string) (*time.Time, error) {
	v := c.Query(name)
	if v == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, v); err == nil {
			return &t, nil
		}
	}
	return nil, BadRequest(name + ": expected RFC 3339 or YYYY-MM-DD")
}
