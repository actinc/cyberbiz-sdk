package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

// titlePrefix marks every record this tool creates so it is easy to find and
// delete from the shop admin.
const titlePrefix = "[TEST]"

// plan holds everything one run creates. Every value is synthetic and carries
// the run id, so two runs never collide on a handle, email or coupon code.
type plan struct {
	RunID       string
	Products    []productPlan
	Customers   []cyberbiz.CustomerCreateRequest
	Discounts   []cyberbiz.DiscountCreateRequest
	Coupons     []cyberbiz.ShopCouponCreateRequest
	Collections []cyberbiz.CustomCollectionCreateRequest
	Blogs       []blogPlan
	Pages       []cyberbiz.PageCreateRequest
}

// productPlan is a product plus the variants created under it. SKU goes on
// the product's default variant: shops with POS reject a product without one.
type productPlan struct {
	Product  cyberbiz.ProductCreateRequest
	SKU      string
	Variants []cyberbiz.ProductVariantCreateRequest
}

// blogPlan is a blog plus the articles created under it.
type blogPlan struct {
	Blog     cyberbiz.BlogCreateRequest
	Articles []cyberbiz.ArticleCreateRequest
}

// newRunID returns a short, sortable id for one run, e.g. "s20261008152104".
func newRunID(now time.Time) string {
	return "s" + now.In(cyberbiz.Taipei).Format("20060102150405")
}

// runTag is the tag added to every taggable record of a run.
func runTag(runID string) string { return "seed-" + runID }

// buildPlan generates n records of each selected kind.
func buildPlan(runID string, n int, kinds kindSet, now time.Time) plan {
	p := plan{RunID: runID}
	if kinds.has(kindProducts) {
		p.Products = buildProducts(runID, n)
	}
	if kinds.has(kindCustomers) {
		p.Customers = buildCustomers(runID, n)
	}
	if kinds.has(kindDiscounts) {
		p.Discounts = buildDiscounts(runID, n, now)
	}
	if kinds.has(kindCoupons) {
		p.Coupons = buildCoupons(runID, n, now)
	}
	if kinds.has(kindCollections) {
		p.Collections = buildCollections(runID, n)
	}
	if kinds.has(kindBlogs) {
		p.Blogs = buildBlogs(runID, n)
	}
	if kinds.has(kindPages) {
		p.Pages = buildPages(runID, n)
	}
	return p
}

func buildProducts(runID string, n int) []productPlan {
	out := make([]productPlan, 0, n)
	for i := 1; i <= n; i++ {
		price := cyberbiz.MoneyFromInt(int64(100 * i))
		out = append(out, productPlan{
			Product: cyberbiz.ProductCreateRequest{
				Title:     fmt.Sprintf("%s 測試商品 %d", titlePrefix, i),
				Handle:    fmt.Sprintf("test-%s-product-%d", runID, i),
				Published: false,
				Price:     price,
				BodyHTML:  ptr("<p>Synthetic product created by the seed tool.</p>"),
				Vendor:    ptr("Seed Vendor"),
				TagsText:  ptr(runTag(runID)),
			},
			SKU:      fmt.Sprintf("TEST-%s-%d", strings.ToUpper(runID), i),
			Variants: buildVariants(runID, i, price),
		})
	}
	return out
}

func buildVariants(runID string, productNo int, price cyberbiz.Money) []cyberbiz.ProductVariantCreateRequest {
	sizes := []string{"S", "M", "L"}
	out := make([]cyberbiz.ProductVariantCreateRequest, 0, len(sizes))
	for i, size := range sizes {
		out = append(out, cyberbiz.ProductVariantCreateRequest{
			Position:            i + 1,
			InventoryManagement: true,
			InventoryQuantity:   10 * (i + 1),
			InventoryPolicy:     cyberbiz.InventoryPolicyDeny,
			RequiresShipping:    true,
			Price:               ptr(price),
			Option1:             ptr(size),
			SKU:                 ptr(fmt.Sprintf("TEST-%s-%d-%s", strings.ToUpper(runID), productNo, size)),
		})
	}
	return out
}

