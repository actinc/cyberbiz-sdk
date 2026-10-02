package cyberbiz

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Taipei is the zone every CYBERBIZ timestamp is expressed in. Taiwan has not
// observed daylight saving time since 1979, so a fixed +08:00 is exact and
// needs no tzdata at runtime.
var Taipei = time.FixedZone("Asia/Taipei", 8*60*60)

const (
	// TimeLayout is the platform's timestamp format. It carries no zone;
	// values are always in [Taipei].
	TimeLayout = "2006-01-02 15:04:05"
	// DateLayout is the platform's date-only format.
	DateLayout = "2006-01-02"
)

// timeLayouts are the formats accepted on input, most common first. ISO 8601
// forms are accepted because CYBERBIZ occasionally emits them; anything
// without a zone is interpreted in Taipei.
var timeLayouts = []string{
	TimeLayout,
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05 -07:00",
	"2006-01-02 15:04:05-0700",
	"2006-01-02 15:04-0700", // seen in POS wallet transactions
	"2006-01-02 15:04",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05.999999",
	DateLayout,
}

// Time is a [time.Time] that marshals to and from the CYBERBIZ timestamp
// format. A JSON null or empty string decodes to the zero Time.
type Time struct {
	time.Time
}

// NewTime converts t to a Time in the Taipei zone.
func NewTime(t time.Time) Time { return Time{t.In(Taipei)} }

// ParseTime parses s in any of the accepted formats.
func ParseTime(s string) (Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Time{}, nil
	}
	for _, layout := range timeLayouts {
		t, err := time.ParseInLocation(layout, s, Taipei)
		if err == nil {
			return Time{t.In(Taipei)}, nil
		}
	}
	return Time{}, fmt.Errorf("cyberbiz: cannot parse %q as a CYBERBIZ timestamp", s)
}

// String formats the time in [TimeLayout]; the zero Time is "".
func (t Time) String() string {
	if t.IsZero() {
		return ""
	}
	return t.In(Taipei).Format(TimeLayout)
}

// MarshalJSONTo implements json.MarshalerTo (encoding/json/v2).
func (t Time) MarshalJSONTo(enc *jsontext.Encoder) error {
	if t.IsZero() {
		return enc.WriteToken(jsontext.Null)
	}
	return enc.WriteToken(jsontext.String(t.String()))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
func (t *Time) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		*t = Time{}
		return nil
	case '"':
		parsed, err := ParseTime(tok.String())
		if err != nil {
			return err
		}
		*t = parsed
		return nil
	}
	return errors.New("cyberbiz: Time must be a JSON string or null")
}

// MarshalJSON implements json.Marshaler (encoding/json v1).
func (t Time) MarshalJSON() ([]byte, error) { return json.Marshal(t) }

// UnmarshalJSON implements json.Unmarshaler (encoding/json v1).
func (t *Time) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, t) }

// Date is a calendar day in the CYBERBIZ "2006-01-02" format, used for
// fields such as a customer's birthday. A JSON null or empty string decodes
// to the zero Date.
type Date struct {
	time.Time
}

// NewDate returns the Date for year, month and day.
func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, Taipei)}
}

// ParseDate parses s as "2006-01-02".
func ParseDate(s string) (Date, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Date{}, nil
	}
	t, err := time.ParseInLocation(DateLayout, s, Taipei)
	if err != nil {
		// Some endpoints return a full timestamp where a date is documented.
		full, ferr := ParseTime(s)
		if ferr != nil {
			return Date{}, fmt.Errorf("cyberbiz: cannot parse %q as a CYBERBIZ date", s)
		}
		t = full.Time
	}
	return Date{time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Taipei)}, nil
}

// String formats the date as "2006-01-02"; the zero Date is "".
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.In(Taipei).Format(DateLayout)
}

// MarshalJSONTo implements json.MarshalerTo (encoding/json/v2).
func (d Date) MarshalJSONTo(enc *jsontext.Encoder) error {
	if d.IsZero() {
		return enc.WriteToken(jsontext.Null)
	}
	return enc.WriteToken(jsontext.String(d.String()))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom (encoding/json/v2).
func (d *Date) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	switch tok.Kind() {
	case 'n':
		*d = Date{}
		return nil
	case '"':
		parsed, err := ParseDate(tok.String())
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	}
	return errors.New("cyberbiz: Date must be a JSON string or null")
}

// MarshalJSON implements json.Marshaler (encoding/json v1).
func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d) }

// UnmarshalJSON implements json.Unmarshaler (encoding/json v1).
func (d *Date) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, d) }
