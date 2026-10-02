package gendocs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// OMap is an insertion-ordered string-keyed map. It is the in-memory form of
// every JSON object the generator reads (Golden Files, Postman bodies, the
// swagger) and every object it writes (OpenAPI nodes, samples), so that key
// order is stable and re-runs produce identical output.
type OMap struct {
	keys []string
	vals map[string]any
}

// NewOMap returns an empty ordered map.
func NewOMap() *OMap {
	return &OMap{vals: map[string]any{}}
}

// Set stores v under k, appending k to the key order if it is new.
func (m *OMap) Set(k string, v any) *OMap {
	if m.vals == nil {
		m.vals = map[string]any{}
	}
	if _, ok := m.vals[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.vals[k] = v
	return m
}

// Get returns the value stored under k.
func (m *OMap) Get(k string) (any, bool) {
	if m == nil {
		return nil, false
	}
	v, ok := m.vals[k]
	return v, ok
}

// Has reports whether k is present.
func (m *OMap) Has(k string) bool {
	_, ok := m.Get(k)
	return ok
}

// Delete removes k, preserving the order of the remaining keys.
func (m *OMap) Delete(k string) {
	if m == nil {
		return
	}
	if _, ok := m.vals[k]; !ok {
		return
	}
	delete(m.vals, k)
	for i, kk := range m.keys {
		if kk == k {
			m.keys = append(m.keys[:i:i], m.keys[i+1:]...)
			break
		}
	}
}

// Keys returns the keys in insertion order.
func (m *OMap) Keys() []string {
	if m == nil {
		return nil
	}
	return append([]string(nil), m.keys...)
}

// Len returns the number of entries.
func (m *OMap) Len() int {
	if m == nil {
		return 0
	}
	return len(m.keys)
}

// SortKeys reorders the map alphabetically; used where the OpenAPI output
// has no natural order (paths, components).
func (m *OMap) SortKeys() {
	sort.Strings(m.keys)
}

// Clone deep-copies the map and any nested OMap / []any values.
func (m *OMap) Clone() *OMap {
	if m == nil {
		return nil
	}
	out := NewOMap()
	for _, k := range m.keys {
		out.Set(k, cloneValue(m.vals[k]))
	}
	return out
}

func cloneValue(v any) any {
	switch t := v.(type) {
	case *OMap:
		return t.Clone()
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = cloneValue(e)
		}
		return out
	default:
		return v
	}
}

// UnmarshalJSON decodes a JSON object preserving key order. Numbers are kept
// as json.Number so that "9999.0" survives as a float literal.
func (m *OMap) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	v, err := decodeOrdered(dec)
	if err != nil {
		return err
	}
	om, ok := v.(*OMap)
	if !ok {
		return fmt.Errorf("omap: expected object, got %T", v)
	}
	*m = *om
	return nil
}

// DecodeJSON decodes arbitrary JSON into OMap / []any / json.Number / string /
// bool / nil, preserving object key order.
func DecodeJSON(b []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	v, err := decodeOrdered(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("omap: trailing data after JSON value")
	}
	return v, nil
}

func decodeOrdered(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			m := NewOMap()
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				k, ok := kt.(string)
				if !ok {
					return nil, fmt.Errorf("omap: object key is %T", kt)
				}
				v, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				m.Set(k, v)
			}
			if _, err := dec.Token(); err != nil { // '}'
				return nil, err
			}
			return m, nil
		case '[':
			arr := []any{}
			for dec.More() {
				v, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			if _, err := dec.Token(); err != nil { // ']'
				return nil, err
			}
			return arr, nil
		}
		return nil, fmt.Errorf("omap: unexpected delimiter %q", t)
	default:
		return tok, nil
	}
}

// MarshalJSON encodes the map in insertion order.
func (m *OMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		vb, err := marshalJSONValue(m.vals[k])
		if err != nil {
			return nil, err
		}
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func marshalJSONValue(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// EncodeJSONIndent renders v as indented JSON without HTML escaping, with a
// trailing newline, in key order.
func EncodeJSONIndent(v any, indent string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", indent)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalYAML renders the map as a YAML mapping node in insertion order.
func (m *OMap) MarshalYAML() (any, error) {
	return toYAMLNode(m)
}

// nullValue is an explicit JSON null in a sample. Go nil is dropped by
// omitempty, so the bare-null responses some lookups return use this type.
type nullValue struct{}

// MarshalJSON renders the JSON null.
func (nullValue) MarshalJSON() ([]byte, error) { return []byte("null"), nil }

// toYAMLNode converts generator values to yaml nodes. json.Number becomes an
// unquoted int/float scalar; multi-line strings use literal block style.
func toYAMLNode(v any) (*yaml.Node, error) {
	switch t := v.(type) {
	case nil, nullValue:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
	case *OMap:
		n := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		if t == nil {
			return n, nil
		}
		for _, k := range t.keys {
			kn, err := toYAMLNode(k)
			if err != nil {
				return nil, err
			}
			vn, err := toYAMLNode(t.vals[k])
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, kn, vn)
		}
		return n, nil
	case []any:
		n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, e := range t {
			en, err := toYAMLNode(e)
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, en)
		}
		return n, nil
	case []string:
		n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, e := range t {
			en, err := toYAMLNode(e)
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, en)
		}
		return n, nil
	case json.Number:
		tag := "!!int"
		if strings.ContainsAny(t.String(), ".eE") {
			tag = "!!float"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: t.String()}, nil
	case string:
		n := &yaml.Node{}
		if err := n.Encode(t); err != nil {
			return nil, err
		}
		if strings.Contains(t, "\n") && !strings.ContainsAny(t, "\t\r") {
			n.Style = yaml.LiteralStyle
		}
		return n, nil
	default:
		n := &yaml.Node{}
		if err := n.Encode(v); err != nil {
			return nil, err
		}
		return n, nil
	}
}

// MarshalYAMLDocument renders a top-level value as a YAML document.
func MarshalYAMLDocument(v any) ([]byte, error) {
	node, err := toYAMLNode(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
