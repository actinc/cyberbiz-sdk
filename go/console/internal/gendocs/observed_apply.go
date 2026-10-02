package gendocs

import (
	"sort"
	"strings"
)

const (
	observedNote = "Observed in live responses; not in the CYBERBIZ swagger."
	nullOnlyNote = "Only null was observed; the type is unknown."
)

// applyObserved reconciles the document with schema_observed.json: for every
// observed GET endpoint the response schema is walked alongside the observed
// JSON paths, and each field takes the observed type, nullability and format.
// Fields seen in Golden Files but absent from the swagger are added.
func applyObserved(doc *Document, obs *observedDoc, log *report) {
	for _, path := range sortedKeys(obs.Endpoints) {
		item, ok := doc.Paths.Get(path)
		if !ok {
			continue
		}
		op := item.(*PathItem).Get
		if op == nil {
			continue
		}
		media := successMedia(op)
		if media == nil {
			continue
		}
		a := &observedApplier{doc: doc, fields: obs.Endpoints[path], log: log, path: path}
		media.Schema = a.applyRoot(media.Schema)
	}
}

// successMedia returns the JSON media type of the first 2xx response.
func successMedia(op *Operation) *MediaType {
	for _, code := range op.Responses.Keys() {
		if !strings.HasPrefix(code, "2") {
			continue
		}
		r, _ := op.Responses.Get(code)
		resp := r.(*Response)
		if resp.Content == nil {
			continue
		}
		if v, ok := resp.Content.Get("application/json"); ok {
			return v.(*MediaType)
		}
	}
	return nil
}

type observedApplier struct {
	doc    *Document
	fields map[string]*observedField
	log    *report
	path   string
	hint   *Schema // item schema of an array that was observed as an object
}

// applyRoot reconciles the response envelope, then every nested path.
func (a *observedApplier) applyRoot(root *Schema) *Schema {
	if root == nil {
		root = &Schema{}
	}
	rootField := a.fields["$"]
	root = a.reconcile("", root, rootField, observedPath{}, root)
	for _, key := range sortedObservedKeys(a.fields) {
		p := parseObservedPath(key)
		if len(p) == 0 {
			continue
		}
		parent := a.resolve(root, p[:len(p)-1])
		if parent == nil {
			continue
		}
		last := p[len(p)-1]
		if last == "[]" {
			if parent.BaseType() != "array" {
				continue
			}
			parent.Items = a.reconcile("", parent.Items, a.fields[key], p, parent.Items)
			continue
		}
		if parent.BaseType() != "object" && parent.Properties == nil {
			continue
		}
		child := parent.Prop(last)
		if child == nil {
			parent.SetProp(last, a.build(p, nil, true))
			a.log.count("observed-field")
			continue
		}
		parent.SetProp(last, a.reconcile(last, child, a.fields[key], p, child))
	}
	return root
}

// resolve walks path from root through refs, items and properties.
func (a *observedApplier) resolve(root *Schema, path observedPath) *Schema {
	cur := a.deref(root)
	for _, tok := range path {
		if cur == nil {
			return nil
		}
		if tok == "[]" {
			cur = a.deref(cur.Items)
			continue
		}
		cur = a.deref(cur.Prop(tok))
	}
	return cur
}

// deref follows a $ref (also inside anyOf) to the shared component schema.
func (a *observedApplier) deref(s *Schema) *Schema {
	for i := 0; i < 8 && s != nil; i++ {
		name := s.RefName()
		if name == "" {
			return s
		}
		v, ok := a.doc.Components.Schemas.Get(name)
		if !ok {
			return nil
		}
		s = v.(*Schema)
	}
	return s
}

