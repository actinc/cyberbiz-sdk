package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/crypto"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// Service implements login, logout, and first-boot admin creation.
type Service struct {
	repo       *Repository
	sessionTTL time.Duration
	now        func() time.Time
}

// NewService returns a Service issuing sessions that live for ttl.
func NewService(repo *Repository, ttl time.Duration) *Service {
	return &Service{repo: repo, sessionTTL: ttl, now: time.Now}
}

// EnsureAdmin creates the admin user when the users table is empty. When
// password is blank a random one is generated and printed once to out.
func (s *Service) EnsureAdmin(ctx context.Context, username, password string, out io.Writer) error {
	n, err := s.repo.CountUsers(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	generated := password == ""
	if generated {
		password, err = randomPassword()
		if err != nil {
			return err
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.CreateUser(ctx, &db.User{Username: username, PasswordHash: string(hash)}); err != nil {
		return fmt.Errorf("auth: creating admin: %w", err)
	}
	log.Info().Str("username", username).Msg("created admin user")
	if generated {
		fmt.Fprintf(out, "\nConsole admin created.\n  username: %s\n  password: %s\n(set CONSOLE_ADMIN_PASSWORD to choose one; this is printed only once)\n\n", username, password)
	}
	return nil
}

// Login verifies the credentials and opens a session.
func (s *Service) Login(ctx context.Context, username, password string) (*db.User, *db.Session, error) {
	u, err := s.repo.UserByName(ctx, username)
	if err != nil {
		return nil, nil, err
	}
	if u == nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, nil, response.Unauthorized("invalid username or password")
	}
	id, err := crypto.RandomHex(32)
	if err != nil {
		return nil, nil, err
	}
	now := s.now()
	sess := &db.Session{ID: id, UserID: u.ID, ExpiresAt: now.Add(s.sessionTTL), CreatedAt: now}
	if err := s.repo.CreateSession(ctx, sess); err != nil {
		return nil, nil, err
	}
	log.Info().Str("username", username).Msg("login")
	return u, sess, nil
}

// Logout deletes the session.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

// Resolve returns the user owning a live session, or nil.
func (s *Service) Resolve(ctx context.Context, sessionID string) (*db.User, error) {
	if sessionID == "" {
		return nil, nil
	}
	sess, err := s.repo.SessionByID(ctx, sessionID, s.now())
	if err != nil || sess == nil {
		return nil, err
	}
	return s.repo.UserByID(ctx, sess.UserID)
}

// PruneSessions deletes expired sessions.
func (s *Service) PruneSessions(ctx context.Context) error {
	return s.repo.DeleteExpiredSessions(ctx, s.now())
}

func randomPassword() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
