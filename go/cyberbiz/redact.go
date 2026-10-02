package cyberbiz

import (
	"net/http"

	"github.com/actinc/cyberbiz-sdk/go/internal/redact"
)

// RedactJSON returns a copy of a JSON payload with personal and secret values
// replaced by placeholders of the same shape. Key order and value types are
// preserved, so a redacted document decodes into the same model as the
// original. Non-JSON input is returned unchanged. It is what the Console uses
// before storing a Golden File.
func RedactJSON(data []byte) ([]byte, error) { return redact.JSON(data) }

// RedactHeaders returns a copy of h with credential-bearing and
// client-identifying headers replaced by a placeholder.
func RedactHeaders(h http.Header) http.Header { return redact.Headers(h) }
