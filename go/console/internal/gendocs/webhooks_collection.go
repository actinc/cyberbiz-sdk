package gendocs

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
)

// collectionEvent is one webhook request of the source-derived webhook
// collection (docs/references/cyberbiz-webhooks.postman_collection.json):
// its event code, the field table under "### Payload" and the sample body,
// whose JSON shape tells lists from objects.
type collectionEvent struct {
	Code   string
	Fields []whField // names may be dotted: "items.sku"
	Body   any       // decoded sample body (values are placeholders)
}

// postmanCollectionItem is the subset of a Postman v2.1 item the webhook
// collection uses.
type postmanCollectionItem struct {
	Name    string                   `json:"name"`
	Item    []postmanCollectionItem  `json:"item"`
	Request postmanCollectionRequest `json:"request"`
}

type postmanCollectionRequest struct {
	Description json.RawMessage `json:"description"`
	Body        struct {
		Raw string `json:"raw"`
	} `json:"body"`
}

var reCollectionRow = regexp.MustCompile("^\\| `([^`]+)` \\| ([^|]*) \\| (.*) \\|$")

// loadWebhookCollection parses the webhook collection. A missing file
// yields no events so the generator still runs without the reference.
// When an event appears in both delivery systems, the first one in the
// collection (Merchant Webhook, then App Store Webhook) wins.
func loadWebhookCollection(path string) (map[string]collectionEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]collectionEvent{}, nil
		}
		return nil, err
	}
	var coll struct {
		Item []postmanCollectionItem `json:"item"`
	}
	if err := json.Unmarshal(data, &coll); err != nil {
		return nil, err
	}
	out := map[string]collectionEvent{}
	var walk func(items []postmanCollectionItem)
	walk = func(items []postmanCollectionItem) {
		for _, it := range items {
			if len(it.Item) > 0 {
				walk(it.Item)
				continue
			}
			code, _, _ := strings.Cut(strings.TrimSpace(it.Name), " ")
			if _, seen := out[code]; seen || code == "" {
				continue
			}
			body := it.Request.bodyJSON()
			fields := payloadTable(it.Request.descriptionText())
			if len(fields) == 0 {
				fields = fieldsFromBody(body, "")
			}
			out[code] = collectionEvent{Code: code, Fields: fields, Body: body}
		}
	}
	walk(coll.Item)
	return out, nil
}

// descriptionText returns a description given as a string or as
// {"content": "..."}.
func (r postmanCollectionRequest) descriptionText() string {
	var s string
	if json.Unmarshal(r.Description, &s) == nil {
		return s
	}
	var obj struct {
		Content string `json:"content"`
	}
	if json.Unmarshal(r.Description, &obj) != nil {
		return ""
	}
	return obj.Content
}

// bodyJSON decodes the raw body, or returns nil when it is not JSON.
func (r postmanCollectionRequest) bodyJSON() any {
	var v any
	if json.Unmarshal([]byte(r.Body.Raw), &v) != nil {
		return nil
	}
	return v
}

// payloadTable reads the field table that follows "### Payload".
func payloadTable(desc string) []whField {
	_, rest, ok := strings.Cut(desc, "### Payload")
	if !ok {
		return nil
	}
	if next := strings.Index(rest, "\n### "); next >= 0 {
		rest = rest[:next]
	}
	var out []whField
	for _, line := range strings.Split(rest, "\n") {
		m := reCollectionRow.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		out = append(out, whField{Name: m[1], Type: strings.TrimSpace(m[2]), Desc: strings.TrimSpace(m[3])})
	}
	return out
}

// fieldsFromBody derives a field table from a sample body when the
// collection documents no table: types come from the JSON values, and a
// string value (which the collection uses to describe the field) becomes the
// description.
func fieldsFromBody(body any, prefix string) []whField {
	obj, ok := body.(map[string]any)
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []whField
	for _, k := range keys {
		name := k
		if prefix != "" {
			name = prefix + "." + k
		}
		v := obj[k]
		f := whField{Name: name, Type: jsonType(v)}
		if s, ok := v.(string); ok {
			f.Desc = s
		}
		out = append(out, f)
		if arr, ok := v.([]any); ok && len(arr) > 0 {
			v = arr[0]
		}
		out = append(out, fieldsFromBody(v, name)...)
	}
	return out
}

