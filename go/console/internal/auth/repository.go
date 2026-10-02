// Package auth owns Console users, login sessions, and the cookie middleware.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"gorm.io/gorm"
)

// Repository is the data access layer for users and sessions.
type Repository struct {
	db *gorm.DB
}

// NewRepository wraps a GORM handle.
func NewRepository(g *gorm.DB) *Repository { return &Repository{db: g} }

// CountUsers returns how many users exist.
func (r *Repository) CountUsers(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&db.User{}).Count(&n).Error
	return n, err
}

// CreateUser inserts u.
func (r *Repository) CreateUser(ctx context.Context, u *db.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// UserByName finds a user by username; nil, nil when absent.
func (r *Repository) UserByName(ctx context.Context, username string) (*db.User, error) {
	var u db.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// UserByID finds a user by id; nil, nil when absent.
func (r *Repository) UserByID(ctx context.Context, id uint) (*db.User, error) {
	var u db.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// CreateSession inserts s.
func (r *Repository) CreateSession(ctx context.Context, s *db.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// SessionByID returns a session that has not expired; nil, nil otherwise.
func (r *Repository) SessionByID(ctx context.Context, id string, now time.Time) (*db.Session, error) {
	var s db.Session
	err := r.db.WithContext(ctx).Where("id = ? AND expires_at > ?", id, now).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

// DeleteSession removes a session; a missing one is not an error.
func (r *Repository) DeleteSession(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&db.Session{}, "id = ?", id).Error
}

// DeleteExpiredSessions prunes sessions past their expiry.
func (r *Repository) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&db.Session{}).Error
}