func buildCustomers(runID string, n int) []cyberbiz.CustomerCreateRequest {
	out := make([]cyberbiz.CustomerCreateRequest, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, cyberbiz.CustomerCreateRequest{
			Name:             fmt.Sprintf("%s 測試顧客 %d", titlePrefix, i),
			Status:           cyberbiz.CustomerStatusEnabled,
			Email:            fmt.Sprintf("seed+%s-%d@example.com", runID, i),
			TagsText:         runTag(runID),
			AcceptsMarketing: ptr(false),
			Note:             "Synthetic customer created by the seed tool.",
		})
	}
	return out
}

func buildDiscounts(runID string, n int, now time.Time) []cyberbiz.DiscountCreateRequest {
	out := make([]cyberbiz.DiscountCreateRequest, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, cyberbiz.DiscountCreateRequest{
			Name:         fmt.Sprintf("%s 測試折扣 %s-%d", titlePrefix, runID, i),
			DiscountType: cyberbiz.DiscountTypeAmount,
			Value:        cyberbiz.MoneyFromInt(int64(50 * i)),
			Threshold:    cyberbiz.MoneyFromInt(1000),
			StartAt:      cyberbiz.NewTime(now),
			EndAt:        cyberbiz.NewTime(now.AddDate(0, 0, 30)),
		})
	}
	return out
}

func buildCoupons(runID string, n int, now time.Time) []cyberbiz.ShopCouponCreateRequest {
	start := now.In(cyberbiz.Taipei)
	end := start.AddDate(0, 0, 30)
	out := make([]cyberbiz.ShopCouponCreateRequest, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, cyberbiz.ShopCouponCreateRequest{
			Title:               fmt.Sprintf("%s 測試優惠券 %d", titlePrefix, i),
			Code:                fmt.Sprintf("TEST%s%d", strings.ToUpper(runID), i),
			CouponType:          cyberbiz.CouponTypeAmount,
			Value:               cyberbiz.MoneyFromInt(100),
			OrderPriceThreshold: cyberbiz.MoneyFromInt(500),
			StartDate:           cyberbiz.NewDate(start.Year(), start.Month(), start.Day()),
			EndDate:             cyberbiz.NewDate(end.Year(), end.Month(), end.Day()),
			UsageLimit:          10,
			Tags:                []string{runTag(runID)},
		})
	}
	return out
}

func buildCollections(runID string, n int) []cyberbiz.CustomCollectionCreateRequest {
	out := make([]cyberbiz.CustomCollectionCreateRequest, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, cyberbiz.CustomCollectionCreateRequest{
			Title:     fmt.Sprintf("%s 測試分類 %d", titlePrefix, i),
			Handle:    fmt.Sprintf("test-%s-collection-%d", runID, i),
			Published: false,
		})
	}
	return out
}

func buildBlogs(runID string, n int) []blogPlan {
	articles := make([]cyberbiz.ArticleCreateRequest, 0, n)
	for i := 1; i <= n; i++ {
		articles = append(articles, cyberbiz.ArticleCreateRequest{
			Title:     fmt.Sprintf("%s 測試文章 %d", titlePrefix, i),
			BodyHTML:  ptr("<p>Synthetic article created by the seed tool.</p>"),
			Author:    ptr("Seed Bot"),
			Published: ptr(false),
			TagsText:  ptr(runTag(runID)),
		})
	}
	return []blogPlan{{
		Blog: cyberbiz.BlogCreateRequest{
			Title:  fmt.Sprintf("%s 測試部落格 %s", titlePrefix, runID),
			Handle: ptr(fmt.Sprintf("test-%s-blog", runID)),
		},
		Articles: articles,
	}}
}

func buildPages(runID string, n int) []cyberbiz.PageCreateRequest {
	out := make([]cyberbiz.PageCreateRequest, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, cyberbiz.PageCreateRequest{
			Title: fmt.Sprintf("%s 測試頁面 %s-%d", titlePrefix, runID, i),
		})
	}
	return out
}

func ptr[T any](v T) *T { return &v }
