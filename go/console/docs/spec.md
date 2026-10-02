# Console specification

The Console is the local web application that sends Outbound requests to the
CYBERBIZ API through the SDK, receives Inbound webhooks, stores both, and lets a
person browse them and turn them into Golden Files. Vocabulary is defined in
the repository `CONTEXT.md`; this document is the contract between the Go
backend (`console/internal`, `console/cmd/console`) and the React UI
(`console/ui`). It is an observation tool: it never forwards, retries, or
replays traffic.

## Stack

- Backend: Go 1.27, Gin, GORM with `github.com/glebarez/sqlite` (pure Go, no
  cgo), zerolog for logs, `golang.org/x/crypto/bcrypt` for passwords,
  AES-256-GCM (standard library) for credentials at rest, `encoding/json/v2`.
- SDK: `github.com/actinc/cyberbiz-sdk/go/cyberbiz` (client) and
  `.../webhook` (inbound parsing); the console module has a `replace` to `../`.
- UI: React 19, TypeScript strict (no `any`), Mantine 8 (`@mantine/core`,
  `@mantine/hooks`, `@mantine/form`, `@mantine/notifications`), TanStack
  Query 5, react-router 7, Vite. Built into `console/internal/ui/dist` and
  served by the Go binary through `go:embed`.
- One binary: `go run ./cmd/console` serves the API, the webhook endpoint,
  and the UI on one port.

## Configuration (environment variables)

| Variable | Default | Meaning |
|---|---|---|
| `CONSOLE_ADDR` | `:8787` | listen address |
| `CONSOLE_DB_PATH` | `./console.db` | SQLite file |
| `CONSOLE_ADMIN_USER` | `admin` | admin username created on first boot |
| `CONSOLE_ADMIN_PASSWORD` | (generated) | admin password on first boot; when unset a random one is generated and printed once to stdout |
| `CONSOLE_ENCRYPTION_KEY` | (required) | 32 bytes, hex or base64; encrypts API Tokens and App Secrets at rest. Startup fails with instructions (`openssl rand -hex 32`) when missing |
| `CONSOLE_SESSION_TTL` | `24h` | login session lifetime |
| `CONSOLE_LOG_RETENTION_DAYS` | `30` | Outbound/Inbound rows older than this are deleted at boot and daily |
| `CONSOLE_GOLDEN_DIR` | `../testdata/golden` | where Outbound Golden Files are written |
| `CONSOLE_WEBHOOK_GOLDEN_DIR` | `../webhook/testdata` | where Inbound Golden Files are written |
| `CONSOLE_PUBLIC_URL` | (empty) | public base URL (e.g. an ngrok URL) shown in the UI so a person can paste the webhook URL into CYBERBIZ |
| `CONSOLE_LOG_LEVEL` | `info` | zerolog level |
| `CYBERBIZ_BASE_URL` | SDK default | override for tests |

`.env` is loaded with `github.com/joho/godotenv` when present.

## Database (GORM models, explicit `TableName()`, snake_case plural)

```
users            id, username (unique), password_hash, created_at, updated_at
sessions         id (random 32-byte hex, PK), user_id, expires_at, created_at
shops            id, name, shop_domain (unique), custom_domain, cyberbiz_shop_id,
                 app_name, app_id, app_secret_enc (bytes), api_token_enc (bytes),
                 token_fingerprint ("sha256:<8 hex> len=N"), created_at, updated_at
outbound_logs    id, shop_id (FK), method, path, query, request_headers (json text,
                 redacted), request_body (text), response_status, response_headers
                 (json text), response_body (text), duration_ms, attempt, request_id,
                 error (text, nullable), created_at
inbound_logs     id, shop_id (FK, nullable), shop_domain, custom_domain, event,
                 signature, domain_signature, signature_valid (bool),
                 domain_signature_valid (bool, nullable when header absent),
                 status ("valid" | "invalid_signature" | "unknown_shop" | "malformed"),
                 headers (json text, redacted), body (text), remote_addr,
                 response_status, duplicate_of (FK self, nullable: an earlier row
                 with the same event and signature), created_at
```

