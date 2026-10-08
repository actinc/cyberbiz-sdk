package main

import (
	"fmt"
	"slices"
	"strings"
)

// kind is one type of record the tool can create.
type kind string

const (
	kindProducts    kind = "products"
	kindCustomers   kind = "customers"
	kindDiscounts   kind = "discounts"
	kindCoupons     kind = "coupons"
	kindCollections kind = "collections"
	kindBlogs       kind = "blogs"
	kindPages       kind = "pages"
)

// allKinds lists every kind in creation order: products come first so
// collections can link them.
var allKinds = []kind{
	kindProducts, kindCustomers, kindDiscounts, kindCoupons,
	kindCollections, kindBlogs, kindPages,
}

// kindSet is the set of kinds selected for one run.
type kindSet map[kind]bool

func (s kindSet) has(k kind) bool { return s[k] }

// parseKinds parses the -only flag: "all" or a comma-separated list of kinds.
func parseKinds(raw string) (kindSet, error) {
	set := kindSet{}
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "all" {
		for _, k := range allKinds {
			set[k] = true
		}
		return set, nil
	}
	for part := range strings.SplitSeq(raw, ",") {
		k := kind(strings.TrimSpace(part))
		if !slices.Contains(allKinds, k) {
			return nil, fmt.Errorf("unknown kind %q (valid: all, %s)", k, joinKinds(allKinds))
		}
		set[k] = true
	}
	return set, nil
}

func joinKinds(ks []kind) string {
	names := make([]string, len(ks))
	for i, k := range ks {
		names[i] = string(k)
	}
	return strings.Join(names, ", ")
}
