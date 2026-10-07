package cyberbiz

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Request describes one call to the CYBERBIZ API. The typed service methods
// build these for you; Do is exported so that tools such as the Console can
// call any endpoint through the same auth, rate limiting, and retry path.
type Request struct {
	Method string
	// Path is relative to the base URL, e.g. "v1/orders/123". A leading
	// slash is tolerated.
	Path   string
	Query  url.Values
	Body   any // encoded as JSON when non-nil
	Header http.Header
}

// Response is the HTTP response of a successful call, with the pagination
// headers parsed and the body already read.
type Response struct {
	*http.Response
	Pagination Pagination
	RequestID  string
	Body       []byte
}

// IsNull reports whether the body is the JSON literal null, which CYBERBIZ
// returns from some lookups instead of a 404.
func (r *Response) IsNull() bool {
	return bytes.Equal(bytes.TrimSpace(r.Body), []byte("null"))
}

// retryableStatus reports whether a response to method is worth retrying. A
// 429 is rejected before the platform acts on the request, so it is repeated
// for every method. A 502/503/504 may arrive after the server already acted
// (an order created, say), so it is repeated only for idempotent methods.
func retryableStatus(method string, code int) bool {
	switch code {
	case http.StatusTooManyRequests:
		return true
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return idempotent(method)
	}
	return false
}

// idempotent reports whether a request may be repeated after a network
// failure without risking a duplicate write.
func idempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false
}

// Do sends req, decodes a 2xx JSON body into out (when out is non-nil), and
// returns the response. Non-2xx statuses and bare error bodies become
// *APIError. Retries and the rate limiter are applied here.
func (c *Client) Do(ctx context.Context, req *Request, out any) (*Response, error) {
	if req == nil {
		return nil, errors.New("cyberbiz: request must not be nil")
	}
	u, err := c.resolve(req.Path, req.Query)
	if err != nil {
		return nil, err
	}
	body, err := c.encodeBody(req.Body)
	if err != nil {
		return nil, err
	}

	var resp *Response
	for attempt := 0; ; attempt++ {
		resp, err = c.send(ctx, req, u, body)
		if err == nil && !retryableStatus(req.Method, resp.StatusCode) {
			break
		}
		if attempt >= c.maxRetries || (err != nil && !idempotent(req.Method)) {
			break
		}
		if err != nil && !isTransportError(err) {
			break
		}
		wait := c.retryDelay(resp, attempt+1)
		c.logger.WarnContext(ctx, "cyberbiz: retrying request",
			slog.String("method", req.Method), slog.String("path", req.Path),
			slog.Int("attempt", attempt+1), slog.Duration("wait", wait), slog.Any("error", err))
		if err := c.wait(ctx, wait); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	if err := c.checkResponse(req, resp); err != nil {
		return resp, err
	}
	if out != nil && len(resp.Body) > 0 && !resp.IsNull() {
		if err := c.decode(resp.Body, out); err != nil {
			return resp, fmt.Errorf("cyberbiz: %s %s: decoding response: %w", req.Method, req.Path, err)
		}
	}
	return resp, nil
}

func (c *Client) resolve(path string, query url.Values) (*url.URL, error) {
	rel, err := url.Parse(strings.TrimPrefix(path, "/"))
	if err != nil {
		return nil, fmt.Errorf("cyberbiz: invalid path %q: %w", path, err)
	}
	u := c.baseURL.ResolveReference(rel)
	if len(query) > 0 {
		merged := u.Query()
		for k, vs := range query {
			for _, v := range vs {
				merged.Add(k, v)
			}
		}
		u.RawQuery = merged.Encode()
	}
	return u, nil
}

func (c *Client) encodeBody(body any) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	if raw, ok := body.([]byte); ok {
		return raw, nil
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cyberbiz: encoding request body: %w", err)
	}
	return b, nil
}

