// Package shops stores Shops with encrypted credentials and hands out one
// cached SDK client per Shop.
package shops

import (
	"context"
	"errors"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"gorm.io/gorm"
)

// Repository is the data access layer for shops.
type Repository struct {
	db *gorm.DB
}

// NewRepository wraps a GORM handle.
func NewRepository(g *gorm.DB) *Repository { return &Repository{db: g} }

// List returns every shop ordered by id.
func (r *Repository) List(ctx context.Context) ([]db.Shop, error) {
	var out []db.Shop
	err := r.db.WithContext(ctx).Order("id").Find(&out).Error
	return out, err
}

// Count returns how many shops exist.
func (r *Repository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&db.Shop{}).Count(&n).Error
	return n, err
}

// Get returns a shop by id; nil, nil when absent.
func (r *Repository) Get(ctx context.Context, id uint) (*db.Shop, error) {
	var s db.Shop
	err := r.db.WithContext(ctx).First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

// GetByDomain returns the shop for a Shop Domain; nil, nil when absent.
func (r *Repository) GetByDomain(ctx context.Context, domain string) (*db.Shop, error) {
	var s db.Shop
	err := r.db.WithContext(ctx).Where("shop_domain = ?", domain).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

// Create inserts s. A duplicate shop_domain yields ErrDuplicateDomain.
func (r *Repository) Create(ctx context.Context, s *db.Shop) error {
	return classify(r.db.WithContext(ctx).Create(s).Error)
}

// Save writes every column of s.
func (r *Repository) Save(ctx context.Context, s *db.Shop) error {
	return classify(r.db.WithContext(ctx).Save(s).Error)
}

// Delete removes the shop row; logs referencing it are kept.
func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&db.Shop{}, id).Error
}

// ErrDuplicateDomain means another shop already has that shop_domain.
var ErrDuplicateDomain = errors.New("shops: shop_domain already exists")

func classify(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrDuplicateDomain
	}
	return err
}
