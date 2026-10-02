package shops

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/crypto"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/response"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/scope"
	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
	"github.com/rs/zerolog/log"
)

// Options tune how SDK clients are built.
type Options struct {
	// Transport records every attempt; it receives the shop id via scope.
	Transport http.RoundTripper
	// BaseURL overrides the CYBERBIZ host (CYBERBIZ_BASE_URL); empty keeps
	// the SDK default.
	BaseURL string
	// PublicURL is CONSOLE_PUBLIC_URL, used to build webhook_url.
	PublicURL string
	// Logger is handed to the SDK through cyberbiz.WithLogger.
	Logger *slog.Logger
	// RateLimit overrides the SDK's client-side limit; nil keeps the default.
	RateLimit *float64
}

// Service manages shops and their SDK clients.
type Service struct {
	repo   *Repository
	cipher *crypto.Cipher
	opts   Options

	mu    sync.Mutex
	cache map[uint]cacheEntry
}

type cacheEntry struct {
	updatedAt time.Time
	client    *cyberbiz.Client
}

// NewService wires the repository, cipher and client options.
func NewService(repo *Repository, cipher *crypto.Cipher, opts Options) *Service {
	return &Service{repo: repo, cipher: cipher, opts: opts, cache: map[uint]cacheEntry{}}
}

// Input is the create/update body. On update every field is optional and a
// blank secret or token keeps the stored one.
type Input struct {
	Name       string `json:"name"`
	ShopDomain string `json:"shop_domain"`
	AppName    string `json:"app_name"`
	AppID      string `json:"app_id"`
	AppSecret  string `json:"app_secret"`
	APIToken   string `json:"api_token"`
}

// List returns every shop.
func (s *Service) List(ctx context.Context) ([]db.Shop, error) { return s.repo.List(ctx) }

// Count returns how many shops exist.
func (s *Service) Count(ctx context.Context) (int64, error) { return s.repo.Count(ctx) }

// Get returns one shop or a 404 error.
func (s *Service) Get(ctx context.Context, id uint) (*db.Shop, error) {
	shop, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if shop == nil {
		return nil, response.NotFound("shop not found")
	}
	return shop, nil
}

// Create stores the shop, then validates the credentials with GET /shop.
// The row is removed again when validation fails.
func (s *Service) Create(ctx context.Context, in Input) (*db.Shop, *cyberbiz.ShopInfo, error) {
	if err := validateNew(in); err != nil {
		return nil, nil, err
	}
	shop := &db.Shop{Name: in.Name, ShopDomain: strings.TrimSpace(in.ShopDomain), AppName: in.AppName, AppID: in.AppID}
	if err := s.applySecrets(shop, in); err != nil {
		return nil, nil, err
	}
	if err := s.repo.Create(ctx, shop); err != nil {
		return nil, nil, mapErr(err)
	}
	info, err := s.Verify(ctx, shop)
	if err != nil {
		if delErr := s.repo.Delete(ctx, shop.ID); delErr != nil {
			log.Error().Err(delErr).Uint("shop_id", shop.ID).Msg("removing unverified shop")
		}
		return nil, nil, err
	}
	log.Info().Uint("shop_id", shop.ID).Str("shop_domain", shop.ShopDomain).Msg("shop created")
	return shop, info, nil
}

// Update changes the given fields and re-verifies when credentials changed.
func (s *Service) Update(ctx context.Context, id uint, in Input) (*db.Shop, error) {
	shop, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	updated := *shop
	setIf(&updated.Name, in.Name)
	setIf(&updated.ShopDomain, strings.TrimSpace(in.ShopDomain))
	setIf(&updated.AppName, in.AppName)
	setIf(&updated.AppID, in.AppID)
	if err := s.applySecrets(&updated, in); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, &updated); err != nil {
		return nil, mapErr(err)
	}
	if in.APIToken != "" {
		if _, err := s.Verify(ctx, &updated); err != nil {
			return nil, err
		}
	}
	log.Info().Uint("shop_id", id).Msg("shop updated")
	return &updated, nil
}

// Delete removes the shop and drops its cached client.
func (s *Service) Delete(ctx context.Context, id uint) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.cache, id)
	s.mu.Unlock()
	log.Info().Uint("shop_id", id).Msg("shop deleted")
	return nil
}

