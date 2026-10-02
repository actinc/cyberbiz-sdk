package cyberbiz

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"testing"
	"time"
)

func TestParseTimeFormats(t *testing.T) {
	want := time.Date(2026, 7, 10, 20, 22, 5, 0, Taipei)
	cases := []string{
		"2026-07-10 20:22:05",
		"2026-07-10T20:22:05+08:00",
		"2026-07-10T12:22:05Z",
		"2026-07-10 20:22:05 +0800",
		"2026-07-10T20:22:05",
	}
	for _, s := range cases {
		got, err := ParseTime(s)
		if err != nil {
			t.Errorf("%q: %v", s, err)
			continue
		}
		if !got.Equal(want) {
			t.Errorf("%q: got %v, want %v", s, got, want)
		}
		if got.Location() != Taipei {
			t.Errorf("%q: location %v, want Taipei", s, got.Location())
		}
	}
	if got, err := ParseTime(""); err != nil || !got.IsZero() {
		t.Errorf("empty: %v %v", got, err)
	}
	if _, err := ParseTime("yesterday"); err == nil {
		t.Error("garbage accepted")
	}
}

func TestTimeJSONRoundTrip(t *testing.T) {
	type doc struct {
		CreatedAt Time  `json:"created_at"`
		ClosedAt  *Time `json:"closed_at"`
		UpdatedAt Time  `json:"updated_at,omitzero"`
	}
	var d doc
	err := json.Unmarshal([]byte(`{"created_at":"2026-07-10 20:22:05","closed_at":null,"updated_at":null}`), &d)
	if err != nil {
		t.Fatal(err)
	}
	if d.CreatedAt.String() != "2026-07-10 20:22:05" {
		t.Errorf("created_at = %q", d.CreatedAt.String())
	}
	if d.ClosedAt != nil {
		t.Errorf("closed_at = %v, want nil", d.ClosedAt)
	}
	if !d.UpdatedAt.IsZero() {
		t.Errorf("updated_at = %v, want zero", d.UpdatedAt)
	}
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"created_at":"2026-07-10 20:22:05","closed_at":null}` {
		t.Errorf("marshal = %s", out)
	}
}

func TestTimeWorksWithEncodingJSONV1(t *testing.T) {
	var tm Time
	if err := jsonv1.Unmarshal([]byte(`"2026-07-10 20:22:05"`), &tm); err != nil {
		t.Fatal(err)
	}
	out, err := jsonv1.Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `"2026-07-10 20:22:05"` {
		t.Errorf("v1 marshal = %s", out)
	}
	if err := jsonv1.Unmarshal([]byte(`null`), &tm); err != nil || !tm.IsZero() {
		t.Errorf("v1 null: %v %v", tm, err)
	}
}

func TestDateJSON(t *testing.T) {
	type doc struct {
		Birthday Date `json:"birthday"`
		Deadline Date `json:"deadline"`
	}
	var d doc
	if err := json.Unmarshal([]byte(`{"birthday":"1990-05-01","deadline":"2026-07-10 20:22:05"}`), &d); err != nil {
		t.Fatal(err)
	}
	if d.Birthday.String() != "1990-05-01" || d.Deadline.String() != "2026-07-10" {
		t.Errorf("got %v %v", d.Birthday, d.Deadline)
	}
	out, _ := json.Marshal(d)
	if string(out) != `{"birthday":"1990-05-01","deadline":"2026-07-10"}` {
		t.Errorf("marshal = %s", out)
	}
	var zero doc
	if err := json.Unmarshal([]byte(`{"birthday":null,"deadline":""}`), &zero); err != nil {
		t.Fatal(err)
	}
	if !zero.Birthday.IsZero() || !zero.Deadline.IsZero() {
		t.Errorf("null/empty should be zero: %+v", zero)
	}
}
