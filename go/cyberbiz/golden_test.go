package cyberbiz

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goldenPath returns the path of a Golden File such as
// "v1/GET_v1_orders_{id}.json".
func goldenPath(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join("..", "..", "testdata", "golden", filepath.FromSlash(name))
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("golden file %s: %v", name, err)
	}
	return p
}

// readGolden returns the body of a Golden File.
func readGolden(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(goldenPath(t, name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// goldenServer serves Golden Files by request path: GET /v1/orders/123 is
// answered with v1/GET_v1_orders_{id}.json. Numeric path segments are
// replaced with {id}; a "suffix" query parameter is appended to the name so
// tests can select variants such as "_include". Unknown paths return 404.
func goldenServer(t *testing.T) *Client {
	t.Helper()
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		name := goldenNameFor(r)
		p := filepath.Join("..", "..", "testdata", "golden", filepath.FromSlash(name))
		data, err := os.ReadFile(p)
		if err != nil {
			http.Error(w, `{"error":["無此資源"]}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Total", "1")
		w.Header().Set("X-Total-Pages", "1")
		w.Header().Set("X-Page", "1")
		_, _ = w.Write(data)
	})
	return c
}

func goldenNameFor(r *http.Request) string {
	segs := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	group := "app"
	if len(segs) > 0 && strings.HasPrefix(segs[0], "v") {
		group = segs[0]
	}
	for i, s := range segs {
		if s != "" && strings.Trim(s, "0123456789") == "" {
			segs[i] = "{id}"
		}
	}
	name := r.Method + "_" + strings.Join(segs, "_") + r.URL.Query().Get("suffix")
	return group + "/" + name + ".json"
}

// decodeGolden decodes a Golden File through the client's lenient decoder
// exactly as a live response would be.
func decodeGolden(t *testing.T, name string, out any) {
	t.Helper()
	c, _ := New("tok")
	if err := c.decode(readGolden(t, name), out); err != nil {
		t.Fatalf("decoding %s: %v", name, err)
	}
}

var testCtx = context.Background()