Migrations: GORM `AutoMigrate` at boot. Indexes: outbound_logs(shop_id,
created_at), outbound_logs(path), inbound_logs(shop_domain, created_at),
inbound_logs(event), inbound_logs(signature).

Credentials: `api_token_enc` and `app_secret_enc` are AES-256-GCM
ciphertexts (12-byte nonce prefix). Decrypted values live only in memory in the
per-shop SDK client cache; the API never returns them, only
`token_fingerprint` and a boolean `has_app_secret`.

## HTTP API

All JSON. Envelope on every response:

```json
{ "success": true, "data": ... }
{ "success": false, "error": "human message", "code": 40401 }
```

Error codes: `40000` bad request, `40100` not logged in, `40300` forbidden,
`40400` not found, `40900` conflict (duplicate shop domain), `42200`
validation, `50000` internal, `50200` CYBERBIZ upstream error (the SDK's
`*APIError`, with `data` carrying `{status, messages, request_id}`).

Auth: cookie `console_session` (HttpOnly, SameSite=Lax, Secure when the
request is HTTPS), value = session id. Every `/api/*` route except
`/api/auth/login` and `/api/health` requires it. `/webhooks/cyberbiz` and the
UI are public.

### Auth

- `POST /api/auth/login` `{username, password}` → `{user:{id, username}}`
- `POST /api/auth/logout` → `{}`
- `GET /api/auth/me` → `{user, setup_required: bool}` (`setup_required` is
  true while no Shop exists; the UI shows the first-run Shop form)

### Shops

- `GET /api/shops` → `[Shop]`
- `POST /api/shops` `{name, shop_domain, app_name, app_id, app_secret, api_token}`
  → `Shop` (validates by calling `GET /shop` through the SDK and stores
  `cyberbiz_shop_id` and `custom_domain` from the reply; the validation call is
  recorded as an Outbound)
- `GET /api/shops/:id` → `Shop`
- `PUT /api/shops/:id` same body, every field optional; blank secret/token
  keeps the stored one
- `DELETE /api/shops/:id` → `{}` (logs are kept; `shop_id` stays as a dangling
  reference and the UI shows "deleted shop")
- `POST /api/shops/:id/verify` → `{shop_info}` (calls `GET /shop`)

`Shop` JSON: `{id, name, shop_domain, custom_domain, cyberbiz_shop_id,
app_name, app_id, has_app_secret, token_fingerprint, webhook_url, created_at,
updated_at}` where `webhook_url` = `CONSOLE_PUBLIC_URL + "/webhooks/cyberbiz"`
(empty when the public URL is unset).

### Catalog (API tester)

- `GET /api/catalog` → `{operations: [Operation], tags: [{name, description}]}`
  built at boot from the OpenAPI files embedded in
  `console/internal/catalog/spec/cyberbiz-openapi-v1.yaml` and `cyberbiz-openapi-v2.yaml`
  (copied from `docs/api/en/` by `go generate`). `Operation` = `{id, method,
  path, tag, summary, description, path_params: [Param], query_params:
  [Param], request_body: {content_type, schema, example} | null, responses:
  [{status, description, example}]}` and `Param` = `{name, in, required,
  type, description, enum, example}`.

### Requests (execute an Outbound)

- `POST /api/shops/:id/requests` `{method, path, query: {name: [values]},
  body: <json or null>, headers: {name: value}}` → `OutboundLog`. The request
  goes through the shop's SDK client `Do` (auth, rate limit, retries) with a
  recording `http.RoundTripper` supplied via `cyberbiz.WithTransport`; every
  attempt is stored as its own row (`attempt` 1..n) and the last one is
  returned. Binary responses (zip labels) are stored base64-encoded with
  `response_headers` carrying the content type; the UI offers them as a
  download link served by `GET /api/outbound/:id/body`.

### Outbound logs

- `GET /api/outbound?shop_id=&method=&path=&status=&from=&to=&q=&page=&per_page=`
  → `{items: [OutboundLogSummary], total, page, per_page}` (summary omits
  bodies)
