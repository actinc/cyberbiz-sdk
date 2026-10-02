package gendocs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// repoRoot finds the repository root from the package directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for d := wd; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "testdata", "golden")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			t.Fatal("repository root not found")
		}
	}
}

// TestFullGeneration runs the whole pipeline on the real (gitignored)
// reference material. It is skipped where docs/references is absent, e.g.
// in CI; the mini-pipeline tests cover the rules there.
func TestFullGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("full generation is slow; run without -short")
	}
	root := repoRoot(t)
	refs := filepath.Join(root, "docs", "references")
	if _, err := os.Stat(filepath.Join(refs, "cyberbiz-v1-swagger.json")); err != nil {
		t.Skip("docs/references not available")
	}
	in, err := loadInputs(Config{RefsDir: refs, GoldenDir: filepath.Join(root, "testdata", "golden")})
	if err != nil {
		t.Fatal(err)
	}
	log := newReport()
	files, err := generate(in, localeEN, log)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{FileOpenAPIV1, FileOpenAPIV2, FilePostmanV1, FilePostmanV2, FileWebhooks} {
		if len(files[name]) == 0 {
			t.Errorf("%s is empty", name)
		}
	}
	for _, r := range correctionsTable {
		if r.ID == "not-found-null" && log.counts[r.ID] == 0 {
			continue // depends on which Golden Files were recorded
		}
		if log.counts[r.ID] == 0 {
			t.Errorf("correction %s was never applied", r.ID)
		}
	}
	if len(log.warnings) > 0 {
		t.Errorf("unexpected warnings: %v", log.warnings)
	}
	v1 := string(files[FileOpenAPIV1])
	for _, want := range []string{"bearerAuth", "X-Total-Pages", "Timestamp:", "Money:", "gift_order_status", "x-cyberbiz-observed: true"} {
		if !strings.Contains(v1, want) {
			t.Errorf("v1 lacks %s", want)
		}
	}
	if strings.Contains(v1, "api.cyberbiz.co") || strings.Contains(v1, "hmac username") {
		t.Error("legacy auth leaked into v1")
	}
	checkWebhookCollection(t, files[FileWebhooks], files[FileWebhookPostman])
	// Regenerating must be byte-identical.
	again, err := generate(in, localeEN, newReport())
	if err != nil {
		t.Fatal(err)
	}
	for name := range files {
		if string(files[name]) != string(again[name]) {
			t.Errorf("%s is not deterministic", name)
		}
	}
}

// checkWebhookCollection asserts the receiver-testing collection has exactly
// one request per webhooks.md event whose body equals the documented payload.
func checkWebhookCollection(t *testing.T, md, collection []byte) {
	t.Helper()
	events := reEventHeading.FindAllStringSubmatch(string(md), -1)
	blocks := reJSONBlock.FindAllStringSubmatch(string(md), -1)
	if len(events) == 0 || len(events) != len(blocks) {
		t.Fatalf("webhooks.md: %d events but %d JSON blocks", len(events), len(blocks))
	}
	want := map[string]any{}
	for i, ev := range events {
		v, err := DecodeJSON([]byte(blocks[i][1]))
		if err != nil {
			t.Fatalf("webhooks.md payload of %s: %v", ev[1], err)
		}
		want[ev[1]] = v
	}
	var col struct {
		Item []struct {
			Item []struct {
				Name    string
				Request struct {
					Header []struct{ Key, Value string }
					Body   struct{ Raw string }
				}
			}
		}
	}
	if err := json.Unmarshal(collection, &col); err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, f := range col.Item {
		for _, it := range f.Item {
			seen[it.Name]++
			body, err := DecodeJSON([]byte(it.Request.Body.Raw))
			if err != nil {
				t.Errorf("%s: body is not JSON: %v", it.Name, err)
				continue
			}
			if !reflect.DeepEqual(body, want[it.Name]) {
				t.Errorf("%s: collection body differs from webhooks.md payload", it.Name)
			}
			if strings.Contains(it.Request.Body.Raw, "{{") {
				t.Errorf("%s: body must not contain variables", it.Name)
			}
			var event string
			for _, h := range it.Request.Header {
				if h.Key == "X-Cyberbiz-Event" {
					event = h.Value
				}
			}
			if event != it.Name {
				t.Errorf("%s: X-Cyberbiz-Event header is %q", it.Name, event)
			}
		}
	}
	for ev := range want {
		if seen[ev] != 1 {
			t.Errorf("event %s has %d requests, want 1", ev, seen[ev])
		}
	}
	if len(seen) != len(want) {
		t.Errorf("collection has %d events, webhooks.md has %d", len(seen), len(want))
	}
}
