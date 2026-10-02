// Package query encodes option structs into URL query parameters using
// `url:"name,omitempty"` struct tags, in the spirit of go-querystring but
// limited to what the CYBERBIZ API needs.
//
// Supported field types: string, bool, all integer and float kinds, slices of
// those (repeated by default, joined with "," when the tag has the comma
// option), pointers to any of them, values implementing fmt.Stringer, and
// embedded or named structs, which are flattened.
package query

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Values encodes v, which must be a struct or a pointer to one. A nil pointer
// encodes to an empty url.Values.
func Values(v any) (url.Values, error) {
	out := url.Values{}
	if v == nil {
		return out, nil
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return out, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("query: expected struct, got %s", rv.Type())
	}
	if err := encodeStruct(rv, out); err != nil {
		return nil, err
	}
	return out, nil
}

func encodeStruct(rv reflect.Value, out url.Values) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() && !field.Anonymous {
			continue
		}
		tag := field.Tag.Get("url")
		if tag == "-" {
			continue
		}
		fv := rv.Field(i)
		name, opts := parseTag(tag)
		if name == "" {
			if isStructLike(fv) {
				if err := encodeNested(fv, out); err != nil {
					return err
				}
				continue
			}
			name = strings.ToLower(field.Name)
		}
		if err := encodeField(name, opts, fv, out); err != nil {
			return fmt.Errorf("query: field %s: %w", field.Name, err)
		}
	}
	return nil
}

func encodeNested(fv reflect.Value, out url.Values) error {
	for fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			return nil
		}
		fv = fv.Elem()
	}
	return encodeStruct(fv, out)
}

func isStructLike(fv reflect.Value) bool {
	t := fv.Type()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct && !t.Implements(stringerType)
}

type tagOptions []string

func (o tagOptions) has(name string) bool {
	for _, opt := range o {
		if opt == name {
			return true
		}
	}
	return false
}

func parseTag(tag string) (string, tagOptions) {
	name, rest, _ := strings.Cut(tag, ",")
	if rest == "" {
		return name, nil
	}
	return name, strings.Split(rest, ",")
}

var stringerType = reflect.TypeFor[fmt.Stringer]()

func encodeField(name string, opts tagOptions, fv reflect.Value, out url.Values) error {
	// A non-nil pointer is never "empty": pointing at false or 0 is how a
	// caller says "send this value explicitly".
	pointer := fv.Kind() == reflect.Pointer
	for fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			return nil
		}
		fv = fv.Elem()
	}
	if opts.has("omitempty") && !pointer && isEmpty(fv) {
		return nil
	}
	if fv.Kind() == reflect.Slice || fv.Kind() == reflect.Array {
		if fv.Len() == 0 {
			return nil
		}
		var parts []string
		for i := 0; i < fv.Len(); i++ {
			s, err := scalar(fv.Index(i))
			if err != nil {
				return err
			}
			parts = append(parts, s)
		}
		if opts.has("comma") {
			out.Add(name, strings.Join(parts, ","))
			return nil
		}
		key := name
		if opts.has("brackets") {
			key += "[]"
		}
		for _, p := range parts {
			out.Add(key, p)
		}
		return nil
	}
	s, err := scalar(fv)
	if err != nil {
		return err
	}
	out.Add(name, s)
	return nil
}

func scalar(fv reflect.Value) (string, error) {
	for fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			return "", nil
		}
		fv = fv.Elem()
	}
	if fv.Type().Implements(stringerType) {
		return fv.Interface().(fmt.Stringer).String(), nil
	}
	switch fv.Kind() {
	case reflect.String:
		return fv.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(fv.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(fv.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(fv.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(fv.Float(), 'f', -1, 64), nil
	}
	return "", fmt.Errorf("unsupported type %s", fv.Type())
}

// isEmpty mirrors encoding/json's omitempty rule, plus IsZero() methods.
func isEmpty(fv reflect.Value) bool {
	if fv.CanInterface() {
		if z, ok := fv.Interface().(interface{ IsZero() bool }); ok {
			return z.IsZero()
		}
	}
	switch fv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return fv.Len() == 0
	case reflect.Bool:
		return !fv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return fv.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return fv.IsNil()
	}
	return fv.IsZero()
}
