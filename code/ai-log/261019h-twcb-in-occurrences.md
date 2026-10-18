# TWCB in search_strongs_occurrences (1.1.2.13)

- Date: 2026-10-19
- Plan ref: §9 (tool behaviour)
- Status: done
- Files touched: internal/bible/strongs.go, internal/tools/tools_search.go, internal/tools/twcb_test.go, assets/tools-list.json, doc/m0/contract/tools-list.json, internal/version/version.go, app_versions.md

## What
`search_strongs_occurrences` gets optional `dictionary` (cbol default / twcb / both), same values as `lookup_strongs` (261019e). Entry goes in `strongs_info` (single) or `strongs_infos` (combined or/and/not/range, max 10).

## Why
Last FHLNUI-review sub-item on TWCB (261019b item 3).

## How
`SearchStrongsOccurrences(..., dictionary, simplified)`; helper `entryIn` reuses `twcbText`/`TWCBEntry`/`StrongsBoth`. Result fields `StrongsInfo`/`StrongsInfos` became `any` (cbol JSON unchanged, goldens pass). Contract JSON edited in both copies (same as 1.1.2.9/.10).

## Decisions & assumptions
Unlike lookup_strongs, a missing TWCB entry under `twcb` is not an error here (strongs_info omitted, occurrences kept) because the occurrences are the main payload. Combined queries with twcb do up to 10 TWCB requests (not cached, TTL 0).

## Verification
go vet, go test ./... pass (new test uses fixtures; both/twcb/missing/combined/invalid). No new live check (endpoints verified in 1.1.2.10).

## Open issues / next
Remaining: retry/rate limit; SN search over lxx/bhs; bhs_code Hebrew search.
