# Greek keyword search in fhlwh (1.1.2.16)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: done
- Files touched: internal/bible/greeksearch.go, internal/bible/greekfold_table.go, internal/bible/hebsearch.go, internal/bible/origtext.go, internal/bible/versions_test.go, internal/version/version.go, app_versions.md

## What
Keyword search with version fhlwh and a Greek query is answered from `bible_fhlwh.json.gz` (261019g); se.php answers "not found" for fhlwh.

## Why
Follow-up left in 261019j.

## How
`origKeyword` (261019j) now serves bhs (Hebrew clusters) and fhlwh (Greek). Greek: fold text and query to bare lower-case α..ω (accents, breathings, subscript dropped, ς=σ, other chars = word boundary), every query word must occur as a substring. Fold table generated from Python unicodedata (data uses precomposed U+1F00-1FFF letters, no combining marks, so no x/text dependency). Rejected: exact-accent matching (keyboards give tonos, FHL data oxia; FHLNUI notes the same trap).

## Decisions & assumptions
Accent-insensitive always (no way to ask for exact accents). lxx keyword search left on se.php (works, exact form) to avoid changing existing behaviour. Queries without Greek letters still go upstream.

## Verification
go vet, go test ./... pass. Temporary live test (deleted): logos 67 verses (FHLNUI recorded 68 for ssn.php `lovgos`, one-verse difference not investigated), capital/tonos form same 67, "agape theou" 32. Not compared with the website.

## Open issues / next
Retry/rate limit is the only FHLNUI item left; lxx accent-insensitive search possible with the same code.
