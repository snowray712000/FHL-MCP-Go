# Version list from uiabv.php (1.1.2.11)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: done
- Files touched: internal/fhlapi/client.go, internal/bible/versions.go, tests, tests/fixtures/upstream/uiabv.json, internal/app/serve_test.go, internal/version/version.go, app_versions.md

## What
`list_bible_versions` / `search_available_versions` now read `uiabv.php` first (fallback `ab.php`). Gains 6 versions and real `can_download` / `version`.

## Why
FHLNUI review item 4 (version flags). Flags `ntonly/otonly/strong` were already exposed from ab.php, but uiabv.php (what FHL's UI uses) is a superset.

## How
Measured: /json/uiabv.php 91 records, /api/ab.php 86. Missing in ab: cbol esv gnt6 lxx thv12h tru (all queryable via qb.php). ab.php lacks candownload/version, so output had them always false/"". Added `Client.UIABV` + optional interface `VersionListAPI`; `ListVersions` prefers it, falls back to `AB` on any error/nil. Rejected: replacing `AB` itself (breaks Python goldens and many client tests); a new tool parameter (no need).

## Decisions & assumptions
Output change vs Python golden is intentional (more versions, real download/version values). Assumes uiabv.php stays on the /json/ base. Frozen contract untouched.

## Verification
go vet, go test ./... pass; fixture is a real response trimmed to 6 records. Live curl compared both endpoints. Not verified: end-to-end against the running server after the change.

## Open issues / next
Remaining FHLNUI items: retry/rate limit, LXX/offline original text, TWCB in search_strongs_occurrences.
