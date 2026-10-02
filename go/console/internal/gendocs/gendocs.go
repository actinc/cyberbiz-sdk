// Package gendocs generates the curated CYBERBIZ API documentation under
// docs/api/en/ from the raw reference material in docs/references/ and the
// Golden Files in testdata/golden/.
package gendocs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Config locates the inputs and the output directory.
type Config struct {
	RefsDir   string    // docs/references
	GoldenDir string    // testdata/golden
	OutRoot   string    // docs/api; one directory per locale is written under it
	Stdout    io.Writer // run report; nil for none
	// UntranslatedFile, when set, receives the descriptions that still
	// contain Chinese after the glossary pass ("<count>\t<text>" per line).
	UntranslatedFile string
}

// inputs is everything the generator reads.
type inputs struct {
	swagger   *swaggerDoc
	observed  *observedDoc
	golden    *goldenSet
	postmanV1 postmanIndex
	webhooks  *webhookSource
}

// Output file names under OutDir.
const (
	FileOpenAPIV1 = "cyberbiz-openapi-v1.yaml"
	FileOpenAPIV2 = "cyberbiz-openapi-v2.yaml"
	FilePostmanV1 = "cyberbiz-v1.postman_collection.json"
	FilePostmanV2 = "cyberbiz-v2.postman_collection.json"
	FileWebhooks  = "webhooks.md"
)

// Run generates every file. Nothing is written unless every document
// validates and passes the forbidden-value scan.
func Run(cfg Config) error {
	in, err := loadInputs(cfg)
	if err != nil {
		return err
	}
	log := newReport()
	outputs := map[string]map[string][]byte{}
	for _, loc := range locales {
		files, err := generate(in, loc, log)
		if err != nil {
			return fmt.Errorf("%s: %w", loc.Code, err)
		}
		outputs[loc.Dir] = files
	}
	for _, loc := range locales {
		dir := filepath.Join(cfg.OutRoot, loc.Dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for _, name := range outputFiles {
			if err := os.WriteFile(filepath.Join(dir, name), outputs[loc.Dir][name], 0o644); err != nil {
				return err
			}
		}
	}
	if cfg.UntranslatedFile != "" {
		if err := os.WriteFile(cfg.UntranslatedFile, []byte(log.untranslatedTSV()), 0o644); err != nil {
			return err
		}
	}
	if cfg.Stdout != nil {
		fmt.Fprint(cfg.Stdout, log.String())
		for _, loc := range locales {
			for _, name := range outputFiles {
				fmt.Fprintf(cfg.Stdout, "wrote %s (%d bytes)\n", filepath.Join(cfg.OutRoot, loc.Dir, name), len(outputs[loc.Dir][name]))
			}
		}
	}
	return nil
}

// outputFiles is the fixed set of files every locale directory contains.
var outputFiles = []string{FileOpenAPIV1, FileOpenAPIV2, FilePostmanV1, FilePostmanV2, FileWebhooks, FileWebhookPostman}

func loadInputs(cfg Config) (*inputs, error) {
	sw, err := loadSwagger(filepath.Join(cfg.RefsDir, "cyberbiz-v1-swagger.json"))
	if err != nil {
		return nil, fmt.Errorf("load swagger: %w", err)
	}
	obs, err := loadObserved(filepath.Join(cfg.RefsDir, "notes", "schema_observed.json"))
	if err != nil {
		return nil, fmt.Errorf("load observed schema: %w", err)
	}
	golden, err := loadGolden(cfg.GoldenDir)
	if err != nil {
		return nil, fmt.Errorf("load golden files: %w", err)
	}
	pm, err := loadPostman(filepath.Join(cfg.RefsDir, "cyberbiz-api-v1.postman_collection.json"))
	if err != nil {
		return nil, fmt.Errorf("load postman: %w", err)
	}
	wh, err := loadWebhookSource(filepath.Join(cfg.RefsDir, "cyberbiz_webhook.md"))
	if err != nil {
		return nil, fmt.Errorf("load webhook reference: %w", err)
	}
	return &inputs{swagger: sw, observed: obs, golden: golden, postmanV1: pm, webhooks: wh}, nil
}

// generate builds, validates and scans every output, returning them by name.
func generate(in *inputs, loc *locale, log *report) (map[string][]byte, error) {
	files := map[string][]byte{}
	v1, err := buildV1(in, loc, log)
	if err != nil {
		return nil, fmt.Errorf("v1: %w", err)
	}
	v2, err := buildV2(in, loc, log)
	if err != nil {
		return nil, fmt.Errorf("v2: %w", err)
	}
	for _, d := range []struct {
		doc          *Document
		yamlName, pm string
		pmTitle      string
	}{
		{v1, FileOpenAPIV1, FilePostmanV1, "CYBERBIZ API v1"},
		{v2, FileOpenAPIV2, FilePostmanV2, "CYBERBIZ API v2"},
	} {
		y, err := MarshalYAMLDocument(toTree(d.doc))
		if err != nil {
			return nil, fmt.Errorf("%s: marshal: %w", d.yamlName, err)
		}
		if err := validateOpenAPI(y); err != nil {
			return nil, fmt.Errorf("%s: %w", d.yamlName, err)
		}
		files[d.yamlName] = y
		pm, err := EncodeJSONIndent(buildPostman(d.doc, d.pmTitle, loc), "  ")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", d.pm, err)
		}
		files[d.pm] = pm
	}
	wh, err := buildWebhooksMarkdown(in.webhooks, in.golden, loc, log)
	if err != nil {
		return nil, fmt.Errorf("webhooks: %w", err)
	}
	files[FileWebhooks] = []byte(wh)
	whpm, err := buildWebhookPostman(in.webhooks, in.golden, loc, log)
	if err != nil {
		return nil, fmt.Errorf("webhook postman: %w", err)
	}
	if files[FileWebhookPostman], err = EncodeJSONIndent(whpm, "  "); err != nil {
		return nil, fmt.Errorf("webhook postman: %w", err)
	}
	for name, content := range files {
		if err := scanOutput(name, content); err != nil {
			return nil, err
		}
	}
	log.notef("%s: v1 %d operations, %d schemas; v2 %d operations, %d schemas",
		loc.Code, countOperations(v1), v1.Components.Schemas.Len(), countOperations(v2), v2.Components.Schemas.Len())
	return files, nil
}

func countOperations(doc *Document) int {
	n := 0
	for _, p := range doc.Paths.Keys() {
		item, _ := doc.Paths.Get(p)
		for _, m := range httpMethods {
			if item.(*PathItem).Operation(m) != nil {
				n++
			}
		}
	}
	return n
}
