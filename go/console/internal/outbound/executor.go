package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/scope"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/shops"
	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
	"github.com/rs/zerolog/log"
)

// ExecuteInput is the body of POST /api/shops/:id/requests.
type ExecuteInput struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Query   map[string][]string `json:"query"`
	Body    json.RawMessage     `json:"body"`
	Headers map[string]string   `json:"headers"`
}

// Executor sends ad-hoc requests through a shop's SDK client.
type Executor struct {
	shops *shops.Service
	repo  *Repository
}

// NewExecutor wires the shop service and log repository.
func NewExecutor(s *shops.Service, repo *Repository) *Executor {
	return &Executor{shops: s, repo: repo}
}

var allowedMethods = map[string]bool{
	http.MethodGet: true, http.MethodPost: true, http.MethodPut: true,
	http.MethodPatch: true, http.MethodDelete: true,
}

// Execute runs the request and returns the row of the last attempt. When
// CYBERBIZ answered (even with an error status) the row is returned and
// the error is nil, so the tester can show the upstream reply.
func (e *Executor) Execute(ctx context.Context, shopID uint, in ExecuteInput) (*db.OutboundLog, error) {
	req, err := buildRequest(in)
	if err != nil {
		return nil, err
	}
	client, _, err := e.shops.ClientByID(ctx, shopID)
	if err != nil {
		return nil, err
	}
	ctx, sc := scope.New(ctx, shopID)
	_, doErr := client.Do(ctx, req, nil)
	if id := sc.LastLog(); id != 0 {
		row, err := e.repo.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		return row, nil
	}
	if doErr == nil {
		return nil, errors.New("outbound: request completed without a recorded attempt")
	}
	log.Warn().Err(doErr).Uint("shop_id", shopID).Msg("request failed before any attempt")
	return nil, response.BadRequest(doErr.Error())
}

func buildRequest(in ExecuteInput) (*cyberbiz.Request, error) {
	method := strings.ToUpper(strings.TrimSpace(in.Method))
	if !allowedMethods[method] {
		return nil, response.Validation("method must be one of GET, POST, PUT, PATCH, DELETE")
	}
	path := strings.TrimSpace(in.Path)
	if path == "" || strings.Contains(path, "://") || strings.HasPrefix(path, "//") {
		return nil, response.Validation("path must be a relative API path such as /v1/orders")
	}
	req := &cyberbiz.Request{Method: method, Path: path, Query: url.Values(in.Query), Header: http.Header{}}
	for k, v := range in.Headers {
		req.Header.Set(k, v)
	}
	if body := strings.TrimSpace(string(in.Body)); body != "" && body != "null" {
		if !json.Valid(in.Body) {
			return nil, response.Validation("body must be valid JSON")
		}
		req.Body = []byte(in.Body)
	}
	return req, nil
}
