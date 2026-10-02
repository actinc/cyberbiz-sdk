package cyberbiz

import (
	"context"
	"encoding/json/jsontext"
)

// ShopService exposes the unprefixed app endpoints: the shop's profile and
// the app's own settings.
type ShopService struct {
	client *Client
}

// ShopInfo is the profile of the shop that owns the token (GET /shop).
type ShopInfo struct {
	ID               int64            `json:"id"`
	Name             string           `json:"name"`
	PrimaryDomain    string           `json:"primary_domain"`
	Email            string           `json:"email"`
	Currency         string           `json:"currency"`
	Language         string           `json:"language"`
	OGImageURL       string           `json:"og_image_url"`
	SMSPrefix        string           `json:"sms_prefix"`
	MerchantLocation string           `json:"merchant_location"`
	MarketLocation   string           `json:"market_location"`
	ShopLine         *ShopLine        `json:"shop_line"`
	ShopLineChatBot  *ShopLineChatBot `json:"shop_line_chat_bot"`
}

// ShopLine describes the shop's LINE login integration.
type ShopLine struct {
	LoginEnable bool   `json:"login_enable"`
	LIFFID      string `json:"liff_id"`
	LIFFEnable  bool   `json:"liff_enable"`
}

// ShopLineChatBot describes the shop's LINE chat bot, when one is linked.
type ShopLineChatBot struct {
	BotID string `json:"bot_id"`
}

type shopInfoEnvelope struct {
	ShopInfo *ShopInfo `json:"shop_info"`
}

// Info returns the shop profile.
func (s *ShopService) Info(ctx context.Context) (*ShopInfo, *Response, error) {
	var env shopInfoEnvelope
	resp, err := s.client.get(ctx, "shop", nil, &env)
	if err != nil {
		return nil, resp, err
	}
	return env.ShopInfo, resp, nil
}

// AppSettings is the installed app's record for this shop (GET /settings).
// Token is the app's own credential and must never be logged or stored in
// plain text.
type AppSettings struct {
	ID           int64          `json:"id"`
	Settings     jsontext.Value `json:"settings"` // shape is defined by the app manifest
	VendorType   string         `json:"vendor_type"`
	Token        string         `json:"token"`
	StartAt      Time           `json:"start_at"`
	EndAt        Time           `json:"end_at"`
	AddOnVersion *AddOnVersion  `json:"add_on_version"`
}

// AddOnVersion is the app version installed on the shop.
type AddOnVersion struct {
	WebhookURL string       `json:"webhook_url"`
	Manifest   *AppManifest `json:"manifest"`
	Embedded   bool         `json:"embedded"`
	Status     string       `json:"status"`
	Scopes     string       `json:"scopes"` // space-separated
}

// AppManifest is the manifest the app was registered with.
type AppManifest struct {
	Name            string         `json:"name"`
	Version         string         `json:"version"`
	Scopes          string         `json:"scopes"` // space-separated
	ManifestVersion int            `json:"manifest_version"`
	Type            string         `json:"type"`
	WebhookEvents   []string       `json:"webhook_events"`
	SettingFields   jsontext.Value `json:"setting_fields,omitzero"`
}

type appSettingsEnvelope struct {
	ShopAddOn *AppSettings `json:"shop_add_on"`
}

// Settings returns the app's settings record for this shop.
func (s *ShopService) Settings(ctx context.Context) (*AppSettings, *Response, error) {
	var env appSettingsEnvelope
	resp, err := s.client.get(ctx, "settings", nil, &env)
	if err != nil {
		return nil, resp, err
	}
	return env.ShopAddOn, resp, nil
}

// SettingValue is one setting field declared in the app manifest.
type SettingValue struct {
	Field string `json:"field"`
	Data  any    `json:"data"`
}

// UpdateSettings writes the given setting fields (PUT /settings).
func (s *ShopService) UpdateSettings(ctx context.Context, values []SettingValue) (*AppSettings, *Response, error) {
	body := struct {
		Settings []SettingValue `json:"settings"`
	}{Settings: values}
	var env appSettingsEnvelope
	resp, err := s.client.put(ctx, "settings", body, &env)
	if err != nil {
		return nil, resp, err
	}
	return env.ShopAddOn, resp, nil
}
