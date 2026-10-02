// Package outbound records every attempt the SDK sends to CYBERBIZ, executes
// ad-hoc requests from the API tester, and exports rows as Golden Files.
package outbound

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/scope"
	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
	"github.com/rs/zerolog/log"
)

// Recorder is an http.RoundTripper that writes one outbound_logs row per
// attempt and tags it with the shop id carried by the request context.
type Recorder struct {
	base http.RoundTripper
	repo *Repository
}

// NewRecorder wraps base (http.DefaultTransport when nil).
func NewRecorder(repo *Repository, base http.RoundTripper) *Recorder {
	if base == nil {
		base = http.DefaultTransport
	}
	return &Recorder{base: base, repo: repo}
}

// RoundTrip performs the request and stores what was sent and received.
func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	sc, _ := scope.From(req.Context())
	row := r.newRow(req, sc)
	reqBody, err := swapBody(&req.Body)
	if err != nil {
		return nil, err
	}
	row.RequestBody = string(reqBody)

	start := time.Now()
	resp, err := r.base.RoundTrip(req)
	row.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		msg := err.Error()
		row.Error = &msg
		r.store(row, sc)
		return nil, err
	}
	respBody, err := swapBody(&resp.Body)
	if err != nil {
		msg := "reading response body: " + err.Error()
		row.Error = &msg
	}
	row.ResponseStatus = resp.StatusCode
	row.ResponseHeaders = db.JSONText(headersJSON(resp.Header))
	row.ResponseBody = EncodeBody(respBody, resp.Header.Get("Content-Type"))
	row.RequestID = resp.Header.Get("X-Request-Id")
	r.store(row, sc)
	return resp, err
}

func (r *Recorder) newRow(req *http.Request, sc *scope.Scope) *db.OutboundLog {
	row := &db.OutboundLog{
		Method:         req.Method,
		Path:           req.URL.Path,
		Query:          req.URL.RawQuery,
		RequestHeaders: db.JSONText(headersJSON(cyberbiz.RedactHeaders(req.Header))),
		Attempt:        1,
	}
	if sc != nil {
		row.ShopID = sc.ShopID
		row.Attempt = sc.NextAttempt()
	}
	return row
}

func (r *Recorder) store(row *db.OutboundLog, sc *scope.Scope) {
	if err := r.repo.Create(row); err != nil {
		log.Error().Err(err).Msg("storing outbound log")
		return
	}
	if sc != nil {
		sc.SetLastLog(row.ID)
	}
	log.Info().Str("method", row.Method).Str("path", row.Path).Int("status", row.ResponseStatus).
		Int64("duration_ms", row.DurationMs).Uint("shop_id", row.ShopID).Int("attempt", row.Attempt).
		Str("request_id", row.RequestID).Msg("outbound")
}

// swapBody reads a body fully and replaces it with a replayable copy.
func swapBody(body *io.ReadCloser) ([]byte, error) {
	if *body == nil || *body == http.NoBody {
		return nil, nil
	}
	data, err := io.ReadAll(*body)
	_ = (*body).Close()
	*body = io.NopCloser(bytes.NewReader(data))
	return data, err
}

func headersJSON(h http.Header) string {
	b, err := json.Marshal(h)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// IsTextual reports whether a content type is stored as text; anything
// else is stored base64-encoded.
func IsTextual(contentType string) bool {
	mt, _, err := mime.ParseMediaType(contentType)
	if err != nil || mt == "" {
		return true
	}
	switch {
	case strings.HasPrefix(mt, "text/"), strings.HasSuffix(mt, "json"), strings.HasSuffix(mt, "xml"),
		mt == "application/javascript", mt == "application/x-www-form-urlencoded":
		return true
	}
	return false
}

// EncodeBody returns the stored form of a response body.
func EncodeBody(body []byte, contentType string) string {
	if IsTextual(contentType) {
		return string(body)
	}
	return base64.StdEncoding.EncodeToString(body)
}

// DecodeBody reverses EncodeBody.
func DecodeBody(stored, contentType string) ([]byte, error) {
	if IsTextual(contentType) {
		return []byte(stored), nil
	}
	return base64.StdEncoding.DecodeString(stored)
}

// ResponseContentType reads Content-Type out of a stored headers JSON.
func ResponseContentType(headersJSON string) string {
	var h http.Header
	if err := json.Unmarshal([]byte(headersJSON), &h); err != nil {
		return ""
	}
	return h.Get("Content-Type")
}
