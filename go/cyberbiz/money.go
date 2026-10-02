package cyberbiz

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Money is a monetary amount in hundredths of the shop currency (TWD for
// every CYBERBIZ shop today), so 199.50 is Money(19950). It is exact under
// addition and comparison, which float64 is not. CYBERBIZ sends amounts as
// JSON numbers such as 9999.0 and occasionally as strings; both decode.
// Values with more than two decimals are rounded half away from zero.
//
// Money deliberately has no IsZero method: encoding/json/v2's omitzero would
// otherwise treat a *Money pointing at 0 as absent, and "send an explicit
// zero" is exactly what a pointer field in a request struct is for.
type Money int64

// MoneyFromFloat converts a float amount, rounding to the nearest hundredth.
func MoneyFromFloat(f float64) Money {
	if f < 0 {
		return -Money(-f*100 + 0.5)
	}
	return Money(f*100 + 0.5)
}

// MoneyFromInt converts a whole-unit amount such as 200 (TWD).
func MoneyFromInt(units int64) Money { return Money(units * 100) }

// ParseMoney parses a decimal string such as "199.5" or "-3" exactly.
func ParseMoney(s string) (Money, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}
	whole, frac, _ := strings.Cut(s, ".")
	if exp := strings.IndexAny(frac, "eE"); exp >= 0 || strings.ContainsAny(whole, "eE") {
		// Exponent notation: fall back to float parsing.
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("cyberbiz: cannot parse %q as Money", s)
		}
		m := MoneyFromFloat(f)
		if neg {
			m = -m
		}
		return m, nil
	}
	if whole == "" {
		whole = "0"
	}
	units, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cyberbiz: cannot parse %q as Money", s)
	}
	cents, err := parseCents(frac)
	if err != nil {
		return 0, fmt.Errorf("cyberbiz: cannot parse %q as Money", s)
	}
	m := Money(units*100 + cents)
	if neg {
		m = -m
	}
	return m, nil
}

// parseCents converts the fractional digits to hundredths, rounding half up.
func parseCents(frac string) (int64, error) {
	if frac == "" {
		return 0, nil
	}
	for _, r := range frac {
		if r < '0' || r > '9' {
			return 0, errors.New("invalid fraction")
		}
	}
	padded := (frac + "000")[:3]
	n, err := strconv.ParseInt(padded, 10, 64)
	if err != nil {
		return 0, err
	}
	cents := n / 10
	if n%10 >= 5 {
		cents++
	}
	return cents, nil
}

// Float64 returns the amount as a float, for callers that need one.
func (m Money) Float64() float64 { return float64(m) / 100 }

// Units returns the whole-unit part, truncated toward zero.
func (m Money) Units() int64 { return int64(m) / 100 }

// String formats the amount with the fewest decimals needed: "200" or
// "199.5" or "0.05".
func (m Money) String() string {
	neg := m < 0
	if neg {
		m = -m
	}
	units, cents := int64(m)/100, int64(m)%100
	var s string
	switch {
	case cents == 0:
		s = strconv.FormatInt(units, 10)
	case cents%10 == 0:
		s = fmt.Sprintf("%d.%d", units, cents/10)
	default:
		s = fmt.Sprintf("%d.%02d", units, cents)
	}
	if neg {
		return "-" + s
	}
	return s
}

// MarshalJSONTo implements json.MarshalerTo (encoding/json/v2). The amount
// is written as a JSON number.
func (m Money) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(jsontext.Value(m.String()))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
func (m *Money) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		*m = 0
		return nil
	case '0', '"':
		parsed, err := ParseMoney(tok.String())
		if err != nil {
			return err
		}
		*m = parsed
		return nil
	}
	return errors.New("cyberbiz: Money must be a JSON number, string or null")
}

// MarshalJSON implements json.Marshaler (encoding/json v1).
func (m Money) MarshalJSON() ([]byte, error) { return json.Marshal(m) }

// UnmarshalJSON implements json.Unmarshaler (encoding/json v1).
func (m *Money) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, m) }
