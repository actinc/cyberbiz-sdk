package cyberbiz

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"strconv"
	"strings"
)

// IDList is a list of ids that several v1 endpoints take as one
// comma-separated string ("12,34,56") instead of a JSON array. Assign a plain
// []int64 to it; it marshals to that string.
type IDList []int64

// String joins the ids with commas.
func (l IDList) String() string {
	parts := make([]string, len(l))
	for i, id := range l {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}

// MarshalJSONTo implements json.MarshalerTo (encoding/json/v2).
func (l IDList) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.String(l.String()))
}

// MarshalJSON implements json.Marshaler (encoding/json v1).
func (l IDList) MarshalJSON() ([]byte, error) { return json.Marshal(l) }
