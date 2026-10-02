package gendocs

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// postmanRequest is what the generator keeps from one request of the
// CYBERBIZ v1 Postman collection: the example query values, the example
// body and the saved example response.
type postmanRequest struct {
	Method   string
	Path     string // /v1/orders/{order_id}
	Query    *OMap  // key -> example value (string)
	PathVars *OMap  // key -> example value (string)
	Body     any    // *OMap for urlencoded / raw JSON, nil when none
	RespCode int
	RespBody any // decoded JSON of the first saved example, nil when none
}

// postmanIndex maps "METHOD path" to the request.
type postmanIndex map[string]*postmanRequest

type pmCollection struct {
	Item []*pmItem `json:"item"`
}

type pmItem struct {
	Name     string        `json:"name"`
	Item     []*pmItem     `json:"item"`
	Request  *pmReq        `json:"request"`
	Response []*pmResponse `json:"response"`
}

type pmReq struct {
	Method string `json:"method"`
	URL    struct {
		Raw      string  `json:"raw"`
		Query    []*pmKV `json:"query"`
		Variable []*pmKV `json:"variable"`
	} `json:"url"`
	Body *struct {
		Mode       string  `json:"mode"`
		Raw        string  `json:"raw"`
		URLEncoded []*pmKV `json:"urlencoded"`
		FormData   []*pmKV `json:"formdata"`
	} `json:"body"`
}

type pmKV struct {
	Key         string `json:"key"`
	Value       any    `json:"value"`
	Description string `json:"description"`
	Disabled    bool   `json:"disabled"`
}

type pmResponse struct {
	Code int    `json:"code"`
	Body string `json:"body"`
}

var rePathVar = regexp.MustCompile(`:([A-Za-z_][A-Za-z0-9_]*)`)

// loadPostman reads a Postman v2.1 collection into an index.
func loadPostman(path string) (postmanIndex, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return postmanIndex{}, nil
		}
		return nil, err
	}
	var col pmCollection
	if err := json.Unmarshal(b, &col); err != nil {
		return nil, fmt.Errorf("postman %s: %w", path, err)
	}
	idx := postmanIndex{}
	var walk func(items []*pmItem)
	walk = func(items []*pmItem) {
		for _, it := range items {
			if it.Request != nil {
				r := convertPostmanRequest(it)
				if _, dup := idx[r.Method+" "+r.Path]; !dup {
					idx[r.Method+" "+r.Path] = r
				}
			}
			walk(it.Item)
		}
	}
	walk(col.Item)
	return idx, nil
}

func convertPostmanRequest(it *pmItem) *postmanRequest {
	raw := it.Request.URL.Raw
	raw = strings.TrimPrefix(raw, "{{baseUrl}}")
	raw = strings.TrimPrefix(raw, "{{base_url}}")
	if i := strings.Index(raw, "?"); i >= 0 {
		raw = raw[:i]
	}
	path := rePathVar.ReplaceAllString(raw, "{$1}")
	path = strings.ReplaceAll(strings.ReplaceAll(path, "{{", "{"), "}}", "}")
	r := &postmanRequest{Method: it.Request.Method, Path: path, Query: NewOMap(), PathVars: NewOMap()}
	for _, q := range it.Request.URL.Query {
		if q.Disabled || q.Value == nil {
			continue
		}
		r.Query.Set(q.Key, fmt.Sprint(q.Value))
	}
	for _, v := range it.Request.URL.Variable {
		if v.Value != nil {
			r.PathVars.Set(v.Key, fmt.Sprint(v.Value))
		}
	}
	if b := it.Request.Body; b != nil {
		switch b.Mode {
		case "raw":
			if v, err := DecodeJSON([]byte(b.Raw)); err == nil {
				r.Body = v
			}
		case "urlencoded", "formdata":
			kvs := b.URLEncoded
			if b.Mode == "formdata" {
				kvs = b.FormData
			}
			body := NewOMap()
			for _, kv := range kvs {
				if kv.Disabled {
					continue
				}
				body.Set(kv.Key, fmt.Sprint(kv.Value))
			}
			r.Body = body
		}
	}
	for _, resp := range it.Response {
		if strings.TrimSpace(resp.Body) == "" {
			continue
		}
		v, err := DecodeJSON([]byte(resp.Body))
		if err != nil {
			continue
		}
		r.RespCode, r.RespBody = resp.Code, v
		break
	}
	return r
}

// lookup returns the Postman request for method and OpenAPI path.
func (idx postmanIndex) lookup(method, path string) *postmanRequest {
	return idx[strings.ToUpper(method)+" "+path]
}
