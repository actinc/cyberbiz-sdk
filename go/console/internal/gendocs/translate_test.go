package gendocs

import (
	"strings"
	"testing"
)

func TestGlossaryFileIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for i, line := range strings.Split(translationsTSV, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		zh, en, ok := strings.Cut(line, "\t")
		if !ok || strings.TrimSpace(zh) == "" || strings.TrimSpace(en) == "" {
			t.Errorf("line %d: expected \"<zh>\\t<en>\": %q", i+1, line)
			continue
		}
		if seen[zh] {
			t.Errorf("line %d: duplicate entry %q", i+1, zh)
		}
		seen[zh] = true
		if containsCJK(en) {
			t.Errorf("line %d: English side still contains CJK: %q", i+1, en)
		}
	}
}

func TestTranslateText(t *testing.T) {
	if got := translateText("訂單ID"); got != "Order ID." {
		t.Errorf("exact: %q", got)
	}
	multi := "訂單狀態，聯集請用逗點分隔\n「已開啟」: open\n「已結案」: closed"
	got := translateText(multi)
	if containsCJK(got) || !strings.Contains(got, "open: open") {
		t.Errorf("line-by-line: %q", got)
	}
	if got := translateText("no chinese here"); got != "no chinese here" {
		t.Error("ASCII text must pass through")
	}
	if got := translateText("未知的字串 xyz"); got != "未知的字串 xyz" {
		t.Error("unknown text must pass through unchanged")
	}
}

func TestObservedPath(t *testing.T) {
	p := parseObservedPath("$[].a.b[][].c")
	if strings.Join(p, "|") != "[]|a|b|[]|[]|c" || p.String() != "$[].a.b[][].c" {
		t.Errorf("parse/round-trip: %v %s", p, p.String())
	}
	if len(parseObservedPath("$")) != 0 {
		t.Error("root path must be empty")
	}
	f := &observedField{Types: []string{"integer", "null", "number"}}
	if f.primaryType() != "number" || !f.hasNull() {
		t.Errorf("primaryType: %s", f.primaryType())
	}
	if (&observedField{Types: []string{"string", "integer"}}).primaryType() != "" {
		t.Error("mixed scalar types have no primary type")
	}
	fields := map[string]*observedField{"$.a": {}, "$.b": {}, "$.b.c": {}, "$.d[]": {}}
	if got := directChildren(fields, observedPath{}); strings.Join(got, ",") != "a,b" {
		t.Errorf("directChildren: %v", got)
	}
}
