package cyberbiz

import (
	"testing"
)

func TestContentGoldenMenus(t *testing.T) {
	var menus []Menu
	decodeGolden(t, "v2/GET_v2_menus.json", &menus)
	if len(menus) != 2 || menus[0].ID != 121299 || menus[0].Title != "主選單" || !menus[1].SystemDefault {
		t.Errorf("menus = %+v", menus)
	}
	var menu Menu
	decodeGolden(t, "v2/GET_v2_menus_{id}.json", &menu)
	if menu.ID != 121299 || menu.Items == nil || len(menu.Items) != 0 {
		t.Errorf("menu = %+v", menu)
	}

	c := goldenServer(t)
	page, err := c.Content.ListMenus(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[1].ID != 121300 {
		t.Errorf("page = %+v", page.Items)
	}
	got, _, err := c.Content.GetMenu(testCtx, 121299)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "主選單" {
		t.Errorf("menu = %+v", got)
	}
}

func TestContentMenuItemsDecodeRecursively(t *testing.T) {
	c, _ := New("tok")
	var menu Menu
	body := `{"id":1221,"title":"主選單","system_default":true,"items":[
		{"id":8357,"title":"最新動態","subject_handle":null,"url":"/blogs/news","position":3,"subtitle":null,"icon_image_url":null,
		 "start_at":"2026-01-01 00:00:00","end_at":null,"item_type":"http","level":1,"items":[
			{"id":12073,"title":"托特包","subject_handle":"tote","url":"/products/tote","position":1,"subtitle":null,"icon_image_url":null,
			 "item_type":"product","level":2,"items":[]}]}]}`
	if err := c.decode([]byte(body), &menu); err != nil {
		t.Fatal(err)
	}
	top := menu.Items[0]
	if top.ItemType != MenuItemTypeHTTP || top.Level != 1 || top.StartAt.IsZero() || !top.EndAt.IsZero() {
		t.Errorf("item = %+v", top)
	}
	if len(top.Items) != 1 || top.Items[0].ItemType != MenuItemTypeProduct || top.Items[0].SubjectHandle != "tote" || top.Items[0].Level != 2 {
		t.Errorf("child = %+v", top.Items)
	}
}

func TestContentGoldenCategories(t *testing.T) {
	var cats []Category
	decodeGolden(t, "v2/GET_v2_categories.json", &cats)
	if len(cats) != 2 || cats[0].ID != 19033 || cats[0].Handle != "ecoupon" || cats[1].FullHandle != "entea" {
		t.Errorf("categories = %+v", cats)
	}
	c := goldenServer(t)
	page, err := c.Content.ListCategories(testCtx, &CategoryListOptions{Q: "茶"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[1].Title != "英茶香" {
		t.Errorf("page = %+v", page.Items)
	}
	var n int
	for _, err := range c.Content.AllCategories(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 2 {
		t.Errorf("walked %d", n)
	}
}

func TestContentGoldenProductFeedsAndShopEmails(t *testing.T) {
	c := goldenServer(t)
	feeds, _, err := c.Content.ProductFeeds(testCtx)
	if err != nil {
		t.Fatal(err)
	}
	if len(feeds) != 4 || feeds[2].MimeType != "text/xml" || feeds[3].MimeType != "application/json" {
		t.Errorf("feeds = %+v", feeds)
	}
	if feeds[0].URL != "https://example.cyberbiz.co/products/fbcatalog/482b9470305c013ccc6f0a430b9d5f3d" {
		t.Errorf("url = %q", feeds[0].URL)
	}
	emails, _, err := c.Content.ShopEmails(testCtx)
	if err != nil {
		t.Fatal(err)
	}
	if emails.ShopEmail != "redacted@example.com" || len(emails.Subscribes) != 1 {
		t.Errorf("emails = %+v", emails)
	}
}

func TestContentRequests(t *testing.T) {
	lc, lcall := discountsSpyClient(t, 200, `[]`)
	if _, err := lc.Content.ListMenus(testCtx, &ListOptions{Page: 2, PerPage: 10}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "GET", "/v2/menus")
	if lcall.Query.Get("page") != "2" || lcall.Query.Get("per_page") != "10" {
		t.Errorf("query = %v", lcall.Query)
	}
	if _, err := lc.Content.ListCategories(testCtx, &CategoryListOptions{ListOptions: ListOptions{PerPage: 4}, Q: "多層級"}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, lcall, "GET", "/v2/categories")
	if lcall.Query.Get("q") != "多層級" || lcall.Query.Get("per_page") != "4" {
		t.Errorf("query = %v", lcall.Query)
	}

	c, call := discountsSpyClient(t, 200, `{"product_feeds":[],"shop_email":"a@b.c","subscribes":[]}`)
	if _, _, err := c.Content.ProductFeeds(testCtx); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v2/product_feeds")
	if _, _, err := c.Content.ShopEmails(testCtx); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v2/shop_emails/shop_emails")
	if _, _, err := c.Content.GetMenu(testCtx, 5); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v2/menus/5")
}

func TestContentPageRequests(t *testing.T) {
	const pageBody = `{"id":1,"title":"Demo Page","handle":"demo-page","status":"published",
		"html_infos":"{\"1714492800_app_store_api\":\"\\u003ch1\\u003eDemo content\\u003c/h1\\u003e\"}",
		"created_at":"2024-05-01 00:00:00","updated_at":"2024-05-01 00:00:00"}`
	c, call := discountsSpyClient(t, 201, pageBody)
	page, _, err := c.Content.CreatePage(testCtx, &PageCreateRequest{Title: "Demo Page"})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "POST", "/v2/pages")
	discountsAssertJSONBody(t, call, `{"title":"Demo Page"}`)
	if page.ID != 1 || page.Status != PageStatusPublished || page.Handle != "demo-page" {
		t.Errorf("page = %+v", page)
	}
	sections, err := page.HTMLSections()
	if err != nil {
		t.Fatal(err)
	}
	if sections["1714492800_app_store_api"] != "<h1>Demo content</h1>" {
		t.Errorf("sections = %q", sections)
	}

	title := "New Title"
	section := "1714492800_app_store_api"
	html := "<h1>New content</h1>"
	if _, _, err := c.Content.UpdatePage(testCtx, 1, &PageUpdateRequest{Title: &title, Status: PageStatusUnpublished, SectionID: &section, SectionHTML: &html}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "PUT", "/v2/pages/1")
	// section_content is the HTML as a JSON string with <, > and & escaped,
	// exactly as the reference example shows.
	discountsAssertJSONBody(t, call, `{"title":"New Title","status":"unpublished","section_id":"1714492800_app_store_api",
		"section_content":"\"\\u003ch1\\u003eNew content\\u003c/h1\\u003e\""}`)
	if string(call.Body) != `{"title":"New Title","status":"unpublished","section_id":"1714492800_app_store_api","section_content":"\"\\u003ch1\\u003eNew content\\u003c/h1\\u003e\""}` {
		t.Errorf("raw body = %s", call.Body)
	}

	empty := ""
	if _, _, err := c.Content.UpdatePage(testCtx, 1, &PageUpdateRequest{Handle: &empty}); err != nil {
		t.Fatal(err)
	}
	discountsAssertJSONBody(t, call, `{"handle":""}`)
}
