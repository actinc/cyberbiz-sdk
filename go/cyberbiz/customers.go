package cyberbiz

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CustomersService exposes customers, their sub-resources (bonus points,
// coupons, custom fields, other-channel orders, customer groups) and the v2
// customer lookups.
type CustomersService struct {
	client *Client
}

// CustomerListOptions filter GET /v1/customers.
type CustomerListOptions struct {
	ListOptions
	UpdatedAtStartTime Time `url:"updated_at_start_time,omitempty"`
	UpdatedAtEndTime   Time `url:"updated_at_end_time,omitempty"`
}

// CustomerInclude names an optional section of GET /v2/customers.
type CustomerInclude string

// Known CustomerInclude values (include_params of GET /v2/customers).
const (
	CustomerIncludeUIDProviders CustomerInclude = "uid_providers" // social login UIDs (Customer.UIDProviders)
	CustomerIncludeTags         CustomerInclude = "tags"          // customer tags (Customer.Tags)
	CustomerIncludeVIPInfo      CustomerInclude = "vip_info"      // VIP level details (Customer.VIPInfo)
)

// CustomerListV2Options filter GET /v2/customers. When IDs is set (at most
// 100) the platform ignores paging and returns exactly those customers.
type CustomerListV2Options struct {
	ListOptions
	IDs     []int64           `url:"ids,omitempty,comma"`
	Include []CustomerInclude `url:"include_params,omitempty,comma"`
}

// CustomerAddressRequest is the default address sent when creating or
// updating a customer.
type CustomerAddressRequest struct {
	Name               string `json:"name,omitzero"` // recipient
	Country            string `json:"country,omitzero"`
	Province           string `json:"province,omitzero"`
	City               string `json:"city,omitzero"`
	District           string `json:"district,omitzero"`
	Address1           string `json:"address1,omitzero"`
	Zip                string `json:"zip,omitzero"`
	CountryCallingCode string `json:"country_calling_code,omitzero"`
	Phone              string `json:"phone,omitzero"`
}

// CustomerCreateRequest is the body of POST /v1/customers. The platform
// parses ConfirmedAt and MobileSMSConfirmedAt as strict ISO 8601, so the SDK
// sends them as RFC 3339 ("2006-01-02T15:04:05+08:00").
type CustomerCreateRequest struct {
	Name                                 string                  `json:"name,omitzero"`
	Status                               CustomerStatus          `json:"status,omitzero"`
	Email                                string                  `json:"email,omitzero"`
	Mobile                               string                  `json:"mobile,omitzero"`
	CountryCallingCode                   string                  `json:"country_calling_code,omitzero"`
	Password                             string                  `json:"password,omitzero"`
	TagsText                             string                  `json:"tags_text,omitzero"` // comma-separated tags
	EnableCVSPickup                      *bool                   `json:"enable_cvs_pickup,omitzero"`
	EnableCVSCOD                         *bool                   `json:"enable_cvs_cod,omitzero"`
	EnableHomeDeliveryCOD                *bool                   `json:"enable_home_delivery_cod,omitzero"`
	AcceptsMarketing                     *bool                   `json:"accepts_marketing,omitzero"`
	Gender                               string                  `json:"gender,omitzero"`
	Birthday                             Date                    `json:"birthday,omitzero"`
	OtherAccumulatedConsumption          *Money                  `json:"other_accumulated_consumption,omitzero"`
	OtherAccumulatedConsumptionExpiredAt Date                    `json:"other_accumulated_consumption_expired_at,omitzero"`
	Note                                 string                  `json:"note,omitzero"`
	ConfirmedAt                          *Time                   `json:"confirmed_at,omitzero"`
	MobileSMSConfirmedAt                 *Time                   `json:"mobile_sms_confirmed_at,omitzero"`
	EnableRegisterGift                   *bool                   `json:"enable_register_gift,omitzero"`
	EnableBirthGift                      *bool                   `json:"enable_birth_gift,omitzero"`
	Address                              *CustomerAddressRequest `json:"address,omitzero"`
}

// CustomerUpdateRequest is the body of PUT /v1/customers/{id}. ConfirmedAt
// and MobileSMSConfirmedAt are sent as RFC 3339, which the platform requires;
// a non-nil pointer to a zero Time sends null, which clears the timestamp.
type CustomerUpdateRequest struct {
	Name                                 string                  `json:"name,omitzero"`
	Status                               CustomerStatus          `json:"status,omitzero"`
	Email                                string                  `json:"email,omitzero"`
	CountryCallingCode                   string                  `json:"country_calling_code,omitzero"`
	Mobile                               string                  `json:"mobile,omitzero"`
	Password                             string                  `json:"password,omitzero"`
	TagsText                             string                  `json:"tags_text,omitzero"` // comma-separated tags
	EnableCVSPickup                      *bool                   `json:"enable_cvs_pickup,omitzero"`
	EnableCVSCOD                         *bool                   `json:"enable_cvs_cod,omitzero"`
	EnableHomeDeliveryCOD                *bool                   `json:"enable_home_delivery_cod,omitzero"`
	AcceptsMarketing                     *bool                   `json:"accepts_marketing,omitzero"`
	Gender                               string                  `json:"gender,omitzero"`
	EnableSendBirthGift                  *bool                   `json:"enable_send_birth_gift,omitzero"`
	Birthday                             Date                    `json:"birthday,omitzero"`
	OtherAccumulatedConsumption          *Money                  `json:"other_accumulated_consumption,omitzero"`
	OtherAccumulatedConsumptionExpiredAt Date                    `json:"other_accumulated_consumption_expired_at,omitzero"`
	Note                                 string                  `json:"note,omitzero"`
	ConfirmedAt                          *Time                   `json:"confirmed_at,omitzero"`
	MobileSMSConfirmedAt                 *Time                   `json:"mobile_sms_confirmed_at,omitzero"`
	Address                              *CustomerAddressRequest `json:"address,omitzero"`
}

// CustomerLookupOptions select customers by email or mobile for LookupIDs.
type CustomerLookupOptions struct {
	Emails  []string `url:"customer_emails,omitempty,comma"`
	Mobiles []string `url:"customer_mobiles,omitempty,comma"`
}

// CustomerConsumeBonusPointsRequest is the body of consume_bonus_points.
// ConsumeAll spends every remaining point; Amount is then ignored.
type CustomerConsumeBonusPointsRequest struct {
	Amount     Money `json:"amount,omitzero"`
	ConsumeAll *bool `json:"consume_all,omitzero"`
}

// CustomerRegisterCodeRequest is the body of
// update_register_code_and_accepts_marketing, used by third-party logins.
type CustomerRegisterCodeRequest struct {
	SecretKey        string `json:"secret_key"`    // the third party's key configured in CYBERBIZ
	RegisterCode     string `json:"register_code"` // the referrer's code
	AcceptsMarketing *bool  `json:"accepts_marketing,omitzero"`
}

// CustomerOAuthAppRequest is the body of POST /v2/customer_oauth: the
// customer-login OAuth application to register for the shop.
type CustomerOAuthAppRequest struct {
	AppName           string `json:"app_name"`
	ScopesDescription string `json:"scopes_description"` // shown to customers on the consent page
	Description       string `json:"description"`
	RedirectURI       string `json:"redirect_uri"`
	// DisplayShopName shows the shop name in the consent page title.
	DisplayShopName *bool `json:"display_shop_name,omitzero"`
}

// CustomerOAuthApp is the credential pair of a registered customer-login
// OAuth application.
type CustomerOAuthApp struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// List returns one page of customers (GET /v1/customers).
func (s *CustomersService) List(ctx context.Context, opts *CustomerListOptions) (*Page[Customer], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Customer](ctx, s.client, "v1/customers", q)
}

