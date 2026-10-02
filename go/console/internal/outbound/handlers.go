package outbound

import (
	"net/http"
	"strconv"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/shops"
	"github.com/gin-gonic/gin"
)

// Handlers exposes the /api/outbound routes and the request executor.
type Handlers struct {
	repo      *Repository
	exec      *Executor
	goldenDir string
}

// NewHandlers wires the outbound routes.
func NewHandlers(repo *Repository, exec *Executor, goldenDir string) *Handlers {
	return &Handlers{repo: repo, exec: exec, goldenDir: goldenDir}
}

// Register mounts the routes on an authenticated group.
func (h *Handlers) Register(r gin.IRoutes) {
	r.POST("/shops/:id/requests", h.execute)
	r.GET("/outbound", h.list)
	r.GET("/outbound/:id", h.get)
	r.GET("/outbound/:id/body", h.body)
	r.POST("/outbound/:id/golden", h.golden)
}

func (h *Handlers) execute(c *gin.Context) {
	shopID, ok := shops.ParseID(c)
	if !ok {
		return
	}
	var in ExecuteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, response.BadRequest("invalid JSON body"))
		return
	}
	row, err := h.exec.Execute(c.Request.Context(), shopID, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, row)
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
	status, _ := strconv.Atoi(c.Query("status"))
	page, perPage := response.Paging(c)
	return Filter{
		ShopID: uint(shopID), Method: c.Query("method"), Path: c.Query("path"), Status: status,
		From: from, To: to, Q: c.Query("q"), Page: page, PerPage: perPage,
	}, nil
}

func (h *Handlers) load(c *gin.Context) (*db.OutboundLog, bool) {
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
		response.Fail(c, response.NotFound("outbound log not found"))
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

func (h *Handlers) body(c *gin.Context) {
	row, ok := h.load(c)
	if !ok {
		return
	}
	ct := ResponseContentType(string(row.ResponseHeaders))
	data, err := DecodeBody(row.ResponseBody, ct)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Data(http.StatusOK, ct, data)
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
