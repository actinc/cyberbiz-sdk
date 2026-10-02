package gendocs

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// maxListSample caps top-level arrays in samples so the documents stay small.
const maxListSample = 2

// exampler attaches request and response Samples to every operation. The
// preference order is Golden File (shape from reality) > Postman saved
// example > generated from the schema; every source passes through the
// synthesizer so values are fake.
type exampler struct {
	doc    *Document
	golden *goldenSet
	pm     postmanIndex
	loc    *locale
	log    *report
}

func attachExamples(doc *Document, golden *goldenSet, pm postmanIndex, loc *locale, log *report) {
	e := &exampler{doc: doc, golden: golden, pm: pm, loc: loc, log: log}
	for _, path := range doc.Paths.Keys() {
		item, _ := doc.Paths.Get(path)
		for _, m := range httpMethods {
			op := item.(*PathItem).Operation(m)
			if op == nil {
				continue
			}
			e.operation(strings.ToUpper(m), path, op)
		}
	}
}

func (e *exampler) operation(method, path string, op *Operation) {
	pm := e.pm.lookup(method, path)
	for _, p := range op.Parameters {
		if p.Ref != "" || p.Example != nil {
			continue
		}
		p.Example = e.parameterExample(p, pm)
	}
	if op.RequestBody != nil {
		e.requestBody(op.RequestBody, pm)
	}
	for _, code := range op.Responses.Keys() {
		if !strings.HasPrefix(code, "2") {
			continue
		}
		r, _ := op.Responses.Get(code)
		e.response(method, path, code, r.(*Response), pm)
	}
}

// parameterExample picks a sample value for a path or query parameter.
func (e *exampler) parameterExample(p *Parameter, pm *postmanRequest) any {
	if p.In == "path" {
		return pathParamExample(p.Name)
	}
	if pm != nil {
		if v, ok := pm.Query.Get(p.Name); ok {
			if s := v.(string); s != "" && !strings.Contains(s, "string") {
				return coerceScalar(newSynthesizer().str(p.Name, s, nil), p.Schema)
			}
		}
	}
	if p.Schema != nil {
		if p.Schema.Default != nil {
			return p.Schema.Default
		}
		if len(p.Schema.Enum) > 0 {
			return p.Schema.Enum[0]
		}
	}
	return queryParamExample(p.Name, p.Schema)
}

var reIDParam = regexp.MustCompile(`_id$`)

func pathParamExample(name string) any {
	switch {
	case reIDParam.MatchString(name):
		return json.Number("1")
	case name == "provider_type":
		return "line"
	case name == "custom_field_name":
		return "vip_level"
	case name == "product_sku":
		return "SKU-001"
	case name == "job_id":
		return "5f1a2b3c4d5e6f7a8b9c0d1e"
	}
	return "sample"
}

// queryParamExample derives a value from the parameter name and type.
func queryParamExample(name string, s *Schema) any {
	base := ""
	if s != nil {
		base = s.BaseType()
	}
	switch {
	case strings.HasSuffix(name, "start_time") || name == "created_at_min" || name == "preorder_start_at":
		return "2026-09-01 00:00:00"
	case strings.HasSuffix(name, "end_time") || name == "created_at_max":
		return "2026-09-30 23:59:59"
	case strings.HasSuffix(name, "start_date"):
		return "2026-09-01"
	case strings.HasSuffix(name, "end_date"):
		return "2026-09-30"
	case name == "statuses" || name == "status":
		return "open"
	case name == "financial_statuses":
		return "paid"
	case name == "fulfillment_statuses":
		return "unshipped"
	case name == "return_statuses":
		return "no_need"
	case name == "tags" || name == "excluded_tags" || name == "tag":
		return "VIP"
	case name == "q" || name == "title" || name == "name":
		return "範例"
	case strings.HasSuffix(name, "_ids") || strings.HasSuffix(name, "_ids[]") || name == "ids" || name == "order_numbers":
		if base == "array" {
			return []any{json.Number("1"), json.Number("2")}
		}
		return "1,2"
	case name == "customer_emails":
		return synthEmail
	case name == "customer_mobiles":
		return synthMobile
	case name == "customer_name":
		return synthName
	case name == "vendor" || name == "author":
		return "範例廠商"
	case name == "data_source":
		return "ec"
	case name == "invoice_no":
		return "AB12345678"
	case name == "uid":
		return "U0000000000000000000000000000001"
	case name == "collection_handle":
		return "frontpage"
	case name == "search_column":
		return "phone"
	case name == "value":
		return "範例"
	}
	switch base {
	case "integer":
		if strings.HasSuffix(name, "_id") {
			return json.Number("1")
		}
		return json.Number("10")
	case "number":
		return json.Number("100.0")
	case "boolean":
		return true
	case "array":
		return []any{"範例"}
	}
	return "範例"
}

