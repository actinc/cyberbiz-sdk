package gendocs

import (
	"sort"
	"strings"
)

// locale selects the language of the free text in the generated documents.
// Structure (paths, names, enums, samples, headers) never depends on it.
//
// English is canonical: every string the tool authors is written in English
// in the Go tables and constants; zh-TW looks them up in localeZhTW.table.
// Descriptions that come from the CYBERBIZ sources are Traditional Chinese
// already: the en locale passes them through the glossary, zh-TW keeps them.
type locale struct {
	Code      string
	Dir       string            // output directory name under docs/api
	table     map[string]string // authored English -> localized
	fragments [][2]string       // substrings of composed strings
	glossary  bool              // translate source descriptions zh -> en
}

var (
	localeEN   = &locale{Code: "en", Dir: "en", glossary: true}
	localeZhTW = &locale{Code: "zh-TW", Dir: "zh-TW", table: zhTWStrings, fragments: zhTWFragments}
	locales    = []*locale{localeEN, localeZhTW}
)

// T localizes a string the tool authored. Unknown strings are returned
// unchanged, so English is always the fallback.
func (l *locale) T(s string) string {
	if l.table == nil || s == "" {
		return s
	}
	if v, ok := l.table[s]; ok {
		return v
	}
	if strings.Contains(s, "\n") {
		lines := strings.Split(s, "\n")
		changed := false
		for i, line := range lines {
			if v, ok := l.table[line]; ok {
				lines[i] = v
				changed = true
			}
		}
		if changed {
			s = strings.Join(lines, "\n")
		}
	}
	for _, f := range l.fragments {
		s = strings.ReplaceAll(s, f[0], f[1])
	}
	return s
}

// Source localizes a description taken from the CYBERBIZ material.
func (l *locale) Source(s string) string {
	if l.glossary {
		return translateText(s)
	}
	return s
}

// Localize is applied to every description of a finished document: authored
// text is looked up, source text goes through the glossary.
func (l *locale) Localize(s string) string {
	return l.Source(l.T(s))
}

// Map localizes every value of a description map.
func (l *locale) Map(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = l.T(v)
	}
	return out
}

// authoredStrings lists every English string the tool writes into the
// documents, so a test can prove the zh-TW table covers them all.
func authoredStrings() []string {
	set := map[string]bool{}
	add := func(ss ...string) {
		for _, s := range ss {
			if s != "" {
				set[s] = true
			}
		}
	}
	for _, v := range tagDescriptions {
		add(v)
	}
	for _, v := range v2FieldDescriptions {
		add(v)
	}
	for _, e := range enumDocs {
		add(e.Description)
	}
	for _, e := range errorStatuses {
		add(e.Name, e.Description)
	}
	for _, h := range paginationHeaders {
		add(h.Description)
	}
	for _, s := range sharedSchemas() {
		add(s.Description)
		if s.Properties != nil {
			for _, k := range s.Properties.Keys() {
				v, _ := s.Properties.Get(k)
				add(v.(*Schema).Description)
			}
		}
	}
	for _, ep := range v2Endpoints {
		add(ep.Summary, ep.Description, ep.RespDesc)
		for _, p := range ep.Params {
			add(p.Description)
		}
	}
	for _, ev := range webhookEvents {
		add(ev.Description)
	}
	for _, n := range releaseNotes {
		add(n.Changes...)
	}
	add(authoredConstants...)
	add(v2InlineStrings...)
	add(webhookPostmanStrings...)
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// missingLocaleStrings returns the authored strings a locale has no entry for.
func missingLocaleStrings(l *locale) []string {
	var out []string
	for _, s := range authoredStrings() {
		if _, ok := l.table[s]; !ok {
			out = append(out, s)
		}
	}
	return out
}
