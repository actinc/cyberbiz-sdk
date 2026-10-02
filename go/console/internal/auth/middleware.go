package auth

import (
	"net/http"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/gin-gonic/gin"
)

// CookieName is the session cookie.
const CookieName = "console_session"

const userKey = "auth.user"

// Require rejects requests without a live session and stores the user in
// the gin context for handlers.
func (s *Service) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Cookie(CookieName)
		u, err := s.Resolve(c.Request.Context(), id)
		if err != nil {
			response.Fail(c, err)
			return
		}
		if u == nil {
			response.Fail(c, response.Unauthorized("not logged in"))
			return
		}
		c.Set(userKey, u)
		c.Next()
	}
}

// CurrentUser returns the user set by Require.
func CurrentUser(c *gin.Context) *db.User {
	u, _ := c.Get(userKey)
	user, _ := u.(*db.User)
	return user
}

// SetCookie writes the session cookie; Secure when the request is HTTPS.
func SetCookie(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(c),
	})
}

func isHTTPS(c *gin.Context) bool {
	return c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
}
