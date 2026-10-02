package gendocs

// This file is the OpenAPI 3.1 output model. Field order in these structs is
// the order the YAML is written in, so keep it stable.

// Document is an OpenAPI 3.1 document.
type Document struct {
	OpenAPI    string                `yaml:"openapi"`
	Info       *Info                 `yaml:"info"`
	Servers    []*Server             `yaml:"servers,omitempty"`
	Security   []map[string][]string `yaml:"security,omitempty"`
	Tags       []*Tag                `yaml:"tags,omitempty"`
	Paths      *OMap                 `yaml:"paths"` // path -> *PathItem
	Components *Components           `yaml:"components,omitempty"`
}

// Info is the OpenAPI info object.
type Info struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description,omitempty"`
}

// Server is one entry of the servers list.
type Server struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description,omitempty"`
}

// Tag groups operations.
type Tag struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// PathItem holds the operations of one path.
type PathItem struct {
	Get    *Operation `yaml:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty"`
	Patch  *Operation `yaml:"patch,omitempty"`
	Delete *Operation `yaml:"delete,omitempty"`
}

// Operation returns the operation stored for an HTTP method.
func (p *PathItem) Operation(method string) *Operation {
	switch method {
	case "get":
		return p.Get
	case "post":
		return p.Post
	case "put":
		return p.Put
	case "patch":
		return p.Patch
	case "delete":
		return p.Delete
	}
	return nil
}

// SetOperation stores op under an HTTP method.
func (p *PathItem) SetOperation(method string, op *Operation) {
	switch method {
	case "get":
		p.Get = op
	case "post":
		p.Post = op
	case "put":
		p.Put = op
	case "patch":
		p.Patch = op
	case "delete":
		p.Delete = op
	}
}

// httpMethods is the fixed output order of methods within a path item.
var httpMethods = []string{"get", "post", "put", "patch", "delete"}

// Operation is one HTTP operation.
type Operation struct {
	Tags        []string     `yaml:"tags,omitempty"`
	Summary     string       `yaml:"summary,omitempty"`
	Description string       `yaml:"description,omitempty"`
	OperationID string       `yaml:"operationId,omitempty"`
	Parameters  []*Parameter `yaml:"parameters,omitempty"`
	RequestBody *RequestBody `yaml:"requestBody,omitempty"`
	Responses   *OMap        `yaml:"responses"` // status -> *Response
}

// Parameter is a path/query/header parameter, or a $ref to a shared one.
type Parameter struct {
	Ref         string  `yaml:"$ref,omitempty"`
	Name        string  `yaml:"name,omitempty"`
	In          string  `yaml:"in,omitempty"`
	Description string  `yaml:"description,omitempty"`
	Required    bool    `yaml:"required,omitempty"`
	Schema      *Schema `yaml:"schema,omitempty"`
	Example     any     `yaml:"example,omitempty"`
}

// RequestBody describes the body of a write operation.
type RequestBody struct {
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Content     *OMap  `yaml:"content"` // media type -> *MediaType
}

// MediaType is the schema and sample for one content type.
type MediaType struct {
	Schema   *Schema `yaml:"schema,omitempty"`
	Example  any     `yaml:"example,omitempty"`
	Examples *OMap   `yaml:"examples,omitempty"` // name -> *Example or ref
}

// Example is a named example object.
type Example struct {
	Ref         string `yaml:"$ref,omitempty"`
	Summary     string `yaml:"summary,omitempty"`
	Description string `yaml:"description,omitempty"`
	Value       any    `yaml:"value,omitempty"`
}

// Response is one documented status code, or a $ref to a shared response.
type Response struct {
	Ref         string `yaml:"$ref,omitempty"`
	Description string `yaml:"description,omitempty"`
	Headers     *OMap  `yaml:"headers,omitempty"` // name -> *Header or ref
	Content     *OMap  `yaml:"content,omitempty"` // media type -> *MediaType
}

// Header is a response header, or a $ref to a shared one.
type Header struct {
	Ref         string  `yaml:"$ref,omitempty"`
	Description string  `yaml:"description,omitempty"`
	Schema      *Schema `yaml:"schema,omitempty"`
	Example     any     `yaml:"example,omitempty"`
}

// Components holds the shared, reusable pieces.
type Components struct {
	SecuritySchemes *OMap `yaml:"securitySchemes,omitempty"`
	Parameters      *OMap `yaml:"parameters,omitempty"`
	Headers         *OMap `yaml:"headers,omitempty"`
	Responses       *OMap `yaml:"responses,omitempty"`
	Examples        *OMap `yaml:"examples,omitempty"`
	Schemas         *OMap `yaml:"schemas,omitempty"`
}

