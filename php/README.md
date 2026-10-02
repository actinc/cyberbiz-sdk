# CYBERBIZ PHP SDK

Planned. This directory will hold the PHP client, with its own
`composer.json`, `src/` and `tests/`.

Ground rules it inherits from the repo (see the root README):

- Endpoints and webhook payloads come from `../docs/api/en/`.
- Model tests decode the shared Golden Files in `../testdata/golden/`; never
  copy them into this directory.
- Its CI runs whenever `php/`, `docs/api/` or `testdata/golden/` changes.
