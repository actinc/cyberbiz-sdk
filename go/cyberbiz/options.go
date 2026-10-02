package cyberbiz

import (
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"
)

// Option configures a [Client] at construction time.
type Option func(*config) error

// BackoffFunc returns how long to wait before retry number attempt (starting
// at 1). The server's Retry-After header, when present, overrides it.
type BackoffFunc func(attempt int) time.Duration

type config struct {
	baseURL    string
	httpClient *http.Client
	rateLimit  float64
	maxRetries int
	backoff    BackoffFunc
	logger     *slog.Logger
	strictJSON bool
	userAgent  string
}

func defaultConfig() config {
	return config{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		rateLimit:  DefaultRateLimit,
		maxRetries: DefaultMaxRetries,
		backoff:    ExponentialBackoff(500*time.Millisecond, 8*time.Second),
		logger:     slog.New(slog.DiscardHandler),
		userAgent:  defaultUserAgent,
	}
}

// WithBaseURL points the client at a different host, for tests or a proxy.
func WithBaseURL(baseURL string) Option {
	return func(c *config) error {
		if baseURL == "" {
			return errors.New("cyberbiz: base URL must not be empty")
		}
		c.baseURL = baseURL
		return nil
	}
}

// WithHTTPClient replaces the default HTTP client. Use it to set timeouts or
// an [http.RoundTripper] that records, traces, or mocks traffic.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) error {
		if hc == nil {
			return errors.New("cyberbiz: HTTP client must not be nil")
		}
		c.httpClient = hc
		return nil
	}
}

// WithTransport wraps the default HTTP client around rt. It is the hook for
// request logging and tracing: every request, including retries, passes
// through rt.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *config) error {
		if rt == nil {
			return errors.New("cyberbiz: transport must not be nil")
		}
		hc := *c.httpClient
		hc.Transport = rt
		c.httpClient = &hc
		return nil
	}
}

// WithRateLimit sets the client-side limit in requests per second. Zero
// disables the limiter, for callers that enforce the platform limit
// elsewhere.
func WithRateLimit(requestsPerSecond float64) Option {
	return func(c *config) error {
		if requestsPerSecond < 0 {
			return errors.New("cyberbiz: rate limit must not be negative")
		}
		c.rateLimit = requestsPerSecond
		return nil
	}
}

// WithMaxRetries sets how many times a request is retried after a 429 or a
// transient 5xx response. Zero disables retries.
func WithMaxRetries(n int) Option {
	return func(c *config) error {
		if n < 0 {
			return errors.New("cyberbiz: max retries must not be negative")
		}
		c.maxRetries = n
		return nil
	}
}

// WithBackoff replaces the default exponential backoff between retries.
func WithBackoff(fn BackoffFunc) Option {
	return func(c *config) error {
		if fn == nil {
			return errors.New("cyberbiz: backoff must not be nil")
		}
		c.backoff = fn
		return nil
	}
}

// WithLogger makes the client log every request at Debug level and every
// retry at Warn level.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) error {
		if logger == nil {
			return errors.New("cyberbiz: logger must not be nil")
		}
		c.logger = logger
		return nil
	}
}

// WithStrictJSON makes the client reject responses that contain invalid
// UTF-8 or duplicate object keys. By default both are tolerated because the
// platform's JSON quality is not guaranteed.
func WithStrictJSON() Option {
	return func(c *config) error {
		c.strictJSON = true
		return nil
	}
}

// WithUserAgent replaces the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(c *config) error {
		if ua == "" {
			return errors.New("cyberbiz: user agent must not be empty")
		}
		c.userAgent = ua
		return nil
	}
}

// ExponentialBackoff returns a BackoffFunc that doubles base on every attempt,
// caps the wait at max, and adds up to 25% jitter so that concurrent clients
// do not retry in lockstep.
func ExponentialBackoff(base, max time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		d := base
		for i := 1; i < attempt && d < max; i++ {
			d *= 2
		}
		if d > max {
			d = max
		}
		jitter := time.Duration(rand.Int64N(int64(d) / 4))
		return d + jitter
	}
}
