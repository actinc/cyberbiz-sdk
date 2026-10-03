# cyberbiz-sdk

Client SDKs for the [CYBERBIZ](https://www.cyberbiz.io) e-commerce platform
API, one directory per language, all built from the same API documentation
and verified against the same Golden Files.

| Language | Directory | Status |
|----------|-----------|--------|
| Go       | [go/](go/README.md)   | v0.x, covers the whole API, plus the Console |
| PHP      | [php/](php/README.md) | v0.x: shop, products, orders, customers, webhooks; `composer require actinc/cyberbiz-sdk` |

## Shared by every SDK

- `docs/api/en/`: OpenAPI 3.1 specs and Postman collections for API v1 and
  v2, a Postman collection for webhook testing, and the webhook reference
  (`webhooks.md`). `docs/api/zh-TW/` is the Traditional Chinese edition.
- `testdata/golden/`: redacted real API responses. Every SDK's model tests
  decode these same files, so all languages agree on the wire format.
- `CONTEXT.md`: the domain vocabulary (Shop, Outbound, Inbound, Event,
  Signature, Golden File, Sample, Console).

## License

MIT, see [LICENSE](LICENSE).
