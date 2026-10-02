package gendocs

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildPostman(t *testing.T) {
	in := miniInputs(t)
	doc, err := buildV1(in, localeEN, newReport())
	if err != nil {
		t.Fatal(err)
	}
	col := buildPostman(doc, "CYBERBIZ API v1", localeEN)
	raw, err := EncodeJSONIndent(col, "  ")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Info struct {
			Version     string `json:"version"`
			Description string `json:"description"`
			ID          string `json:"_postman_id"`
		} `json:"info"`
		Auth struct {
			Type   string                   `json:"type"`
			Bearer []struct{ Value string } `json:"bearer"`
		} `json:"auth"`
		Item []struct {
			Name string `json:"name"`
			Item []struct {
				Name    string `json:"name"`
				Request struct {
					Method string `json:"method"`
					URL    struct {
						Raw   string                        `json:"raw"`
						Query []struct{ Key, Value string } `json:"query"`
					} `json:"url"`
					Body *struct {
						Mode string `json:"mode"`
						Raw  string `json:"raw"`
					} `json:"body"`
				} `json:"request"`
				Response []struct {
					Code int    `json:"code"`
					Body string `json:"body"`
				} `json:"response"`
			} `json:"item"`
		} `json:"item"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Info.Version != DocVersion || !strings.Contains(parsed.Info.Description, "## Release Notes") {
		t.Error("collection info must carry version and release notes")
	}
	if parsed.Auth.Type != "bearer" || parsed.Auth.Bearer[0].Value != "{{CyberbizToken}}" {
		t.Error("collection must use bearer auth with {{CyberbizToken}}")
	}
	if len(parsed.Item) != 1 || parsed.Item[0].Name != "orders" || len(parsed.Item[0].Item) != 3 {
		t.Fatalf("unexpected folder layout: %+v", parsed.Item)
	}
	var sawBody, sawList bool
	for _, it := range parsed.Item[0].Item {
		if len(it.Response) == 0 {
			t.Errorf("%s has no saved example response", it.Name)
		}
		if it.Request.Method == "PUT" {
			sawBody = it.Request.Body != nil && it.Request.Body.Mode == "raw" && strings.Contains(it.Request.Body.Raw, "note")
			if !strings.Contains(it.Request.URL.Raw, ":order_id") {
				t.Error("path parameters must use the :name form")
			}
		}
		if it.Request.Method == "GET" {
			sawList = strings.Contains(it.Request.URL.Raw, "page=1") && it.Response[0].Code == 200 && strings.HasPrefix(it.Response[0].Body, "[")
		}
	}
	if !sawBody || !sawList {
		t.Errorf("body example or list example missing (body=%v list=%v)", sawBody, sawList)
	}
	if stableUUID("a") != stableUUID("a") || stableUUID("a") == stableUUID("b") || len(stableUUID("a")) != 36 {
		t.Error("stableUUID must be deterministic and UUID-shaped")
	}
	if raw2, _ := EncodeJSONIndent(buildPostman(doc, "CYBERBIZ API v1", localeEN), "  "); string(raw) != string(raw2) {
		t.Error("postman output must be deterministic")
	}
}
