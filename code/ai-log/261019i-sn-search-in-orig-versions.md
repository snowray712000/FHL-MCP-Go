# Strong's number search in lxx / bhs / fhlwh (1.1.2.14)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: done
- Files touched: internal/bible/origtext.go, internal/bible/snquery.go, internal/bible/versions_test.go, internal/version/version.go, app_versions.md

## What
`search_bible` / `search_bible_advanced` with `search_type` greek_number / hebrew_number and version lxx / bhs / fhlwh are answered from the pre-built files (261019g) instead of se.php.

## Why
se.php returns `Fail: Empty Query` for these versions (measured). Follow-up listed in 261019g.

## How
`evalSN` (SN algebra of 1.1.2.7) asks `origSNRecords` per number first: substring scan for `<WG{n}>` / `<WH{n}>` over the parsed rows, filtered by Range (1 NT, 2 OT, 3 book range). `searchSNCombined` treats a single number as evaluable for these versions. Records are `SERecord`, so paging/output are unchanged. No 5000 cap offline.

## Decisions & assumptions
Exact numbers only (3608 != 3608a). Wrong-testament number type finds nothing instead of an error. If the client lacks `OrigDataAPI`, behaviour is as before (upstream failure). A failed file load is returned as an error (no fallback exists).

## Verification
go vet, go test ./... pass. Temporary live test (deleted): bhs H430 2247 verses, fhlwh G3056 317, lxx "G2316 not G3588" 61; first hits sane. Not verified: the counts against any independent source.

## Open issues / next
bhs_code Hebrew keyword search; retry/rate limit. Keyword search over lxx/bhs unchanged (se.php).
