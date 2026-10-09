package cyberbiz

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Version is the SDK version reported in the User-Agent header.
const Version = "0.3.0"

const (
	// DefaultBaseURL is the CYBERBIZ API host shared by every shop.
	DefaultBaseURL = "https://app-store-api.cyberbiz.io/"

	// DefaultRateLimit is the platform limit of requests per second.
	DefaultRateLimit = 5.0

	// DefaultMaxRetries is the number of times a request is retried after a
	// 429 or a transient 5xx response.
	DefaultMaxRetries = 3

	// DefaultTimeout is the per-request timeout of the default HTTP client.
	DefaultTimeout = 30 * time.Second

	defaultUserAgent = "cyberbiz-sdk-go/" + Version
)

// Client talks to the CYBERBIZ API on behalf of exactly one shop. It is safe
// for concurrent use and must be created with [New]. A Client is immutable:
// every setting is fixed by the [Option] values given to New.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
	limiter    *rate.Limiter // nil when rate limiting is disabled
	maxRetries int
	backoff    BackoffFunc
	logger     *slog.Logger
	strictJSON bool
	userAgent  string
	sleep      func(time.Duration) // swapped in tests

	// Services, one per API resource group. All share this client's auth,
	// rate limiting, retries and transport.
	Shop         *ShopService         // /shop, /settings
	Orders       *OrdersService       // /v1/orders, /v1/order_etickets, /v2/orders
	Customers    *CustomersService    // /v1/customers, /v1/customer_groups, /v2/customers, /v2/customer_oauth
	Products     *ProductsService     // /v1/products, /v2/products
	Collections  *CollectionsService  // /v1/*_collections, /v2/*_collections
	BranchStores *BranchStoresService // /v1/branch_stores
	PosShops     *PosShopsService     // /v1/pos_shops, /v1/pos_shop_coupons, /v2/pos_wallets
	Stock        *StockService        // /v1/stock_*, /v1/inventory_sync_groups
	Discounts    *DiscountsService    // /v1/discounts, /v1/shop_coupons
	VIPGroups    *VIPGroupsService    // /v1/vip_groups
	BonusRule    *BonusRuleService    // /v1/bonus_rule
	Einvoices    *EinvoicesService    // /v1/einvoices, /v1/offline_einvoices
	CustomFields *CustomFieldsService // /v1/custom_fields, /v1/custom_field_types
	Assets       *AssetsService       // /v1/assets
	Periodic     *PeriodicService     // /v1/periodic_orders
	Blogs        *BlogsService        // /v1/blogs
	Content      *ContentService      // /v2/pages, /v2/menus, /v2/categories, /v2/product_feeds, /v2/shop_emails
	Affiliates   *AffiliatesService   // /v2/affiliate_vendor_orders
}

// New returns a Client for the shop that owns token.
func New(token string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("cyberbiz: token must not be empty")
	}
	cfg := defaultConfig()
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	base, err := url.Parse(cfg.baseURL)
	if err != nil {
		return nil, errors.New("cyberbiz: invalid base URL: " + err.Error())
	}
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}

	c := &Client{
		baseURL:    base,
		token:      token,
		httpClient: cfg.httpClient,
		maxRetries: cfg.maxRetries,
		backoff:    cfg.backoff,
		logger:     cfg.logger,
		strictJSON: cfg.strictJSON,
		userAgent:  cfg.userAgent,
		sleep:      time.Sleep,
	}
	if cfg.rateLimit > 0 {
		c.limiter = rate.NewLimiter(rate.Limit(cfg.rateLimit), 1)
	}
	c.initServices()
	return c, nil
}

// BaseURL returns the API host this client sends requests to.
func (c *Client) BaseURL() *url.URL {
	u := *c.baseURL
	return &u
}

// initServices wires every resource service to the client.
func (c *Client) initServices() {
	c.Shop = &ShopService{client: c}
	c.Orders = &OrdersService{client: c}
	c.Customers = &CustomersService{client: c}
	c.Products = &ProductsService{client: c}
	c.Collections = &CollectionsService{client: c}
	c.BranchStores = &BranchStoresService{client: c}
	c.PosShops = &PosShopsService{client: c}
	c.Stock = &StockService{client: c}
	c.Discounts = &DiscountsService{client: c}
	c.VIPGroups = &VIPGroupsService{client: c}
	c.BonusRule = &BonusRuleService{client: c}
	c.Einvoices = &EinvoicesService{client: c}
	c.CustomFields = &CustomFieldsService{client: c}
	c.Assets = &AssetsService{client: c}
	c.Periodic = &PeriodicService{client: c}
	c.Blogs = &BlogsService{client: c}
	c.Content = &ContentService{client: c}
	c.Affiliates = &AffiliatesService{client: c}
}
