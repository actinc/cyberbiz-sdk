package gendocs

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// whField is one row of a payload field table in cyberbiz_webhook.md.
type whField struct {
	Name string
	Type string // Integer, String, Float, Boolean, Object, [Object], Hash, [String], Date, ...
	Desc string // zh-TW description as written
}

// whEvent is one row of an event table.
type whEvent struct {
	Code string
	Desc string // zh-TW
}

// webhookSource is the parsed reference: field tables keyed by the English
// name in the section heading ("Order", "Line Item", ...) and the events.
type webhookSource struct {
	Sections map[string][]whField
	Order    []string // section names in document order
	Events   []whEvent
}

var (
	reHeading  = regexp.MustCompile(`^#{2,4}\s+(.*)$`)
	reParen    = regexp.MustCompile(`\(([^)]*)\)\s*$`)
	reTableRow = regexp.MustCompile(`^\|(.*)\|$`)
	reCode     = regexp.MustCompile("`([^`]*)`")
)

// loadWebhookSource parses docs/references/cyberbiz_webhook.md. A missing
// file yields an empty source so the generator can still run in tests.
func loadWebhookSource(path string) (*webhookSource, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &webhookSource{Sections: map[string][]whField{}}, nil
		}
		return nil, err
	}
	defer f.Close()
	src := &webhookSource{Sections: map[string][]whField{}}
	section := ""
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if m := reHeading.FindStringSubmatch(line); m != nil {
			section = sectionName(m[1])
			continue
		}
		m := reTableRow.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		cells := splitCells(m[1])
		if len(cells) < 2 || strings.HasPrefix(cells[0], "---") || cells[0] == "屬性" || cells[0] == "事件代碼" || cells[0] == "Header 名稱" || cells[0] == "代碼" {
			continue
		}
		switch len(cells) {
		case 2:
			if strings.Contains(cells[0], "/") {
				src.Events = append(src.Events, whEvent{Code: cells[0], Desc: cells[1]})
			}
		default:
			if section == "" || section == "HTTP Headers" {
				continue
			}
			if _, seen := src.Sections[section]; !seen {
				src.Order = append(src.Order, section)
			}
			src.Sections[section] = append(src.Sections[section], whField{Name: cells[0], Type: cells[1], Desc: cells[2]})
		}
	}
	return src, sc.Err()
}

// sectionName extracts the English name from a heading such as
// "3. 訂單 (Order)" or "VIP 階層 / 當前 VIP 層級 / 下一個 VIP 層級".
func sectionName(h string) string {
	h = strings.TrimSpace(h)
	if m := reParen.FindStringSubmatch(h); m != nil {
		return strings.TrimSpace(m[1])
	}
	switch {
	case strings.Contains(h, "VIP 階層 / 當前"):
		return "VIP Level"
	case strings.Contains(h, "細部設定"):
		return "Gift Setting"
	case strings.Contains(h, "restrict_campaigns"):
		return "restrict_campaigns"
	}
	return h
}

// splitCells splits a markdown table row and unwraps `code` spans.
func splitCells(row string) []string {
	parts := strings.Split(row, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if m := reCode.FindStringSubmatch(p); m != nil && strings.HasPrefix(p, "`") && strings.HasSuffix(p, "`") && strings.Count(p, "`") == 2 {
			p = m[1]
		}
		out = append(out, p)
	}
	return out
}

// fields returns the table of a section, if any.
func (s *webhookSource) fields(section string) []whField {
	return s.Sections[section]
}
