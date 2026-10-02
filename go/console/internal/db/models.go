package db

import (
	"encoding/json"
	"time"
)

// JSONText is a column that holds JSON text, such as a header map. It is
// stored as text but serialized as the JSON it contains, so API clients get
// an object instead of a string; empty or invalid text serializes as null.
type JSONText string

// MarshalJSON implements json.Marshaler.
func (t JSONText) MarshalJSON() ([]byte, error) {
	if t == "" || !json.Valid([]byte(t)) {
		return []byte("null"), nil
	}
	return []byte(t), nil
}

// UnmarshalJSON implements json.Unmarshaler; it keeps the JSON text as is.
func (t *JSONText) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*t = ""
		return nil
	}
	*t = JSONText(data)
	return nil
}

// User is a Console login.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName implements the explicit snake_case plural convention.
func (User) TableName() string { return "users" }

// Session is one login; the ID is the cookie value.
type Session struct {
	ID        string    `gorm:"primaryKey;size:64"`
	UserID    uint      `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

// TableName implements the explicit snake_case plural convention.
func (Session) TableName() string { return "sessions" }

// Shop is one CYBERBIZ merchant store with its App credentials and API
// Token, both stored encrypted.
type Shop struct {
	ID               uint   `gorm:"primaryKey"`
	Name             string `gorm:"size:255"`
	ShopDomain       string `gorm:"uniqueIndex;size:255"`
	CustomDomain     string `gorm:"size:255"`
	CyberbizShopID   int64  `gorm:"column:cyberbiz_shop_id"`
	AppName          string `gorm:"size:255"`
	AppID            string `gorm:"column:app_id;size:255"`
	AppSecretEnc     []byte `gorm:"column:app_secret_enc"`
	APITokenEnc      []byte `gorm:"column:api_token_enc"`
	TokenFingerprint string `gorm:"size:64"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TableName implements the explicit snake_case plural convention.
func (Shop) TableName() string { return "shops" }

// OutboundLog is one HTTP attempt the SDK sent to the CYBERBIZ API.
type OutboundLog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ShopID          uint      `gorm:"index:idx_outbound_shop_created,priority:1" json:"shop_id"`
	Method          string    `gorm:"size:16" json:"method"`
	Path            string    `gorm:"index;size:1024" json:"path"`
	Query           string    `json:"query"`
	RequestHeaders  JSONText  `json:"request_headers"`
	RequestBody     string    `json:"request_body"`
	ResponseStatus  int       `json:"response_status"`
	ResponseHeaders JSONText  `json:"response_headers"`
	ResponseBody    string    `json:"response_body"`
	DurationMs      int64     `json:"duration_ms"`
	Attempt         int       `json:"attempt"`
	RequestID       string    `gorm:"size:128" json:"request_id"`
	Error           *string   `json:"error"`
	CreatedAt       time.Time `gorm:"index:idx_outbound_shop_created,priority:2" json:"created_at"`
}

// TableName implements the explicit snake_case plural convention.
func (OutboundLog) TableName() string { return "outbound_logs" }

// InboundLog is one webhook request CYBERBIZ sent us, whatever its outcome.
type InboundLog struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	ShopID               *uint     `json:"shop_id"`
	ShopDomain           string    `gorm:"index:idx_inbound_domain_created,priority:1;size:255" json:"shop_domain"`
	CustomDomain         string    `gorm:"size:255" json:"custom_domain"`
	Event                string    `gorm:"index;size:128" json:"event"`
	Signature            string    `gorm:"index;size:128" json:"signature"`
	DomainSignature      string    `gorm:"size:128" json:"domain_signature"`
	SignatureValid       bool      `json:"signature_valid"`
	DomainSignatureValid *bool     `json:"domain_signature_valid"`
	Status               string    `gorm:"size:32" json:"status"`
	Headers              JSONText  `json:"headers"`
	Body                 string    `json:"body"`
	RemoteAddr           string    `gorm:"size:64" json:"remote_addr"`
	ResponseStatus       int       `json:"response_status"`
	DuplicateOf          *uint     `json:"duplicate_of"`
	CreatedAt            time.Time `gorm:"index:idx_inbound_domain_created,priority:2" json:"created_at"`
}

// TableName implements the explicit snake_case plural convention.
func (InboundLog) TableName() string { return "inbound_logs" }

// Inbound statuses.
const (
	InboundValid            = "valid"
	InboundInvalidSignature = "invalid_signature"
	InboundUnknownShop      = "unknown_shop"
	InboundMalformed        = "malformed"
)
