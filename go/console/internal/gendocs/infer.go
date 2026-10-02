package gendocs

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	reTimestampValue = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}(:\d{2})?( ?\+0800)?$`)
	reDateValue      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	reBoolKey        = regexp.MustCompile(`^(enable_|is_|accepts_|has_|can_)|(_enabled|_management|_sync|published|searchable)$`)
)

// inferSchema derives a JSON Schema from a sample value. Object keys keep
// their order; arrays take the shape of their first element; null values
// get a type guessed from the key name and are marked nullable. descs
// supplies field descriptions by key.
func inferSchema(v any, key string, descs map[string]string) *Schema {
	s := inferType(v, key, descs)
	if d, ok := descs[key]; ok && key != "" && s.Description == "" {
		s.Description = d
	}
	return s
}

func inferType(v any, key string, descs map[string]string) *Schema {
	switch t := v.(type) {
	case *OMap:
		s := &Schema{Type: "object", Properties: NewOMap()}
		for _, k := range t.Keys() {
			cv, _ := t.Get(k)
			s.SetProp(k, inferSchema(cv, k, descs))
		}
		if t.Len() == 0 {
			s.Properties = nil
			s.AdditionalProperties = true
		}
		return s
	case []any:
		s := &Schema{Type: "array"}
		if len(t) > 0 {
			s.Items = inferSchema(t[0], key, descs)
			s.Items.Description = ""
		} else {
			s.Items = guessFromKey(key)
		}
		return s
	case json.Number:
		if strings.ContainsAny(t.String(), ".eE") || isMoneyKey(key) {
			if isMoneyKey(key) {
				return &Schema{Ref: schemaRef("Money")}
			}
			return &Schema{Type: "number"}
		}
		return &Schema{Type: "integer"}
	case string:
		switch {
		case reTimestampValue.MatchString(t):
			return &Schema{Ref: schemaRef("Timestamp")}
		case reDateValue.MatchString(t):
			return &Schema{Ref: schemaRef("Date")}
		}
		return &Schema{Type: "string"}
	case bool:
		return &Schema{Type: "boolean"}
	case nil, nullValue:
		return markNullable(guessFromKey(key))
	}
	return &Schema{}
}

// guessFromKey picks a plausible type for a field only ever seen as null.
func guessFromKey(key string) *Schema {
	switch {
	case key == "":
		return &Schema{Description: nullOnlyNote}
	case reIDKey.MatchString(key) || key == "customer_id":
		return &Schema{Type: "integer"}
	case reTimeSuffix.MatchString(key):
		return &Schema{Ref: schemaRef("Timestamp")}
	case key == "birthday" || strings.HasSuffix(key, "_date"):
		return &Schema{Ref: schemaRef("Date")}
	case isMoneyKey(key):
		return &Schema{Ref: schemaRef("Money")}
	case reBoolKey.MatchString(key):
		return &Schema{Type: "boolean"}
	case strings.HasSuffix(key, "_quantity") || strings.HasSuffix(key, "_count") || key == "position" || key == "quantity":
		return &Schema{Type: "integer"}
	case key == "shop_discount" || key == "coupon_discount" || key == "special_collection" || key == "pos_shop" || key == "einvoice" || key == "branch_store" || key == "shop_line_chat_bot":
		return &Schema{Type: "object", AdditionalProperties: true}
	}
	return &Schema{Type: "string"}
}

// decodeSample parses one of the embedded JSON samples; a syntax error is a
// programming mistake, so it panics.
func decodeSample(src string) any {
	v, err := DecodeJSON([]byte(src))
	if err != nil {
		panic("gendocs: bad embedded sample: " + err.Error())
	}
	return v
}
