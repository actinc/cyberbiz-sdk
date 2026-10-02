package gendocs

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
)

// toTree converts the typed OpenAPI model into a plain OMap / []any / scalar
// tree honouring `yaml:"name,omitempty"` tags, so that one marshaller
// (toYAMLNode) renders everything and json.Number samples stay numbers.
func toTree(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case nullValue:
		return t
	case *OMap:
		if t == nil {
			return nil
		}
		out := NewOMap()
		for _, k := range t.Keys() {
			val, _ := t.Get(k)
			out.Set(k, toTree(val))
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = toTree(e)
		}
		return out
	case []string:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = e
		}
		return out
	case json.Number, string, bool, int, int64, float64:
		return t
	case *float64:
		if t == nil {
			return nil
		}
		return *t
	case map[string][]string:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := NewOMap()
		for _, k := range keys {
			out.Set(k, toTree(t[k]))
		}
		return out
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return nil
		}
		return toTree(rv.Elem().Interface())
	case reflect.Slice:
		out := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out[i] = toTree(rv.Index(i).Interface())
		}
		return out
	case reflect.Struct:
		return structToTree(rv)
	}
	return v
}

func structToTree(rv reflect.Value) *OMap {
	out := NewOMap()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("yaml")
		if tag == "-" {
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		fv := rv.Field(i)
		if strings.Contains(opts, "omitempty") && isEmptyValue(fv) {
			continue
		}
		out.Set(name, toTree(fv.Interface()))
	}
	return out
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return true
		}
		if v.Kind() == reflect.Slice {
			return v.Len() == 0
		}
		if om, ok := v.Interface().(*OMap); ok {
			return om.Len() == 0
		}
		return false
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int64:
		return v.Int() == 0
	case reflect.Float64:
		return v.Float() == 0
	}
	return false
}
