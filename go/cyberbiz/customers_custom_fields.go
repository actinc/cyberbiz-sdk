package cyberbiz

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CustomerCustomFieldCreateRequest is the body of
// POST /v1/customers/{id}/custom_fields.
type CustomerCustomFieldCreateRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CustomerCustomFieldUpdateRequest is the body of
// PUT /v1/customers/{id}/custom_fields/{custom_field_id}.
type CustomerCustomFieldUpdateRequest struct {
	Name  *string `json:"name,omitzero"`
	Value *string `json:"value,omitzero"`
}

// CustomerCustomFieldInput is one entry of the bulk upsert
// (PUT /v1/customers/{id}/custom_fields). With ID set the field is updated,
// without it a new field is created.
type CustomerCustomFieldInput struct {
	ID    int64  `json:"id,omitzero"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ListCustomFields returns one page of the customer's custom fields
// (GET /v1/customers/{id}/custom_fields).
func (s *CustomersService) ListCustomFields(ctx context.Context, id int64, opts *ListOptions) (*Page[CustomerCustomField], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerCustomField](ctx, s.client, fmt.Sprintf("v1/customers/%d/custom_fields", id), q)
}

// GetCustomField returns one custom field
// (GET /v1/customers/{id}/custom_fields/{custom_field_id}).
func (s *CustomersService) GetCustomField(ctx context.Context, customerID, fieldID int64) (*CustomerCustomField, *Response, error) {
	var out CustomerCustomField
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/customers/%d/custom_fields/%d", customerID, fieldID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ID == 0 {
		out.ID = fieldID
	}
	return &out, resp, nil
}

// CreateCustomField adds one custom field to the customer
// (POST /v1/customers/{id}/custom_fields).
func (s *CustomersService) CreateCustomField(ctx context.Context, id int64, req *CustomerCustomFieldCreateRequest) (*CustomerCustomField, *Response, error) {
	var out CustomerCustomField
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/customers/%d/custom_fields", id), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpsertCustomFields creates or updates several custom fields at once and
// returns the customer's resulting fields (PUT /v1/customers/{id}/custom_fields).
func (s *CustomersService) UpsertCustomFields(ctx context.Context, id int64, fields []CustomerCustomFieldInput) ([]CustomerCustomField, *Response, error) {
	body := struct {
		CustomFields []CustomerCustomFieldInput `json:"custom_fields"`
	}{CustomFields: fields}
	out := []CustomerCustomField{}
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/customers/%d/custom_fields", id), body, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// UpdateCustomField changes one custom field
// (PUT /v1/customers/{id}/custom_fields/{custom_field_id}).
func (s *CustomersService) UpdateCustomField(ctx context.Context, customerID, fieldID int64, req *CustomerCustomFieldUpdateRequest) (*CustomerCustomField, *Response, error) {
	var out CustomerCustomField
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/customers/%d/custom_fields/%d", customerID, fieldID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.ID == 0 {
		out.ID = fieldID
	}
	return &out, resp, nil
}

// DeleteCustomField removes one custom field
// (DELETE /v1/customers/{id}/custom_fields/{custom_field_id}).
func (s *CustomersService) DeleteCustomField(ctx context.Context, customerID, fieldID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/customers/%d/custom_fields/%d", customerID, fieldID), nil)
}

// DeleteCustomFields removes several custom fields by id
// (DELETE /v1/customers/{id}/custom_fields?custom_field_ids[]=...).
func (s *CustomersService) DeleteCustomFields(ctx context.Context, customerID int64, fieldIDs []int64) (*Response, error) {
	q := url.Values{}
	for _, fid := range fieldIDs {
		q.Add("custom_field_ids[]", strconv.FormatInt(fid, 10))
	}
	return s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("v1/customers/%d/custom_fields", customerID),
		Query:  q,
	}, nil)
}
