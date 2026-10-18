# Page se.php searches when limit exceeds 500 (1.1.2.6)

- Date: 2026-10-18
- Plan ref: §9 D12 (related)
- Status: done
- Files touched: internal/bible/search.go, internal/tools/m2c_test.go, internal/version/version.go, app_versions.md

## What
`search_bible`, `search_bible_advanced`, `search_strongs_occurrences` accept `limit` up to 5000; above 500 the service walks `offset` pages of 500 and merges.

## Why
User recalled a 500 cap. Live probe: se.php with limit>=501 (or 0) answers `Fail: record count > 500`; offset paging works; `record_count` = page size. FHLNUI (`docs/api_reference.md` §8) recurses on offset until `record_count < 500`. Go previously surfaced a bare ErrUpstream.

## How
`Service.searchPaged`; limit<=500 unchanged (single call, goldens untouched). Walk stops on a short page. Cap 5000 -> InvalidParameterError (assumption, mine).

## Decisions & assumptions
5000 cap and error wording are assumptions. Checked qb.php: chapter-scoped, no cap issue (31 verses fine). Not changed: ssc.php (D12 error), sesub/seaf.

## Verification
`go vet`, `go test ./...` pass, incl. new mock test (1200 records, 3 pages). Not run live end to end with limit>500.

## Open issues / next
Could add default-limit guidance in tool descriptions (contract is frozen, so not done).
