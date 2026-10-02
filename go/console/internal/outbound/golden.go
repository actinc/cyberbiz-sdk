package outbound

import (
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

var (
	numericSeg = regexp.MustCompile(`^\d+$`)
	versionSeg = regexp.MustCompile(`^v\d+$`)
)

// GoldenName returns "<group>/<METHOD>_<path>_<suffix>.json" in the layout
// internal/tools/goldenimport and the SDK's golden tests use: numeric path
// segments become {id}; the group is the version prefix or "app".
func GoldenName(method, path, suffix string) string {
	segs := strings.Split(strings.Trim(path, "/"), "/")
	group := "app"
	if len(segs) > 0 && versionSeg.MatchString(segs[0]) {
		group = segs[0]
	}
	for i, s := range segs {
		if numericSeg.MatchString(s) {
			segs[i] = "{id}"
		}
	}
	if suffix != "" && !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}
	return group + "/" + strings.ToUpper(method) + "_" + strings.Join(segs, "_") + suffix + ".json"
}

// ExportGolden writes the redacted response body and headers of row under
// dir and returns the body path. An existing file is refused unless
// overwrite is set.
func ExportGolden(dir string, row *db.OutboundLog, suffix string, overwrite bool) (string, error) {
	name := GoldenName(row.Method, row.Path, suffix)
	dst := filepath.Join(dir, filepath.FromSlash(name))
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return "", response.Conflict("golden file already exists: " + name + " (set overwrite to replace it)")
		}
	}
	body, err := DecodeBody(row.ResponseBody, ResponseContentType(string(row.ResponseHeaders)))
	if err != nil {
		return "", err
	}
	clean, err := cyberbiz.RedactJSON(body)
	if err != nil {
		return "", err
	}
	headers, err := RedactedHeaders(string(row.ResponseHeaders))
	if err != nil {
		return "", err
	}
	if err := WriteGolden(dst, clean, headers); err != nil {
		return "", err
	}
	return dst, nil
}

// RedactedHeaders parses a stored headers JSON and redacts it.
func RedactedHeaders(stored string) (http.Header, error) {
	var h http.Header
	if stored == "" {
		return http.Header{}, nil
	}
	if err := json.Unmarshal([]byte(stored), &h); err != nil {
		return nil, errors.New("stored headers are not valid JSON")
	}
	return cyberbiz.RedactHeaders(h), nil
}

// WriteGolden writes body to dst and headers to the matching .headers.json.
func WriteGolden(dst string, body []byte, headers http.Header) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, append(body, '\n'), 0o644); err != nil {
		return err
	}
	encoded, err := jsonv2.Marshal(headers, jsonv2.Deterministic(true))
	if err != nil {
		return err
	}
	return os.WriteFile(strings.TrimSuffix(dst, ".json")+".headers.json", append(encoded, '\n'), 0o644)
}
