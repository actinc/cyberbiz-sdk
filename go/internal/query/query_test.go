package query

import (
	"testing"
	"time"
)

type listOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

type stamp struct{ time.Time }

func (s stamp) String() string { return s.Format("2006-01-02 15:04:05") }

type orderOptions struct {
	listOptions
	Statuses   []string `url:"statuses,omitempty,comma"`
	Tags       []string `url:"tags,omitempty"`
	IDs        []int64  `url:"ids,omitempty,brackets"`
	Vendor     string   `url:"vendor,omitempty"`
	OnlyValid  bool     `url:"only_valid,omitempty"`
	Published  *bool    `url:"published,omitempty"`
	Since      stamp    `url:"updated_at_start_time,omitempty"`
	Threshold  float64  `url:"threshold,omitempty"`
	Skipped    string   `url:"-"`
	unexported string
}

func TestValues(t *testing.T) {
	no := false
	opts := &orderOptions{
		listOptions: listOptions{Page: 2},
		Statuses:    []string{"open", "closed"},
		Tags:        []string{"a", "b"},
		IDs:         []int64{1, 2},
		Published:   &no,
		Since:       stamp{time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)},
		Threshold:   1.5,
		Skipped:     "x",
		unexported:  "y",
	}
	v, err := Values(opts)
	if err != nil {
		t.Fatal(err)
	}
	want := "ids%5B%5D=1&ids%5B%5D=2&page=2&published=false&statuses=open%2Cclosed&tags=a&tags=b&threshold=1.5&updated_at_start_time=2026-01-02+03%3A04%3A05"
	if got := v.Encode(); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestValuesNilAndZero(t *testing.T) {
	var nilOpts *orderOptions
	if v, err := Values(nilOpts); err != nil || len(v) != 0 {
		t.Errorf("nil: %v %v", v, err)
	}
	if v, err := Values(orderOptions{}); err != nil || len(v) != 0 {
		t.Errorf("zero: %v %v", v, err)
	}
	if _, err := Values(42); err == nil {
		t.Error("non-struct accepted")
	}
}