// requestBody fills the example of every media type of a request body.
func (e *exampler) requestBody(rb *RequestBody, pm *postmanRequest) {
	for _, mt := range rb.Content.Keys() {
		v, _ := rb.Content.Get(mt)
		media := v.(*MediaType)
		if media.Example != nil {
			continue
		}
		if pm != nil && pm.Body != nil {
			body := newSynthesizer().Value("", pm.Body, nil)
			if om, ok := body.(*OMap); ok {
				body = coerceObject(nestFormValues(om), media.Schema, e.doc)
			}
			media.Example = body
			continue
		}
		media.Example = e.fromSchema(media.Schema, "", 0)
	}
}

// response fills the example of a 2xx response.
func (e *exampler) response(method, path, code string, resp *Response, pm *postmanRequest) {
	if resp.Content == nil {
		return
	}
	v, ok := resp.Content.Get("application/json")
	if !ok {
		return
	}
	media := v.(*MediaType)
	if media.Example != nil {
		return
	}
	if gf := e.golden.success(method, path); gf != nil {
		if gf.isNullBody() {
			media.Example = nullValue{}
			resp.Description = e.loc.Source(resp.Description) + e.loc.T(nullNote)
			e.log.count("not-found-null")
			return
		}
		media.Example = truncateLists(newSynthesizer().Value("", gf.Body, nil))
		return
	}
	if pm != nil && pm.RespBody != nil {
		media.Example = truncateLists(newSynthesizer().Value("", pm.RespBody, nil))
		return
	}
	media.Example = e.fromSchema(media.Schema, "", 0)
}

// truncateLists keeps at most maxListSample items in a top-level array.
func truncateLists(v any) any {
	if arr, ok := v.([]any); ok && len(arr) > maxListSample {
		return arr[:maxListSample]
	}
	return v
}

// fromSchema generates a sample value from a schema.
func (e *exampler) fromSchema(s *Schema, key string, depth int) any {
	if s == nil || depth > 6 {
		return nullValue{}
	}
	if name := s.RefName(); name != "" {
		switch name {
		case "Timestamp":
			return synthTimestamp
		case "Date":
			return synthDate
		case "Money":
			return json.Number("100.0")
		}
		target, ok := e.doc.Components.Schemas.Get(name)
		if !ok {
			return nullValue{}
		}
		return e.fromSchema(target.(*Schema), key, depth+1)
	}
	if len(s.AnyOf) > 0 {
		return e.fromSchema(s.AnyOf[0], key, depth)
	}
	if len(s.Enum) > 0 {
		return s.Enum[0]
	}
	if s.Default != nil {
		return s.Default
	}
	switch s.BaseType() {
	case "object":
		out := NewOMap()
		if s.Properties != nil {
			for _, k := range s.Properties.Keys() {
				p, _ := s.Properties.Get(k)
				out.Set(k, e.fromSchema(p.(*Schema), k, depth+1))
			}
		}
		return out
	case "array":
		return []any{e.fromSchema(s.Items, key, depth+1)}
	case "integer":
		if reIDKey.MatchString(key) || key == "customer_id" {
			return json.Number("1")
		}
		return json.Number("1")
	case "number":
		return json.Number("100.0")
	case "boolean":
		return true
	case "string":
		if s.Format == "binary" {
			return "(binary file contents)"
		}
		syn := newSynthesizer()
		if v, ok := syn.keyed(key, "sample", nil); ok {
			return v
		}
		return syn.placeholder(key)
	}
	return nullValue{}
}

