package cyberbiz

import "encoding/json/v2"

// PageStatus is the publication state of a custom page.
type PageStatus string

// Known PageStatus values (status of a custom page).
const (
	PageStatusPublished   PageStatus = "published"   // visible on the storefront
	PageStatusUnpublished PageStatus = "unpublished" // hidden from the storefront
)

// CustomPage is a storefront page created through the API (POST /v2/pages).
// Such pages can only be previewed, not edited, in the admin; changes go
// through UpdatePage. No Golden File exists; the shape follows the Notion
// reference.
type CustomPage struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// Handle is the URL slug, as in /pages/{handle}.
	Handle string     `json:"handle"`
	Status PageStatus `json:"status"`
	// HTMLInfos is a JSON string mapping section id to section HTML; see
	// HTMLSections.
	HTMLInfos string `json:"html_infos"`
	CreatedAt Time   `json:"created_at"`
	UpdatedAt Time   `json:"updated_at"`
}

// HTMLSections decodes HTMLInfos into section id to HTML content.
func (p *CustomPage) HTMLSections() (map[string]string, error) {
	out := map[string]string{}
	if p.HTMLInfos == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(p.HTMLInfos), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Menu is a storefront navigation menu, called a linklist in the admin
// (GET /v2/menus). Items is only filled by GetMenu.
type Menu struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// SystemDefault marks the menus every shop starts with.
	SystemDefault bool       `json:"system_default"`
	Items         []MenuItem `json:"items"`
}

// MenuItemType is what a menu item links to.
type MenuItemType string

// Known MenuItemType values (item_type of GET /v2/menus entries).
const (
	MenuItemTypeFrontpage      MenuItemType = "frontpage"       // the shop home page
	MenuItemTypeCollectionsAll MenuItemType = "collections_all" // the all-products listing
	MenuItemTypeHTTP           MenuItemType = "http"            // an arbitrary URL
	MenuItemTypeProduct        MenuItemType = "product"         // a single product page
	MenuItemTypePage           MenuItemType = "page"            // a custom page
	MenuItemTypeContact        MenuItemType = "contact"         // the contact form
)

// MenuItem is one entry of a menu; Items holds its sub-entries recursively.
type MenuItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// SubjectHandle is the handle of the linked product or page, when the
	// item links to one.
	SubjectHandle string `json:"subject_handle"`
	URL           string `json:"url"`
	Position      int    `json:"position"`
	Subtitle      string `json:"subtitle"`
	IconImageURL  string `json:"icon_image_url"`
	// StartAt and EndAt bound when the item is shown, when set.
	StartAt  Time         `json:"start_at"`
	EndAt    Time         `json:"end_at"`
	ItemType MenuItemType `json:"item_type"`
	// Level is the nesting depth, 1 for top-level items.
	Level int        `json:"level"`
	Items []MenuItem `json:"items"`
}

// Category is a multi-level product category (GET /v2/categories).
type Category struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// Handle is this category's own slug; FullHandle includes every parent,
	// e.g. "clothing/mens/shirts", and builds /categories/{full_handle}.
	Handle     string `json:"handle"`
	FullHandle string `json:"full_handle"`
}

// ProductFeedName identifies a product feed's target platform.
type ProductFeedName string

// Known ProductFeedName values (name of GET /v2/product_feeds entries).
const (
	ProductFeedFacebook   ProductFeedName = "facebook"   // Facebook catalog feed
	ProductFeedGoogle     ProductFeedName = "google"     // Google Merchant Center feed
	ProductFeedShopDotCom ProductFeedName = "shopdotcom" // SHOP.COM (美安) feed
	ProductFeedLine       ProductFeedName = "line"       // LINE Shopping feed
)

// ProductFeed is a product feed URL the shop publishes (GET /v2/product_feeds).
type ProductFeed struct {
	Name ProductFeedName `json:"name"`
	URL  string          `json:"url"`
	// MimeType is the feed format: text/csv, text/xml or application/json.
	MimeType string `json:"mime_type"`
}

// ShopEmails is the shop's contact address and the addresses that receive
// order notifications (GET /v2/shop_emails/shop_emails).
type ShopEmails struct {
	ShopEmail  string   `json:"shop_email"`
	Subscribes []string `json:"subscribes"`
}
