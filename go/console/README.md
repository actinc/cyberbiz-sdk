# Console

Local web application for exercising the CYBERBIZ API through the SDK and
inspecting inbound webhooks. It sends Outbound requests, receives Inbound
webhooks, stores both in SQLite, and turns them into Golden Files. See the
repository `CONTEXT.md` for terms and `docs/spec.md` for the backend/UI
contract. It is its own Go module so the SDK stays dependency-free.

## Run

```sh
cd console
export CONSOLE_ENCRYPTION_KEY=$(openssl rand -hex 32)   # or put it in .env
go run ./cmd/console
```

On the first boot an `admin` user is created. When `CONSOLE_ADMIN_PASSWORD`
is unset a random password is generated and printed once to stdout; copy it,
open <http://localhost:8787>, and log in. The first-run screen asks for a
Shop: its CYBERBIZ Shop Domain (`example.cyberbiz.co`), App name/ID/secret,
and API Token. Saving calls `GET /shop` through the SDK to validate the token.

The React UI is served from `internal/ui/dist` through `go:embed`. The built
UI is committed, so `go run` works from a fresh clone without Node. After
changing anything under `console/ui`, rebuild it with `npm run build` in
`console/ui` (Node 22) and commit the new `dist/`.

## Environment variables

`.env` in the working directory is loaded automatically.

| Variable | Default | Meaning |
|---|---|---|
| `CONSOLE_ADDR` | `:8787` | listen address |
| `CONSOLE_DB_PATH` | `./console.db` | SQLite file |
| `CONSOLE_ADMIN_USER` | `admin` | admin username created on first boot |
| `CONSOLE_ADMIN_PASSWORD` | (generated) | admin password on first boot |
| `CONSOLE_ENCRYPTION_KEY` | (required) | 32 bytes, hex or base64; encrypts API Tokens and App Secrets at rest |
| `CONSOLE_SESSION_TTL` | `24h` | login session lifetime |
| `CONSOLE_LOG_RETENTION_DAYS` | `30` | Outbound/Inbound rows older than this are deleted at boot and daily |
| `CONSOLE_GOLDEN_DIR` | `../testdata/golden` | where Outbound Golden Files are written |
| `CONSOLE_WEBHOOK_GOLDEN_DIR` | `../webhook/testdata` | where Inbound Golden Files are written |
| `CONSOLE_PUBLIC_URL` | (empty) | public base URL shown in the UI as the webhook URL |
| `CONSOLE_LOG_LEVEL` | `info` | zerolog level |
| `CYBERBIZ_BASE_URL` | SDK default | override the API host (tests, proxies) |

## Receiving webhooks locally

CYBERBIZ needs a public HTTPS URL. Expose the Console with a tunnel and set
`CONSOLE_PUBLIC_URL` to the URL it prints so the dashboard shows the full
webhook URL to paste into CYBERBIZ:

```sh
ngrok http 8787
# or
cloudflared tunnel --url http://localhost:8787
```

The endpoint is `POST <public url>/webhooks/cyberbiz`. Every delivery is
stored, including ones that fail verification (`invalid_signature`,
`unknown_shop`, `malformed`), so you can see exactly what CYBERBIZ sent.

## HTTP API

Every route under `/api` except `/api/health` and `/api/auth/login` needs
the `console_session` cookie. Responses use the envelope
`{"success": true, "data": ...}` / `{"success": false, "error": "...", "code": 40400}`.
The full route list and error codes are in `docs/spec.md`.

```sh
curl -c jar -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"..."}' localhost:8787/api/auth/login
curl -b jar localhost:8787/api/shops
curl -b jar -H 'Content-Type: application/json' \
  -d '{"method":"GET","path":"/v1/orders","query":{"page":["1"]}}' \
  localhost:8787/api/shops/1/requests
curl -b jar -X POST localhost:8787/api/outbound/1/golden
```

## Development

```sh
go generate ./internal/catalog/...   # copy docs/api/en/openapi-*.yaml into the embedded catalog
gofmt -l . && go vet ./... && go test -race ./...
```

Tests run against a temporary SQLite file and a fake CYBERBIZ server; they
never touch the network.
