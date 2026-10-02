package cyberbiz

// AssetPicture is an image in the shop's CKEditor picture library
// (GET /v1/assets/pictures). The test shop's library is empty, so the shape
// follows the swagger and Postman examples.
type AssetPicture struct {
	ID int64 `json:"id"`
	// URL is the original image; URLThumb and URLContent are the thumbnail
	// and content-width renditions.
	URL        string `json:"url"`
	URLThumb   string `json:"url_thumb"`
	URLContent string `json:"url_content"`
}
