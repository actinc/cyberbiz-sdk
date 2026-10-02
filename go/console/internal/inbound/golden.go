package inbound

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/outbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

// GoldenName returns "<event with / replaced by _><suffix>.json".
func GoldenName(event, suffix string) string {
	name := strings.ReplaceAll(strings.Trim(event, "/"), "/", "_")
	if name == "" {
		name = "unknown_event"
	}
	if suffix != "" && !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}
	return name + suffix + ".json"
}

// ExportGolden writes the redacted body and headers of row under dir.
func ExportGolden(dir string, row *db.InboundLog, suffix string, overwrite bool) (string, error) {
	name := GoldenName(row.Event, suffix)
	dst := filepath.Join(dir, name)
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return "", response.Conflict("golden file already exists: " + name + " (set overwrite to replace it)")
		}
	}
	clean, err := cyberbiz.RedactJSON([]byte(row.Body))
	if err != nil {
		return "", err
	}
	headers, err := outbound.RedactedHeaders(string(row.Headers))
	if err != nil {
		return "", err
	}
	if err := outbound.WriteGolden(dst, clean, headers); err != nil {
		return "", err
	}
	return dst, nil
}
