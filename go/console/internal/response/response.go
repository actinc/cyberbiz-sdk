// Package response writes the JSON envelope every Console API route uses:
// {success, data} on success and {success, error, code} on failure.
package response

import (
	"errors"
	"net/http"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Error codes from console/docs/spec.md.
const (
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeConflict     = 40900
	CodeValidation   = 42200
	CodeInternal     = 50000
	CodeUpstream     = 50200
)

// Envelope is the wire shape of every response.
type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// Error is a failure the handlers can return; Fail maps it onto the
// envelope.
type Error struct {
	Status  int
	Code    int
	Message string
	Data    any
}

func (e *Error) Error() string { return e.Message }

// Constructors for the common failures.
func BadRequest(msg string) *Error { return &Error{http.StatusBadRequest, CodeBadRequest, msg, nil} }
func Unauthorized(msg string) *Error {
	return &Error{http.StatusUnauthorized, CodeUnauthorized, msg, nil}
}
func Forbidden(msg string) *Error { return &Error{http.StatusForbidden, CodeForbidden, msg, nil} }
func NotFound(msg string) *Error  { return &Error{http.StatusNotFound, CodeNotFound, msg, nil} }
func Conflict(msg string) *Error  { return &Error{http.StatusConflict, CodeConflict, msg, nil} }
func Validation(msg string) *Error {
	return &Error{http.StatusUnprocessableEntity, CodeValidation, msg, nil}
}

// Upstream wraps an SDK *APIError as a 502 with {status, messages, request_id}.
func Upstream(apiErr *cyberbiz.APIError) *Error {
	return &Error{
		Status:  http.StatusBadGateway,
		Code:    CodeUpstream,
		Message: apiErr.Error(),
		Data: gin.H{
			"status":     apiErr.StatusCode,
			"messages":   apiErr.Messages,
			"request_id": apiErr.RequestID,
		},
	}
}

// OK writes a success envelope.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

// Fail writes a failure envelope for err, classifying it: *Error as is,
// *cyberbiz.APIError as upstream, anything else as internal.
func Fail(c *gin.Context, err error) {
	var e *Error
	var apiErr *cyberbiz.APIError
	switch {
	case errors.As(err, &e):
	case errors.As(err, &apiErr):
		e = Upstream(apiErr)
	default:
		log.Error().Err(err).Str("path", c.FullPath()).Msg("internal error")
		e = &Error{http.StatusInternalServerError, CodeInternal, "internal error", nil}
	}
	c.AbortWithStatusJSON(e.Status, Envelope{Success: false, Error: e.Message, Code: e.Code, Data: e.Data})
}
