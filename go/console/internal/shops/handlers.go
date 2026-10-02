package shops

import (
	"strconv"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/gin-gonic/gin"
)

// Handlers exposes the /api/shops routes.
type Handlers struct {
	svc *Service
}

// NewHandlers wires the shop routes.
func NewHandlers(svc *Service) *Handlers { return &Handlers{svc: svc} }

// Register mounts the routes on an authenticated group.
func (h *Handlers) Register(r gin.IRoutes) {
	r.GET("/shops", h.list)
	r.POST("/shops", h.create)
	r.GET("/shops/:id", h.get)
	r.PUT("/shops/:id", h.update)
	r.DELETE("/shops/:id", h.delete)
	r.POST("/shops/:id/verify", h.verify)
}

// JSON is the public Shop shape: credentials never leave the server.
type JSON struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	ShopDomain       string    `json:"shop_domain"`
	CustomDomain     string    `json:"custom_domain"`
	CyberbizShopID   int64     `json:"cyberbiz_shop_id"`
	AppName          string    `json:"app_name"`
	AppID            string    `json:"app_id"`
	HasAppSecret     bool      `json:"has_app_secret"`
	TokenFingerprint string    `json:"token_fingerprint"`
	WebhookURL       string    `json:"webhook_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToJSON converts a row to its API shape.
func (s *Service) ToJSON(shop *db.Shop) JSON {
	url := ""
	if s.opts.PublicURL != "" {
		url = s.opts.PublicURL + "/webhooks/cyberbiz"
	}
	return JSON{
		ID: shop.ID, Name: shop.Name, ShopDomain: shop.ShopDomain, CustomDomain: shop.CustomDomain,
		CyberbizShopID: shop.CyberbizShopID, AppName: shop.AppName, AppID: shop.AppID,
		HasAppSecret: len(shop.AppSecretEnc) > 0, TokenFingerprint: shop.TokenFingerprint,
		WebhookURL: url, CreatedAt: shop.CreatedAt, UpdatedAt: shop.UpdatedAt,
	}
}

// ParseID reads the :id path parameter.
func ParseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.Fail(c, response.BadRequest("invalid id"))
		return 0, false
	}
	return uint(id), true
}

func (h *Handlers) list(c *gin.Context) {
	shops, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	out := make([]JSON, 0, len(shops))
	for i := range shops {
		out = append(out, h.svc.ToJSON(&shops[i]))
	}
	response.OK(c, out)
}

func (h *Handlers) create(c *gin.Context) {
	var in Input
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, response.BadRequest("invalid JSON body"))
		return
	}
	shop, _, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, h.svc.ToJSON(shop))
}

func (h *Handlers) get(c *gin.Context) {
	id, ok := ParseID(c)
	if !ok {
		return
	}
	shop, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, h.svc.ToJSON(shop))
}

func (h *Handlers) update(c *gin.Context) {
	id, ok := ParseID(c)
	if !ok {
		return
	}
	var in Input
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, response.BadRequest("invalid JSON body"))
		return
	}
	shop, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, h.svc.ToJSON(shop))
}

func (h *Handlers) delete(c *gin.Context) {
	id, ok := ParseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *Handlers) verify(c *gin.Context) {
	id, ok := ParseID(c)
	if !ok {
		return
	}
	shop, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	info, err := h.svc.Verify(c.Request.Context(), shop)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"shop_info": info})
}
