package redact

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// forbidden lists what must never appear in a committed Golden File: real
// e-mail addresses, Taiwanese mobile numbers, JWTs and merchant hostnames.
// The names of the integrators whose shops recorded the files are checked
// too, but that list is private: see [Denylist].
var forbidden = []struct {
	name string
	re   *regexp.Regexp
}{
	{"e-mail outside example.com", regexp.MustCompile(`[A-Za-z0-9._%+-]+@(?:[A-Za-z0-9-]+\.)+[A-Za-z]{2,}`)},
	{"Taiwanese mobile number", regexp.MustCompile(`(?:\+?886-?|\b0)9\d{2}-?\d{3}-?\d{3}\b`)},
	{"JWT", regexp.MustCompile(`eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.`)},
	{"shop hostname", regexp.MustCompile(`\b[a-z0-9-]+\.cyberbiz\.(?:co|io)\b`)},
	{"URL query token", regexp.MustCompile(`(?i)[?&][a-z_]*token=[^&#"\s]+`)},
}

// allowed are the synthetic values the redaction rules write on purpose.
var allowed = map[string]bool{
	"redacted@example.com": true, "0912345678": true, "example.cyberbiz.co": true,
	"app-store-api.cyberbiz.io": true, "api.cyberbiz.co": true, "api-doc.cyberbiz.co": true,
	"www.cyberbiz.co": true, "app.cyberbiz.co": true, "eyJhbGciOiJIUzI1NiJ9.REDACTED.": true,
}

// TestGoldenFilesAreRedacted scans every committed Golden File and webhook
// sample so that a pull request cannot smuggle real data into the repo.
func TestGoldenFilesAreRedacted(t *testing.T) {
	names := integratorNames(t)
	dirs := []string{
		filepath.Join("..", "..", "..", "testdata", "golden"),
		filepath.Join("..", "..", "webhook", "testdata"),
	}
	files := 0
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			files++
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			scan(t, path, string(data), names)
			return nil
		})
		if err != nil {
			t.Fatalf("%s: %v", dir, err)
		}
	}
	if files == 0 {
		t.Fatal("no files scanned")
	}
	t.Logf("scanned %d files", files)
}

// synthetic reports values that are obviously made up: any example.com
// address, the 09000000xx phone block used by the webhook samples, a
// redacted or synthetic URL query token, and the redacted JWT placeholder.
func synthetic(m string) bool {
	return strings.HasSuffix(m, "@example.com") ||
		strings.HasSuffix(m, "token="+Placeholder) || strings.HasSuffix(m, "token=SYNTHETIC") ||
		strings.HasPrefix(m, "09000000") ||
		strings.HasPrefix(m, "eyJhbGciOiJIUzI1NiJ9.REDACTED.")
}

// integratorNames loads the private name list. CI must provide it; a local
// run without it skips only the name check.
func integratorNames(t *testing.T) *regexp.Regexp {
	t.Helper()
	re, err := Denylist()
	if err != nil {
		t.Fatal(err)
	}
	if re == nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("%s must be set in CI", DenylistEnv)
		}
		t.Logf("%s not set: integrator-name check skipped", DenylistEnv)
	}
	return re
}

func scan(t *testing.T, path, text string, names *regexp.Regexp) {
	t.Helper()
	if names != nil {
		for _, m := range names.FindAllString(text, -1) {
			t.Errorf("%s: integrator name: %q", path, m)
		}
	}
	for _, f := range forbidden {
		for _, m := range f.re.FindAllString(text, -1) {
			if allowed[m] || synthetic(m) {
				continue
			}
			t.Errorf("%s: %s: %q", path, f.name, m)
		}
	}
}
