// Package scope carries the Shop an outbound request belongs to through the
// context, so the recording transport can tag every attempt with it and
// the executor can find the row of the last attempt.
package scope

import (
	"context"
	"sync"
)

type key struct{}

// Scope is the per-request state shared by the executor and the recorder.
type Scope struct {
	ShopID uint

	mu      sync.Mutex
	attempt int
	lastLog uint
}

// New attaches a fresh Scope for shopID to ctx.
func New(ctx context.Context, shopID uint) (context.Context, *Scope) {
	s := &Scope{ShopID: shopID}
	return context.WithValue(ctx, key{}, s), s
}

// From returns the Scope stored in ctx, if any.
func From(ctx context.Context) (*Scope, bool) {
	s, ok := ctx.Value(key{}).(*Scope)
	return s, ok
}

// NextAttempt increments and returns the attempt counter (1-based).
func (s *Scope) NextAttempt() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempt++
	return s.attempt
}

// SetLastLog records the id of the most recent outbound_logs row.
func (s *Scope) SetLastLog(id uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastLog = id
}

// LastLog returns the id of the most recent outbound_logs row, 0 when none.
func (s *Scope) LastLog() uint {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastLog
}
