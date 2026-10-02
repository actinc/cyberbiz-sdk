package gendocs

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// converter turns the CYBERBIZ Swagger 2.0 file into an OpenAPI 3.1 Document.
// It only translates structure; every fix of the swagger's content lives in
// corrections.go so the two concerns stay separate and testable.
type converter struct {
	sw    *swaggerDoc
	names map[string]string // swagger definition -> component schema name
	log   *report
}

// convertSwagger builds the v1 document skeleton (schemas, paths, tags).
func convertSwagger(sw *swaggerDoc, log *report) (*Document, error) {
	c := &converter{sw: sw, log: log}
	if err := c.buildNames(); err != nil {
		return nil, err
	}
	doc := &Document{
		OpenAPI: "3.1.0",
		Paths:   NewOMap(),
		Components: &Components{
			Schemas: NewOMap(),
		},
	}
	for _, def := range sw.sortedDefinitions() {
		doc.Components.Schemas.Set(c.names[def], c.schema(sw.Definitions[def]))
	}
	for _, path := range sw.sortedPaths() {
		item := &PathItem{}
		for _, method := range httpMethods {
			op, ok := sw.Paths[path][method]
			if !ok {
				continue
			}
			item.SetOperation(method, c.operation(method, path, op))
		}
		doc.Paths.Set(path, item)
	}
	for _, t := range sw.Tags {
		doc.Tags = append(doc.Tags, &Tag{Name: t.Name, Description: stripHTML(t.Description)})
	}
	return doc, nil
}

var (
	reDefPrefix = regexp.MustCompile(`^(Cyberbiz_Entities_V1_|AppStore_Api_Entities_)`)
	reNonAlnum  = regexp.MustCompile(`[^A-Za-z0-9]+`)
)

// buildNames maps every swagger definition to a short unique component name:
// Cyberbiz_Entities_V1_Order_LineItemsEntity -> OrderLineItems,
// postV1Customers (a request body) -> PostV1CustomersBody.
func (c *converter) buildNames() error {
	c.names = map[string]string{}
	used := map[string]string{}
	for _, def := range c.sw.sortedDefinitions() {
		name := componentName(def)
		if prev, clash := used[name]; clash {
			return fmt.Errorf("schema name %q maps both %s and %s", name, prev, def)
		}
		used[name] = def
		c.names[def] = name
	}
	return nil
}

func componentName(def string) string {
	if reDefPrefix.MatchString(def) {
		name := reDefPrefix.ReplaceAllString(def, "")
		name = strings.TrimSuffix(name, "Entity")
		return reNonAlnum.ReplaceAllString(name, "")
	}
	// Request-body definitions are named after the operationId (postV1Customers).
	name := reNonAlnum.ReplaceAllString(def, "")
	return strings.ToUpper(name[:1]) + name[1:] + "Body"
}

// schema converts one swagger schema (recursively).
func (c *converter) schema(sw *swSchema) *Schema {
	if sw == nil {
		return nil
	}
	if sw.Ref != "" {
		name, ok := c.names[strings.TrimPrefix(sw.Ref, "#/definitions/")]
		if !ok {
			c.log.warnf("unknown $ref %s", sw.Ref)
			return &Schema{Description: unresolvedNote + sw.Ref}
		}
		return &Schema{Ref: schemaRef(name), Description: stripHTML(sw.Description)}
	}
	out := &Schema{Description: stripHTML(sw.Description), Enum: sw.Enum, Default: sw.Default}
	switch strings.ToLower(sw.Type) {
	case "interger": // typo in the swagger
		out.Type = "integer"
	case "file":
		out.Type, out.Format = "string", "binary"
	case "":
		if sw.Properties != nil {
			out.Type = "object"
		}
	default:
		out.Type = sw.Type
	}
	if sw.Format == "float" || sw.Format == "date" || sw.Format == "date-time" {
		out.Format = sw.Format
	}
	if sw.Properties != nil {
		for _, k := range sw.Properties.Keys() {
			out.SetProp(k, c.schema(sw.Properties.Get(k)))
		}
	} else if out.Type == "object" {
		out.AdditionalProperties = true
	}
	if sw.Items != nil {
		out.Items = c.schema(sw.Items)
	} else if out.Type == "array" {
		out.Items = &Schema{}
	}
	out.Required = sw.Required
	return out
}