// reconcile returns the schema that should sit at path given what was
// observed there. candidate is the schema currently documented at the path.
func (a *observedApplier) reconcile(key string, cur *Schema, f *observedField, path observedPath, candidate *Schema) *Schema {
	if cur == nil {
		return a.build(path, candidate, true)
	}
	if f == nil {
		return cur
	}
	obsType := f.primaryType()
	curType := a.effectiveType(cur)
	switch {
	case obsType == "" || obsType == curType:
	case curType == "object" && obsType == "array":
		cur = &Schema{Type: "array", Items: cur, Description: cur.Description}
		a.log.count("observed-array")
	case curType == "array" && obsType == "object":
		a.hint = cur.Items
		cur = a.build(path, cur.Items, false)
		a.hint = nil
		a.log.count("observed-object")
	case isScalar(obsType) && (isScalar(curType) || curType == ""):
		cur.Type = replaceBaseType(cur.Type, obsType)
		if cur.Ref != "" {
			cur.Ref = ""
			cur.Type = obsType
		}
		cur.Format = ""
		a.log.count("observed-type")
		if key == "order_number" {
			a.log.count("order-number")
		}
	default:
		a.log.warnf("%s %s: swagger %s vs observed %s not reconciled", a.path, path, curType, obsType)
	}
	if fm := f.format(); fm != "" && cur.BaseType() == "string" && cur.RefName() == "" {
		if name := formatSchemaName(fm); name != "" {
			cur = refSchema(name, cur)
			a.log.count("observed-format")
		}
	}
	if f.hasNull() && !cur.IsNullable() {
		cur = markNullable(cur)
		a.log.count("observed-null")
	}
	return cur
}

func (a *observedApplier) effectiveType(s *Schema) string {
	if s.RefName() != "" {
		switch s.RefName() {
		case "Timestamp", "Date":
			return "string"
		case "Money":
			return "number"
		}
		return a.effectiveType(a.deref(s))
	}
	if t := s.BaseType(); t != "" {
		return t
	}
	if s.Properties != nil {
		return "object"
	}
	return ""
}

func isScalar(t string) bool {
	switch t {
	case "string", "integer", "number", "boolean":
		return true
	}
	return false
}

// formatSchemaName maps an observed string format to a shared schema.
func formatSchemaName(format string) string {
	switch format {
	case "YYYY-MM-DD HH:MM:SS":
		return "Timestamp"
	case "YYYY-MM-DD":
		return "Date"
	}
	return ""
}

// build creates a schema for an observed path from the observed sub-tree.
// candidate, when it is a component ref whose properties cover everything
// observed, is reused instead of an inline object. added marks fields the
// swagger does not know about.
func (a *observedApplier) build(path observedPath, candidate *Schema, added bool) *Schema {
	f := a.fields[path.String()]
	var s *Schema
	switch t := f.primaryType(); t {
	case "object":
		s = a.buildObject(path, candidate, added)
	case "array":
		var itemCandidate *Schema
		if candidate != nil {
			itemCandidate = candidate.Items
		}
		s = &Schema{Type: "array", Items: a.build(path.child("[]"), itemCandidate, added)}
	case "":
		if candidate != nil {
			s = candidate.Clone()
		} else {
			s = &Schema{Description: nullOnlyNote}
		}
	default:
		s = &Schema{Type: t}
		if name := formatSchemaName(f.format()); name != "" {
			s = &Schema{Ref: schemaRef(name)}
		}
	}
	if added {
		s.XObserved = true
		if s.Description == "" {
			s.Description = observedNote
		}
	}
	if f.hasNull() {
		s = markNullable(s)
	}
	return s
}

func (a *observedApplier) buildObject(path observedPath, candidate *Schema, added bool) *Schema {
	children := directChildren(a.fields, path)
	if candidate == nil && a.hint != nil && len(path) > 0 {
		candidate = a.hint
	}
	if candidate != nil && candidate.RefName() != "" && a.covers(candidate, children) {
		return &Schema{Ref: candidate.Ref}
	}
	s := &Schema{Type: "object", Properties: NewOMap()}
	for _, c := range children {
		s.SetProp(c, a.build(path.child(c), nil, added))
	}
	return s
}

// covers reports whether the referenced component has every listed property.
func (a *observedApplier) covers(ref *Schema, props []string) bool {
	target := a.deref(ref)
	if target == nil {
		return false
	}
	for _, p := range props {
		if target.Prop(p) == nil {
			return false
		}
	}
	return true
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