// jsonType names the reference type of a decoded JSON value.
func jsonType(v any) string {
	switch t := v.(type) {
	case bool:
		return "Boolean"
	case float64:
		if t == float64(int64(t)) {
			return "Integer"
		}
		return "Float"
	case map[string]any:
		return "Object"
	case []any:
		if len(t) > 0 {
			return "[" + jsonType(t[0]) + "]"
		}
		return "[String]"
	}
	return "String"
}

// addCollectionSections adds, for every event whose payload section the
// reference lacks, a section built from the collection's field table.
func addCollectionSections(src *webhookSource, coll map[string]collectionEvent, events []webhookEvent) {
	if src.Links == nil {
		src.Links = map[string]string{}
	}
	for _, ev := range events {
		if _, ok := src.Sections[ev.Payload]; ok {
			continue
		}
		ce, ok := coll[ev.Code]
		if !ok || len(ce.Fields) == 0 {
			continue
		}
		addSection(src, ev.Payload, ce)
	}
}

// addSection turns a flat field table into sections. A dotted field
// "items.sku" becomes field "sku" of child section "<name> items", linked
// from its parent through src.Links; a field is a list when the sample body
// holds an array there.
func addSection(src *webhookSource, name string, ce collectionEvent) {
	ensureSection(src, name)
	for _, f := range ce.Fields {
		path := strings.Split(f.Name, ".")
		parent := name
		for i := 1; i < len(path); i++ {
			child := name + " " + strings.Join(path[:i], " ")
			src.Links[parent+"/"+path[i-1]] = child
			ensureSection(src, child)
			markObject(src, parent, path[i-1], ce.Body, path[:i])
			parent = child
		}
		src.Sections[parent] = append(src.Sections[parent], whField{
			Name: path[len(path)-1],
			Type: collectionType(f.Type, ce.Body, path),
			Desc: f.Desc,
		})
	}
}

// markObject makes the parent field that owns a child section an object or
// a list of objects.
func markObject(src *webhookSource, parent, field string, body any, path []string) {
	for i, f := range src.Sections[parent] {
		if f.Name != field {
			continue
		}
		typ := "Object"
		if strings.HasPrefix(f.Type, "[") || isListAt(body, path) {
			typ = "[Object]"
		}
		src.Sections[parent][i].Type = typ
	}
}

func ensureSection(src *webhookSource, name string) {
	if _, ok := src.Sections[name]; !ok {
		src.Sections[name] = nil
		src.Order = append(src.Order, name)
	}
}

// collectionType maps a collection type to the reference vocabulary. A
// field is a list when the type says Array<...>, when the sample body holds
// an array there, or when an entity is exposed under a plural name (the
// collection's static samples render has-many entities as one object).
func collectionType(t string, body any, path []string) string {
	base := strings.TrimSpace(t)
	list := isListAt(body, path)
	if inner, ok := strings.CutPrefix(base, "Array<"); ok {
		base, list = strings.TrimSuffix(inner, ">"), true
	} else if strings.HasPrefix(base, "[") {
		base, list = strings.Trim(base, "[]"), true
	}
	field := path[len(path)-1]
	switch {
	case strings.HasSuffix(base, "Entity"):
		base = "Object"
		list = list || strings.HasSuffix(field, "s")
	case strings.Contains(base, "Boolean"):
		base = "Boolean"
	case base == "Hash", base == "Array", base == "JSON":
		base = "Object"
	case base == "Decimal", base == "BigDecimal", base == "Numeric":
		base = "Float"
	case base == "DateTime", base == "Time":
		base = "Date"
	case base == "String" && isNumberAt(body, path):
		base = "Integer" // the table says String but the sample is a number
	}
	if list {
		return "[" + base + "]"
	}
	return base
}

// isNumberAt reports whether the sample body holds a number at path.
func isNumberAt(body any, path []string) bool {
	cur := body
	for _, p := range path {
		if arr, ok := cur.([]any); ok && len(arr) > 0 {
			cur = arr[0]
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return false
		}
		cur = obj[p]
	}
	_, ok := cur.(float64)
	return ok
}

// isListAt reports whether the value at path in body is a JSON array.
func isListAt(body any, path []string) bool {
	cur := body
	for _, p := range path {
		if arr, ok := cur.([]any); ok {
			if len(arr) == 0 {
				return false
			}
			cur = arr[0]
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return false
		}
		cur = obj[p]
	}
	_, ok := cur.([]any)
	return ok
}
