package redact

import (
	"fmt"
	"os"
	"regexp"
)

// DenylistEnv names the environment variable that holds a regular
// expression of integrator and merchant names (e.g. "acme|globex") that must
// never appear in a committed file. The list identifies customers, so it is
// kept out of the source: CI sets it as a secured variable and maintainers
// set it locally.
const DenylistEnv = "CYBERBIZ_REDACT_DENYLIST"

// Denylist compiles the expression in [DenylistEnv], case-insensitively.
// It returns nil when the variable is unset or empty.
func Denylist() (*regexp.Regexp, error) {
	expr := os.Getenv(DenylistEnv)
	if expr == "" {
		return nil, nil
	}
	re, err := regexp.Compile("(?i)" + expr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", DenylistEnv, err)
	}
	return re, nil
}
