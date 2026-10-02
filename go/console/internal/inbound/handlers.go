package inbound

import (
	"strconv"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/shops"
	"github.com/gin-gonic/gin"
)

// Handlers exposes the /api/inbound routes.
type Handlers struct {
	repo      *Repository
	goldenDir string
}

// NewHandlers wires the inbound routes.
func NewHandlers(repo *Repository, goldenDir string) *Handlers {
	return &Handlers{repo: repo, goldenDir: goldenDir}
}

// Register mounts the routes on an authenticated group.
func (h *Handlers) Register(r gin.IRoutes) {
	r.GET("/inbound", h.list)
	r.GET("/inbound/:id", h.get)
	r.POST("/inbound/:id/golden", h.golden)
}

func (h *Handlers) list(c *gin.Context) {
	f, err := filterFrom(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	items, total, err := h.repo.List(c.Request.Context(), f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.Page[Summary]{Items: items, Total: total, Page: f.Page, PerPage: f.PerPage})
}

func filterFrom(c *gin.Context) (Filter, error) {
	from, err := response.TimeParam(c, "from")
	if err != nil {
		return Filter{}, err
	}
	to, err := response.TimeParam(c, "to")
	if err != nil {
		return Filter{}, err
	}
	shopID, _ := strconv.ParseUint(c.Query("shop_id"), 10, 32)
	page, perPage := response.Paging(c)
	return Filter{
		ShopID: uint(shopID), Event: c.Query("event"), Status: c.Query("status"),
		From: from, To: to, Q: c.Query("q"), Page: page, PerPage: perPage,
	}, nil
}

func (h *Handlers) load(c *gin.Context) (*db.InboundLog, bool) {
	id, ok := shops.ParseID(c)
	if !ok {
		return nil, false
	}
	row, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return nil, false
	}
	if row == nil {
		response.Fail(c, response.NotFound("inbound log not found"))
		return nil, false
	}
	return row, true
}

func (h *Handlers) get(c *gin.Context) {
	row, ok := h.load(c)
	if !ok {
		return
	}
	response.OK(c, row)
}

type goldenRequest struct {
	Name      string `json:"name"`
	Overwrite bool   `json:"overwrite"`
}

func (h *Handlers) golden(c *gin.Context) {
	row, ok := h.load(c)
	if !ok {
		return
	}
	var in goldenRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&in); err != nil {
			response.Fail(c, response.BadRequest("invalid JSON body"))
			return
		}
	}
	path, err := ExportGolden(h.goldenDir, row, in.Name, in.Overwrite)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"path": path})
}
