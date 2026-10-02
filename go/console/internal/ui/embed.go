// Package ui embeds the built React application (console/ui → dist/) and
// serves it with a single-page-application fallback.
package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist
var dist embed.FS

// FS returns the built UI files. Before the React build runs it contains
// only a placeholder index.html.
func FS() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// Handler serves static files and falls back to index.html for any path
// that is not an API or webhook route, so client-side routing works.
func Handler() gin.HandlerFunc {
	files := FS()
	server := http.FileServer(http.FS(files))
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/webhooks/") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found", "code": 40400})
			return
		}
		name := strings.TrimPrefix(p, "/")
		if name != "" {
			if f, err := files.Open(name); err == nil {
				_ = f.Close()
				server.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		c.Request.URL.Path = "/"
		server.ServeHTTP(c.Writer, c.Request)
	}
}
