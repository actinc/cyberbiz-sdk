// Package inbound receives CYBERBIZ webhooks, stores every outcome, and
// exports rows as Golden Files for the webhook package.
package inbound

import (
	"context"
	"errors"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"gorm.io/gorm"
)

// Repository is the data access layer for inbound_logs.
type Repository struct {
	db *gorm.DB
}

// NewRepository wraps a GORM handle.
func NewRepository(g *gorm.DB) *Repository { return &Repository{db: g} }

// Create inserts a row.
func (r *Repository) Create(ctx context.Context, row *db.InboundLog) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// Get returns one row; nil, nil when absent.
func (r *Repository) Get(ctx context.Context, id uint) (*db.InboundLog, error) {
	var row db.InboundLog
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// EarliestBySignature returns the id of the first row with that signature,
// 0 when none. Only rows with the same Event count.
func (r *Repository) EarliestBySignature(ctx context.Context, event, signature string) (uint, error) {
	if signature == "" {
		return 0, nil
	}
	// A redelivery repeats the same Event with the same body. Two different
	// Events can legitimately share a body (and therefore a signature), e.g.
	// orders/create followed by orders/paid, so the Event is part of the key.
	var row db.InboundLog
	err := r.db.WithContext(ctx).Select("id").Where("event = ? AND signature = ?", event, signature).Order("id").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return row.ID, err
}

// Filter narrows List.
type Filter struct {
	ShopID  uint
	Event   string
	Status  string
	From    *time.Time
	To      *time.Time
	Q       string
	Page    int
	PerPage int
}

// Summary is a row without its headers and body.
type Summary struct {
	ID                   uint      `json:"id"`
	ShopID               *uint     `json:"shop_id"`
	ShopDomain           string    `json:"shop_domain"`
	CustomDomain         string    `json:"custom_domain"`
	Event                string    `json:"event"`
	SignatureValid       bool      `json:"signature_valid"`
	DomainSignatureValid *bool     `json:"domain_signature_valid"`
	Status               string    `json:"status"`
	RemoteAddr           string    `json:"remote_addr"`
	ResponseStatus       int       `json:"response_status"`
	DuplicateOf          *uint     `json:"duplicate_of"`
	CreatedAt            time.Time `json:"created_at"`
}

// List returns one page of summaries, newest first, and the total count.
func (r *Repository) List(ctx context.Context, f Filter) ([]Summary, int64, error) {
	q := applyFilter(r.db.WithContext(ctx).Model(&db.InboundLog{}), f)
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
	if f.Event != "" {
		q = q.Where("event = ?", f.Event)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}
	if f.Q != "" {
		like := "%" + f.Q + "%"
		q = q.Where("shop_domain LIKE ? OR event LIKE ? OR body LIKE ?", like, like, like)
	}
	return q
}

// CountSince counts rows created at or after t, optionally only invalid ones.
func (r *Repository) CountSince(ctx context.Context, t time.Time, onlyInvalid bool) (int64, error) {
	q := r.db.WithContext(ctx).Model(&db.InboundLog{}).Where("created_at >= ?", t)
	if onlyInvalid {
		q = q.Where("status <> ?", db.InboundValid)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

// DeleteBefore removes rows older than t and returns how many.
func (r *Repository) DeleteBefore(ctx context.Context, t time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("created_at < ?", t).Delete(&db.InboundLog{})
	return res.RowsAffected, res.Error
}
