package cyberbiz

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"testing"
)

func TestParseMoney(t *testing.T) {
	cases := map[string]Money{
		"200":     20000,
		"9999.0":  999900,
		"199.5":   19950,
		"199.50":  19950,
		"0.05":    5,
		"-3":      -300,
		"-0.5":    -50,
		"":        0,
		"1.005":   101, // half away from zero
		"1.004":   100,
		"2.5e2":   25000,
		".5":      50,
		" 12.34 ": 1234,
	}
	for in, want := range cases {
		got, err := ParseMoney(in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("%q: got %d, want %d", in, got, want)
		}
	}
	for _, bad := range []string{"abc", "1.2.3", "1,000"} {
		if _, err := ParseMoney(bad); err == nil {
			t.Errorf("%q: expected error", bad)
		}
	}
}

func TestMoneyString(t *testing.T) {
	cases := map[Money]string{
		20000:  "200",
		19950:  "199.5",
		5:      "0.05",
		-50:    "-0.5",
		0:      "0",
		123456: "1234.56",
	}
	for in, want := range cases {
		if got := in.String(); got != want {
			t.Errorf("%d: got %q, want %q", in, got, want)
		}
	}
	if MoneyFromFloat(9999.0) != 999900 || MoneyFromFloat(-1.235) != -124 || MoneyFromInt(7) != 700 {
		t.Error("constructors")
	}
	if Money(19950).Float64() != 199.5 || Money(19950).Units() != 199 {
		t.Error("accessors")
	}
}

func TestMoneyJSON(t *testing.T) {
	type doc struct {
		Price Money  `json:"price"`
		Cost  Money  `json:"cost"`
		Fee   Money  `json:"fee"`
		Tip   *Money `json:"tip"`
	}
	var d doc
	err := json.Unmarshal([]byte(`{"price":9999.0,"cost":"199.5","fee":null,"tip":null}`), &d)
	if err != nil {
		t.Fatal(err)
	}
	if d.Price != 999900 || d.Cost != 19950 || d.Fee != 0 || d.Tip != nil {
		t.Errorf("got %+v", d)
	}
	out, _ := json.Marshal(d)
	if string(out) != `{"price":9999,"cost":199.5,"fee":0,"tip":null}` {
		t.Errorf("marshal = %s", out)
	}
	if err := json.Unmarshal([]byte(`{"price":true}`), &d); err == nil {
		t.Error("bool accepted")
	}
	// encoding/json v1 compatibility.
	var m Money
	if err := jsonv1.Unmarshal([]byte(`12.34`), &m); err != nil || m != 1234 {
		t.Errorf("v1 unmarshal: %v %v", m, err)
	}
	if b, _ := jsonv1.Marshal(m); string(b) != "12.34" {
		t.Errorf("v1 marshal = %s", b)
	}
}