// operation converts one swagger operation.
func (c *converter) operation(method, path string, op *swOperation) *Operation {
	desc := stripHTML(op.Description)
	out := &Operation{
		Tags:        op.Tags,
		Summary:     firstLine(desc),
		Description: desc,
		OperationID: op.OperationID,
		Responses:   NewOMap(),
	}
	if out.OperationID == "" {
		out.OperationID = operationIDFor(method, path)
	}
	var form []*swParameter
	for _, p := range op.Parameters {
		switch p.In {
		case "path", "query", "header":
			out.Parameters = append(out.Parameters, c.parameter(p))
		case "formData":
			form = append(form, p)
		case "body":
			out.RequestBody = &RequestBody{Required: p.Required, Content: NewOMap()}
			out.RequestBody.Content.Set("application/json", &MediaType{Schema: c.schema(p.Schema)})
		}
	}
	if len(form) > 0 {
		out.RequestBody = c.formBody(form)
	}
	codes := make([]string, 0, len(op.Responses))
	for code := range op.Responses {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		r := op.Responses[code]
		resp := &Response{Description: stripHTML(r.Description)}
		if resp.Description == "" {
			resp.Description = successNote
		}
		if r.Schema != nil && code != "204" {
			resp.Content = NewOMap()
			resp.Content.Set("application/json", &MediaType{Schema: c.schema(r.Schema)})
		}
		out.Responses.Set(code, resp)
	}
	return out
}

// parameter converts a path/query parameter.
func (c *converter) parameter(p *swParameter) *Parameter {
	s := &Schema{Enum: p.Enum, Default: p.Default}
	switch p.Type {
	case "array":
		s.Type = "array"
		s.Items = c.schema(p.Items)
		if s.Items == nil {
			s.Items = &Schema{Type: "string"}
		}
	case "":
		s.Type = "string"
	default:
		s.Type = p.Type
	}
	if p.Format == "date-time" || p.Format == "date" {
		// The swagger says date-time but the API takes "YYYY-MM-DD hh:mm:ss".
		s.Type = "string"
	}
	return &Parameter{
		Name:        p.Name,
		In:          p.In,
		Description: stripHTML(p.Description),
		Required:    p.Required || p.In == "path",
		Schema:      s,
	}
}

// formBody turns swagger formData parameters into a request body. Bracketed
// names (billing_address[city], tags[], extra_infos[products][][name]) become
// nested properties. The platform accepts JSON for every write; file uploads
// are multipart only.
func (c *converter) formBody(form []*swParameter) *RequestBody {
	root := &Schema{Type: "object", Properties: NewOMap()}
	hasFile := false
	for _, p := range form {
		if p.Type == "file" {
			hasFile = true
		}
		leaf := c.parameter(p).Schema
		leaf.Description = stripHTML(p.Description)
		if p.Type == "file" {
			leaf.Type, leaf.Format = "string", "binary"
		}
		if leaf.Type == "string" && leaf.Format == "" && p.Format != "" && p.Format != "date-time" && p.Format != "date" {
			leaf.Format = p.Format
		}
		nestFormField(root, p.Name, leaf, p.Required)
	}
	body := &RequestBody{Required: true, Content: NewOMap()}
	if hasFile {
		body.Content.Set("multipart/form-data", &MediaType{Schema: root})
		return body
	}
	body.Content.Set("application/json", &MediaType{Schema: root})
	body.Content.Set("application/x-www-form-urlencoded", &MediaType{Schema: root})
	return body
}

// nestFormField inserts leaf at the bracket path of name under root.
func nestFormField(root *Schema, name string, leaf *Schema, required bool) {
	segs := splitBrackets(name)
	cur := root
	for i, seg := range segs {
		last := i == len(segs)-1
		if seg == "" { // "[]": array element
			if cur.Type != "array" {
				cur.Type = "array"
				cur.Properties = nil
				cur.AdditionalProperties = nil
			}
			if cur.Items == nil {
				cur.Items = &Schema{Type: "object", Properties: NewOMap()}
			}
			if last {
				cur.Items = leaf
				return
			}
			cur = cur.Items
			continue
		}
		if last {
			cur.SetProp(seg, leaf)
			if required && cur == root {
				cur.Required = append(cur.Required, seg)
			}
			return
		}
		next := cur.Prop(seg)
		if next == nil {
			next = &Schema{Type: "object", Properties: NewOMap()}
			cur.SetProp(seg, next)
		}
		cur = next
	}
}

// splitBrackets splits "a[b][][c]" into ["a","b","","c"].
func splitBrackets(name string) []string {
	if !strings.Contains(name, "[") {
		return []string{name}
	}
	head, rest, _ := strings.Cut(name, "[")
	out := []string{head}
	for _, part := range strings.Split(rest, "[") {
		out = append(out, strings.TrimSuffix(part, "]"))
	}
	return out
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	if r := []rune(line); len(r) > 100 {
		return string(r[:100]) + "…"
	}
	return line
}

// operationIDFor derives an operationId from method and path.
func operationIDFor(method, path string) string {
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '{' || r == '}' || r == '_' })
	var b strings.Builder
	b.WriteString(method)
	for _, p := range parts {
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}