// All walks every page of customers (GET /v1/customers).
func (s *CustomersService) All(ctx context.Context, opts *CustomerListOptions) iter.Seq2[Customer, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Customer, error) bool) { yield(Customer{}, err) }
	}
	return listAll[Customer](ctx, s.client, "v1/customers", q)
}

// Get returns one customer (GET /v1/customers/{id}).
func (s *CustomersService) Get(ctx context.Context, id int64) (*Customer, *Response, error) {
	var out Customer
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// Create creates a customer (POST /v1/customers).
func (s *CustomersService) Create(ctx context.Context, req *CustomerCreateRequest) (*Customer, *Response, error) {
	body, err := req.wire()
	if err != nil {
		return nil, nil, fmt.Errorf("cyberbiz: encoding customer: %w", err)
	}
	var out Customer
	resp, err := s.client.post(ctx, "v1/customers", body, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// customerUpdateWire shadows the two clearable timestamps so that a pointer
// to a zero Time reaches the platform as an explicit null.
type customerUpdateWire struct {
	*CustomerUpdateRequest
	ConfirmedAt          jsontext.Value `json:"confirmed_at,omitzero"`
	MobileSMSConfirmedAt jsontext.Value `json:"mobile_sms_confirmed_at,omitzero"`
}

// customerCreateWire shadows the two verification timestamps so they are
// sent as RFC 3339.
type customerCreateWire struct {
	*CustomerCreateRequest
	ConfirmedAt          jsontext.Value `json:"confirmed_at,omitzero"`
	MobileSMSConfirmedAt jsontext.Value `json:"mobile_sms_confirmed_at,omitzero"`
}

func (r *CustomerCreateRequest) wire() (*customerCreateWire, error) {
	if r == nil {
		return nil, nil
	}
	confirmed, err := customersNullableTime(r.ConfirmedAt)
	if err != nil {
		return nil, err
	}
	mobile, err := customersNullableTime(r.MobileSMSConfirmedAt)
	if err != nil {
		return nil, err
	}
	return &customerCreateWire{CustomerCreateRequest: r, ConfirmedAt: confirmed, MobileSMSConfirmedAt: mobile}, nil
}

// customersNullableTime encodes a *Time as omitted (nil), null (zero) or an
// RFC 3339 string: the platform parses these fields with Ruby's
// Time.iso8601, which rejects the space-separated API layout.
func customersNullableTime(t *Time) (jsontext.Value, error) {
	if t == nil {
		return nil, nil
	}
	if t.IsZero() {
		return jsontext.Value("null"), nil
	}
	return json.Marshal(t.Format(time.RFC3339))
}

func (r *CustomerUpdateRequest) wire() (*customerUpdateWire, error) {
	if r == nil {
		return nil, nil
	}
	confirmed, err := customersNullableTime(r.ConfirmedAt)
	if err != nil {
		return nil, err
	}
	mobile, err := customersNullableTime(r.MobileSMSConfirmedAt)
	if err != nil {
		return nil, err
	}
	return &customerUpdateWire{CustomerUpdateRequest: r, ConfirmedAt: confirmed, MobileSMSConfirmedAt: mobile}, nil
}

// Update changes a customer (PUT /v1/customers/{id}).
func (s *CustomersService) Update(ctx context.Context, id int64, req *CustomerUpdateRequest) (*Customer, *Response, error) {
	body, err := req.wire()
	if err != nil {
		return nil, nil, fmt.Errorf("cyberbiz: encoding customer update: %w", err)
	}
	var out Customer
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/customers/%d", id), body, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ID == 0 {
		out.ID = id
	}
	return &out, resp, nil
}

// LookupIDs finds customer ids by email or mobile. The platform answers 206
// with an empty array when nothing matches; that is returned as an empty
// slice, not an error (GET /v1/customers/get_customer_id).
func (s *CustomersService) LookupIDs(ctx context.Context, opts *CustomerLookupOptions) ([]CustomerIDMatch, *Response, error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, nil, err
	}
	out := []CustomerIDMatch{}
	resp, err := s.client.get(ctx, "v1/customers/get_customer_id", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// LookupIDsByName finds customers whose name starts with prefix; limit caps
// the hits at up to 50, zero means the platform default
// (GET /v1/customers/get_customer_id_by_name).
func (s *CustomersService) LookupIDsByName(ctx context.Context, prefix string, limit int) ([]CustomerNameMatch, *Response, error) {
	q := url.Values{"customer_name": {prefix}}
	if limit > 0 {
		q.Set("limit", fmt.Sprint(limit))
	}
	out := []CustomerNameMatch{}
	resp, err := s.client.get(ctx, "v1/customers/get_customer_id_by_name", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// DefaultGenderOptions returns the shop's extra gender options
// (GET /v1/customers/default_gender_options).
func (s *CustomersService) DefaultGenderOptions(ctx context.Context) ([]CustomerGenderOption, *Response, error) {
	out := []CustomerGenderOption{}
	resp, err := s.client.get(ctx, "v1/customers/default_gender_options", nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// ListTags returns one page of customer tags (GET /v1/customers/tags).
func (s *CustomersService) ListTags(ctx context.Context, opts *ListOptions) (*Page[CustomerTag], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerTag](ctx, s.client, "v1/customers/tags", q)
}

// AccountActivationURL returns the customer's account activation link
// (GET /v1/customers/{id}/account_activation_url).
func (s *CustomersService) AccountActivationURL(ctx context.Context, id int64) (*CustomerActivationURL, *Response, error) {
	var out CustomerActivationURL
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d/account_activation_url", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ConsumeBonusPoints deducts bonus points from the customer
// (POST /v1/customers/{id}/consume_bonus_points).
func (s *CustomersService) ConsumeBonusPoints(ctx context.Context, id int64, req *CustomerConsumeBonusPointsRequest) (*Response, error) {
	return s.client.post(ctx, fmt.Sprintf("v1/customers/%d/consume_bonus_points", id), req, nil)
}

// UpdateRegisterCode records the referrer code and marketing consent of a
// customer who registered through a third-party login
// (POST /v1/customers/{id}/update_register_code_and_accepts_marketing).
func (s *CustomersService) UpdateRegisterCode(ctx context.Context, id int64, req *CustomerRegisterCodeRequest) (*Response, error) {
	return s.client.post(ctx, fmt.Sprintf("v1/customers/%d/update_register_code_and_accepts_marketing", id), req, nil)
}

// GetUIDProvider returns the customer's external UID for providerType
// (GET /v1/customers/{id}/uid_providers/{provider_type}).
func (s *CustomersService) GetUIDProvider(ctx context.Context, id int64, providerType CustomerUIDProviderType) (*CustomerUIDLookup, *Response, error) {
	var out CustomerUIDLookup
	path := fmt.Sprintf("v1/customers/%d/uid_providers/%s", id, url.PathEscape(string(providerType)))
	resp, err := s.client.getOne(ctx, path, nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// SetUIDProvider creates or replaces the customer's external UID for
// providerType (PUT /v1/customers/{id}/uid_providers/{provider_type}).
func (s *CustomersService) SetUIDProvider(ctx context.Context, id int64, providerType CustomerUIDProviderType, uid string) (*Response, error) {
	body := struct {
		UID string `json:"uid"`
	}{UID: uid}
	path := fmt.Sprintf("v1/customers/%d/uid_providers/%s", id, url.PathEscape(string(providerType)))
	return s.client.put(ctx, path, body, nil)
}

// VIPInfo returns the customer's VIP membership state
// (GET /v1/customers/{id}/vip_info).
func (s *CustomersService) VIPInfo(ctx context.Context, id int64) (*CustomerVIPInfo, *Response, error) {
	var out CustomerVIPInfo
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d/vip_info", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CustomerSpendingOptions bound SpendingOverview; both dates are required.
type CustomerSpendingOptions struct {
	StartDate Date `url:"start_date"`
	EndDate   Date `url:"end_date"`
}

// SpendingOverview summarises the customer's paid and valid orders between
// two dates (GET /v1/customers/{id}/spending_overview).
func (s *CustomersService) SpendingOverview(ctx context.Context, id int64, opts *CustomerSpendingOptions) (*CustomerSpendingOverview, *Response, error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, nil, err
	}
	var out CustomerSpendingOverview
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d/spending_overview", id), q, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CartItems returns the variants waiting in the customer's web-shop cart
// (GET /v1/customers/{id}/customer_cart_items).
func (s *CustomersService) CartItems(ctx context.Context, id int64) ([]CustomerCartItem, *Response, error) {
	out := []CustomerCartItem{}
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/customers/%d/customer_cart_items", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// ListMessagePosts returns one page of the customer's service threads
// (GET /v1/customers/{id}/message_posts).
func (s *CustomersService) ListMessagePosts(ctx context.Context, id int64, opts *ListOptions) (*Page[CustomerMessagePost], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerMessagePost](ctx, s.client, fmt.Sprintf("v1/customers/%d/message_posts", id), q)
}

// ListV2 returns one page of customers through the v2 endpoint, which can
// select by id and include VIP info (GET /v2/customers).
func (s *CustomersService) ListV2(ctx context.Context, opts *CustomerListV2Options) (*Page[Customer], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Customer](ctx, s.client, "v2/customers", q)
}

// AllV2 walks every page of the v2 customer list (GET /v2/customers).
func (s *CustomersService) AllV2(ctx context.Context, opts *CustomerListV2Options) iter.Seq2[Customer, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(Customer, error) bool) { yield(Customer{}, err) }
	}
	return listAll[Customer](ctx, s.client, "v2/customers", q)
}

// GetByUIDProvider finds the customer linked to an external identity. The
// platform declares and filters on "provider_type"; "provider" is sent too
// for compatibility with deployments the SDK was first tested against, where
// the lookup appeared to work with it (the platform ignores undeclared
// parameters, so that call matched the uid under any provider). The
// platform replies "200 null" for an unknown uid, which becomes ErrNotFound
// (GET /v2/customers/by_uid_provider).
func (s *CustomersService) GetByUIDProvider(ctx context.Context, providerType CustomerUIDProviderType, uid string) (*Customer, *Response, error) {
	q := url.Values{"uid": {uid}, "provider_type": {string(providerType)}, "provider": {string(providerType)}}
	var out Customer
	resp, err := s.client.getOne(ctx, "v2/customers/by_uid_provider", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateOAuthApp registers a customer-login OAuth application for the shop
// and returns its client id and secret. It needs the customer_oauth scope.
// A failure the platform reports as {"success":false,"errors":[...]} is
// returned as an *APIError (POST /v2/customer_oauth).
func (s *CustomersService) CreateOAuthApp(ctx context.Context, req *CustomerOAuthAppRequest) (*CustomerOAuthApp, *Response, error) {
	const path = "v2/customer_oauth"
	var out struct {
		CustomerOAuthApp
		Success *bool    `json:"success"`
		Errors  []string `json:"errors"`
	}
	resp, err := s.client.post(ctx, path, req, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ClientID == "" && out.Success != nil && !*out.Success {
		return nil, resp, &APIError{StatusCode: resp.StatusCode, Method: http.MethodPost, Path: path,
			RequestID: resp.RequestID, Messages: out.Errors, Body: resp.Body}
	}
	return &out.CustomerOAuthApp, resp, nil
}
