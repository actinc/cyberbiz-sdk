package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
	"github.com/rs/zerolog/log"
)

// manifest records the id of every record a run created, so a failed or
// finished run can be cleaned up.
type manifest struct {
	RunID       string  `json:"run_id"`
	Shop        string  `json:"shop"`
	Products    []int64 `json:"products,omitempty"`
	Variants    []int64 `json:"variants,omitempty"`
	Customers   []int64 `json:"customers,omitempty"`
	Discounts   []int64 `json:"discounts,omitempty"`
	Coupons     []int64 `json:"coupons,omitempty"`
	Collections []int64 `json:"collections,omitempty"`
	Blogs       []int64 `json:"blogs,omitempty"`
	Articles    []int64 `json:"articles,omitempty"`
	Pages       []int64 `json:"pages,omitempty"`
}

// seeder creates the records of a plan through one client. It stops at the
// first error; m keeps whatever was created before it.
type seeder struct {
	c *cyberbiz.Client
	m *manifest
}

// run creates every kind in order. A failing kind stops only itself, so one
// shop-specific rule (e.g. a required field) does not block the other kinds;
// the errors of all failed kinds are returned together.
func (s *seeder) run(ctx context.Context, p plan) error {
	steps := []func(context.Context, plan) error{
		s.products, s.customers, s.discounts, s.coupons,
		s.collections, s.blogs, s.pages,
	}
	var errs []error
	for _, step := range steps {
		if err := step(ctx, p); err != nil {
			log.Error().Err(err).Msg("kind failed; continuing with the next kind")
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// productCreateBody adds the default variant's SKU to the documented product
// body. The field is not in the v1 swagger, but shops with POS reject a
// product whose default variant has no SKU.
type productCreateBody struct {
	cyberbiz.ProductCreateRequest
	SKU string `json:"sku,omitzero"`
}

func (s *seeder) createProduct(ctx context.Context, pp productPlan) (*cyberbiz.Product, error) {
	var prod cyberbiz.Product
	req := &cyberbiz.Request{
		Method: http.MethodPost,
		Path:   "v1/products",
		Body:   productCreateBody{ProductCreateRequest: pp.Product, SKU: pp.SKU},
	}
	if _, err := s.c.Do(ctx, req, &prod); err != nil {
		return nil, err
	}
	return &prod, nil
}

func (s *seeder) products(ctx context.Context, p plan) error {
	for _, pp := range p.Products {
		prod, err := s.createProduct(ctx, pp)
		if err != nil {
			return fmt.Errorf("create product %q: %w", pp.Product.Handle, err)
		}
		s.m.Products = append(s.m.Products, prod.ID)
		log.Info().Int64("id", prod.ID).Str("title", pp.Product.Title).Msg("created product")
		for _, v := range pp.Variants {
			pv, _, err := s.c.Products.CreateVariant(ctx, prod.ID, &v)
			if err != nil {
				return fmt.Errorf("create variant %s of product %d: %w", *v.SKU, prod.ID, err)
			}
			s.m.Variants = append(s.m.Variants, pv.ID)
		}
	}
	return nil
}

func (s *seeder) customers(ctx context.Context, p plan) error {
	for _, req := range p.Customers {
		cust, _, err := s.c.Customers.Create(ctx, &req)
		if err != nil {
			return fmt.Errorf("create customer %q: %w", req.Email, err)
		}
		s.m.Customers = append(s.m.Customers, cust.ID)
		log.Info().Int64("id", cust.ID).Str("email", req.Email).Msg("created customer")
	}
	return nil
}

func (s *seeder) discounts(ctx context.Context, p plan) error {
	for _, req := range p.Discounts {
		d, _, err := s.c.Discounts.Create(ctx, &req)
		if err != nil {
			return fmt.Errorf("create discount %q: %w", req.Name, err)
		}
		s.m.Discounts = append(s.m.Discounts, d.ID)
		log.Info().Int64("id", d.ID).Str("name", req.Name).Msg("created discount")
	}
	return nil
}

func (s *seeder) coupons(ctx context.Context, p plan) error {
	for _, req := range p.Coupons {
		cp, _, err := s.c.Discounts.CreateCoupon(ctx, &req)
		if err != nil {
			return fmt.Errorf("create coupon %q: %w", req.Code, err)
		}
		s.m.Coupons = append(s.m.Coupons, cp.ID)
		log.Info().Int64("id", cp.ID).Str("code", req.Code).Msg("created coupon")
	}
	return nil
}

// collections creates the custom collections and puts every product of this
// run into each of them.
func (s *seeder) collections(ctx context.Context, p plan) error {
	for _, req := range p.Collections {
		col, _, err := s.c.Collections.CreateCustom(ctx, &req)
		if err != nil {
			return fmt.Errorf("create collection %q: %w", req.Handle, err)
		}
		s.m.Collections = append(s.m.Collections, col.ID)
		log.Info().Int64("id", col.ID).Str("title", req.Title).Msg("created collection")
		if len(s.m.Products) == 0 {
			continue
		}
		if _, _, err := s.c.Collections.AddCustomProducts(ctx, col.ID, s.m.Products); err != nil {
			return fmt.Errorf("add products to collection %d: %w", col.ID, err)
		}
	}
	return nil
}

func (s *seeder) blogs(ctx context.Context, p plan) error {
	for _, bp := range p.Blogs {
		blog, _, err := s.c.Blogs.Create(ctx, &bp.Blog)
		if err != nil {
			return fmt.Errorf("create blog %q: %w", bp.Blog.Title, err)
		}
		s.m.Blogs = append(s.m.Blogs, blog.ID)
		log.Info().Int64("id", blog.ID).Str("title", bp.Blog.Title).Msg("created blog")
		for _, a := range bp.Articles {
			art, _, err := s.c.Blogs.CreateArticle(ctx, blog.ID, &a)
			if err != nil {
				return fmt.Errorf("create article %q: %w", a.Title, err)
			}
			s.m.Articles = append(s.m.Articles, art.ID)
		}
	}
	return nil
}

func (s *seeder) pages(ctx context.Context, p plan) error {
	for _, req := range p.Pages {
		pg, _, err := s.c.Content.CreatePage(ctx, &req)
		if err != nil {
			return fmt.Errorf("create page %q: %w", req.Title, err)
		}
		s.m.Pages = append(s.m.Pages, pg.ID)
		log.Info().Int64("id", pg.ID).Str("title", req.Title).Msg("created page")
	}
	return nil
}
