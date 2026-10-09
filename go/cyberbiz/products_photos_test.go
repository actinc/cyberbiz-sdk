package cyberbiz

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"
)

func TestProductPhotoGolden(t *testing.T) {
	var list []ProductPhoto
	decodeGolden(t, "v1/GET_v1_products_{id}_product_photos.json", &list)
	if list == nil || len(list) != 0 {
		t.Errorf("photos = %v", list)
	}
	c := goldenServer(t)
	list, _, err := c.Products.ListPhotos(testCtx, 69458894)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("photos = %v", list)
	}
}

// productsMultipart returns the plain fields and the "photo" file of a recorded
// multipart body.
func productsMultipart(t *testing.T, rec *productsRecorded) (map[string]string, string, []byte) {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(rec.ContentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content-type = %q: %v", rec.ContentType, err)
	}
	form, err := multipart.NewReader(bytes.NewReader(rec.Body), params["boundary"]).ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{}
	for k, v := range form.Value {
		fields[k] = v[0]
	}
	files := form.File["photo"]
	if len(files) != 1 {
		t.Fatalf("photo parts = %d", len(files))
	}
	f, err := files[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return fields, files[0].Filename, data
}

func TestProductsCreatePhotoMultipart(t *testing.T) {
	// The platform stores the photo in a background job and answers only
	// {"success": true}.
	c, rec := productsRecorder(t, 201, `{"success":true}`)
	resp, err := c.Products.CreatePhoto(testCtx, 5, &ProductPhotoUploadRequest{
		Photo: strings.NewReader("PNGDATA"), Filename: "x.png", Position: productsPtr(2), ProductVariantIDs: []int64{1, 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products/5/product_photos", "")
	fields, name, data := productsMultipart(t, rec)
	if fields["position"] != "2" || fields["product_variant_ids"] != "1,2" || len(fields) != 2 {
		t.Errorf("fields = %v", fields)
	}
	if name != "x.png" || string(data) != "PNGDATA" {
		t.Errorf("file = %q %q", name, data)
	}
	if resp == nil || string(resp.Body) != `{"success":true}` {
		t.Errorf("resp = %+v", resp)
	}

	// Optional fields are omitted entirely.
	if _, err := c.Products.CreatePhoto(testCtx, 5, &ProductPhotoUploadRequest{Photo: strings.NewReader("x")}); err != nil {
		t.Fatal(err)
	}
	fields, name, _ = productsMultipart(t, rec)
	if len(fields) != 0 || name != "photo" {
		t.Errorf("fields = %v name = %q", fields, name)
	}
	if _, err := c.Products.CreatePhoto(testCtx, 5, nil); err == nil {
		t.Error("nil request accepted")
	}
}

func TestProductsPhotoBatchAndDelete(t *testing.T) {
	c, rec := productsRecorder(t, 201, ``)
	_, err := c.Products.BatchCreatePhotos(testCtx, &ProductPhotoBatchCreateRequest{
		Photo: strings.NewReader("IMG"), Filename: "a.jpg", Type: ProductPhotoBatchTypeQC, Value: "WARE0014",
	})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products/batch_create_product_photo", "")
	fields, _, data := productsMultipart(t, rec)
	if fields["type"] != "qc" || fields["value"] != "WARE0014" || string(data) != "IMG" {
		t.Errorf("fields = %v data = %q", fields, data)
	}
	if _, err := c.Products.BatchCreatePhotos(testCtx, nil); err == nil {
		t.Error("nil request accepted")
	}

	c, rec = productsRecorder(t, 204, ``)
	if _, err := c.Products.BatchDeletePhotos(testCtx, ProductPhotoBatchTypeSKU, "471 1"); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "DELETE", "/v1/products/batch_delete_product_photo", "type=sku&value=471+1")

	if _, err := c.Products.DeletePhoto(testCtx, 5, 9); err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "DELETE", "/v1/products/5/product_photos/9", "")
}

func TestProductDescriptionsRequests(t *testing.T) {
	c, rec := productsRecorder(t, 200, `[{"setting_name":"product_description_section_intro","body_html":"<p>hi</p>","id":3}]`)
	list, _, err := c.Products.ListDescriptions(testCtx, 5)
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/5/product_descriptions", "")
	if len(list) != 1 || list[0].ID != 3 {
		t.Errorf("descriptions = %+v", list)
	}

	c, rec = productsRecorder(t, 200, `{"setting_name":"product_description_section_intro","body_html":"<p>hi</p>"}`)

	d, _, err := c.Products.GetDescription(testCtx, 5, 3)
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "GET", "/v1/products/5/product_descriptions/3", "")
	if d.ID != 3 || d.SettingName != ProductDescriptionSettingIntro || d.BodyHTML != "<p>hi</p>" {
		t.Errorf("description = %+v", d)
	}

	d, _, err = c.Products.CreateDescription(testCtx, 5, &ProductDescriptionCreateRequest{SettingName: ProductDescriptionSettingIntro})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "POST", "/v1/products/5/product_descriptions", "")
	rec.assertJSON(t, `{"setting_name":"product_description_section_intro"}`)
	if d.BodyHTML == "" {
		t.Errorf("description = %+v", d)
	}

	d, _, err = c.Products.UpdateDescription(testCtx, 5, 3, &ProductDescriptionUpdateRequest{BodyHTML: productsPtr("")})
	if err != nil {
		t.Fatal(err)
	}
	rec.assert(t, "PUT", "/v1/products/5/product_descriptions/3", "")
	rec.assertJSON(t, `{"body_html":""}`)
	if d.ID != 3 {
		t.Errorf("id not set from path: %d", d.ID)
	}

	c, _ = productsRecorder(t, 404, string(readGolden(t, "errors/GET_v1_products_{id}_product_descriptions_{id}.json")))
	if _, _, err := c.Products.GetDescription(testCtx, 5, 3); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v", err)
	}
}
