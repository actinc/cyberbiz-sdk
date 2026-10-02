// Package catalog builds the API tester's operation list from the OpenAPI
// documents embedded under spec/ (copied from docs/api/en by go generate).
package catalog

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:generate sh -c "rm -f spec/*.yaml && cp ../../../../docs/api/en/cyberbiz-openapi-v1.yaml ../../../../docs/api/en/cyberbiz-openapi-v2.yaml spec/"

//go:embed spec/*.yaml
var specs embed.FS

// Catalog is the parsed operation list.
type Catalog struct {
	Operations []Operation `json:"operations"`
	Tags       []Tag       `json:"tags"`
}

// Tag groups operations.
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Operation is one method + path from the specs.
type Operation struct {
	ID          string       `json:"id"`
	Method      string       `json:"method"`
	Path        string       `json:"path"`
	Tag         string       `json:"tag"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	PathParams  []Param      `json:"path_params"`
	QueryParams []Param      `json:"query_params"`
	RequestBody *RequestBody `json:"request_body"`
	Responses   []Response   `json:"responses"`
}

// Param is a path or query parameter.
type Param struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Enum        []any  `json:"enum"`
	Example     any    `json:"example"`
}

// RequestBody is the JSON body an operation accepts.
type RequestBody struct {
	ContentType string `json:"content_type"`
	Schema      any    `json:"schema"`
	Example     any    `json:"example"`
}

// Response is one documented status code.
type Response struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	Example     any    `json:"example"`
}

var methodOrder = []string{"get", "post", "put", "patch", "delete", "head", "options"}

// Load parses every embedded spec.
func Load() (*Catalog, error) {
	entries, err := specs.ReadDir("spec")
	if err != nil {
		return nil, err
	}
	var docs [][]byte
	for _, e := range entries {
		data, err := specs.ReadFile("spec/" + e.Name())
		if err != nil {
			return nil, err
		}
		docs = append(docs, data)
	}
	return Parse(docs...)
}

// Parse merges one or more OpenAPI 3 YAML documents.
func Parse(docs ...[]byte) (*Catalog, error) {
	cat := &Catalog{Operations: []Operation{}, Tags: []Tag{}}
	seenTags := map[string]bool{}
	ids := map[string]int{}
	for i, data := range docs {
		var doc map[string]any
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("catalog: spec %d: %w", i, err)
		}
		p := &parser{doc: doc}
		cat.Operations = append(cat.Operations, p.operations(ids)...)
		for _, t := range p.tags() {
			if !seenTags[t.Name] {
				seenTags[t.Name] = true
				cat.Tags = append(cat.Tags, t)
			}
		}
	}
	for _, op := range cat.Operations {
		if op.Tag != "" && !seenTags[op.Tag] {
			seenTags[op.Tag] = true
			cat.Tags = append(cat.Tags, Tag{Name: op.Tag})
		}
	}
	return cat, nil
}

type parser struct {
	doc map[string]any
}

func (p *parser) tags() []Tag {
	var out []Tag
	for _, t := range asSlice(p.doc["tags"]) {
		m := asMap(t)
		out = append(out, Tag{Name: asString(m["name"]), Description: asString(m["description"])})
	}
	return out
}

func (p *parser) operations(ids map[string]int) []Operation {
	paths := asMap(p.doc["paths"])
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []Operation
	for _, path := range keys {
		item := asMap(p.resolve(paths[path], 0))
		shared := asSlice(item["parameters"])
		for _, method := range methodOrder {
			raw, ok := item[method]
			if !ok {
				continue
			}
			out = append(out, p.operation(strings.ToUpper(method), path, asMap(raw), shared, ids))
		}
	}
	return out
}

func (p *parser) operation(method, path string, op map[string]any, shared []any, ids map[string]int) Operation {
	o := Operation{
		Method:      method,
		Path:        path,
		Summary:     asString(op["summary"]),
		Description: asString(op["description"]),
		PathParams:  []Param{},
		QueryParams: []Param{},
		Responses:   []Response{},
	}
	if tags := asSlice(op["tags"]); len(tags) > 0 {
		o.Tag = asString(tags[0])
	}
	o.ID = uniqueID(ids, asString(op["operationId"]), method, path)
	for _, raw := range append(append([]any{}, shared...), asSlice(op["parameters"])...) {
		prm := p.param(asMap(p.resolve(raw, 0)))
		switch prm.In {
		case "path":
			o.PathParams = append(o.PathParams, prm)
		case "query":
			o.QueryParams = append(o.QueryParams, prm)
		}
	}
	o.RequestBody = p.requestBody(op["requestBody"])
	o.Responses = p.responses(op["responses"])
	return o
}

func (p *parser) param(m map[string]any) Param {
	schema := asMap(p.resolve(m["schema"], 0))
	prm := Param{
		Name:        asString(m["name"]),
		In:          asString(m["in"]),
		Required:    asBool(m["required"]),
		Type:        asString(schema["type"]),
		Description: asString(m["description"]),
		Enum:        asSlice(schema["enum"]),
		Example:     m["example"],
	}
	if prm.Example == nil {
		prm.Example = schema["example"]
	}
	return prm
}

func (p *parser) requestBody(raw any) *RequestBody {
	if raw == nil {
		return nil
	}
	body := asMap(p.resolve(raw, 0))
	ct, media := firstMedia(asMap(body["content"]))
	if media == nil {
		return nil
	}
	schema := p.resolve(media["schema"], 0)
	return &RequestBody{ContentType: ct, Schema: schema, Example: exampleOf(media, asMap(schema))}
}

func (p *parser) responses(raw any) []Response {
	m := asMap(raw)
	codes := make([]string, 0, len(m))
	for k := range m {
		codes = append(codes, k)
	}
	sort.Strings(codes)
	out := make([]Response, 0, len(codes))
	for _, code := range codes {
		r := asMap(p.resolve(m[code], 0))
		_, media := firstMedia(asMap(r["content"]))
		var example any
		if media != nil {
			example = exampleOf(media, asMap(p.resolve(media["schema"], 0)))
		}
		out = append(out, Response{Status: code, Description: asString(r["description"]), Example: example})
	}
	return out
}

// firstMedia prefers application/json, else the first content type.
func firstMedia(content map[string]any) (string, map[string]any) {
	if m, ok := content["application/json"]; ok {
		return "application/json", asMap(m)
	}
	keys := make([]string, 0, len(content))
	for k := range content {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		return k, asMap(content[k])
	}
	return "", nil
}

// exampleOf reads example, then the first of examples, then schema.example.
func exampleOf(media, schema map[string]any) any {
	if ex, ok := media["example"]; ok {
		return ex
	}
	examples := asMap(media["examples"])
	keys := make([]string, 0, len(examples))
	for k := range examples {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		return asMap(examples[k])["value"]
	}
	return schema["example"]
}

func uniqueID(ids map[string]int, opID, method, path string) string {
	id := opID
	if id == "" {
		id = strings.ToLower(method) + "_" + strings.NewReplacer("/", "_", "{", "", "}", "").Replace(strings.Trim(path, "/"))
	}
	ids[id]++
	if n := ids[id]; n > 1 {
		return fmt.Sprintf("%s_%d", id, n)
	}
	return id
}
