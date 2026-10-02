package outbound

import (
	"context"
	"errors"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"gorm.io/gorm"
)

// Repository is the data access layer for outbound_logs.
type Repository struct {
	db *gorm.DB
}

// NewRepository wraps a GORM handle.
func NewRepository(g *gorm.DB) *Repository { return &Repository{db: g} }

// Create inserts a row.
func (r *Repository) Create(row *db.OutboundLog) error {
	return r.db.Create(row).Error
}

// Get returns one row; nil, nil when absent.
func (r *Repository) Get(ctx context.Context, id uint) (*db.OutboundLog, error) {
	var row db.OutboundLog
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Filter narrows List.
type Filter struct {
	ShopID  uint
	Method  string
	Path    string
	Status  int
	From    *time.Time
	To      *time.Time
	Q       string
	Page    int
	PerPage int
}

// Summary is a row without its bodies.
type Summary struct {
	ID             uint      `json:"id"`
	ShopID         uint      `json:"shop_id"`
	Method         string    `json:"method"`
	Path           string    `json:"path"`
	Query          string    `json:"query"`
	ResponseStatus int       `json:"response_status"`
	DurationMs     int64     `json:"duration_ms"`
	Attempt        int       `json:"attempt"`
	RequestID      string    `json:"request_id"`
	Error          *string   `json:"error"`
	CreatedAt      time.Time `json:"created_at"`
}

// List returns one page of summaries, newest first, and the total count.
func (r *Repository) List(ctx context.Context, f Filter) ([]Summary, int64, error) {
	q := r.db.WithContext(ctx).Model(&db.OutboundLog{})
	q = applyFilter(q, f)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []Summary
	err := q.Order("id DESC").Offset((f.Page - 1) * f.PerPage).Limit(f.PerPage).Find(&items).Error
	if items == nil {
		items = []Summary{}
	}
	return items, total, err
}

func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	if f.ShopID != 0 {
		q = q.Where("shop_id = ?", f.ShopID)
	}
	if f.Method != "" {
		q = q.Where("method = ?", f.Method)
	}
	if f.Path != "" {
		q = q.Where("path LIKE ?", "%"+f.Path+"%")
	}
	if f.Status != 0 {
		q = q.Where("response_status = ?", f.Status)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}
	if f.Q != "" {
		like := "%" + f.Q + "%"
		q = q.Where("path LIKE ? OR query LIKE ? OR request_id LIKE ?", like, like, like)
	}
	return q
}

// CountSince counts rows created at or after t.
func (r *Repository) CountSince(ctx context.Context, t time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&db.OutboundLog{}).Where("created_at >= ?", t).Count(&n).Error
	return n, err
}

// DeleteBefore removes rows older than t and returns how many.
func (r *Repository) DeleteBefore(ctx context.Context, t time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("created_at < ?", t).Delete(&db.OutboundLog{})
	return res.RowsAffected, res.Error
}
