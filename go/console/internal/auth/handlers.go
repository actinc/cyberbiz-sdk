package auth

import (
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/gin-gonic/gin"
)

// Handlers exposes the /api/auth routes.
type Handlers struct {
	svc           *Service
	setupRequired func(*gin.Context) (bool, error)
}

// NewHandlers wires the auth routes; setupRequired reports whether no Shop
// exists yet.
func NewHandlers(svc *Service, setupRequired func(*gin.Context) (bool, error)) *Handlers {
	return &Handlers{svc: svc, setupRequired: setupRequired}
}

// Register mounts the routes: login is public, the rest require a session.
func (h *Handlers) Register(public, private gin.IRoutes) {
	public.POST("/auth/login", h.login)
	private.POST("/auth/logout", h.logout)
	private.GET("/auth/me", h.me)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userJSON struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func (h *Handlers) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.BadRequest("invalid JSON body"))
		return
	}
	if req.Username == "" || req.Password == "" {
		response.Fail(c, response.Validation("username and password are required"))
		return
	}
	u, sess, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	SetCookie(c, sess.ID, int(h.svc.sessionTTL.Seconds()))
	response.OK(c, gin.H{"user": userJSON{ID: u.ID, Username: u.Username}})
}

func (h *Handlers) logout(c *gin.Context) {
	id, _ := c.Cookie(CookieName)
	if err := h.svc.Logout(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	SetCookie(c, "", -1)
	response.OK(c, gin.H{})
}

func (h *Handlers) me(c *gin.Context) {
	u := CurrentUser(c)
	setup, err := h.setupRequired(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"user": userJSON{ID: u.ID, Username: u.Username}, "setup_required": setup})
}