// send performs exactly one HTTP round trip.
func (c *Client) send(ctx context.Context, req *Request, u *url.URL, body []byte) (*Response, error) {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("cyberbiz: building request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	for k, vs := range req.Header {
		httpReq.Header[k] = append([]string(nil), vs...)
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &transportError{err: err}
	}
	data, err := io.ReadAll(httpResp.Body)
	closeQuietly(httpResp.Body)
	if err != nil {
		return nil, &transportError{err: fmt.Errorf("reading response body: %w", err)}
	}
	c.logger.DebugContext(ctx, "cyberbiz: request",
		slog.String("method", req.Method), slog.String("path", req.Path),
		slog.Int("status", httpResp.StatusCode), slog.Duration("duration", time.Since(start)),
		slog.String("request_id", httpResp.Header.Get("X-Request-Id")))
	return &Response{
		Response:   httpResp,
		Pagination: parsePagination(httpResp.Header),
		RequestID:  httpResp.Header.Get("X-Request-Id"),
		Body:       data,
	}, nil
}

// closeQuietly closes a fully consumed response body. Once the body has been
// read to EOF the Close error carries no information the caller can act on, so
// it is deliberately ignored.
func closeQuietly(body io.Closer) {
	_ = body.Close()
}

// transportError marks a failure that happened before a response arrived.
type transportError struct{ err error }

func (e *transportError) Error() string { return "cyberbiz: " + e.err.Error() }
func (e *transportError) Unwrap() error { return e.err }

func isTransportError(err error) bool {
	var te *transportError
	return errors.As(err, &te)
}

// retryDelay honours Retry-After when the server sent one, else backs off.
func (c *Client) retryDelay(resp *Response, attempt int) time.Duration {
	if resp != nil {
		if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			return d
		}
	}
	return c.backoff(attempt)
}

func parseRetryAfter(v string) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		return max(time.Until(t), 0), true
	}
	return 0, false
}

func (c *Client) wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	done := make(chan struct{})
	go func() {
		c.sleep(d)
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// checkResponse converts error statuses and bare error bodies into *APIError.
func (c *Client) checkResponse(req *Request, resp *Response) error {
	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	if ok && !isBareErrorObject(resp.Body) {
		return nil
	}
	return &APIError{
		StatusCode: resp.StatusCode,
		Method:     req.Method,
		Path:       req.Path,
		RequestID:  resp.RequestID,
		Messages:   errorMessages(resp.Body),
		Body:       resp.Body,
	}
}

func (c *Client) decode(data []byte, out any) error {
	if c.strictJSON {
		return json.Unmarshal(data, out)
	}
	return json.Unmarshal(data, out,
		jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true))
}

// Convenience wrappers used by the resource services.

func (c *Client) get(ctx context.Context, path string, query url.Values, out any) (*Response, error) {
	return c.Do(ctx, &Request{Method: http.MethodGet, Path: path, Query: query}, out)
}

func (c *Client) post(ctx context.Context, path string, body, out any) (*Response, error) {
	return c.Do(ctx, &Request{Method: http.MethodPost, Path: path, Body: body}, out)
}

func (c *Client) put(ctx context.Context, path string, body, out any) (*Response, error) {
	return c.Do(ctx, &Request{Method: http.MethodPut, Path: path, Body: body}, out)
}

func (c *Client) delete(ctx context.Context, path string, out any) (*Response, error) {
	return c.Do(ctx, &Request{Method: http.MethodDelete, Path: path}, out)
}

// getOne fetches a single resource and maps the platform's "200 null" reply
// onto ErrNotFound.
func (c *Client) getOne(ctx context.Context, path string, query url.Values, out any) (*Response, error) {
	resp, err := c.get(ctx, path, query, out)
	if err != nil {
		return resp, err
	}
	if resp.IsNull() {
		return resp, &APIError{StatusCode: http.StatusNotFound, Method: http.MethodGet, Path: path,
			RequestID: resp.RequestID, Messages: []string{"resource is null"}, Body: resp.Body}
	}
	return resp, nil
}

// list fetches one page of T from a paginated endpoint.
func list[T any](ctx context.Context, c *Client, path string, query url.Values) (*Page[T], error) {
	var items []T
	resp, err := c.get(ctx, path, query, &items)
	if err != nil {
		return nil, err
	}
	return &Page[T]{Items: items, Pagination: resp.Pagination, Response: resp}, nil
}

// listAll walks every page of a paginated endpoint.
func listAll[T any](ctx context.Context, c *Client, path string, query url.Values) iter.Seq2[T, error] {
	start, _ := strconv.Atoi(query.Get("page"))
	return allPages(ctx, start, func(ctx context.Context, page int) (*Page[T], error) {
		return list[T](ctx, c, path, withPage(query, page))
	})
}
