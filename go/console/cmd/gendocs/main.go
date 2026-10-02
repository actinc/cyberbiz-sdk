// Command gendocs regenerates docs/api/en from docs/references and
// testdata/golden. Run it from the console module:
//
//	cd console && go run ./cmd/gendocs
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/gendocs"
)

func main() {
	root := flag.String("root", "", "repository root (default: parent of the console module)")
	out := flag.String("out", "", "output root; en/ and zh-TW/ are written under it (default: <root>/docs/api)")
	untranslated := flag.String("untranslated", "", "write the descriptions that still contain Chinese to this file")
	flag.Parse()

	if *root == "" {
		wd, err := os.Getwd()
		if err != nil {
			fatal(err)
		}
		*root = findRoot(wd)
	}
	if *out == "" {
		*out = filepath.Join(*root, "docs", "api")
	}
	err := gendocs.Run(gendocs.Config{
		RefsDir:   filepath.Join(*root, "docs", "references"),
		GoldenDir: filepath.Join(*root, "testdata", "golden"),
		OutRoot:   *out,
		Stdout:    os.Stdout,

		UntranslatedFile: *untranslated,
	})
	if err != nil {
		fatal(err)
	}
}

// findRoot walks up from dir until it finds testdata/golden (the repo root).
func findRoot(dir string) string {
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "testdata", "golden")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			return dir
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "gendocs:", err)
	os.Exit(1)
}
