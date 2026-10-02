package cyberbiz

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"mime/multipart"
	"net/http"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// AssetsService exposes the shop's picture library (/v1/assets/pictures).
type AssetsService struct {
	client *Client
}

// ListPictures returns one page of library pictures (GET /v1/assets/pictures).
func (s *AssetsService) ListPictures(ctx context.Context, opts *ListOptions) (*Page[AssetPicture], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[AssetPicture](ctx, s.client, "v1/assets/pictures", q)
}

// AllPictures walks every page of library pictures (GET /v1/assets/pictures).
func (s *AssetsService) AllPictures(ctx context.Context, opts *ListOptions) iter.Seq2[AssetPicture, error] {
	q, err := query.Values(opts)
	if err != nil {
		return func(yield func(AssetPicture, error) bool) { yield(AssetPicture{}, err) }
	}
	return listAll[AssetPicture](ctx, s.client, "v1/assets/pictures", q)
}

// UploadPicture adds an image to the library as multipart form data; filename
// is the name reported to the platform (POST /v1/assets/pictures).
func (s *AssetsService) UploadPicture(ctx context.Context, filename string, image io.Reader) (*AssetPicture, *Response, error) {
	req, err := assetsMultipartRequest("v1/assets/pictures", "picture", filename, image)
	if err != nil {
		return nil, nil, err
	}
	var out AssetPicture
	resp, err := s.client.Do(ctx, req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// DeletePicture removes a library picture (DELETE /v1/assets/pictures/{id}).
func (s *AssetsService) DeletePicture(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/assets/pictures/%d", id), nil)
}

// assetsMultipartRequest builds a POST whose body is one file field, for the few
// endpoints that take an upload instead of JSON.
func assetsMultipartRequest(path, field, filename string, r io.Reader) (*Request, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		return nil, fmt.Errorf("cyberbiz: building upload: %w", err)
	}
	if _, err := io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("cyberbiz: reading upload: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("cyberbiz: building upload: %w", err)
	}
	return &Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   buf.Bytes(),
		Header: http.Header{"Content-Type": {w.FormDataContentType()}},
	}, nil
}
