package gendocs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// swaggerDoc is the subset of a Swagger 2.0 document the generator reads.
type swaggerDoc struct {
	Info struct {
		Title       string `json:"title"`
		Version     string `json:"version"`
		Description string `json:"description"`
	} `json:"info"`
	Tags        []swTag                            `json:"tags"`
	Paths       map[string]map[string]*swOperation `json:"paths"`
	Definitions map[string]*swSchema               `json:"definitions"`
}

type swTag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type swOperation struct {
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	OperationID string                 `json:"operationId"`
	Tags        []string               `json:"tags"`
	Parameters  []*swParameter         `json:"parameters"`
	Responses   map[string]*swResponse `json:"responses"`
	Consumes    []string               `json:"consumes"`
}

type swParameter struct {
	In          string    `json:"in"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Format      string    `json:"format"`
	Required    bool      `json:"required"`
	Schema      *swSchema `json:"schema"`
	Items       *swSchema `json:"items"`
	Enum        []any     `json:"enum"`
	Default     any       `json:"default"`
}

type swResponse struct {
	Description string    `json:"description"`
	Schema      *swSchema `json:"schema"`
}

// swSchema is a Swagger 2.0 schema; Properties keeps source order.
type swSchema struct {
	Ref         string    `json:"$ref"`
	Type        string    `json:"type"`
	Format      string    `json:"format"`
	Description string    `json:"description"`
	Properties  *swProps  `json:"properties"`
	Items       *swSchema `json:"items"`
	Enum        []any     `json:"enum"`
	Required    []string  `json:"required"`
	Default     any       `json:"default"`
}

// swProps is an ordered property map.
type swProps struct {
	keys []string
	m    map[string]*swSchema
}

func (p *swProps) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("swagger: properties is not an object")
	}
	p.m = map[string]*swSchema{}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return err
		}
		k := kt.(string)
		var s swSchema
		if err := dec.Decode(&s); err != nil {
			return fmt.Errorf("swagger: property %q: %w", k, err)
		}
		if _, dup := p.m[k]; !dup {
			p.keys = append(p.keys, k)
		}
		p.m[k] = &s
	}
	return nil
}

// Keys returns property names in source order.
func (p *swProps) Keys() []string {
	if p == nil {
		return nil
	}
	return p.keys
}

// Get returns one property.
func (p *swProps) Get(k string) *swSchema {
	if p == nil {
		return nil
	}
	return p.m[k]
}

// loadSwagger reads the CYBERBIZ v1 Swagger 2.0 file.
func loadSwagger(path string) (*swaggerDoc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc swaggerDoc
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("swagger %s: %w", path, err)
	}
	return &doc, nil
}

// sortedPaths returns the swagger paths in a stable order.
func (d *swaggerDoc) sortedPaths() []string {
	out := make([]string, 0, len(d.Paths))
	for p := range d.Paths {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// sortedDefinitions returns definition names in a stable order.
func (d *swaggerDoc) sortedDefinitions() []string {
	out := make([]string, 0, len(d.Definitions))
	for n := range d.Definitions {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// operationCount returns the number of operations across all paths.
func (d *swaggerDoc) operationCount() int {
	n := 0
	for _, ops := range d.Paths {
		n += len(ops)
	}
	return n
}
