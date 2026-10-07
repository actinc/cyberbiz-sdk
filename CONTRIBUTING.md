# Contributing

Thanks for helping with the CYBERBIZ Go SDK. This page is short on purpose;
the detailed rules live in the files it points to.

## Before you start

- Read `CONTEXT.md` for the vocabulary (Shop, Outbound, Inbound, Event,
  Signature, Golden File, Sample, Console) and use it in code, tests, and
  pull requests.

## Ground rules

- **Golden Files are the source of truth.** When a Golden File in
  `testdata/golden/` and any document disagree, the Golden File wins. Do not
  hand-write or hand-edit a Golden File: record it through the Console
  ("Save as Golden File") or import a raw response with
  `go run ./internal/tools/goldenimport`, both of which redact it. CI rejects
  Golden Files that still contain e-mail addresses, phone numbers, tokens, or
  merchant hostnames. Maintainers also set `CYBERBIZ_REDACT_DENYLIST` (a
  private regular expression of customer and integrator names) so the same
  check rejects those names; CI requires it.
- **Never commit credentials or raw reference material.** `docs/references/`
  is gitignored and stays that way.
- **The SDK module stays dependency-free** (standard library and
  `golang.org/x/*` only).
- **Docs are generated.** Do not edit `docs/api/` by hand; both locales are
  regenerated together.

## Workflow

1. Fork and branch from `main`.
2. Make the change with tests: every service method has a request-shape test,
   every model has a Golden File decode test.
3. Run `make lint test` in `go/`. To run every CI check that your change
   touches, run `scripts/ci-local.sh` (`--all` runs every check). To have
   `git push` run it first and stop when a check fails, enable the hook once
   per clone with `scripts/install-hooks.sh`. It uses the tool versions CI
   uses: your installed go, golangci-lint, node or php when the version
   matches CI's, otherwise CI's Docker image, so Docker is required. Use
   `git push --no-verify` to skip it for one push.
4. Open a pull request. CI runs formatting, vet, golangci-lint, and the unit
   tests for the SDK and the Console. Live tests against the CYBERBIZ API run
   only after a maintainer merges to `main`; they are read-only and need a
   token that pull requests never receive.
5. Use Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `chore:`,
   `ci:`), scoped when useful (`feat(sdk):`, `feat(console):`).

## Reporting API discrepancies

If the live API disagrees with `docs/api/` or with a model, that is the most
valuable kind of issue. Include the endpoint, what the documentation says,
what the API returned (redacted), and if possible a Golden File recorded
through the Console.

## Releasing (maintainers)

Go releases are tagged `go/vX.Y.Z`, because the module lives in `go/`, and are
made by pull request:

1. Open a PR that sets `go/VERSION` to the new version (e.g. `0.3.0`),
   titled with the ticket key and the version.
2. Merge it. The `main` build publishes the code to GitHub, tags that commit
   `go/v0.3.0`, and tags the matching GitHub commit. A version that is not
   above the newest release fails the build. A version counts as released
   only when its tag is on both Bitbucket and GitHub: a tag found only on
   Bitbucket is published to GitHub, or the build fails naming the tag and
   its commit when GitHub has no commit with that tag's public tree.

GitHub only changes on a release: merges that do not bump a `VERSION` file
stay on Bitbucket until the next one. To publish `main` sooner without
releasing (a documentation fix, say), run the `sync-github` pipeline on
`main`.

PHP releases work the same way with `php/VERSION` and plain `vX.Y.Z` tags,
which is what Packagist reads (it ignores `go/v*`). The SDKs are
versioned independently. A PHP release PR also updates `Client::VERSION`
in `php/src/Client.php`; a test fails when the two disagree.

Java releases work the same way with `java/VERSION` and `java/vX.Y.Z` tags.
A version ending in `-SNAPSHOT` (as `java/VERSION` does until the first Java
release) is skipped, so the SDK can sit in `main` unreleased.
A Java release PR also updates the version in `java/pom.xml` and
`CyberbizClient.VERSION`; a test fails when they disagree. The Java SDK is
not published to Maven Central yet.

If a tag ever needs to be republished to GitHub, run the `release-tag`
pipeline on `main` with `TAG=go/vX.Y.Z`, `TAG=vX.Y.Z` or `TAG=java/vX.Y.Z`.
It publishes `main` first; an existing GitHub tag is never moved.