// Verify calls GET /shop and stores cyberbiz_shop_id and custom_domain.
func (s *Service) Verify(ctx context.Context, shop *db.Shop) (*cyberbiz.ShopInfo, error) {
	client, err := s.Client(shop)
	if err != nil {
		return nil, err
	}
	ctx, _ = scope.New(ctx, shop.ID)
	info, _, err := client.Shop.Info(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	if info == nil {
		return nil, response.Upstream(&cyberbiz.APIError{StatusCode: 200, Method: "GET", Path: "shop", Messages: []string{"empty shop_info"}})
	}
	shop.CyberbizShopID = info.ID
	shop.CustomDomain = info.PrimaryDomain
	if err := s.repo.Save(ctx, shop); err != nil {
		return nil, mapErr(err)
	}
	return info, nil
}

// Client returns the cached SDK client for shop, rebuilding it when the
// shop row changed since it was built.
func (s *Service) Client(shop *db.Shop) (*cyberbiz.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.cache[shop.ID]; ok && e.updatedAt.Equal(shop.UpdatedAt) {
		return e.client, nil
	}
	client, err := s.build(shop)
	if err != nil {
		return nil, err
	}
	s.cache[shop.ID] = cacheEntry{updatedAt: shop.UpdatedAt, client: client}
	return client, nil
}

// ClientByID loads the shop and returns its client.
func (s *Service) ClientByID(ctx context.Context, id uint) (*cyberbiz.Client, *db.Shop, error) {
	shop, err := s.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	client, err := s.Client(shop)
	return client, shop, err
}

func (s *Service) build(shop *db.Shop) (*cyberbiz.Client, error) {
	token, err := s.cipher.Decrypt(shop.APITokenEnc)
	if err != nil {
		return nil, fmt.Errorf("shops: decrypting token: %w", err)
	}
	opts := []cyberbiz.Option{cyberbiz.WithTransport(s.opts.Transport)}
	if s.opts.Logger != nil {
		opts = append(opts, cyberbiz.WithLogger(s.opts.Logger))
	}
	if s.opts.BaseURL != "" {
		opts = append(opts, cyberbiz.WithBaseURL(s.opts.BaseURL))
	}
	if s.opts.RateLimit != nil {
		opts = append(opts, cyberbiz.WithRateLimit(*s.opts.RateLimit))
	}
	client, err := cyberbiz.New(string(token), opts...)
	if err != nil {
		return nil, response.Validation(err.Error())
	}
	return client, nil
}

// WebhookSecret implements webhook.SecretResolver against the shops table.
func (s *Service) WebhookSecret(ctx context.Context, shopDomain string) (string, error) {
	shop, err := s.repo.GetByDomain(ctx, shopDomain)
	if err != nil {
		return "", err
	}
	if shop == nil {
		return "", errors.New("no shop with that domain")
	}
	secret, err := s.cipher.Decrypt(shop.AppSecretEnc)
	if err != nil {
		return "", err
	}
	return string(secret), nil
}

// ByDomain returns the shop for a Shop Domain; nil, nil when unknown.
func (s *Service) ByDomain(ctx context.Context, domain string) (*db.Shop, error) {
	return s.repo.GetByDomain(ctx, domain)
}

func (s *Service) applySecrets(shop *db.Shop, in Input) error {
	if in.AppSecret != "" {
		enc, err := s.cipher.Encrypt([]byte(in.AppSecret))
		if err != nil {
			return err
		}
		shop.AppSecretEnc = enc
	}
	if in.APIToken != "" {
		enc, err := s.cipher.Encrypt([]byte(in.APIToken))
		if err != nil {
			return err
		}
		shop.APITokenEnc = enc
		shop.TokenFingerprint = crypto.Fingerprint(in.APIToken)
	}
	return nil
}

func validateNew(in Input) error {
	var missing []string
	for name, v := range map[string]string{"name": in.Name, "shop_domain": in.ShopDomain, "api_token": in.APIToken} {
		if strings.TrimSpace(v) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)
		return response.Validation("missing required fields: " + strings.Join(missing, ", "))
	}
	return nil
}

func setIf(dst *string, v string) {
	if v != "" {
		*dst = v
	}
}

func mapErr(err error) error {
	if errors.Is(err, ErrDuplicateDomain) {
		return response.Conflict("a shop with that shop_domain already exists")
	}
	return err
}
