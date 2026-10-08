// Command seed creates synthetic test data in a CYBERBIZ test shop:
// products (with variants), customers, discounts, shop coupons, custom
// collections, a blog with articles, and pages.
//
// Every record is titled "[TEST] ..." and tagged "seed-<run id>" where the
// resource has tags. The ids of everything created are written to
// seed-<run id>.json for cleanup.
//
// It writes to a real shop, which cannot be undone, so it refuses to run
// unless CYBERBIZ_SHOP names the shop that owns CYBERBIZ_API_TOKEN:
//
//	export CYBERBIZ_API_TOKEN=...       # token of the test shop
//	export CYBERBIZ_SHOP=example.cyberbiz.co
//	go run ./cmd/seed -dry-run          # print the plan, no API calls
//	go run ./cmd/seed -n 5 -only products,customers
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/logging"
	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
	"github.com/rs/zerolog/log"
)

// maxPerKind caps -n so a typo cannot flood the shop.
const maxPerKind = 50

// options are the parsed flags and environment of one invocation.
type options struct {
	n        int
	kinds    kindSet
	dryRun   bool
	manifest string // manifest path; empty means seed-<run id>.json
	token    string
	shop     string
	baseURL  string // overrides the API host; used by tests
}

func main() {
	if err := logging.Setup("info"); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
	opts, err := parseOptions(os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := run(ctx, opts, time.Now(), os.Stdout); err != nil {
		log.Error().Err(err).Msg("seed failed")
		os.Exit(1)
	}
}

func parseOptions(args []string, getenv func(string) string) (options, error) {
	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	n := fs.Int("n", 3, "records to create per kind (max 50)")
	only := fs.String("only", "all", "comma-separated kinds: "+joinKinds(allKinds))
	dryRun := fs.Bool("dry-run", false, "print the plan as JSON and call no API")
	manifestPath := fs.String("manifest", "", "where to write created ids (default seed-<run id>.json)")
	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if *n < 1 || *n > maxPerKind {
		return options{}, fmt.Errorf("-n must be between 1 and %d", maxPerKind)
	}
	kinds, err := parseKinds(*only)
	if err != nil {
		return options{}, err
	}
	o := options{
		n: *n, kinds: kinds, dryRun: *dryRun, manifest: *manifestPath,
		token:   getenv("CYBERBIZ_API_TOKEN"),
		shop:    normalizeDomain(getenv("CYBERBIZ_SHOP")),
		baseURL: getenv("CYBERBIZ_BASE_URL"),
	}
	if !o.dryRun && (o.token == "" || o.shop == "") {
		return options{}, errors.New("CYBERBIZ_API_TOKEN and CYBERBIZ_SHOP must be set (or use -dry-run)")
	}
	return o, nil
}

func run(ctx context.Context, o options, now time.Time, out io.Writer) error {
	runID := newRunID(now)
	p := buildPlan(runID, o.n, o.kinds, now)
	if o.dryRun {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(p)
	}
	c, err := newClient(o)
	if err != nil {
		return err
	}
	if err := checkShop(ctx, c, o.shop); err != nil {
		return err
	}
	m := &manifest{RunID: runID, Shop: o.shop}
	log.Info().Str("shop", o.shop).Str("run_id", runID).Int("per_kind", o.n).Msg("seeding")
	seedErr := (&seeder{c: c, m: m}).run(ctx, p)
	path := o.manifest
	if path == "" {
		path = "seed-" + runID + ".json"
	}
	if err := writeManifest(path, m); err != nil {
		return errors.Join(seedErr, err)
	}
	log.Info().Str("manifest", path).Msg("wrote ids of created records")
	return seedErr
}

func newClient(o options) (*cyberbiz.Client, error) {
	opts := []cyberbiz.Option{cyberbiz.WithLogger(logging.Slog(log.Logger))}
	if o.baseURL != "" {
		opts = append(opts, cyberbiz.WithBaseURL(o.baseURL))
	}
	c, err := cyberbiz.New(o.token, opts...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	return c, nil
}

// checkShop refuses to continue unless the token belongs to the shop the
// caller named, so a token for another shop is never written to by mistake.
func checkShop(ctx context.Context, c *cyberbiz.Client, want string) error {
	info, _, err := c.Shop.Info(ctx)
	if err != nil {
		return fmt.Errorf("read shop info: %w", err)
	}
	got := normalizeDomain(info.PrimaryDomain)
	if got != want {
		return fmt.Errorf("token belongs to shop %q, not CYBERBIZ_SHOP %q; refusing to write", got, want)
	}
	return nil
}

// normalizeDomain strips scheme, path and case so "https://Shop.cyberbiz.co/"
// and "shop.cyberbiz.co" compare equal.
func normalizeDomain(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	return s
}

func writeManifest(path string, m *manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}
