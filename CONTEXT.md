# CYBERBIZ SDK for Go

A Go client for the CYBERBIZ e-commerce platform API, plus a local Console for
exercising the API and inspecting webhook traffic. This file is the glossary
for the whole repo; it defines terms, not implementation.

## Language

### Platform

**CYBERBIZ API**:
The single HTTP API served at `app-store-api.cyberbiz.io`, authenticated with a
Bearer token. Versions are path prefixes (`/v1`, `/v2`) plus a small set of
unprefixed app endpoints (`/shop`, `/settings`), not separate APIs.
_Avoid_: v1 API and v2 API as if they were different services, legacy API

**Shop**:
One CYBERBIZ merchant store, identified by its `shop_id` and Shop Domain. One
Shop has exactly one set of App credentials and one API Token.
_Avoid_: Store, merchant, tenant, account

**Shop Domain**:
The Shop's CYBERBIZ-issued hostname, e.g. `example.cyberbiz.co`, sent on every
Inbound as `X-Cyberbiz-Domain`. It is the key for finding a Shop's credentials.
_Avoid_: Custom domain (that is the merchant's own storefront hostname, sent as
`X-Cyberbiz-Shop-Domain`, and is not an identifier)

**App**:
The CYBERBIZ app installation that grants access to a Shop. Its credentials are
the App Name, App ID, and App Secret; the App Secret signs Inbound webhooks.
No two Shops share an App Secret, but one Shop can install several Apps. An
Inbound carries no App identifier, so its App is the one whose App Secret
verifies the Signature (ADR-0008).
_Avoid_: Integration, connector

**API Token**:
The Bearer JWT that authorises Outbound requests for one Shop. It carries the
Shop's id, domain, and scopes.
_Avoid_: Access token, key, credential (ambiguous with App Secret)

**Scope**:
A permission carried by an API Token that gates a resource group, e.g.
`read_orders`, `write_products`.

### Traffic

**Outbound**:
An HTTP request the SDK sends to the CYBERBIZ API, together with its response.
_Avoid_: API call, request log, egress

**Inbound**:
An HTTP request CYBERBIZ sends to a webhook endpoint for a Shop event, together
with the response we returned.
_Avoid_: Webhook log, notification, ingress

**Event**:
The kind of Shop change an Inbound announces, named `resource/action`, e.g.
`orders/paid`, `customers/create`.
_Avoid_: Topic, hook type, message type

**Signature**:
The HMAC-SHA256 of an Inbound's raw body, keyed by the App Secret, sent by
CYBERBIZ in `X-Cyberbiz-Hmac-Sha256`. Verifying it proves the Inbound came
from CYBERBIZ and was not altered.
_Avoid_: Hash, digest, checksum

**Domain Signature**:
The HMAC-SHA256 of the Shop Domain, keyed by the App Secret, sent in
`X-Cyberbiz-Domain-Hmac-Sha256`. It binds the Inbound to one Shop; it is
recorded but the Signature is what authenticates the payload.

### Tooling

**Console**:
The local web application in this repo that sends Outbound requests through
the SDK, receives Inbound webhooks, stores both, and lets a person browse them.
It is an observation tool: it never forwards, retries, or replays traffic.
_Avoid_: Playground, inspector, devtool, admin, dashboard

**Golden File**:
A recorded real API response, with sensitive values redacted, kept in the repo
as the reference shape for a data model and its contract test.
_Avoid_: Fixture, snapshot, sample (a sample is the synthetic example in docs)

**Sample**:
A synthetic example request or response in the curated API docs, realistic in
shape but never containing real data.
_Avoid_: Example payload, mock data
