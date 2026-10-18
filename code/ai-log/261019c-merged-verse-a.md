# Merged-verse marker "a" -> placeholder (1.1.2.8)

- Date: 2026-10-19
- Plan ref: §9
- Status: done
- Files touched: internal/bible/verse.go, internal/bible/search.go, internal/bible/verse_test.go, tests/fixtures/upstream/qb_col_2_20_21_merged.json, internal/version/version.go, app_versions.md

## What
Upstream sends `bible_text:"a"` for a verse merged into the previous one (西 2:20-21). `newVerseItem` now emits text `（併入上節）` plus `merged_into_previous:true` (omitempty). Used by `GetVerse` (verse/chapter/citation) and search `hits`.

## Why
FHLNUI `docs/z261002b`; ai-log 261019b item 2. Clients otherwise saw a literal "a".

## How
Helper keyed on `TrimSpace(text)=="a"`. Extra field is omitempty so existing goldens are untouched. Rejected: extending previous verse label to `20-21` (verse is an int in the contract).

## Decisions & assumptions
Placeholder wording and the extra field name are mine. Not applied to apocrypha/apostolic (`scripture.go`) since no evidence they use "a".

## Verification
`go vet`, `go test ./...` pass incl. new `TestMergedVersePlaceholder`. Fixture is hand-written from the FHLNUI doc, not a live response; live API not called.

## Open issues / next
Confirm with a live qb.php call for 西 2:21 (also with `strong=1`); decide about scripture.go and `search_commentary`-like paths.
