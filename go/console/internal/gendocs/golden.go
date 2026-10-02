package gendocs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// goldenFile is one redacted real response from testdata/golden/<group>/.
// The file name encodes METHOD_path with {id} for path parameters.
type goldenFile struct {
	Group   string // v1, v2, app, errors
	Method  string // GET, POST, ...
	Name    string // e.g. v1_orders_{id}
	Body    any    // decoded JSON (OMap / []any / scalars / nil)
	Headers map[string][]string
	Status  int // 200 for v1/v2/app; inferred for errors
}

// goldenSet indexes Golden Files by group.
type goldenSet struct {
	files []*goldenFile
}

// loadGolden reads every Golden File under root.
func loadGolden(root string) (*goldenSet, error) {
	set := &goldenSet{}
	groups, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return set, nil
		}
		return nil, err
	}
	for _, g := range groups {
		if !g.IsDir() {
			continue
		}
		dir := filepath.Join(root, g.Name())
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".headers.json") {
				continue
			}
			gf, err := readGoldenFile(dir, g.Name(), name)
			if err != nil {
				return nil, err
			}
			set.files = append(set.files, gf)
		}
	}
	sort.Slice(set.files, func(i, j int) bool {
		a, b := set.files[i], set.files[j]
		if a.Group != b.Group {
			return a.Group < b.Group
		}
		if a.Method != b.Method {
			return a.Method < b.Method
		}
		return a.Name < b.Name
	})
	return set, nil
}

func readGoldenFile(dir, group, name string) (*goldenFile, error) {
	base := strings.TrimSuffix(name, ".json")
	method, rest, ok := strings.Cut(base, "_")
	if !ok {
		return nil, fmt.Errorf("golden %s/%s: name must be METHOD_path", group, name)
	}
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	body, err := DecodeJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("golden %s/%s: %w", group, name, err)
	}
	gf := &goldenFile{Group: group, Method: method, Name: rest, Body: body, Status: 200}
	if hb, err := os.ReadFile(filepath.Join(dir, base+".headers.json")); err == nil {
		if err := json.Unmarshal(hb, &gf.Headers); err != nil {
			return nil, fmt.Errorf("golden %s/%s headers: %w", group, name, err)
		}
	}
	if group == "errors" {
		gf.Status = inferErrorStatus(body)
	}
	return gf, nil
}

// inferErrorStatus maps an error body to the status the sweep recorded for
// that shape (see schema_observed.error_bodies_by_status). Bodies that are
// not error envelopes return 0 and are ignored.
func inferErrorStatus(body any) int {
	m, ok := body.(*OMap)
	if !ok {
		return 0
	}
	if v, ok := m.Get("error"); ok {
		switch t := v.(type) {
		case []any:
			for _, e := range t {
				s, _ := e.(string)
				switch {
				case s == "無權使用該 API":
					return 401
				case strings.HasPrefix(s, "無此"):
					return 404
				}
			}
			return 422
		case string:
			return 403
		}
	}
	if _, ok := m.Get("messages"); ok {
		return 403
	}
	if _, ok := m.Get("message"); ok {
		return 422
	}
	return 0
}

// goldenPathPattern builds a regexp matching the golden Name of an operation
// path: literal segments joined by "_", path parameters matching "{id}" or a
// literal value (e.g. uid_providers_line for {provider_type}).
func goldenPathPattern(path string) *regexp.Regexp {
	segs := strings.Split(strings.Trim(path, "/"), "/")
	var b strings.Builder
	b.WriteString("^")
	for i, s := range segs {
		if i > 0 {
			b.WriteString("_")
		}
		if strings.HasPrefix(s, "{") {
			b.WriteString(`(\{id\}|[^_]+)`)
		} else {
			b.WriteString(regexp.QuoteMeta(s))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// find returns the Golden Files recorded for method+path in the given groups,
// preferring an exact "{id}" spelling over literal path-parameter values.
func (g *goldenSet) find(method, path string, groups ...string) []*goldenFile {
	if g == nil {
		return nil
	}
	exact := strings.ReplaceAll(strings.Trim(path, "/"), "/", "_")
	exact = regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(exact, "{id}")
	pat := goldenPathPattern(path)
	var exactHits, looseHits []*goldenFile
	for _, f := range g.files {
		if f.Method != method || !containsString(groups, f.Group) {
			continue
		}
		switch {
		case f.Name == exact:
			exactHits = append(exactHits, f)
		case pat.MatchString(f.Name):
			looseHits = append(looseHits, f)
		}
	}
	return append(exactHits, looseHits...)
}

// success returns the first 2xx Golden File for method+path, if any.
func (g *goldenSet) success(method, path string) *goldenFile {
	hits := g.find(method, path, "v1", "v2", "app")
	if len(hits) == 0 {
		return nil
	}
	return hits[0]
}

// errorsFor returns the error Golden Files for method+path keyed by status.
func (g *goldenSet) errorsFor(method, path string) map[int]*goldenFile {
	out := map[int]*goldenFile{}
	for _, f := range g.find(method, path, "errors") {
		if f.Status == 0 {
			continue
		}
		if _, dup := out[f.Status]; !dup {
			out[f.Status] = f
		}
	}
	return out
}

// isNullBody reports whether the recorded body is the bare JSON null that the
// platform returns for some not-found lookups.
func (f *goldenFile) isNullBody() bool {
	return f != nil && f.Body == nil
}
