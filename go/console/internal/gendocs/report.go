package gendocs

import (
	"fmt"
	"sort"
	"strings"
)

// report collects warnings and per-rule correction counts for the summary
// the tool prints at the end of a run.
type report struct {
	warnings     []string
	counts       map[string]int
	notes        []string
	untranslated map[string]int // zh-TW strings without a glossary entry
}

func newReport() *report {
	return &report{counts: map[string]int{}, untranslated: map[string]int{}}
}

// untranslatedTSV lists the untranslated strings, most frequent first, in
// the "<count>\t<zh-TW>" form used to extend translations.tsv.
func (r *report) untranslatedTSV() string {
	keys := make([]string, 0, len(r.untranslated))
	for k := range r.untranslated {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if r.untranslated[keys[i]] != r.untranslated[keys[j]] {
			return r.untranslated[keys[i]] > r.untranslated[keys[j]]
		}
		return keys[i] < keys[j]
	})
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%d\t%s\n", r.untranslated[k], k)
	}
	return b.String()
}

func (r *report) warnf(format string, args ...any) {
	r.warnings = append(r.warnings, fmt.Sprintf(format, args...))
}

// count records one application of a correction rule.
func (r *report) count(rule string) {
	r.counts[rule]++
}

// countN records n applications of a correction rule.
func (r *report) countN(rule string, n int) {
	r.counts[rule] += n
}

func (r *report) notef(format string, args ...any) {
	r.notes = append(r.notes, fmt.Sprintf(format, args...))
}

// String renders the report as plain text.
func (r *report) String() string {
	var b strings.Builder
	b.WriteString("corrections applied:\n")
	ids := make([]string, 0, len(r.counts))
	for id := range r.counts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Fprintf(&b, "  %-28s %6d\n", id, r.counts[id])
	}
	for _, n := range r.notes {
		fmt.Fprintf(&b, "note: %s\n", n)
	}
	if len(r.warnings) > 0 {
		fmt.Fprintf(&b, "warnings (%d):\n", len(r.warnings))
		for _, w := range r.warnings {
			fmt.Fprintf(&b, "  %s\n", w)
		}
	}
	return b.String()
}
