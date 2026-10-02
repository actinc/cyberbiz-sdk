package cyberbiz

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestAssetPicturesGolden(t *testing.T) {
	var out []AssetPicture
	decodeGolden(t, "v1/GET_v1_assets_pictures.json", &out)
	if out == nil || len(out) != 0 {
		t.Errorf("pictures = %#v", out)
	}
	c := goldenServer(t)
	page, err := c.Assets.ListPictures(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 || page.Pagination.Total != 1 {
		t.Errorf("page = %+v", page)
	}
}

func TestAssetPictureDecodesPostmanShape(t *testing.T) {
	c, _ := New("tok")
	var out AssetPicture
	body := `{"id":4650,"url":"https://cdn.example/a.jpg","url_thumb":"https://cdn.example/a_thumb.jpg","url_content":"https://cdn.example/a_content.jpg"}`
	if err := c.decode([]byte(body), &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != 4650 || out.URLThumb != "https://cdn.example/a_thumb.jpg" {
		t.Errorf("picture = %+v", out)
	}
}

func TestAssetsRequests(t *testing.T) {
	c, call := discountsSpyClient(t, 200, `[]`)
	if _, err := c.Assets.ListPictures(testCtx, &ListOptions{PerPage: 20, Offset: 5}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/assets/pictures")
	if call.Query.Get("per_page") != "20" || call.Query.Get("offset") != "5" {
		t.Errorf("query = %v", call.Query)
	}

	d, dcall := discountsSpyClient(t, 200, `{"id":1}`)
	if _, err := d.Assets.DeletePicture(testCtx, 4650); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, dcall, "DELETE", "/v1/assets/pictures/4650")
}

func TestAssetsUploadPictureSendsMultipart(t *testing.T) {
	c, call := discountsSpyClient(t, 201, `{"id":77,"url":"u","url_thumb":"t","url_content":"c"}`)
	pic, _, err := c.Assets.UploadPicture(testCtx, "logo.png", strings.NewReader("PNGDATA"))
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v1/assets/pictures")
	if pic.ID != 77 {
		t.Errorf("picture = %+v", pic)
	}
	mediaType, params, err := mime.ParseMediaType(call.ContentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content type = %q (%v)", call.ContentType, err)
	}
	part, err := multipart.NewReader(bytes.NewReader(call.Body), params["boundary"]).NextPart()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(part)
	if part.FormName() != "picture" || part.FileName() != "logo.png" || string(data) != "PNGDATA" {
		t.Errorf("part = %s %s %q", part.FormName(), part.FileName(), data)
	}
}

func TestAssetsUploadIsNotRetriedOnTransportError(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	srv.Close()
	if _, _, err := c.Assets.UploadPicture(testCtx, "x.png", strings.NewReader("x")); err == nil {
		t.Fatal("expected error")
	}
}
