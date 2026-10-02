package catalog

import "strings"

const maxDepth = 12

// resolve replaces $ref pointers (local "#/..." only) recursively, with a
// depth guard so self-referential schemas terminate.
func (p *parser) resolve(node any, depth int) any {
	if depth > maxDepth {
		return nil
	}
	switch v := node.(type) {
	case map[string]any:
		if ref, ok := v["$ref"].(string); ok {
			return p.resolve(p.lookup(ref), depth+1)
		}
		out := make(map[string]any, len(v))
		for k, child := range v {
			out[k] = p.resolve(child, depth+1)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			out[i] = p.resolve(child, depth+1)
		}
		return out
	}
	return node
}

func (p *parser) lookup(ref string) any {
	if !strings.HasPrefix(ref, "#/") {
		return map[string]any{"$unresolved": ref}
	}
	var cur any = p.doc
	for _, seg := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		seg = strings.NewReplacer("~1", "/", "~0", "~").Replace(seg)
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[seg]
	}
	return cur
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func asSlice(v any) []any {
	s, _ := v.([]any)
	return s
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asBool(v any) bool {
	b, _ := v.(bool)
	return b
}
