package gendocs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
)

// validateOpenAPI parses YAML with libopenapi, builds the v3 model (which
// resolves every $ref) and validates the document against the OpenAPI 3.1
// meta-schema. Any problem fails the run: invalid output must never be
// written to docs/api/.
func validateOpenAPI(yamlBytes []byte) error {
	doc, err := libopenapi.NewDocument(yamlBytes)
	if err != nil {
		return fmt.Errorf("openapi: parse: %w", err)
	}
	if _, err := doc.BuildV3Model(); err != nil {
		return fmt.Errorf("openapi: build model: %w", err)
	}
	v, errs := validator.NewValidator(doc)
	if len(errs) > 0 {
		return fmt.Errorf("openapi: validator: %w", errors.Join(errs...))
	}
	ok, verrs := v.ValidateDocument()
	if ok {
		return nil
	}
	var b strings.Builder
	for i, ve := range verrs {
		if i >= 20 {
			fmt.Fprintf(&b, "... and %d more", len(verrs)-i)
			break
		}
		fmt.Fprintf(&b, "%s: %s\n", ve.Message, ve.Reason)
		for j, sve := range ve.SchemaValidationErrors {
			if j >= 10 {
				break
			}
			fmt.Fprintf(&b, "  %s\n", sve.Reason)
		}
	}
	return fmt.Errorf("openapi: document invalid:\n%s", b.String())
}
