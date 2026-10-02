package gendocs

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// observedDoc is docs/references/notes/schema_observed.json: for every
// endpoint the sweep called, the JSON types seen at each JSON path.
type observedDoc struct {
	Endpoints   map[string]map[string]*observedField `json:"endpoints"`
	Envelope    map[string][]string                  `json:"envelope"`
	ErrorBodies map[string]map[string]*observedField `json:"error_bodies_by_status"`
}

type observedField struct {
	Types    []string `json:"types"`
	Examples []any    `json:"examples"`
	Formats  []string `json:"formats"`
	Count    int      `json:"count"`
}

// loadObserved reads schema_observed.json; a missing file yields an empty doc
// so the generator still runs (with no observed corrections).
func loadObserved(path string) (*observedDoc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &observedDoc{Endpoints: map[string]map[string]*observedField{}}, nil
		}
		return nil, err
	}
	var doc observedDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("schema_observed %s: %w", path, err)
	}
	if doc.Endpoints == nil {
		doc.Endpoints = map[string]map[string]*observedField{}
	}
	return &doc, nil
}

// hasNull reports whether null was observed.
func (f *observedField) hasNull() bool {
	return f != nil && containsString(f.Types, "null")
}

// nonNullTypes returns the observed types other than null, sorted.
func (f *observedField) nonNullTypes() []string {
	if f == nil {
		return nil
	}
	var out []string
	for _, t := range f.Types {
		if t != "null" {
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out
}

// primaryType collapses the observed non-null types into one JSON Schema type:
// integer+number -> number; anything mixed otherwise -> "".
func (f *observedField) primaryType() string {
	ts := f.nonNullTypes()
	switch len(ts) {
	case 0:
		return ""
	case 1:
		return ts[0]
	case 2:
		if ts[0] == "integer" && ts[1] == "number" {
			return "number"
		}
	}
	return ""
}

// format returns the single observed string format, if any.
func (f *observedField) format() string {
	if f == nil || len(f.Formats) != 1 {
		return ""
	}
	return f.Formats[0]
}

// observedPath is a parsed "$[].customer.address" key: a list of tokens where
// "[]" means array items and anything else is a property name.
type observedPath []string

// parseObservedPath tokenises "$[].a.b[][].c" into ["[]","a","b","[]","[]","c"].
func parseObservedPath(s string) observedPath {
	s = strings.TrimPrefix(s, "$")
	var out observedPath
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '.':
			flush()
		case s[i] == '[' && i+1 < len(s) && s[i+1] == ']':
			flush()
			out = append(out, "[]")
			i++
		default:
			cur.WriteByte(s[i])
		}
	}
	flush()
	return out
}

// String renders the path back to the sweep notation.
func (p observedPath) String() string {
	var b strings.Builder
	b.WriteString("$")
	for _, t := range p {
		if t == "[]" {
			b.WriteString("[]")
		} else {
			b.WriteString(".")
			b.WriteString(t)
		}
	}
	return b.String()
}

// child returns the key of a direct child path.
func (p observedPath) child(tok string) observedPath {
	out := make(observedPath, 0, len(p)+1)
	out = append(out, p...)
	return append(out, tok)
}

// sortedKeys returns the observed paths of one endpoint, shortest first so
// parents are visited before children, then alphabetically.
func sortedObservedKeys(fields map[string]*observedField) []string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		li, lj := len(parseObservedPath(keys[i])), len(parseObservedPath(keys[j]))
		if li != lj {
			return li < lj
		}
		return keys[i] < keys[j]
	})
	return keys
}

// directChildren lists the property names observed directly under parent
// (for objects) in a stable order.
func directChildren(fields map[string]*observedField, parent observedPath) []string {
	prefix := parent.String()
	var out []string
	for k := range fields {
		if !strings.HasPrefix(k, prefix+".") {
			continue
		}
		rest := parseObservedPath(k)[len(parent):]
		if len(rest) == 1 && rest[0] != "[]" {
			out = append(out, rest[0])
		}
	}
	sort.Strings(out)
	return out
}

func containsString(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
