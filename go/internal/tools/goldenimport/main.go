// Command goldenimport turns raw API responses recorded by a sweep or by the
// Console into redacted Golden Files under testdata/golden.
//
// Input layout (what the Console and the sweep script both write):
//
//	<in>/<group>/<NAME>.json          response body
//	<in>/<group>/<NAME>.headers.txt   "Key: Value" lines, optional
//
// Output layout:
//
//	testdata/golden/<group>/<NAME>.json           redacted body
//	testdata/golden/<group>/<NAME>.headers.json   redacted headers as JSON
//
// Usage: go run ./internal/tools/goldenimport -in <dir> -out ../testdata/golden
package main

import (
	"encoding/json/v2"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/internal/redact"
)

func main() {
	in := flag.String("in", "", "directory holding <group>/<NAME>.json files")
	out := flag.String("out", "../testdata/golden", "destination directory")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "goldenimport: -in is required")
		os.Exit(2)
	}
	n, err := run(*in, *out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "goldenimport:", err)
		os.Exit(1)
	}
	fmt.Printf("goldenimport: wrote %d golden files to %s\n", n, *out)
}

func run(in, out string) (int, error) {
	count := 0
	err := filepath.WalkDir(in, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".headers.json") {
			return err
		}
		rel, err := filepath.Rel(in, path)
		if err != nil {
			return err
		}
		if err := importOne(path, filepath.Join(out, rel)); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		count++
		return nil
	})
	return count, err
}

func importOne(src, dst string) error {
	body, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	clean, err := redact.JSON(body)
	if err != nil {
		return fmt.Errorf("redacting body: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, append(clean, '\n'), 0o644); err != nil {
		return err
	}
	headers, err := readHeaders(strings.TrimSuffix(src, ".json") + ".headers.txt")
	if err != nil {
		return err
	}
	if headers == nil {
		// Re-importing an existing Golden File directory: re-redact the
		// stored headers in place so new rules apply to them too.
		headers, err = readHeadersJSON(strings.TrimSuffix(src, ".json") + ".headers.json")
		if err != nil || headers == nil {
			return err
		}
	}
	encoded, err := json.Marshal(redact.Headers(headers), json.Deterministic(true))
	if err != nil {
		return err
	}
	return os.WriteFile(strings.TrimSuffix(dst, ".json")+".headers.json", append(encoded, '\n'), 0o644)
}

// readHeadersJSON reads a previously written .headers.json; a missing file
// yields nil, nil.
func readHeadersJSON(path string) (http.Header, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var h http.Header
	if err := json.Unmarshal(raw, &h); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return h, nil
}

// readHeaders parses "Key: Value" lines; a missing file yields nil, nil.
func readHeaders(path string) (http.Header, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	h := http.Header{}
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "HTTP/") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		h.Add(strings.TrimSpace(k), strings.TrimSpace(v))
	}
	return h, nil
}
