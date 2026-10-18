# Hebrew keyword search in bhs (1.1.2.15)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: done
- Files touched: internal/bible/hebsearch.go, internal/bible/origtext.go, internal/bible/search.go, internal/bible/versions_test.go, internal/version/version.go, app_versions.md

## What
`search_bible` / `search_bible_advanced` keyword search with version bhs and a Hebrew query is answered locally from `bible_bhs.json.gz` (261019g) instead of se.php (which stores FHL ASCII code and returns 0 for Unicode).

## Why
FHLNUI review item 5 follow-up; user chose "bhs_code Hebrew keyword search".

## How
Cluster-based matching (letter + points), rules from FHLNUI z260927a: sofit folded; dagesh/shin/sin only required when typed; vowels exact when typed, free otherwise; cantillation/meteg ignored; words split on space/maqaf, all query words must occur, substring allowed. Hook at the top of `searchPaged` (after the 5000 limit check). Deliberate deviation: not bible_bhs_code.json.gz. That file is the old ASCII code and would need the umscode/regex port; the Unicode bhs file we already load supports the same rules directly. ASCII-only queries still go to se.php.

## Decisions & assumptions
Matching semantics copied from FHLNUI, not from FHL's own search; counts not compared with the website. Results show text without `<WH>` markers. fhlwh Greek keyword search not done.

## Verification
go vet, go test ./... pass. Temporary live test (deleted): 'elohim' 1083 verses, 'YHWH elohim' 257, one fully pointed word 6. Not verified: against an independent count.

## Open issues / next
Retry/rate limit (only remaining FHLNUI item); fhlwh Greek keyword search.
