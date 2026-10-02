package gendocs

import (
	"html"
	"regexp"
	"strings"
)

var (
	reBreakTag = regexp.MustCompile(`(?i)<\s*(br|/p|/div|/li|/tr|/h[1-6]|/pre|/ol|/ul|/table)\s*/?>`)
	reListItem = regexp.MustCompile(`(?i)<\s*li[^>]*>`)
	reAnyTag   = regexp.MustCompile(`<[^>]+>`)
	reSpaces   = regexp.MustCompile(`[ \t\x{3000}]+`)
	reBlank    = regexp.MustCompile(`\n{2,}`)
)

// stripHTML turns the HTML fragments CYBERBIZ embeds in descriptions into
// plain text: <br> and block closers become line breaks, <li> becomes a
// bullet, every other tag is dropped, and entities are unescaped.
func stripHTML(s string) string {
	if !strings.ContainsAny(s, "<&") {
		return normalizeSpace(s)
	}
	s = reBreakTag.ReplaceAllString(s, "\n")
	s = reListItem.ReplaceAllString(s, "\n- ")
	s = reAnyTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return normalizeSpace(s)
}

// normalizeSpace trims each line, collapses runs of spaces and blank lines.
func normalizeSpace(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(reSpaces.ReplaceAllString(l, " "))
	}
	s = strings.Join(lines, "\n")
	s = reBlank.ReplaceAllString(s, "\n")
	return strings.TrimSpace(s)
}
