package gendocs

import (
	_ "embed"
	"strings"
	"sync"
)

// translations.tsv is the zh-TW -> English glossary applied to every
// description that comes from the CYBERBIZ swagger or webhook reference.
// One entry per line: "<zh-TW>\t<English>". Lookups are exact, first on the
// whole description and then line by line, so multi-line descriptions
// (enum lists) translate as long as each line is in the glossary.
//
//go:embed translations.tsv
var translationsTSV string

var (
	glossaryOnce sync.Once
	glossary     map[string]string
)

func loadGlossary() map[string]string {
	glossaryOnce.Do(func() {
		glossary = map[string]string{}
		for _, line := range strings.Split(translationsTSV, "\n") {
			zh, en, ok := strings.Cut(line, "\t")
			if !ok || strings.HasPrefix(zh, "#") {
				continue
			}
			glossary[strings.TrimSpace(zh)] = strings.TrimSpace(en)
		}
	})
	return glossary
}

// translateText translates a description through the glossary. Strings
// without a translation are returned unchanged.
func translateText(s string) string {
	g := loadGlossary()
	if s == "" || !containsCJK(s) {
		return s
	}
	if en, ok := g[strings.TrimSpace(s)]; ok {
		return en
	}
	lines := strings.Split(s, "\n")
	changed := false
	for i, l := range lines {
		if en, ok := g[strings.TrimSpace(l)]; ok {
			lines[i] = en
			changed = true
		}
	}
	if changed {
		return strings.Join(lines, "\n")
	}
	return s
}

// containsCJK reports whether s has any CJK character.
func containsCJK(s string) bool {
	for _, r := range s {
		if r >= 0x2E80 && r <= 0x9FFF || r >= 0xF900 && r <= 0xFAFF || r >= 0xFF00 && r <= 0xFFEF {
			return true
		}
	}
	return false
}

// localizeDocument applies the locale to every free-text string of doc:
// authored text is looked up in the locale table, source text goes through
// the glossary (en only). It records how many strings stayed untranslated.
func localizeDocument(doc *Document, loc *locale, log *report) {
	translated, remaining := 0, 0
	tr := func(s *string) {
		if *s == "" {
			return
		}
		out := loc.Localize(*s)
		if out != *s {
			translated++
			*s = out
		}
		if loc.glossary && containsCJK(out) {
			remaining++
			for _, line := range strings.Split(out, "\n") {
				if containsCJK(line) {
					log.untranslated[strings.TrimSpace(line)]++
				}
			}
		}
	}
	tr(&doc.Info.Title)
	tr(&doc.Info.Description)
	for _, s := range doc.Servers {
		tr(&s.Description)
	}
	for _, t := range doc.Tags {
		tr(&t.Description)
	}
	walkDocumentSchemas(doc, func(_ string, s *Schema) *Schema {
		tr(&s.Description)
		tr(&s.Title)
		return s
	})
	for _, path := range doc.Paths.Keys() {
		item, _ := doc.Paths.Get(path)
		for _, m := range httpMethods {
			op := item.(*PathItem).Operation(m)
			if op == nil {
				continue
			}
			tr(&op.Summary)
			tr(&op.Description)
			for _, p := range op.Parameters {
				tr(&p.Description)
			}
			if op.RequestBody != nil {
				tr(&op.RequestBody.Description)
			}
			localizeResponses(op.Responses, tr)
		}
	}
	localizeComponents(doc.Components, tr)
	if loc.glossary {
		log.countN("translate", translated)
	} else {
		log.countN("localize-"+loc.Code, translated)
	}
	if remaining > 0 {
		log.notef("%s: %d descriptions still contain Chinese (add them to translations.tsv)", doc.Info.Title, remaining)
	}
}

func localizeResponses(responses *OMap, tr func(*string)) {
	if responses == nil {
		return
	}
	for _, code := range responses.Keys() {
		r, _ := responses.Get(code)
		resp := r.(*Response)
		tr(&resp.Description)
		if resp.Headers != nil {
			for _, k := range resp.Headers.Keys() {
				h, _ := resp.Headers.Get(k)
				tr(&h.(*Header).Description)
			}
		}
		if resp.Content != nil {
			for _, k := range resp.Content.Keys() {
				m, _ := resp.Content.Get(k)
				localizeExamples(m.(*MediaType).Examples, tr)
			}
		}
	}
}

func localizeExamples(examples *OMap, tr func(*string)) {
	if examples == nil {
		return
	}
	for _, k := range examples.Keys() {
		e, _ := examples.Get(k)
		tr(&e.(*Example).Summary)
		tr(&e.(*Example).Description)
	}
}

func localizeComponents(c *Components, tr func(*string)) {
	if c == nil {
		return
	}
	if c.Parameters != nil {
		for _, k := range c.Parameters.Keys() {
			p, _ := c.Parameters.Get(k)
			tr(&p.(*Parameter).Description)
		}
	}
	if c.Headers != nil {
		for _, k := range c.Headers.Keys() {
			h, _ := c.Headers.Get(k)
			tr(&h.(*Header).Description)
		}
	}
	localizeResponses(c.Responses, tr)
	localizeExamples(c.Examples, tr)
	if c.SecuritySchemes != nil {
		for _, k := range c.SecuritySchemes.Keys() {
			s, _ := c.SecuritySchemes.Get(k)
			tr(&s.(*SecurityScheme).Description)
		}
	}
}
