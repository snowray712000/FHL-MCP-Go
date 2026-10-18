# Combined Strong's number queries: or / and / not / range (1.1.2.7)

- Date: 2026-10-19
- Plan ref: §9 (search tools)
- Status: done
- Files touched: internal/bible/snquery.go, internal/bible/search.go, internal/bible/strongs.go, internal/tools/snquery_test.go, internal/version/version.go, app_versions.md

## What
`search_strongs_occurrences` (`number`), `search_bible` / `search_bible_advanced` (`query` with greek_number/hebrew_number) accept expressions like `G34 or G35`, `G34 G35` (and), `G34 not G35`, `G3114-G3120`.

## Why
User: se.php only does AND-like behaviour; OR is needed to view cognate words together. FHLNUI `docs/z261011a` does it client-side (`snQuery.es2023.js`): one se.php request per number, local set algebra.

## How
`parseSNQuery` -> groups(or) of items(and, not flag) of terms (range expanded). `evalSN` fetches each distinct number once (limit 5000, error if a number hits the cap), intersect/union/subtract by (book id, chap, sec), sort canonical, then offset/limit. Single plain number keeps the old path (goldens untouched). Combined occurrences result carries `strongs_infos` (<=10 lookups, failures skipped) instead of `strongs_info`.

## Decisions & assumptions
Limits 60 terms / 200 per range (copied from FHLNUI); 10 dictionary lookups mine. Tool descriptions not changed (frozen contract), so clients must know the syntax. total_count of a combined result = matched verses.

## Verification
`go vet`, `go test ./...` pass with a mock SE (or/and/not/range/limit/errors/advanced range). Not run against the live upstream.

## Open issues / next
Consider documenting the syntax in tool descriptions if the contract freeze is lifted; sequential requests (up to 60) rely on the cache.
