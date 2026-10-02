package cyberbiz

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors that an *APIError matches through errors.Is, keyed on the
// HTTP status CYBERBIZ returned.
var (
	// ErrUnauthorized is a 401. CYBERBIZ uses it both for an invalid token
	// and for a feature the shop has not licensed; the message tells which.
	ErrUnauthorized = errors.New("cyberbiz: unauthorized")
	// ErrForbidden is a 403: a missing token scope or plugin, or an action
	// the resource's current state does not allow.
	ErrForbidden = errors.New("cyberbiz: forbidden")
	// ErrNotFound is a 404, or a 2xx whose body means "no such resource".
	ErrNotFound = errors.New("cyberbiz: not found")
	// ErrValidation is a 422: a rejected parameter or a feature the shop
	// has not enabled.
	ErrValidation = errors.New("cyberbiz: validation failed")
	// ErrRateLimited is a 429 that survived every retry.
	ErrRateLimited = errors.New("cyberbiz: rate limited")
	// ErrServer is any 5xx that survived every retry.
	ErrServer = errors.New("cyberbiz: server error")
)

// APIError is an error response from the CYBERBIZ API. It is returned for
// every non-2xx status and for the platform quirk of a 2xx whose body is only
// an error object.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	RequestID  string   // the X-Request-Id header, useful when contacting CYBERBIZ
	Messages   []string // human-readable messages, usually Traditional Chinese
	Body       []byte   // the raw response body
}

func (e *APIError) Error() string {
	msg := strings.Join(e.Messages, "; ")
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	return fmt.Sprintf("cyberbiz: %s %s: %d %s", e.Method, e.Path, e.StatusCode, msg)
}

// Is maps the status code onto the package sentinels so that callers can
// write errors.Is(err, cyberbiz.ErrNotFound).
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrValidation:
		return e.StatusCode == http.StatusUnprocessableEntity
	case ErrRateLimited:
		return e.StatusCode == http.StatusTooManyRequests
	case ErrServer:
		return e.StatusCode >= 500
	}
	return false
}

// errorMessages extracts human-readable messages from the four body shapes
// CYBERBIZ uses: {"error":[...]}, {"error":"..."}, {"message":"..."} and
// {"messages":"..."} (each value may be a string or an array of strings).
func errorMessages(body []byte) []string {
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil
	}
	var out []string
	for _, key := range []string{"error", "errors", "message", "messages"} {
		out = append(out, flattenStrings(obj[key])...)
	}
	return out
}

// isBareErrorObject reports whether a 2xx body is nothing but an error
// object, which CYBERBIZ sometimes returns instead of a proper status.
func isBareErrorObject(body []byte) bool {
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil || len(obj) != 1 {
		return false
	}
	for _, key := range []string{"error", "errors", "messages"} {
		if _, ok := obj[key]; ok {
			return true
		}
	}
	return false
}

func flattenStrings(v any) []string {
	switch x := v.(type) {
	case string:
		if x == "" {
			return nil
		}
		return []string{x}
	case []any:
		var out []string
		for _, item := range x {
			out = append(out, flattenStrings(item)...)
		}
		return out
	case map[string]any:
		var out []string
		for field, item := range x {
			for _, s := range flattenStrings(item) {
				out = append(out, field+": "+s)
			}
		}
		return out
	}
	return nil
}
