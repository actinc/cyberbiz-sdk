package cyberbiz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ProductPhotoUploadRequest is the multipart body of CreatePhoto. Photo and
// Filename are required.
type ProductPhotoUploadRequest struct {
	Photo             io.Reader
	Filename          string
	Position          *int    // display order
	ProductVariantIDs []int64 // variants the photo belongs to (enterprise feature)
}

// ProductPhotoBatchCreateRequest is the multipart body of BatchCreatePhotos.
// Every field is required; the photo is attached to every variant whose SKU
// or QC equals Value.
type ProductPhotoBatchCreateRequest struct {
	Photo    io.Reader
	Filename string
	Type     ProductPhotoBatchType
	Value    string
}

// ProductDescriptionCreateRequest is the body of CreateDescription.
// SettingName is required.
type ProductDescriptionCreateRequest struct {
	SettingName ProductDescriptionSetting `json:"setting_name"`
	BodyHTML    *string                   `json:"body_html,omitzero"`
}

// ProductDescriptionUpdateRequest is the body of UpdateDescription.
type ProductDescriptionUpdateRequest struct {
	SettingName ProductDescriptionSetting `json:"setting_name,omitzero"`
	BodyHTML    *string                   `json:"body_html,omitzero"`
}

// ListPhotos returns a product's photos (GET /v1/products/{id}/product_photos).
func (s *ProductsService) ListPhotos(ctx context.Context, productID int64) ([]ProductPhoto, *Response, error) {
	var out []ProductPhoto
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/products/%d/product_photos", productID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// CreatePhoto uploads a photo to a product (POST /v1/products/{id}/product_photos).
func (s *ProductsService) CreatePhoto(ctx context.Context, productID int64, req *ProductPhotoUploadRequest) (*ProductPhoto, *Response, error) {
	if req == nil || req.Photo == nil {
		return nil, nil, errors.New("cyberbiz: photo must not be nil")
	}
	fields := map[string]string{}
	if req.Position != nil {
		fields["position"] = strconv.Itoa(*req.Position)
	}
	if len(req.ProductVariantIDs) > 0 {
		fields["product_variant_ids"] = productsJoinInt64s(req.ProductVariantIDs)
	}
	body, contentType, err := productsMultipartBody(req.Photo, req.Filename, fields)
	if err != nil {
		return nil, nil, err
	}
	var out ProductPhoto
	resp, err := s.client.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   fmt.Sprintf("v1/products/%d/product_photos", productID),
		Body:   body,
		Header: http.Header{"Content-Type": {contentType}},
	}, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// DeletePhoto removes a photo (DELETE /v1/products/{id}/product_photos/{id}).
func (s *ProductsService) DeletePhoto(ctx context.Context, productID, photoID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/products/%d/product_photos/%d", productID, photoID), nil)
}

// BatchCreatePhotos uploads one photo to every variant matched by SKU or QC
// (POST /v1/products/batch_create_product_photo).
func (s *ProductsService) BatchCreatePhotos(ctx context.Context, req *ProductPhotoBatchCreateRequest) (*Response, error) {
	if req == nil || req.Photo == nil {
		return nil, errors.New("cyberbiz: photo must not be nil")
	}
	fields := map[string]string{"type": string(req.Type), "value": req.Value}
	body, contentType, err := productsMultipartBody(req.Photo, req.Filename, fields)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, &Request{
		Method: http.MethodPost,
		Path:   "v1/products/batch_create_product_photo",
		Body:   body,
		Header: http.Header{"Content-Type": {contentType}},
	}, nil)
}

// BatchDeletePhotos removes the photos of every variant matched by SKU or QC
// (DELETE /v1/products/batch_delete_product_photo).
func (s *ProductsService) BatchDeletePhotos(ctx context.Context, typ ProductPhotoBatchType, value string) (*Response, error) {
	q := url.Values{"type": {string(typ)}, "value": {value}}
	return s.client.Do(ctx, &Request{
		Method: http.MethodDelete,
		Path:   "v1/products/batch_delete_product_photo",
		Query:  q,
	}, nil)
}

// ListDescriptions returns a product's description sections
// (GET /v1/products/{id}/product_descriptions).
func (s *ProductsService) ListDescriptions(ctx context.Context, productID int64) ([]ProductDescription, *Response, error) {
	var out []ProductDescription
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/products/%d/product_descriptions", productID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// GetDescription returns one description section
// (GET /v1/products/{id}/product_descriptions/{id}).
func (s *ProductsService) GetDescription(ctx context.Context, productID, descriptionID int64) (*ProductDescription, *Response, error) {
	var out ProductDescription
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/products/%d/product_descriptions/%d", productID, descriptionID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = descriptionID
	return &out, resp, nil
}

// CreateDescription adds a description section to a product
// (POST /v1/products/{id}/product_descriptions).
func (s *ProductsService) CreateDescription(ctx context.Context, productID int64, req *ProductDescriptionCreateRequest) (*ProductDescription, *Response, error) {
	var out ProductDescription
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/products/%d/product_descriptions", productID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateDescription changes a description section
// (PUT /v1/products/{id}/product_descriptions/{id}).
func (s *ProductsService) UpdateDescription(ctx context.Context, productID, descriptionID int64, req *ProductDescriptionUpdateRequest) (*ProductDescription, *Response, error) {
	var out ProductDescription
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/products/%d/product_descriptions/%d", productID, descriptionID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	out.ID = descriptionID
	return &out, resp, nil
}

// productsMultipartBody encodes a file part named "photo" plus plain fields, and
// returns the body and its Content-Type (which carries the boundary).
func productsMultipartBody(photo io.Reader, filename string, fields map[string]string) ([]byte, string, error) {
	if filename == "" {
		filename = "photo"
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for name, value := range fields {
		if err := w.WriteField(name, value); err != nil {
			return nil, "", fmt.Errorf("cyberbiz: encoding multipart field %s: %w", name, err)
		}
	}
	part, err := w.CreateFormFile("photo", filename)
	if err != nil {
		return nil, "", fmt.Errorf("cyberbiz: encoding multipart file: %w", err)
	}
	if _, err := io.Copy(part, photo); err != nil {
		return nil, "", fmt.Errorf("cyberbiz: reading photo: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, "", fmt.Errorf("cyberbiz: encoding multipart body: %w", err)
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

func productsJoinInt64s(ids []int64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}