- `GET /api/outbound/:id` → `OutboundLog` (full)
- `GET /api/outbound/:id/body` → raw response body with its original content
  type (for binary downloads)
- `POST /api/outbound/:id/golden` `{name?}` → `{path}`: writes the redacted
  response body to `CONSOLE_GOLDEN_DIR/<v1|v2|app>/<METHOD>_<path with numeric
  segments replaced by {id}><name suffix>.json` plus `.headers.json` (same
  format as `internal/tools/goldenimport`; use `internal/redact` from the SDK
  module — it is `internal`, so the console module CANNOT import it: the SDK
  exposes it as `cyberbiz.RedactJSON([]byte) ([]byte, error)` and
  `cyberbiz.RedactHeaders(http.Header) http.Header`). Refuses to overwrite an
  existing file unless `overwrite: true`.

### Inbound logs

- `POST /webhooks/cyberbiz` (public): reads the body (2 MB cap), looks up the
  Shop by `X-Cyberbiz-Domain`, verifies with `webhook.Parse` using a
  `SecretResolver` backed by the shops table; stores a row in every case:
  `valid` → 200 `{"ok":true}`; `invalid_signature` → 401; `unknown_shop` →
  401; `malformed` (missing headers, too large, bad JSON) → 400 / 413. Marks
  `duplicate_of` when an earlier row has the same `event` and `signature`
  (a redelivery; two different events may share a body). Never logs the
  body at info level.
- `GET /api/inbound?shop_id=&event=&status=&from=&to=&q=&page=&per_page=` →
  `{items: [InboundLogSummary], total, page, per_page}`
- `GET /api/inbound/:id` → `InboundLog`
- `POST /api/inbound/:id/golden` `{name?}` → `{path}`: writes the redacted
  body to `CONSOLE_WEBHOOK_GOLDEN_DIR/<event with / replaced by _><suffix>.json`
  and the redacted headers alongside as `.headers.json`.

### Misc

- `GET /api/health` → `{status: "ok", version}` (public)
- `GET /api/stats` → `{shops, outbound_24h, inbound_24h, inbound_invalid_24h}`

## UI pages (react-router)

- `/login`: username/password.
- `/setup`: shown when `setup_required`; the Shop form (name, shop domain,
  app name, app id, app secret, api token) with a "Verify" button that calls
  `POST /api/shops` and shows the returned shop info.
- `/`: dashboard with the stats and the webhook URL to paste into CYBERBIZ
  (with a note that a tunnel such as ngrok or cloudflared is needed locally).
- `/shops`: list, add, edit, delete, verify; shows the token fingerprint only.
- `/requests`: the API tester. Left: tag filter + searchable operation list
  from `/api/catalog`. Right: the selected operation's docs, a form for path
  and query parameters, a JSON editor (textarea with validation) prefilled
  with the example body, a Shop selector, "Send". Below: status, duration,
  headers, pretty-printed body, "Save as Golden File".
- `/outbound`: filterable table (time, shop, method, path, status, duration);
  row click opens a drawer with request/response headers and bodies (JSON
  pretty-printed, copy buttons) and "Save as Golden File".
- `/inbound`: filterable table (time, shop domain, event, status, duplicate
  badge); drawer with headers, body, verification result, "Save as Golden
  File".

Mobile: tables collapse to cards below `sm`. Theme: Mantine defaults, light
and dark. No charts, no export.

## Logging and safety

- zerolog structured logs: every Outbound at info (method, path, status,
  duration, shop id, request id); every Inbound at info (event, shop domain,
  status); tokens, secrets and bodies never at info or above.
- The SDK client per shop is built once and cached; rebuilt when the shop's
  credentials change. `cyberbiz.WithLogger` receives a `slog.Logger` bridged
  to zerolog.
- Retention job runs at boot and every 24h.

## Tests

- Backend: `httptest` integration tests against a temporary SQLite file for
  auth, shops (with a fake CYBERBIZ server), request execution and recording,
  webhook receipt for every status, golden export (to a temp dir), retention.
- UI: `vitest` unit tests for the API client and the catalog form builder;
  `tsc --noEmit` and ESLint must pass.
