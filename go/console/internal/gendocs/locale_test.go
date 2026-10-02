package gendocs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestZhTWCoversAuthoredStrings(t *testing.T) {
	if missing := missingLocaleStrings(localeZhTW); len(missing) > 0 {
		t.Errorf("%d authored strings have no zh-TW entry, e.g.:\n  %s", len(missing), strings.Join(missing[:min(5, len(missing))], "\n  "))
	}
	for en, zh := range zhTWStrings {
		if strings.TrimSpace(zh) == "" {
			t.Errorf("empty zh-TW entry for %q", en)
		}
	}
	if got := localeZhTW.T("Created at."); got != "建立時間。" {
		t.Errorf("T: %q", got)
	}
	if got := localeZhTW.T("透過訂單 ID 取得該訂單資料" + nullNote); !strings.HasSuffix(got, "資源不存在時回應 `null`。") || !strings.HasPrefix(got, "透過訂單") {
		t.Errorf("fragment: %q", got)
	}
	if got := localeEN.Localize("訂單ID"); got != "Order ID." {
		t.Errorf("en Localize: %q", got)
	}
	if got := localeZhTW.Localize("訂單ID"); got != "訂單ID" {
		t.Errorf("zh-TW must keep source text: %q", got)
	}
}

// TestLocalesShareStructure proves docs/api/en and docs/api/zh-TW differ
// only in free text: same paths, operations, parameters, schemas, samples,
// status codes, Postman requests and webhook events / headers / payloads.
func TestLocalesShareStructure(t *testing.T) {
	root := repoRoot(t)
	enDir := filepath.Join(root, "docs", "api", "en")
	zhDir := filepath.Join(root, "docs", "api", "zh-TW")
	if _, err := os.Stat(filepath.Join(zhDir, FileOpenAPIV1)); err != nil {
		t.Skip("docs/api/zh-TW not generated")
	}
	for _, name := range []string{FileOpenAPIV1, FileOpenAPIV2} {
		en, zh := loadYAMLTree(t, filepath.Join(enDir, name)), loadYAMLTree(t, filepath.Join(zhDir, name))
		compareTrees(t, name, "$", en, zh, false)
	}
	for _, name := range []string{FilePostmanV1, FilePostmanV2, FileWebhookPostman} {
		en, zh := postmanShape(t, filepath.Join(enDir, name)), postmanShape(t, filepath.Join(zhDir, name))
		if !reflect.DeepEqual(en, zh) {
			t.Errorf("%s: request/response structure differs between locales", name)
		}
	}
	en, zh := webhookShape(t, filepath.Join(enDir, FileWebhooks)), webhookShape(t, filepath.Join(zhDir, FileWebhooks))
	if !reflect.DeepEqual(en, zh) {
		t.Errorf("%s: events, headers or payloads differ between locales", FileWebhooks)
	}
}

func loadYAMLTree(t *testing.T, path string) any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return v
}

// freeTextKeys are the only keys whose string values may differ.
var freeTextKeys = map[string]bool{"description": true, "summary": true, "title": true}

// strictKeys start subtrees (samples, defaults, enums) where even free-text
// keys must be identical. Every other key is compared strictly anyway; only
// freeTextKeys outside these subtrees may differ. A property that happens to
// be named "value" lives under "properties" and is not a strict subtree.
var strictKeys = map[string]bool{"example": true, "value": true, "default": true, "enum": true}

func compareTrees(t *testing.T, file, path string, a, b any, strict bool) {
	compareTreesIn(t, file, path, "", a, b, strict)
}

func compareTreesIn(t *testing.T, file, path, parent string, a, b any, strict bool) {
	t.Helper()
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok {
			t.Errorf("%s %s: map vs %T", file, path, b)
			return
		}
		ak, bk := sortedMapKeys(av), sortedMapKeys(bv)
		if !reflect.DeepEqual(ak, bk) {
			t.Errorf("%s %s: keys differ\n en: %v\n zh: %v", file, path, ak, bk)
			return
		}
		for _, k := range ak {
			_, as := av[k].(string)
			_, bs := bv[k].(string)
			if !strict && freeTextKeys[k] && as && bs {
				continue
			}
			compareTreesIn(t, file, path+"."+k, k, av[k], bv[k], strict || (strictKeys[k] && parent != "properties"))
		}
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			t.Errorf("%s %s: list length differs", file, path)
			return
		}
		for i := range av {
			compareTreesIn(t, file, fmt.Sprintf("%s[%d]", path, i), parent, av[i], bv[i], strict)
		}
	default:
		if !reflect.DeepEqual(a, b) {
			t.Errorf("%s %s: %v != %v", file, path, a, b)
		}
	}
}

func sortedMapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// postmanShape reduces a collection to its structural facts.
func postmanShape(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var col struct {
		Info     struct{ Version string }
		Variable []struct{ Key, Value string }
		Event    []struct {
			Listen string
			Script struct{ Exec []string }
		}
		Item []struct {
			Name string
			Item []struct {
				Name    string
				Request struct {
					Method string
					URL    struct{ Raw string }
					Header []struct{ Key, Value string }
					Body   *struct{ Raw string }
				}
				Response []struct {
					Code int
					Body string
				}
			}
		}
	}
	if err := json.Unmarshal(b, &col); err != nil {
		t.Fatal(err)
	}
	out := []string{"version=" + col.Info.Version}
	for _, v := range col.Variable {
		out = append(out, "var:"+v.Key+"="+v.Value)
	}
	for _, e := range col.Event {
		out = append(out, "script:"+e.Listen+":"+strings.Join(e.Script.Exec, "\n"))
	}
	for _, f := range col.Item {
		for _, it := range f.Item {
			line := f.Name + "|" + it.Request.Method + " " + it.Request.URL.Raw
			for _, h := range it.Request.Header {
				line += "|" + h.Key + ": " + h.Value
			}
			if it.Request.Body != nil {
				line += "|body=" + it.Request.Body.Raw
			}
			for _, r := range it.Response {
				line += fmt.Sprintf("|%d:%s", r.Code, r.Body)
			}
			out = append(out, line)
		}
	}
	return out
}

var (
	reEventHeading = regexp.MustCompile("(?m)^### `([^`]+)`$")
	reHTTPBlock    = regexp.MustCompile("(?s)```http\n(.*?)```")
	reJSONBlock    = regexp.MustCompile("(?s)```json\n(.*?)```")
	reHeaderName   = regexp.MustCompile(`(?m)^([A-Za-z-]+): `)
	reVersionLine  = regexp.MustCompile(`(?m)^Version: .*$`)
)

// webhookShape reduces webhooks.md to its events, header names and payloads.
func webhookShape(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	md := string(b)
	out := []string{reVersionLine.FindString(md)}
	for _, m := range reEventHeading.FindAllStringSubmatch(md, -1) {
		out = append(out, "event:"+m[1])
	}
	for _, blk := range reHTTPBlock.FindAllStringSubmatch(md, -1) {
		names := []string{}
		for _, h := range reHeaderName.FindAllStringSubmatch(blk[1], -1) {
			names = append(names, h[1])
		}
		out = append(out, "headers:"+strings.Join(names, ","))
	}
	for _, blk := range reJSONBlock.FindAllStringSubmatch(md, -1) {
		out = append(out, "json:"+blk[1])
	}
	return out
}
