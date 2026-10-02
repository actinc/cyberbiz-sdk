package gendocs

import (
	"fmt"
	"regexp"
	"strings"
)

// Patterns that must never appear in generated docs: real e-mail addresses,
// real Taiwanese mobile numbers, JWT fragments, redaction markers and the
// hostnames of the shops the Golden Files were recorded from. Integrator and
// merchant names are checked as well, from the private list in
// redact.DenylistEnv, so the names themselves never enter the source.
var forbiddenPatterns = []struct {
	Name string
	Re   *regexp.Regexp
	OK   func(match string) bool
}{
	{"email", regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`), func(m string) bool {
		return strings.HasSuffix(strings.ToLower(m), "@example.com")
	}},
	{"tw-mobile", regexp.MustCompile(`\b09\d{8}\b`), func(m string) bool { return m == synthMobile }},
	{"jwt", regexp.MustCompile(`eyJ[A-Za-z0-9_-]{8,}`), nil},
	{"redacted", regexp.MustCompile(`(?i)<?redacted>?`), nil},
	{"shop-host", reShopHost, platformHost},
	{"confirmation-token", regexp.MustCompile(`confirmation_token=[A-Za-z0-9]+`), func(m string) bool { return m == "confirmation_token=SYNTHETIC" }},
}

// scanOutput returns an error listing every forbidden match in content.
func scanOutput(name string, content []byte) error {
	var problems []string
	names, err := denylist()
	if err != nil {
		return err
	}
	if names != nil {
		for _, m := range names.FindAll(content, -1) {
			problems = append(problems, fmt.Sprintf("integrator-name: %s", m))
		}
	}
	for _, p := range forbiddenPatterns {
		if p.Re == nil {
			continue
		}
		for _, m := range p.Re.FindAll(content, -1) {
			if p.OK != nil && p.OK(string(m)) {
				continue
			}
			problems = append(problems, fmt.Sprintf("%s: %s", p.Name, string(m)))
			if len(problems) > 20 {
				break
			}
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s contains forbidden values:\n  %s", name, strings.Join(problems, "\n  "))
}
