package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/url"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CustomFieldsService exposes custom field definitions (/v1/custom_fields)
// and custom field types (/v1/custom_field_types).
type CustomFieldsService struct {
	client *Client
}

// CustomFieldTypeCreateRequest is the body of CreateType.
type CustomFieldTypeCreateRequest struct {
	Model CustomFieldTypeModel `json:"model"`
	Name  string               `json:"name"`
}

// List returns every custom field definition of owner; the endpoint is not
// paginated (GET /v1/custom_fields).
func (s *CustomFieldsService) List(ctx context.Context, owner CustomFieldOwner) ([]CustomFieldSetting, *Response, error) {
	var out []CustomFieldSetting
	q := url.Values{"custom_field_type": {string(owner)}}
	resp, err := s.client.get(ctx, "v1/custom_fields", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// Get returns one custom field definition by its name. The platform answers
// an unknown name with 200 null, which is reported as ErrNotFound
// (GET /v1/custom_fields/{name}).
func (s *CustomFieldsService) Get(ctx context.Context, owner CustomFieldOwner, name string) (*CustomFieldSetting, *Response, error) {
	var out CustomFieldSetting
	q := url.Values{"custom_field_type": {string(owner)}}
	resp, err := s.client.getOne(ctx, "v1/custom_fields/"+url.PathEscape(name), q, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// ListTypes returns one page of custom field types (GET /v1/custom_field_types).
func (s *CustomFieldsService) ListTypes(ctx context.Context, opts *ListOptions) (*Page[CustomFieldType], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomFieldType](ctx, s.client, "v1/custom_field_types", q)
}

// AllTypes walks every page of custom field types (GET /v1/custom_field_types).
func (s *CustomFieldsService) AllTypes(ctx context.Context, opts *ListOptions) iter.Seq2[CustomFieldType, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(CustomFieldType, error) bool) { yield(CustomFieldType{}, err) }
	}
	return listAll[CustomFieldType](ctx, s.client, "v1/custom_field_types", q)
}

// GetType returns one custom field type (GET /v1/custom_field_types/{id}).
func (s *CustomFieldsService) GetType(ctx context.Context, id int64) (*CustomFieldType, *Response, error) {
	var out CustomFieldType
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/custom_field_types/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = id
	return &out, resp, nil
}

// CreateType creates a custom field type (POST /v1/custom_field_types).
func (s *CustomFieldsService) CreateType(ctx context.Context, req *CustomFieldTypeCreateRequest) (*CustomFieldType, *Response, error) {
	var out CustomFieldType
	resp, err := s.client.post(ctx, "v1/custom_field_types", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// DeleteType removes a custom field type (DELETE /v1/custom_field_types/{id}).
func (s *CustomFieldsService) DeleteType(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/custom_field_types/%d", id), nil)
}