// SecurityScheme is an entry of components.securitySchemes.
type SecurityScheme struct {
	Type         string `yaml:"type"`
	Scheme       string `yaml:"scheme,omitempty"`
	BearerFormat string `yaml:"bearerFormat,omitempty"`
	Description  string `yaml:"description,omitempty"`
}

// Schema is a JSON Schema (2020-12 dialect as used by OpenAPI 3.1).
// Type is a string or a []string (e.g. ["string", "null"]).
type Schema struct {
	Ref                  string    `yaml:"$ref,omitempty"`
	Title                string    `yaml:"title,omitempty"`
	Description          string    `yaml:"description,omitempty"`
	Type                 any       `yaml:"type,omitempty"`
	Format               string    `yaml:"format,omitempty"`
	Enum                 []any     `yaml:"enum,omitempty"`
	Default              any       `yaml:"default,omitempty"`
	Pattern              string    `yaml:"pattern,omitempty"`
	Minimum              *float64  `yaml:"minimum,omitempty"`
	Maximum              *float64  `yaml:"maximum,omitempty"`
	Required             []string  `yaml:"required,omitempty"`
	Properties           *OMap     `yaml:"properties,omitempty"` // name -> *Schema
	AdditionalProperties any       `yaml:"additionalProperties,omitempty"`
	Items                *Schema   `yaml:"items,omitempty"`
	AnyOf                []*Schema `yaml:"anyOf,omitempty"`
	Examples             []any     `yaml:"examples,omitempty"`
	XObserved            bool      `yaml:"x-cyberbiz-observed,omitempty"`
}

// Clone deep-copies a schema.
func (s *Schema) Clone() *Schema {
	if s == nil {
		return nil
	}
	out := *s
	out.Enum = append([]any(nil), s.Enum...)
	out.Required = append([]string(nil), s.Required...)
	out.Examples = append([]any(nil), s.Examples...)
	if s.Properties != nil {
		out.Properties = NewOMap()
		for _, k := range s.Properties.Keys() {
			v, _ := s.Properties.Get(k)
			out.Properties.Set(k, v.(*Schema).Clone())
		}
	}
	out.Items = s.Items.Clone()
	if s.AnyOf != nil {
		out.AnyOf = make([]*Schema, len(s.AnyOf))
		for i, a := range s.AnyOf {
			out.AnyOf[i] = a.Clone()
		}
	}
	if ap, ok := s.AdditionalProperties.(*Schema); ok {
		out.AdditionalProperties = ap.Clone()
	}
	return &out
}

// Prop returns the named property schema, if any.
func (s *Schema) Prop(name string) *Schema {
	if s == nil || s.Properties == nil {
		return nil
	}
	v, ok := s.Properties.Get(name)
	if !ok {
		return nil
	}
	return v.(*Schema)
}

// SetProp stores a property schema.
func (s *Schema) SetProp(name string, p *Schema) {
	if s.Properties == nil {
		s.Properties = NewOMap()
	}
	s.Properties.Set(name, p)
}

// BaseType returns the non-null type name of a schema ("" for refs/anyOf).
func (s *Schema) BaseType() string {
	switch t := s.Type.(type) {
	case string:
		return t
	case []string:
		for _, e := range t {
			if e != "null" {
				return e
			}
		}
	}
	return ""
}

// IsNullable reports whether null is an allowed value.
func (s *Schema) IsNullable() bool {
	if ts, ok := s.Type.([]string); ok {
		for _, e := range ts {
			if e == "null" {
				return true
			}
		}
	}
	for _, a := range s.AnyOf {
		if a.BaseType() == "null" {
			return true
		}
	}
	return false
}

// RefName returns the component name behind a $ref (direct or via anyOf).
func (s *Schema) RefName() string {
	if s == nil {
		return ""
	}
	if s.Ref != "" {
		return refName(s.Ref)
	}
	for _, a := range s.AnyOf {
		if a.Ref != "" {
			return refName(a.Ref)
		}
	}
	return ""
}

func refName(ref string) string {
	const p = "#/components/schemas/"
	if len(ref) > len(p) && ref[:len(p)] == p {
		return ref[len(p):]
	}
	return ""
}

func schemaRef(name string) string {
	return "#/components/schemas/" + name
}