// nestFormValues turns bracketed Postman form keys into nested objects.
func nestFormValues(om *OMap) *OMap {
	out := NewOMap()
	for _, k := range om.Keys() {
		v, _ := om.Get(k)
		if !strings.Contains(k, "[") {
			out.Set(k, v)
			continue
		}
		segs := splitBrackets(k)
		cur := out
		for i, seg := range segs {
			last := i == len(segs)-1
			if seg == "" {
				existing, _ := cur.Get(segs[i-1])
				arr, _ := existing.([]any)
				if last {
					cur.Set(segs[i-1], append(arr, v))
					break
				}
				var item *OMap
				if len(arr) > 0 {
					item, _ = arr[len(arr)-1].(*OMap)
				}
				if item == nil {
					item = NewOMap()
					cur.Set(segs[i-1], append(arr, item))
				}
				cur = item
				continue
			}
			if last {
				cur.Set(seg, v)
				break
			}
			if i+1 < len(segs) && segs[i+1] == "" {
				if !cur.Has(seg) {
					cur.Set(seg, []any{})
				}
				continue
			}
			next, _ := cur.Get(seg)
			nm, ok := next.(*OMap)
			if !ok {
				nm = NewOMap()
				cur.Set(seg, nm)
			}
			cur = nm
		}
	}
	return out
}

// coerceObject converts string form values to the JSON types the schema
// declares (integers, numbers, booleans, arrays).
func coerceObject(om *OMap, s *Schema, doc *Document) *OMap {
	if s == nil {
		return om
	}
	out := NewOMap()
	for _, k := range om.Keys() {
		v, _ := om.Get(k)
		out.Set(k, coerceValue(v, derefIn(doc, s.Prop(k)), doc))
	}
	return out
}

func coerceValue(v any, s *Schema, doc *Document) any {
	if s == nil {
		return v
	}
	switch t := v.(type) {
	case *OMap:
		return coerceObject(t, s, doc)
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = coerceValue(e, derefIn(doc, s.Items), doc)
		}
		return out
	case string:
		return coerceScalar(t, s)
	}
	return v
}

// coerceScalar parses a string into the scalar type of s when possible.
func coerceScalar(v any, s *Schema) any {
	str, ok := v.(string)
	if !ok || s == nil {
		return v
	}
	base := s.BaseType()
	if base == "" && s.RefName() == "Money" {
		base = "number"
	}
	switch base {
	case "integer":
		if _, err := strconv.ParseInt(str, 10, 64); err == nil {
			return json.Number(str)
		}
		return json.Number("1")
	case "number":
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return json.Number(strconv.FormatFloat(math.Round(f*100)/100, 'f', -1, 64))
		}
		return json.Number("100.0")
	case "boolean":
		if b, err := strconv.ParseBool(str); err == nil {
			return b
		}
		return true
	case "array":
		if str == "" {
			return []any{}
		}
		return []any{coerceScalar(str, s.Items)}
	}
	return v
}

func derefIn(doc *Document, s *Schema) *Schema {
	for i := 0; i < 8 && s != nil; i++ {
		name := s.RefName()
		if name == "" {
			return s
		}
		switch name {
		case "Money":
			return &Schema{Type: "number"}
		case "Timestamp", "Date":
			return &Schema{Type: "string"}
		}
		v, ok := doc.Components.Schemas.Get(name)
		if !ok {
			return nil
		}
		s = v.(*Schema)
	}
	return s
}
